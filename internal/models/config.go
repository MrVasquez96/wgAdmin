package models

import (
	"fmt"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Interface represents a WireGuard interface for UI display.
type Interface struct {
	Name   string
	IP     string
	Active bool
}

// Config represents a complete WireGuard configuration.
type Config struct {
	Name      string
	Interface InterfaceConfig
	Peers     []PeerConfig
}

// InterfaceConfig represents the [Interface] section of a WireGuard config.
type InterfaceConfig struct {
	PrivateKey wgtypes.Key
	Address    []net.IPNet // Multiple addresses supported (IPv4 + IPv6)
	ListenPort *int
	DNS        []net.IP
	MTU        int
	Table      string // "auto", "off", or a routing table number
	PreUp      string
	PostUp     string
	PreDown    string
	PostDown   string
	FwMark     *int
}

// PeerConfig represents a [Peer] section of a WireGuard config.
type PeerConfig struct {
	Name                string
	PublicKey           wgtypes.Key
	PresharedKey        *wgtypes.Key
	Endpoint            *net.UDPAddr
	AllowedIPs          []net.IPNet
	PersistentKeepalive time.Duration
}
// HasDefaultRoute returns true if any peer has 0.0.0.0/0 or ::/0 in AllowedIPs.
func (c *Config) HasDefaultRoute() bool {
	for _, peer := range c.Peers {
		for _, ip := range peer.AllowedIPs {
			ones, bits := ip.Mask.Size()
			if ones == 0 && (bits == 32 || bits == 128) {
				return true
			}
		}
	}
	return false
}

// Validate performs additional validation on the config.
func (c *Config) Validate() error {
	if len(c.Interface.Address) == 0 {
		return fmt.Errorf("at least one Address is required in Interface section")
	}

	for i, peer := range c.Peers {
		if len(peer.AllowedIPs) == 0 {
			return fmt.Errorf("peer %d: AllowedIPs is required", i)
		}
	}

	return nil
}
