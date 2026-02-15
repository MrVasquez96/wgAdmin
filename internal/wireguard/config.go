// Package wireguard provides a production-ready userspace WireGuard VPN client.
package wireguard

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"wgAdmin/internal/models"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/ini.v1"
)
 

// parseClientConfig reads and parses a WireGuard configuration file into typed config for the tunnel Client.
func parseClientConfig(path string) (*models.Config, error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		AllowNonUniqueSections: true,
	}, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	config := &models.Config{
		Interface: models.InterfaceConfig{
			MTU:   1420, // WireGuard default
			Table: "auto",
		},
	}

	// Parse Interface section
	ifaceSection := cfg.Section("Interface")
	if err := parseClientInterface(ifaceSection, &config.Interface); err != nil {
		return nil, fmt.Errorf("failed to parse Interface section: %w", err)
	}

	// Parse all Peer sections
	for _, section := range cfg.Sections() {
		if section.Name() == "Peer" {
			peer := models.PeerConfig{}
			if err := parseClientPeer(section, &peer); err != nil {
				return nil, fmt.Errorf("failed to parse Peer section: %w", err)
			}
			config.Peers = append(config.Peers, peer)
		}
	}

	if len(config.Peers) == 0 {
		return nil, fmt.Errorf("no peers defined in configuration")
	}

	return config, nil
}

func parseClientInterface(section *ini.Section, iface *models.InterfaceConfig) error {
	// PrivateKey (required)
	privKeyStr := section.Key("PrivateKey").String()
	if privKeyStr == "" {
		return fmt.Errorf("PrivateKey is required")
	}
	privKey, err := wgtypes.ParseKey(privKeyStr)
	if err != nil {
		return fmt.Errorf("invalid PrivateKey: %w", err)
	}
	iface.PrivateKey = privKey

	// Address (required for client operation)
	addressStr := section.Key("Address").String()
	if addressStr != "" {
		addresses := strings.Split(addressStr, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
			// Handle addresses without CIDR notation
			if !strings.Contains(addr, "/") {
				if strings.Contains(addr, ":") {
					addr += "/128"
				} else {
					addr += "/32"
				}
			}
			_, ipNet, err := net.ParseCIDR(addr)
			if err != nil {
				return fmt.Errorf("invalid Address %q: %w", addr, err)
			}
			iface.Address = append(iface.Address, *ipNet)
		}
	}

	// ListenPort (optional)
	if portStr := section.Key("ListenPort").String(); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid ListenPort: %w", err)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("ListenPort out of range: %d", port)
		}
		iface.ListenPort = &port
	}

	// DNS (optional)
	if dnsStr := section.Key("DNS").String(); dnsStr != "" {
		dnsServers := strings.Split(dnsStr, ",")
		for _, dns := range dnsServers {
			dns = strings.TrimSpace(dns)
			if dns == "" {
				continue
			}
			ip := net.ParseIP(dns)
			if ip == nil {
				return fmt.Errorf("invalid DNS address: %s", dns)
			}
			iface.DNS = append(iface.DNS, ip)
		}
	}

	// MTU (optional)
	if mtuStr := section.Key("MTU").String(); mtuStr != "" {
		mtu, err := strconv.Atoi(mtuStr)
		if err != nil {
			return fmt.Errorf("invalid MTU: %w", err)
		}
		if mtu < 576 || mtu > 65535 {
			return fmt.Errorf("MTU out of range: %d", mtu)
		}
		iface.MTU = mtu
	}

	// Table (optional)
	if table := section.Key("Table").String(); table != "" {
		iface.Table = table
	}

	// FwMark (optional)
	if fwMarkStr := section.Key("FwMark").String(); fwMarkStr != "" {
		// Handle hex format (0x...)
		var fwMark int
		var err error
		if strings.HasPrefix(fwMarkStr, "0x") {
			var mark int64
			mark, err = strconv.ParseInt(fwMarkStr[2:], 16, 32)
			fwMark = int(mark)
		} else {
			fwMark, err = strconv.Atoi(fwMarkStr)
		}
		if err != nil {
			return fmt.Errorf("invalid FwMark: %w", err)
		}
		iface.FwMark = &fwMark
	}

	// Script hooks (optional)
	iface.PreUp = section.Key("PreUp").String()
	iface.PostUp = section.Key("PostUp").String()
	iface.PreDown = section.Key("PreDown").String()
	iface.PostDown = section.Key("PostDown").String()

	return nil
}

func parseClientPeer(section *ini.Section, peer *models.PeerConfig) error {
	// PublicKey (required)
	pubKeyStr := section.Key("PublicKey").String()
	if pubKeyStr == "" {
		return fmt.Errorf("PublicKey is required")
	}
	pubKey, err := wgtypes.ParseKey(pubKeyStr)
	if err != nil {
		return fmt.Errorf("invalid PublicKey: %w", err)
	}
	peer.PublicKey = pubKey

	// PresharedKey (optional)
	if pskStr := section.Key("PresharedKey").String(); pskStr != "" {
		psk, err := wgtypes.ParseKey(pskStr)
		if err != nil {
			return fmt.Errorf("invalid PresharedKey: %w", err)
		}
		peer.PresharedKey = &psk
	}

	// Endpoint (optional for server, typically required for client)
	if endpointStr := section.Key("Endpoint").String(); endpointStr != "" {
		endpoint, err := net.ResolveUDPAddr("udp", endpointStr)
		if err != nil {
			return fmt.Errorf("invalid Endpoint %q: %w", endpointStr, err)
		}
		peer.Endpoint = endpoint
	}

	// AllowedIPs (required)
	allowedIPsStr := section.Key("AllowedIPs").String()
	if allowedIPsStr == "" {
		return fmt.Errorf("AllowedIPs is required")
	}
	allowedIPs := strings.Split(allowedIPsStr, ",")
	for _, cidr := range allowedIPs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid AllowedIPs %q: %w", cidr, err)
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *ipNet)
	}

	// PersistentKeepalive (optional)
	if keepaliveStr := section.Key("PersistentKeepalive").String(); keepaliveStr != "" {
		keepalive, err := strconv.Atoi(keepaliveStr)
		if err != nil {
			return fmt.Errorf("invalid PersistentKeepalive: %w", err)
		}
		if keepalive < 0 || keepalive > 65535 {
			return fmt.Errorf("PersistentKeepalive out of range: %d", keepalive)
		}
		peer.PersistentKeepalive = time.Duration(keepalive) * time.Second
	}

	return nil
}
