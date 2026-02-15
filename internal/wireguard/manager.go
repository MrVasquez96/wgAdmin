package wireguard

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"wgAdmin/internal/models"
)

const ConfigDir = "/etc/wireguard"

// Client registry tracks running WireGuard clients by interface name.
var (
	clientsMu sync.Mutex
	clients   = make(map[string]*Client)
)

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
	return ""
}

// IsInterfaceActive checks if a WireGuard interface is up
func IsInterfaceActive(name string) bool {
	clientsMu.Lock()
	_, running := clients[name]
	clientsMu.Unlock()
	if running {
		return true
	}

	// Fallback: check if the OS knows about the interface
	_, err := net.InterfaceByName(name)
	return err == nil
}

// ToggleInterface brings interface up or down using the native WireGuard client
func ToggleInterface(name string, up bool) error {
	if up {
		return startInterface(name)
	}
	return stopInterface(name)
}

func startInterface(name string) error {
	clientsMu.Lock()
	if _, exists := clients[name]; exists {
		clientsMu.Unlock()
		return fmt.Errorf("interface %s is already running", name)
	}
	clientsMu.Unlock()

	configPath := GetConfigPath(name)
	client, err := NewClientFromFile(configPath, WithInterfaceName(name))
	if err != nil {
		return fmt.Errorf("failed to create WireGuard client for %s: %w", name, err)
	}

	if err := client.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start interface %s: %w", name, err)
	}

	clientsMu.Lock()
	clients[name] = client
	clientsMu.Unlock()

	return nil
}

func stopInterface(name string) error {
	clientsMu.Lock()
	client, exists := clients[name]
	if !exists {
		clientsMu.Unlock()
		return fmt.Errorf("interface %s is not running", name)
	}
	delete(clients, name)
	clientsMu.Unlock()

	if err := client.Stop(); err != nil {
		return fmt.Errorf("failed to stop interface %s: %w", name, err)
	}

	return nil
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
