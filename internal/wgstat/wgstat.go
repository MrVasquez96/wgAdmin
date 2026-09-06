// Package wgstat reads live state from the running WireGuard interfaces.
//
// The rest of the application works from the .conf files on disk, which say
// what a tunnel *should* be. This package says what it is actually doing right
// now: negotiated listen port, per-peer last handshake, resolved endpoint and
// the byte counters. It talks to the kernel through wgctrl and degrades to an
// empty, clearly-marked snapshot when that is not possible (typically because
// the process is not root), so callers can always render something sensible.
package wgstat

import (
	"net"
	"sort"
	"strconv"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// PeerState summarises how recently a peer completed a handshake.
type PeerState int

const (
	// StateNever means the peer has never completed a handshake.
	StateNever PeerState = iota
	// StateOnline means a handshake happened within OnlineWindow.
	StateOnline
	// StateIdle means the last handshake is recent but outside OnlineWindow.
	StateIdle
	// StateStale means the last handshake is older than IdleWindow.
	StateStale
)

const (
	// OnlineWindow is how recent a handshake must be for a peer to count as
	// online. WireGuard rekeys about every two minutes when traffic flows, so
	// three minutes avoids flapping between refreshes.
	OnlineWindow = 3 * time.Minute
	// IdleWindow is the cutoff past which a peer is considered stale.
	IdleWindow = 15 * time.Minute
)

// Peer is the live state of a single peer.
type Peer struct {
	PublicKey  string
	Endpoint   string
	AllowedIPs []string
	Handshake  time.Time
	RX, TX     int64
	Keepalive  time.Duration
}

// Device is the live state of one WireGuard interface.
type Device struct {
	Name       string
	PublicKey  string
	ListenPort int
	RX, TX     int64 // summed across peers
	Peers      []Peer
}

// Snapshot is one reading of every live interface.
type Snapshot struct {
	Devices map[string]Device
	// Available reports whether the kernel could be queried at all. When it is
	// false the UI should say telemetry is unavailable rather than show zeroes,
	// which would wrongly imply an idle tunnel.
	Available bool
	Err       error
	Taken     time.Time
}

// Read takes a snapshot of the live WireGuard interfaces. It never panics and
// never blocks for long; on total failure it returns a snapshot with
// Available=false and the cause in Err.
//
// names, when given, are the interfaces the caller knows about. They matter
// because wgctrl.Devices() enumerates every interface and fails as a whole if
// even one cannot be read — with userspace implementations that means a single
// unreadable socket in /var/run/wireguard hides the telemetry of every other
// tunnel. When the bulk call fails we therefore retry interface by interface
// and keep whatever answers, so one broken tunnel costs only its own row.
func Read(names ...string) Snapshot {
	snap := Snapshot{
		Devices: map[string]Device{},
		Taken:   time.Now(),
	}

	client, err := wgctrl.New()
	if err != nil {
		snap.Err = err
		return snap
	}
	defer client.Close()

	devices, err := client.Devices()
	if err == nil {
		snap.Available = true
		for _, d := range devices {
			snap.Devices[d.Name] = convertDevice(d)
		}
		return snap
	}
	snap.Err = err

	// Bulk enumeration failed — fall back to the interfaces we were told about.
	for _, name := range names {
		d, err := client.Device(name)
		if err != nil {
			continue // inactive, or not readable by this user
		}
		snap.Available = true
		snap.Devices[name] = convertDevice(d)
	}
	if snap.Available {
		snap.Err = nil
	}

	return snap
}

// convertDevice maps a wgctrl device onto the package's own types.
func convertDevice(d *wgtypes.Device) Device {
	dev := Device{
		Name:       d.Name,
		PublicKey:  d.PublicKey.String(),
		ListenPort: d.ListenPort,
		Peers:      make([]Peer, 0, len(d.Peers)),
	}
	for _, p := range d.Peers {
		peer := Peer{
			PublicKey: p.PublicKey.String(),
			Handshake: p.LastHandshakeTime,
			RX:        p.ReceiveBytes,
			TX:        p.TransmitBytes,
			Keepalive: p.PersistentKeepaliveInterval,
		}
		if p.Endpoint != nil {
			peer.Endpoint = p.Endpoint.String()
		}
		for _, ipnet := range p.AllowedIPs {
			peer.AllowedIPs = append(peer.AllowedIPs, cidrString(ipnet))
		}
		dev.RX += p.ReceiveBytes
		dev.TX += p.TransmitBytes
		dev.Peers = append(dev.Peers, peer)
	}

	// Newest handshake first: the peers an operator cares about float up.
	sort.SliceStable(dev.Peers, func(i, j int) bool {
		return dev.Peers[i].Handshake.After(dev.Peers[j].Handshake)
	})

	return dev
}

// Device returns the live state for name.
func (s Snapshot) Device(name string) (Device, bool) {
	d, ok := s.Devices[name]
	return d, ok
}

// Peer returns the live peer with the given public key.
func (d Device) Peer(publicKey string) (Peer, bool) {
	for _, p := range d.Peers {
		if p.PublicKey == publicKey {
			return p, true
		}
	}
	return Peer{}, false
}

// OnlinePeers counts the peers that handshook within OnlineWindow.
func (d Device) OnlinePeers(now time.Time) int {
	n := 0
	for _, p := range d.Peers {
		if p.State(now) == StateOnline {
			n++
		}
	}
	return n
}

// State classifies the peer by how recent its last handshake is.
func (p Peer) State(now time.Time) PeerState {
	if p.Handshake.IsZero() {
		return StateNever
	}
	switch age := now.Sub(p.Handshake); {
	case age <= OnlineWindow:
		return StateOnline
	case age <= IdleWindow:
		return StateIdle
	default:
		return StateStale
	}
}

// cidrString renders an IPNet, keeping IPv6 in its canonical form.
func cidrString(n net.IPNet) string {
	ones, _ := n.Mask.Size()
	return n.IP.String() + "/" + strconv.Itoa(ones)
}
