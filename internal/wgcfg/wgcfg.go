// Package wgcfg wraps go-wg's config parser with the metadata handling wgAdmin
// needs.
//
// go-wg's writer stores two pieces of wgAdmin state in the config's comment
// block — the tunnel name and the public endpoint clients dial — but its parser
// only recognises a "# Endpoint: " spelling the writer never emits. The
// endpoint therefore survives exactly one save and is invisible from then on,
// which is why the form asked for an endpoint that nothing downstream could
// see. The writer lives in the dependency, so the gap is closed here, on the
// reading side.
//
// The same round trip also fed every generated comment back in as a user
// comment, so a file's header grew on each save. Normalize drops those.
package wgcfg

import (
	"strings"

	"github.com/MrVasquez96/go-wg/wg/config"
)

// Comment lines go-wg's writer emits. These are structure, not user text.
const (
	metaName        = "Config name:"
	metaEndpointOut = "Server endpoint:" // what the writer emits
	metaEndpointIn  = "Endpoint:"        // what the parser looks for
	metaComments    = "Comment(s):"
	commentBullet   = "- "
)

// Load parses a config file and restores the metadata go-wg drops.
func Load(path string) (*config.Config, error) {
	cfg, err := config.ParseConfig(path)
	if err != nil {
		return nil, err
	}
	Normalize(cfg)
	return cfg, nil
}

// LoadString is Load for config text held in memory.
func LoadString(text string) (*config.Config, error) {
	cfg, err := config.ParseConfigString(text)
	if err != nil {
		return nil, err
	}
	Normalize(cfg)
	return cfg, nil
}

// Normalize lifts wgAdmin's metadata out of the parsed comment block and leaves
// cfg.Comments holding only what a human actually wrote.
//
// The parser collects every "#" line it sees, including the header it wrote
// itself and the name comment above each [Peer]. Writing those back out turns
// them into ordinary comments, so a file re-saved a few times accumulates a
// header of its own history.
func Normalize(cfg *config.Config) {
	if cfg == nil {
		return
	}

	// Peer sections are introduced by a "# <name>" line, which is
	// indistinguishable from a user comment once parsed — except that it
	// matches a peer's name.
	peerNames := map[string]bool{"Unknown Peer": true, "Unknown": true}
	for _, p := range cfg.Peers {
		if p.Name != "" {
			peerNames[p.Name] = true
		}
	}

	var name string
	var comments []string

	for _, raw := range cfg.Comments {
		line := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "#"))

		switch {
		case strings.HasPrefix(line, metaName):
			name = strings.TrimSpace(strings.TrimPrefix(line, metaName))
		case strings.HasPrefix(line, metaEndpointOut):
			cfg.PublicEndpoint = strings.TrimSpace(strings.TrimPrefix(line, metaEndpointOut))
		case strings.HasPrefix(line, metaEndpointIn):
			cfg.PublicEndpoint = strings.TrimSpace(strings.TrimPrefix(line, metaEndpointIn))
		case line == "" || line == metaComments || peerNames[line]:
			// Structure, not content.
		default:
			comments = append(comments, strings.TrimPrefix(line, commentBullet))
		}
	}

	cfg.Comments = comments

	// The parser takes whatever line sits directly above [Interface] as the
	// config name — which is the endpoint header in a file wgAdmin wrote, and
	// any stray comment in a hand-written one. Only the name comment is a real
	// name; with it cleared, the writer fills the name in from the file name.
	cfg.Name = name
}

// TunnelName returns the interface name from a config's name metadata, which
// go-wg stores as "[wg0] Public key: …".
func TunnelName(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	name := strings.TrimSpace(cfg.Name)
	if strings.HasPrefix(name, "[") {
		if i := strings.Index(name, "]"); i > 1 {
			return name[1:i]
		}
	}
	return name
}

// PublicEndpoint returns the endpoint a client should dial to reach this
// tunnel: the tunnel-wide default when one is set, otherwise the endpoint its
// peers agree on. It is empty when the tunnel has none.
func PublicEndpoint(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if cfg.PublicEndpoint != "" {
		return cfg.PublicEndpoint
	}
	if eps := PeerEndpoints(cfg); len(eps) == 1 {
		return eps[0]
	}
	return ""
}

// PeerEndpoints lists the distinct endpoints configured on peers, in file
// order.
func PeerEndpoints(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range cfg.Peers {
		ep := strings.TrimSpace(p.Endpoint)
		if ep == "" || seen[ep] {
			continue
		}
		seen[ep] = true
		out = append(out, ep)
	}
	return out
}

// Endpoints lists every endpoint associated with a tunnel — its own default
// first, then any peer that overrides it. A hub whose peers all dial the same
// host yields one entry; a config with per-peer endpoints yields each of them.
func Endpoints(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	var out []string
	if cfg.PublicEndpoint != "" {
		out = append(out, cfg.PublicEndpoint)
	}
	for _, ep := range PeerEndpoints(cfg) {
		if ep != cfg.PublicEndpoint {
			out = append(out, ep)
		}
	}
	return out
}

// EndpointFor returns the endpoint to write into a peer's client config: the
// peer's own override when it has one, and the tunnel default otherwise.
//
// The fallback is the tunnel's own field rather than PublicEndpoint's derived
// value: in a config where one peer overrides the endpoint and another has
// none, the override belongs to that peer alone.
func EndpointFor(cfg *config.Config, peer config.PeerConfig) string {
	if ep := strings.TrimSpace(peer.Endpoint); ep != "" {
		return ep
	}
	if cfg == nil {
		return ""
	}
	return cfg.PublicEndpoint
}
