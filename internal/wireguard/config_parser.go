package wireguard

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"wgAdmin/internal/models"
)

// ParseConfig reads a .conf file and returns WireGuardConfig
func ParseConfig(path string) (*models.WireGuardConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseConfigReader(bufio.NewScanner(file))
}

// ParseConfigString parses config from a string
func ParseConfigString(content string) (*models.WireGuardConfig, error) {
	return parseConfigReader(bufio.NewScanner(strings.NewReader(content)))
}

func parseConfigReader(scanner *bufio.Scanner) (*models.WireGuardConfig, error) {
	config := &models.WireGuardConfig{}
	var currentPeer *models.Peer
	inPeerSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if line == "[Interface]" {
			inPeerSection = false
			currentPeer = nil
			continue
		}

		if line == "[Peer]" {
			inPeerSection = true
			config.Peers = append(config.Peers, models.Peer{})
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
			switch key {
			case "PrivateKey":
				config.PrivateKey = value
			case "Address":
				config.Address = value
			case "DNS":
				config.DNS = value
			case "ListenPort":
				config.ListenPort, _ = strconv.Atoi(value)
			case "MTU":
				config.MTU, _ = strconv.Atoi(value)
			}
		} else if currentPeer != nil {
			// Peer section
			switch key {
			case "PublicKey":
				currentPeer.PublicKey = value
			case "AllowedIPs":
				currentPeer.AllowedIPs = value
			case "Endpoint":
				currentPeer.Endpoint = value
			case "PersistentKeepalive":
				currentPeer.PersistentKeepalive, _ = strconv.Atoi(value)
			case "PresharedKey":
				currentPeer.PresharedKey = value
			}
		}
	}

	return config, scanner.Err()
}
