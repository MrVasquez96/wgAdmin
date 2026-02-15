package wireguard

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"wgAdmin/internal/models"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ParseConfig reads a .conf file and returns a typed Config.
func ParseConfig(path string) (*models.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseConfigReader(bufio.NewScanner(file))
}

// ParseConfigString parses config from a string.
func ParseConfigString(content string) (*models.Config, error) {
	return parseConfigReader(bufio.NewScanner(strings.NewReader(content)))
}

func parseConfigReader(scanner *bufio.Scanner) (*models.Config, error) {
	config := &models.Config{
		Interface: models.InterfaceConfig{
			MTU:   1420,
			Table: "auto",
		},
	}
	var currentPeer *models.PeerConfig
	inPeerSection := false
	var lastLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if len(lastLines) == 3 {
			lastLines = lastLines[1:]
		}
		lastLines = append(lastLines, line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if line == "[Interface]" {
			configName := "Unknown"
			if len(lastLines) > 1 {
				configName = lastLines[len(lastLines)-2]
			}
			config.Name = removeHashtag(configName)
			inPeerSection = false
			currentPeer = nil
			continue
		}

		if line == "[Peer]" {
			inPeerSection = true

			peer := models.PeerConfig{}
			if len(lastLines) > 1 {
				peerName := lastLines[len(lastLines)-2]
				if peerName != "" {
					peer.Name = removeHashtag(peerName)
				} else {
					peer.Name = "Unknown"
				}
			}
			config.Peers = append(config.Peers, peer)
			currentPeer = &config.Peers[len(config.Peers)-1]
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if !inPeerSection {
			// Interface section
			if err := parseInterfaceField(config, key, value); err != nil {
				return nil, fmt.Errorf("interface field %s: %w", key, err)
			}
		} else if currentPeer != nil {
			// Peer section
			if err := parsePeerField(currentPeer, key, value); err != nil {
				return nil, fmt.Errorf("peer field %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

func parseInterfaceField(config *models.Config, key, value string) error {
	switch key {
	case "PrivateKey":
		privKey, err := wgtypes.ParseKey(value)
		if err != nil {
			return fmt.Errorf("invalid PrivateKey: %w", err)
		}
		config.Interface.PrivateKey = privKey
	case "Address":
		addresses := strings.Split(value, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
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
			config.Interface.Address = append(config.Interface.Address, *ipNet)
		}
	case "DNS":
		dnsServers := strings.Split(value, ",")
		for _, dns := range dnsServers {
			dns = strings.TrimSpace(dns)
			if dns == "" {
				continue
			}
			ip := net.ParseIP(dns)
			if ip == nil {
				return fmt.Errorf("invalid DNS address: %s", dns)
			}
			config.Interface.DNS = append(config.Interface.DNS, ip)
		}
	case "ListenPort":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid ListenPort: %w", err)
		}
		if port < 0 || port > 65535 {
			return fmt.Errorf("ListenPort out of range: %d", port)
		}
		config.Interface.ListenPort = &port
	case "MTU":
		mtu, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid MTU: %w", err)
		}
		config.Interface.MTU = mtu
	case "Table":
		config.Interface.Table = value
	case "FwMark":
		var fwMark int
		var err error
		if strings.HasPrefix(value, "0x") {
			var mark int64
			mark, err = strconv.ParseInt(value[2:], 16, 32)
			fwMark = int(mark)
		} else {
			fwMark, err = strconv.Atoi(value)
		}
		if err != nil {
			return fmt.Errorf("invalid FwMark: %w", err)
		}
		config.Interface.FwMark = &fwMark
	case "PreUp":
		config.Interface.PreUp = value
	case "PostUp":
		config.Interface.PostUp = value
	case "PreDown":
		config.Interface.PreDown = value
	case "PostDown":
		config.Interface.PostDown = value
	}
	return nil
}

func parsePeerField(peer *models.PeerConfig, key, value string) error {
	switch key {
	case "PublicKey":
		pubKey, err := wgtypes.ParseKey(value)
		if err != nil {
			return fmt.Errorf("invalid PublicKey: %w", err)
		}
		peer.PublicKey = pubKey
	case "AllowedIPs":
		cidrs := strings.Split(value, ",")
		for _, cidr := range cidrs {
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
	case "Endpoint":
		endpoint, err := net.ResolveUDPAddr("udp", value)
		if err != nil {
			return fmt.Errorf("invalid Endpoint %q: %w", value, err)
		}
		peer.Endpoint = endpoint
	case "PersistentKeepalive":
		keepalive, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid PersistentKeepalive: %w", err)
		}
		if keepalive < 0 || keepalive > 65535 {
			return fmt.Errorf("PersistentKeepalive out of range: %d", keepalive)
		}
		peer.PersistentKeepalive = time.Duration(keepalive) * time.Second
	case "PresharedKey":
		psk, err := wgtypes.ParseKey(value)
		if err != nil {
			return fmt.Errorf("invalid PresharedKey: %w", err)
		}
		peer.PresharedKey = &psk
	}
	return nil
}

func removeHashtag(s string) string {
	if strings.Contains(s, "#") {
		if len(s) >= 2 && s[:2] == "# " {
			s = s[2:]
		} else if len(s) >= 1 && s[:1] == "#" {
			s = s[1:]
		}
	}
	return s
}
