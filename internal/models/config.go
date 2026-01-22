package models

// WireGuardConfig represents a complete WireGuard configuration
type WireGuardConfig struct {
	// Interface section
	PrivateKey string
	Address    string
	DNS        string
	ListenPort int
	MTU        int

	// Peers
	Peers []Peer
}
