//go:build !linux

package wireguard

import (
	"fmt"
	"runtime"
	"wgAdmin/internal/models"
)

// NetworkConfig handles OS-level network configuration for WireGuard.
type NetworkConfig struct {
	interfaceName string
	config        *models.Config
}

// NewNetworkConfig creates a new NetworkConfig for the given interface.
func NewNetworkConfig(ifaceName string, config *models.Config) *NetworkConfig {
	return &NetworkConfig{
		interfaceName: ifaceName,
		config:        config,
	}
}

// Apply configures the network interface, addresses, routes, and DNS.
// On non-Linux platforms, this is a stub that returns an error.
func (n *NetworkConfig) Apply() error {
	return fmt.Errorf("automatic network configuration is not supported on %s; please configure interface manually", runtime.GOOS)
}

// Remove cleans up network configuration.
// On non-Linux platforms, this is a no-op.
func (n *NetworkConfig) Remove() error {
	return nil
}
