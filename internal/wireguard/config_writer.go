package wireguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wgAdmin/internal/models"
)

// WriteConfig writes a Config to a .conf file.
func WriteConfig(name string, config *models.Config) error {
	path := filepath.Join(ConfigDir, name+".conf")
	content := GenerateConfigString(config)
	return os.WriteFile(path, []byte(content), 0600)
}

// GenerateConfigString creates the INI-format string for a config.
func GenerateConfigString(config *models.Config) string {
	var sb strings.Builder

	// Interface section
	if config.Name != "" {
		sb.WriteString(fmt.Sprintf("# %s\n", config.Name))
	}
	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", config.Interface.PrivateKey.String()))

	if len(config.Interface.Address) > 0 {
		addrs := make([]string, len(config.Interface.Address))
		for i, addr := range config.Interface.Address {
			addrs[i] = addr.String()
		}
		sb.WriteString(fmt.Sprintf("Address = %s\n", strings.Join(addrs, ", ")))
	}

	if len(config.Interface.DNS) > 0 {
		dnsAddrs := make([]string, len(config.Interface.DNS))
		for i, dns := range config.Interface.DNS {
			dnsAddrs[i] = dns.String()
		}
		sb.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(dnsAddrs, ", ")))
	}

	if config.Interface.ListenPort != nil {
		sb.WriteString(fmt.Sprintf("ListenPort = %d\n", *config.Interface.ListenPort))
	}

	if config.Interface.MTU > 0 && config.Interface.MTU != 1420 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", config.Interface.MTU))
	}

	if config.Interface.Table != "" && config.Interface.Table != "auto" {
		sb.WriteString(fmt.Sprintf("Table = %s\n", config.Interface.Table))
	}

	if config.Interface.FwMark != nil {
		sb.WriteString(fmt.Sprintf("FwMark = %d\n", *config.Interface.FwMark))
	}

	if config.Interface.PreUp != "" {
		sb.WriteString(fmt.Sprintf("PreUp = %s\n", config.Interface.PreUp))
	}
	if config.Interface.PostUp != "" {
		sb.WriteString(fmt.Sprintf("PostUp = %s\n", config.Interface.PostUp))
	}
	if config.Interface.PreDown != "" {
		sb.WriteString(fmt.Sprintf("PreDown = %s\n", config.Interface.PreDown))
	}
	if config.Interface.PostDown != "" {
		sb.WriteString(fmt.Sprintf("PostDown = %s\n", config.Interface.PostDown))
	}

	// Peer sections
	for _, peer := range config.Peers {
		if peer.Name != "" {
			sb.WriteString(fmt.Sprintf("\n# %s\n", peer.Name))
		} else {
			sb.WriteString("\n# Unknown Peer\n")
		}
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey.String()))

		if len(peer.AllowedIPs) > 0 {
			ips := make([]string, len(peer.AllowedIPs))
			for i, ip := range peer.AllowedIPs {
				ips[i] = ip.String()
			}
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(ips, ", ")))
		}

		if peer.Endpoint != nil {
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint.String()))
		}

		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", int(peer.PersistentKeepalive.Seconds())))
		}

		if peer.PresharedKey != nil {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey.String()))
		}
	}

	return sb.String()
}
