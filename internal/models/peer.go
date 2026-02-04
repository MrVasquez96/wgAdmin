package models

// Peer represents a WireGuard peer configuration
type Peer struct {
	Name                string
	PublicKey           string
	AllowedIPs          string
	Endpoint            string
	PersistentKeepalive int
	PresharedKey        string
}
