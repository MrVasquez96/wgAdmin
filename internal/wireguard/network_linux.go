//go:build linux

package wireguard

import (
	"fmt"
	"net"
	"os"
	"strings"

	"wgAdmin/internal/models"
	"github.com/vishvananda/netlink"
)

// NetworkConfig handles OS-level network configuration for WireGuard.
type NetworkConfig struct {
	interfaceName string
	config        *models.Config
	link          netlink.Link
	addedRoutes   []netlink.Route
	originalDNS   []byte
	dnsConfigured bool
}

// NewNetworkConfig creates a new NetworkConfig for the given interface.
func NewNetworkConfig(ifaceName string, config *models.Config) *NetworkConfig {
	return &NetworkConfig{
		interfaceName: ifaceName,
		config:        config,
	}
}

// Apply configures the network interface, addresses, routes, and DNS.
func (n *NetworkConfig) Apply() error {
	// Find the interface
	link, err := netlink.LinkByName(n.interfaceName)
	if err != nil {
		return fmt.Errorf("failed to find interface %s: %w", n.interfaceName, err)
	}
	n.link = link

	// Set MTU if specified
	if n.config.Interface.MTU > 0 {
		if err := netlink.LinkSetMTU(link, n.config.Interface.MTU); err != nil {
			return fmt.Errorf("failed to set MTU: %w", err)
		}
	}

	// Add IP addresses
	for _, addr := range n.config.Interface.Address {
		nlAddr := &netlink.Addr{
			IPNet: &addr,
		}
		if err := netlink.AddrAdd(link, nlAddr); err != nil {
			// Ignore "file exists" errors (address already assigned)
			if !os.IsExist(err) && !strings.Contains(err.Error(), "file exists") {
				return fmt.Errorf("failed to add address %s: %w", addr.String(), err)
			}
		}
	}

	// Bring interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring up interface: %w", err)
	}

	// Add routes for AllowedIPs
	if n.config.Interface.Table != "off" {
		if err := n.configureRoutes(); err != nil {
			return fmt.Errorf("failed to configure routes: %w", err)
		}
	}

	// Configure DNS if specified
	if len(n.config.Interface.DNS) > 0 {
		if err := n.configureDNS(); err != nil {
			// DNS configuration is not critical, log but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to configure DNS: %v\n", err)
		}
	}

	return nil
}

// configureRoutes adds routes for all AllowedIPs.
func (n *NetworkConfig) configureRoutes() error {
	tableID := 0 // Main table

	// Parse table setting
	if n.config.Interface.Table != "auto" && n.config.Interface.Table != "" {
		var err error
		tableID, err = parseTableID(n.config.Interface.Table)
		if err != nil {
			return fmt.Errorf("invalid Table setting: %w", err)
		}
	}

	// Collect all unique AllowedIPs from all peers
	routeSet := make(map[string]net.IPNet)
	for _, peer := range n.config.Peers {
		for _, allowedIP := range peer.AllowedIPs {
			key := allowedIP.String()
			routeSet[key] = allowedIP
		}
	}

	// Add routes
	for _, dest := range routeSet {
		route := netlink.Route{
			LinkIndex: n.link.Attrs().Index,
			Dst:       &dest,
			Table:     tableID,
			Protocol:  4, // RTPROT_STATIC
		}

		// For default route (0.0.0.0/0 or ::/0), we may need special handling
		ones, bits := dest.Mask.Size()
		if ones == 0 {
			// This is a default route
			if bits == 32 {
				// IPv4 default route - add with lower metric to allow existing default
				route.Priority = 1
			} else {
				// IPv6 default route
				route.Priority = 1
			}
		}

		if err := netlink.RouteAdd(&route); err != nil {
			// Ignore "file exists" errors
			if !os.IsExist(err) && !strings.Contains(err.Error(), "file exists") {
				return fmt.Errorf("failed to add route %s: %w", dest.String(), err)
			}
		} else {
			n.addedRoutes = append(n.addedRoutes, route)
		}
	}

	return nil
}

// configureDNS sets up DNS resolution using systemd-resolved or /etc/resolv.conf.
func (n *NetworkConfig) configureDNS() error {
	// Try systemd-resolved first
	if n.trySystemdResolved() {
		return nil
	}

	// Fall back to /etc/resolv.conf modification
	return n.configureResolvConf()
}

// trySystemdResolved attempts to configure DNS using systemd-resolved.
func (n *NetworkConfig) trySystemdResolved() bool {
	// Check if systemd-resolved is running by checking for resolvectl
	conn, err := dbusCon()
	if err != nil {
		return false
	}
	defer conn.Close()

	// Get interface index
	ifIndex := n.link.Attrs().Index

	// Build DNS server list
	var dnsServers []interface{}
	for _, dns := range n.config.Interface.DNS {
		if ip4 := dns.To4(); ip4 != nil {
			// IPv4: family 2, address bytes
			dnsServers = append(dnsServers, []interface{}{int32(2), ip4})
		} else {
			// IPv6: family 10, address bytes
			dnsServers = append(dnsServers, []interface{}{int32(10), dns.To16()})
		}
	}

	// Call org.freedesktop.resolve1.Manager.SetLinkDNS
	obj := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
	call := obj.Call("org.freedesktop.resolve1.Manager.SetLinkDNS", 0, int32(ifIndex), dnsServers)
	if call.Err != nil {
		return false
	}

	// If this is a full tunnel (default route), set as default route for DNS too
	if n.config.HasDefaultRoute() {
		// SetLinkDefaultRoute
		call = obj.Call("org.freedesktop.resolve1.Manager.SetLinkDefaultRoute", 0, int32(ifIndex), true)
		if call.Err != nil {
			// Non-fatal, continue
		}
	}

	n.dnsConfigured = true
	return true
}

// configureResolvConf modifies /etc/resolv.conf directly.
func (n *NetworkConfig) configureResolvConf() error {
	// Backup original resolv.conf
	original, err := os.ReadFile("/etc/resolv.conf")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}
	n.originalDNS = original

	// Build new nameserver entries
	var lines []string
	lines = append(lines, "# Generated by WireGuard client")
	for _, dns := range n.config.Interface.DNS {
		lines = append(lines, fmt.Sprintf("nameserver %s", dns.String()))
	}

	// Append original content (commented out if we're the full tunnel)
	if n.config.HasDefaultRoute() {
		for _, line := range strings.Split(string(original), "\n") {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
				lines = append(lines, "# "+line)
			}
		}
	} else {
		lines = append(lines, string(original))
	}

	newContent := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile("/etc/resolv.conf", []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write /etc/resolv.conf: %w", err)
	}

	n.dnsConfigured = true
	return nil
}

// Remove cleans up network configuration.
func (n *NetworkConfig) Remove() error {
	var errs []error

	// Restore DNS
	if n.dnsConfigured {
		if n.originalDNS != nil {
			if err := os.WriteFile("/etc/resolv.conf", n.originalDNS, 0644); err != nil {
				errs = append(errs, fmt.Errorf("failed to restore /etc/resolv.conf: %w", err))
			}
		}

		// Try to revert systemd-resolved
		if conn, err := dbusCon(); err == nil {
			defer conn.Close()
			if n.link != nil {
				obj := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
				obj.Call("org.freedesktop.resolve1.Manager.RevertLink", 0, int32(n.link.Attrs().Index))
			}
		}
	}

	// Remove routes (in reverse order)
	for i := len(n.addedRoutes) - 1; i >= 0; i-- {
		if err := netlink.RouteDel(&n.addedRoutes[i]); err != nil {
			// Ignore "no such process" errors (route already removed)
			if !strings.Contains(err.Error(), "no such process") {
				errs = append(errs, fmt.Errorf("failed to remove route: %w", err))
			}
		}
	}

	// Link will be removed when TUN device is closed

	if len(errs) > 0 {
		return fmt.Errorf("errors during cleanup: %v", errs)
	}
	return nil
}

// parseTableID parses a routing table identifier.
func parseTableID(table string) (int, error) {
	// Could be a number or a name from /etc/iproute2/rt_tables
	var id int
	if _, err := fmt.Sscanf(table, "%d", &id); err == nil {
		return id, nil
	}

	// Try to look up in rt_tables
	data, err := os.ReadFile("/etc/iproute2/rt_tables")
	if err != nil {
		return 0, fmt.Errorf("cannot parse table name %q", table)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == table {
			if _, err := fmt.Sscanf(fields[0], "%d", &id); err == nil {
				return id, nil
			}
		}
	}

	return 0, fmt.Errorf("table %q not found", table)
}
