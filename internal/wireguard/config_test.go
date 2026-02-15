package wireguard

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"wgAdmin/internal/models"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestparseClientConfig(t *testing.T) {
	// Create a temporary config file
	content := `[Interface]
PrivateKey = WG8jSCtXPbZ2nhL1+YBQRWE9jM3d/zj/BZu6xwEQqWs=
Address = 10.0.0.2/24, fd00::2/64
DNS = 1.1.1.1, 8.8.8.8
ListenPort = 51820
MTU = 1400
FwMark = 0x200

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
PresharedKey = FpCyhws9cxwWoV4xELtfJvjJN+zQVRi32YulM0ieCGQ=
Endpoint = 203.0.113.1:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Parse the config
	config, err := parseClientConfig(configPath)
	if err != nil {
		t.Fatalf("parseClientConfig failed: %v", err)
	}

	// Verify Interface section
	if config.Interface.PrivateKey.String() == "" {
		t.Error("PrivateKey should not be empty")
	}

	if len(config.Interface.Address) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(config.Interface.Address))
	}

	if len(config.Interface.DNS) != 2 {
		t.Errorf("Expected 2 DNS servers, got %d", len(config.Interface.DNS))
	}

	if config.Interface.ListenPort == nil || *config.Interface.ListenPort != 51820 {
		t.Error("ListenPort should be 51820")
	}

	if config.Interface.MTU != 1400 {
		t.Errorf("MTU should be 1400, got %d", config.Interface.MTU)
	}

	if config.Interface.FwMark == nil || *config.Interface.FwMark != 0x200 {
		t.Error("FwMark should be 0x200")
	}

	// Verify Peer section
	if len(config.Peers) != 1 {
		t.Fatalf("Expected 1 peer, got %d", len(config.Peers))
	}

	peer := config.Peers[0]

	if peer.PublicKey.String() == "" {
		t.Error("Peer PublicKey should not be empty")
	}

	if peer.PresharedKey == nil {
		t.Error("Peer PresharedKey should not be nil")
	}

	if peer.Endpoint == nil || peer.Endpoint.String() != "203.0.113.1:51820" {
		t.Errorf("Peer Endpoint mismatch")
	}

	if len(peer.AllowedIPs) != 2 {
		t.Errorf("Expected 2 AllowedIPs, got %d", len(peer.AllowedIPs))
	}

	if peer.PersistentKeepalive != 25*time.Second {
		t.Errorf("PersistentKeepalive should be 25s, got %v", peer.PersistentKeepalive)
	}
}

func TestparseClientConfigMinimal(t *testing.T) {
	content := `[Interface]
PrivateKey = WG8jSCtXPbZ2nhL1+YBQRWE9jM3d/zj/BZu6xwEQqWs=
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
AllowedIPs = 10.0.0.0/24
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := parseClientConfig(configPath)
	if err != nil {
		t.Fatalf("parseClientConfig failed: %v", err)
	}

	// Check defaults
	if config.Interface.MTU != 1420 {
		t.Errorf("Default MTU should be 1420, got %d", config.Interface.MTU)
	}

	if config.Interface.Table != "auto" {
		t.Errorf("Default Table should be 'auto', got %s", config.Interface.Table)
	}
}

func TestparseClientConfigMultiplePeers(t *testing.T) {
	content := `[Interface]
PrivateKey = WG8jSCtXPbZ2nhL1+YBQRWE9jM3d/zj/BZu6xwEQqWs=
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
AllowedIPs = 10.0.0.0/24
Endpoint = 203.0.113.1:51820

[Peer]
PublicKey = Gu3xYHe/b6jKAQDxrWH7YfVqL5r5XYNLsQUVbPzRVik=
AllowedIPs = 10.0.1.0/24
Endpoint = 203.0.113.2:51820
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "multi-peer.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := parseClientConfig(configPath)
	if err != nil {
		t.Fatalf("parseClientConfig failed: %v", err)
	}

	if len(config.Peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(config.Peers))
	}
}

func TestparseClientConfigMissingPrivateKey(t *testing.T) {
	content := `[Interface]
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
AllowedIPs = 10.0.0.0/24
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-privkey.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := parseClientConfig(configPath)
	if err == nil {
		t.Error("Expected error for missing PrivateKey")
	}
}

func TestparseClientConfigNoPeers(t *testing.T) {
	content := `[Interface]
PrivateKey = WG8jSCtXPbZ2nhL1+YBQRWE9jM3d/zj/BZu6xwEQqWs=
Address = 10.0.0.2/24
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "no-peers.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := parseClientConfig(configPath)
	if err == nil {
		t.Error("Expected error for no peers")
	}
}

func TestparseClientConfigInvalidKey(t *testing.T) {
	content := `[Interface]
PrivateKey = invalid-key
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
AllowedIPs = 10.0.0.0/24
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid-key.conf")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := parseClientConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

func TestHasDefaultRoute(t *testing.T) {
	tests := []struct {
		name       string
		allowedIPs string
		expected   bool
	}{
		{"IPv4 default", "0.0.0.0/0", true},
		{"IPv6 default", "::/0", true},
		{"Both defaults", "0.0.0.0/0, ::/0", true},
		{"Subnet only", "10.0.0.0/24", false},
		{"Mixed", "10.0.0.0/24, 0.0.0.0/0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := `[Interface]
PrivateKey = WG8jSCtXPbZ2nhL1+YBQRWE9jM3d/zj/BZu6xwEQqWs=
Address = 10.0.0.2/24

[Peer]
PublicKey = xTIBA5rboUvnH4htodjb60Y7YAf21J7YQMlNGC8HQ14=
AllowedIPs = ` + tt.allowedIPs

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test.conf")
			if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			config, err := parseClientConfig(configPath)
			if err != nil {
				t.Fatalf("parseClientConfig failed: %v", err)
			}

			if config.HasDefaultRoute() != tt.expected {
				t.Errorf("HasDefaultRoute() = %v, want %v", config.HasDefaultRoute(), tt.expected)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	// Valid config
	privKey, _ := wgtypes.GeneratePrivateKey()
	pubKey := privKey.PublicKey()

	validConfig := &models.Config{
		Interface: models.InterfaceConfig{
			PrivateKey: privKey,
			Address:    []net.IPNet{{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)}},
		},
		Peers: []models.PeerConfig{
			{
				PublicKey:  pubKey,
				AllowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			},
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Missing address
	noAddrConfig := &models.Config{
		Interface: models.InterfaceConfig{
			PrivateKey: privKey,
		},
		Peers: []models.PeerConfig{
			{
				PublicKey:  pubKey,
				AllowedIPs: []net.IPNet{{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(24, 32)}},
			},
		},
	}

	if err := noAddrConfig.Validate(); err == nil {
		t.Error("Config without address should fail validation")
	}
}
