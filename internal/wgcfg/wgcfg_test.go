package wgcfg

import (
	"net"
	"strings"
	"testing"

	"github.com/MrVasquez96/go-wg/wg/config"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// newTestConfig builds a hub config with two peers, one of which overrides the
// endpoint.
func newTestConfig(t *testing.T) *config.Config {
	t.Helper()

	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	peerA, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	peerB, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	_, netA, _ := net.ParseCIDR("10.0.0.2/32")
	_, netB, _ := net.ParseCIDR("10.0.0.3/32")

	return &config.Config{
		Name:           "wg0",
		PublicEndpoint: "vpn.example.com:51820",
		Comments:       []string{"hand written note"},
		Interface: config.InterfaceConfig{
			PrivateKey: priv,
			Address:    []net.IPNet{{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)}},
		},
		Peers: []config.PeerConfig{
			{Name: "laptop", PublicKey: peerA.PublicKey(), AllowedIPs: []net.IPNet{*netA}},
			{Name: "phone", PublicKey: peerB.PublicKey(), AllowedIPs: []net.IPNet{*netB},
				Endpoint: "backup.example.com:51821"},
		},
	}
}

// TestPublicEndpointSurvivesRoundTrip is the whole reason this package exists:
// go-wg writes the endpoint as "# Server endpoint:" but its parser only reads
// "# Endpoint:", so without Normalize the value is lost the moment a config is
// read back, and the tunnel form reopens with an empty field.
func TestPublicEndpointSurvivesRoundTrip(t *testing.T) {
	original := newTestConfig(t)

	got, err := LoadString(config.GenerateConfigString(original))
	if err != nil {
		t.Fatalf("LoadString: %v", err)
	}

	if got.PublicEndpoint != original.PublicEndpoint {
		t.Errorf("PublicEndpoint = %q, want %q", got.PublicEndpoint, original.PublicEndpoint)
	}
	if got.Name != original.Name {
		t.Errorf("Name = %q, want %q", got.Name, original.Name)
	}
}

// TestPeerEndpointSurvivesRoundTrip covers the per-peer override, which is what
// lets one client dial a different host or port from the rest.
func TestPeerEndpointSurvivesRoundTrip(t *testing.T) {
	original := newTestConfig(t)

	got, err := LoadString(config.GenerateConfigString(original))
	if err != nil {
		t.Fatalf("LoadString: %v", err)
	}

	if len(got.Peers) != 2 {
		t.Fatalf("got %d peers, want 2", len(got.Peers))
	}
	if got.Peers[0].Endpoint != "" {
		t.Errorf("peer without an override has Endpoint %q, want empty", got.Peers[0].Endpoint)
	}
	if want := "backup.example.com:51821"; got.Peers[1].Endpoint != want {
		t.Errorf("peer Endpoint = %q, want %q", got.Peers[1].Endpoint, want)
	}
}

// TestCommentsDoNotAccumulate guards the other half of the round trip: the
// parser collects the writer's own header lines as comments, so re-saving a
// config used to grow its header by the whole previous header each time.
func TestCommentsDoNotAccumulate(t *testing.T) {
	cfg := newTestConfig(t)

	text := config.GenerateConfigString(cfg)
	for i := 0; i < 3; i++ {
		reloaded, err := LoadString(text)
		if err != nil {
			t.Fatalf("LoadString (pass %d): %v", i, err)
		}
		text = config.GenerateConfigString(reloaded)
	}

	final, err := LoadString(text)
	if err != nil {
		t.Fatalf("LoadString: %v", err)
	}

	if len(final.Comments) != 1 || final.Comments[0] != "hand written note" {
		t.Errorf("Comments = %#v, want the single original note", final.Comments)
	}
	if n := strings.Count(text, "Server endpoint:"); n != 1 {
		t.Errorf("endpoint header appears %d times after re-saving, want 1", n)
	}
	if n := strings.Count(text, "Config name:"); n != 1 {
		t.Errorf("name header appears %d times after re-saving, want 1", n)
	}
	if final.PublicEndpoint != cfg.PublicEndpoint {
		t.Errorf("PublicEndpoint = %q, want %q", final.PublicEndpoint, cfg.PublicEndpoint)
	}
}

func TestEndpointFor(t *testing.T) {
	cfg := newTestConfig(t)

	if got, want := EndpointFor(cfg, cfg.Peers[0]), "vpn.example.com:51820"; got != want {
		t.Errorf("peer without an override: got %q, want the tunnel default %q", got, want)
	}
	if got, want := EndpointFor(cfg, cfg.Peers[1]), "backup.example.com:51821"; got != want {
		t.Errorf("peer with an override: got %q, want %q", got, want)
	}

	// A peer's override must not leak to a peer that has none.
	noDefault := &config.Config{Peers: cfg.Peers}
	if got := EndpointFor(noDefault, noDefault.Peers[0]); got != "" {
		t.Errorf("got %q, want empty: another peer's override is not this peer's endpoint", got)
	}
}

func TestEndpoints(t *testing.T) {
	cfg := newTestConfig(t)

	got := Endpoints(cfg)
	want := []string{"vpn.example.com:51820", "backup.example.com:51821"}
	if len(got) != len(want) {
		t.Fatalf("Endpoints() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Endpoints()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// A client config carries no tunnel-wide endpoint of its own; the peer's is
	// the only one there is, and the detail pane still has to show it.
	client := &config.Config{Peers: []config.PeerConfig{{Endpoint: "vpn.example.com:51820"}}}
	if got := Endpoints(client); len(got) != 1 || got[0] != "vpn.example.com:51820" {
		t.Errorf("Endpoints() = %#v, want the peer's endpoint", got)
	}
}

// TestHandWrittenClientConfig covers a file wgAdmin did not write: no metadata
// header, the endpoint on the peer where WireGuard itself puts it. The detail
// pane has to find that endpoint, and the leading comment must not be mistaken
// for the tunnel's name.
func TestHandWrittenClientConfig(t *testing.T) {
	priv, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	peer, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	text := "# my laptop's tunnel\n" +
		"[Interface]\n" +
		"PrivateKey = " + priv.String() + "\n" +
		"Address = 10.123.12.3/32\n" +
		"MTU = 1420\n\n" +
		"[Peer]\n" +
		"PublicKey = " + peer.PublicKey().String() + "\n" +
		"AllowedIPs = 10.123.12.0/24\n" +
		"Endpoint = wgcloud.example.com:51820\n"

	cfg, err := LoadString(text)
	if err != nil {
		t.Fatalf("LoadString: %v", err)
	}

	if got := Endpoints(cfg); len(got) != 1 || got[0] != "wgcloud.example.com:51820" {
		t.Errorf("Endpoints() = %#v, want the peer's endpoint", got)
	}
	if cfg.Name != "" {
		t.Errorf("Name = %q, want empty: a stray comment is not a tunnel name", cfg.Name)
	}
	if len(cfg.Comments) != 1 || cfg.Comments[0] != "my laptop's tunnel" {
		t.Errorf("Comments = %#v, want the original note", cfg.Comments)
	}
}

func TestTunnelName(t *testing.T) {
	cases := map[string]string{
		"[wg0] Public key: abc=": "wg0",
		"wg0":                    "wg0",
		"":                       "",
	}
	for in, want := range cases {
		if got := TunnelName(&config.Config{Name: in}); got != want {
			t.Errorf("TunnelName(%q) = %q, want %q", in, got, want)
		}
	}
}
