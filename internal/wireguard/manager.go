package wireguard

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"wgAdmin/internal/models"
)

const ConfigDir = "/etc/wireguard"

// ListInterfaces returns all WireGuard interfaces from /etc/wireguard/*.conf
func ListInterfaces() ([]models.Interface, error) {
	pattern := filepath.Join(ConfigDir, "*.conf")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var interfaces []models.Interface
	for _, path := range matches {
		name := strings.TrimSuffix(filepath.Base(path), ".conf")
		active := IsInterfaceActive(name)
		ip := GetInterfaceIP(name)
		if ip == "" {
			ip = "unknown"
		}

		interfaces = append(interfaces, models.Interface{
			Name:   name,
			IP:     ip,
			Active: active,
		})
	}
	return interfaces, nil
}

// GetInterfaceIP gets the IPv4 address of a network interface
func GetInterfaceIP(name string) string {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return ""
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			return ipNet.IP.String()
		}
	}

	// Fallback to ip command
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(`ip -4 address show %s | grep -oP '(?<=inet\s)\d+(\.\d+){3}/\d+' | cut -d/ -f1 | head -n1`, shellQuote(name)))
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// IsInterfaceActive checks if a WireGuard interface is up
func IsInterfaceActive(name string) bool {
	cmd := exec.Command("ip", "link", "show", name)
	return cmd.Run() == nil
}

// ToggleInterface brings interface up or down via wg-quick
func ToggleInterface(name string, up bool) error {
	action := "down"
	if up {
		action = "up"
	}
	cmd := exec.Command("wg-quick", action, name)
	return cmd.Run()
}

// DeleteInterface removes the .conf file, optionally creating a backup
func DeleteInterface(name string, backup bool) error {
	path := filepath.Join(ConfigDir, name+".conf")

	if backup {
		backupDir := filepath.Join(os.Getenv("HOME"), ".wgadmin-backups")
		if err := os.MkdirAll(backupDir, 0700); err != nil {
			return fmt.Errorf("failed to create backup directory: %w", err)
		}
		backupPath := filepath.Join(backupDir, name+".conf.bak")
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read config for backup: %w", err)
		}
		if err := os.WriteFile(backupPath, data, 0600); err != nil {
			return fmt.Errorf("failed to write backup: %w", err)
		}
	}

	return os.Remove(path)
}

// GetConfigPath returns the path to a config file
func GetConfigPath(name string) string {
	return filepath.Join(ConfigDir, name+".conf")
}

// ConfigExists checks if a config file exists
func ConfigExists(name string) bool {
	_, err := os.Stat(GetConfigPath(name))
	return err == nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
