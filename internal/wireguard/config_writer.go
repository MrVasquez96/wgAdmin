package wireguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wgAdmin/internal/models"
)

// WriteConfig writes a WireGuardConfig to a .conf file
func WriteConfig(name string, config *models.WireGuardConfig) error {
	path := filepath.Join(ConfigDir, name+".conf")
	content := GenerateConfigString(config)
	return os.WriteFile(path, []byte(content), 0600)
}

// GenerateConfigString creates the INI-format string for a config
func GenerateConfigString(config *models.WireGuardConfig) string {
	var sb strings.Builder

	// Interface section
	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", config.Address))

	if config.DNS != "" {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", config.DNS))
	}
	if config.ListenPort > 0 {
		sb.WriteString(fmt.Sprintf("ListenPort = %d\n", config.ListenPort))
	}
	if config.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", config.MTU))
	}

	// Peer sections
	for _, peer := range config.Peers {
		if peer.Name != "" {
			sb.WriteString(fmt.Sprintf("\n# %s\n", peer.Name))
		} else {
			sb.WriteString("\n# Unknown Peer\n")
		}
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs))

		if peer.Endpoint != "" {
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint))
		}
		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
	}

	return sb.String()
}
