// Command preview runs the real wgAdmin console against mock tunnels with
// simulated live traffic, so the main window can be developed and
// screenshotted without root, pkexec or a real WireGuard setup.
//
//	go run ./cmd/preview                       # dark, live mock data
//	go run ./cmd/preview -theme light -select office
//	go run ./cmd/preview -noroot               # telemetry-unavailable state
//
// Nothing is written outside a temporary directory, and nothing touches the
// kernel: the interface list and the WireGuard snapshot are faked.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	ui "wgAdmin/internal/ui/views"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"
)

type mockPeer struct {
	name, allowed, endpoint string
	key                     wgtypes.Key
	online                  bool    // has a recent handshake
	rate                    float64 // bytes/s while up
}

type mockTunnel struct {
	name, address, endpoint string
	port, mtu               int
	dns                     string
	active                  bool
	peers                   []mockPeer
}

func key() wgtypes.Key {
	k, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	return k
}

// tunnels is the invented setup the preview shows. Every name, address and
// endpoint here is made up (example.* domains, arbitrary private ranges), so
// screenshots taken from it reveal nothing about anyone's real network.
func tunnels() []mockTunnel {
	return []mockTunnel{
		{name: "home-lab", address: "10.8.0.1/24", endpoint: "vpn.example.com:51820", port: 51820, mtu: 1420, active: true,
			peers: []mockPeer{
				{name: "laptop", allowed: "10.8.0.2/32", endpoint: "vpn.example.com:51820", online: true, rate: 92e3},
				{name: "phone", allowed: "10.8.0.3/32", endpoint: "vpn.example.com:51820", online: true, rate: 21e3},
				{name: "nas", allowed: "10.8.0.4/32, 192.168.50.0/24", endpoint: "vpn.example.com:51820", rate: 0},
				{name: "tablet", allowed: "10.8.0.5/32", endpoint: "vpn.example.com:51820", rate: 0},
			}},
		{name: "office", address: "10.20.0.7/24", endpoint: "gw.example.org:51820", dns: "10.20.0.1", active: true,
			peers: []mockPeer{
				{name: "office-gateway", allowed: "10.20.0.0/16, 172.16.40.0/24", endpoint: "gw.example.org:51820", online: true, rate: 310e3},
			}},
		{name: "datacenter", address: "10.30.1.2/32", endpoint: "dc.example.net:51820",
			peers: []mockPeer{
				{allowed: "10.30.0.0/22, 10.30.4.0/22", endpoint: "dc.example.net:51820"},
				{allowed: "10.30.8.0/22", endpoint: "dc-backup.example.net:51820"},
			}},
		{name: "staging", address: "10.40.0.1/24", port: 51900, mtu: 1380},
		{name: "travel", address: "10.50.0.9/32", endpoint: "edge.example.com:51820",
			peers: []mockPeer{{name: "edge-router", allowed: "0.0.0.0/0", endpoint: "edge.example.com:51820"}}},
	}
}

// writeConfig renders a tunnel as a wg-quick .conf, plus a client config for
// every named peer under clientDir, the way wgAdmin stores the configs it
// generates — so those peers can show a working QR code.
func writeConfig(dir, clientDir string, t *mockTunnel) error {
	var b strings.Builder
	fmt.Fprintf(&b, "[Interface]\nPrivateKey = %s\nAddress = %s\n", key(), t.address)
	if t.port > 0 {
		fmt.Fprintf(&b, "ListenPort = %d\n", t.port)
	}
	if t.mtu > 0 {
		fmt.Fprintf(&b, "MTU = %d\n", t.mtu)
	}
	if t.dns != "" {
		fmt.Fprintf(&b, "DNS = %s\n", t.dns)
	}
	for i := range t.peers {
		p := &t.peers[i]
		priv := key()
		p.key = priv.PublicKey()
		if p.name != "" {
			cdir := filepath.Join(clientDir, t.name)
			if err := os.MkdirAll(cdir, 0o700); err != nil {
				return err
			}
			client := fmt.Sprintf("[Interface]\nPrivateKey = %s\nAddress = %s\n", priv, strings.Split(p.allowed, ",")[0])
			if err := os.WriteFile(filepath.Join(cdir, p.name+".conf"), []byte(client), 0o600); err != nil {
				return err
			}
		}
		// go-wg takes the comment line directly above [Peer] as the name.
		if p.name != "" {
			fmt.Fprintf(&b, "\n# %s\n[Peer]\n", p.name)
		} else {
			b.WriteString("\n[Peer]\n")
		}
		fmt.Fprintf(&b, "PublicKey = %s\nAllowedIPs = %s\n", p.key, p.allowed)
		if p.endpoint != "" {
			fmt.Fprintf(&b, "Endpoint = %s\n", p.endpoint)
		}
		b.WriteString("PersistentKeepalive = 25\n")
	}
	return os.WriteFile(filepath.Join(dir, t.name+".conf"), []byte(b.String()), 0o600)
}

func main() {
	variant := flag.String("theme", "dark", "dark or light")
	sel := flag.String("select", "home-lab", "tunnel to show")
	scroll := flag.Float64("scroll", 0, "scroll the detail pane down by this much, for screenshots")
	openWhat := flag.String("open", "", "open a screen at start: new, edit, peers, config, backups, settings, scan, palette, shortcuts, delete, busy, qr")
	noRoot := flag.Bool("noroot", false, "simulate running without root (no telemetry)")
	compact := flag.Bool("compact", false, "use the compact density")
	noTips := flag.Bool("notips", false, "disable tooltips, for screenshots")
	width := flag.Float64("w", 1280, "window width")
	height := flag.Float64("h", 860, "window height")
	flag.Parse()

	// A fixed, readable path: it shows in the UI, and screenshots of a random
	// temp name look like noise.
	root := filepath.Join(os.TempDir(), "wgadmin-preview")
	dir := filepath.Join(root, "wireguard")
	_ = os.RemoveAll(root)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)

	mocks := tunnels()
	byName := map[string]*mockTunnel{}
	for i := range mocks {
		if err := writeConfig(dir, filepath.Join(dir, "clients"), &mocks[i]); err != nil {
			panic(err)
		}
		byName[mocks[i].name] = &mocks[i]
	}

	start := time.Now()
	ui.ListInterfaces = func(ctrl *wg.WG) ([]config.Interface, error) {
		list, err := ctrl.ListInterfaces()
		for i := range list {
			if m, ok := byName[list[i].Name]; ok {
				list[i].Active = m.active
				list[i].IP = strings.Split(m.address, "/")[0]
			}
		}
		return list, err
	}
	ui.IsRoot = func() bool { return !*noRoot }
	wgwidget.TooltipsEnabled = !*noTips
	ui.ReadSnapshot = func(names ...string) wgstat.Snapshot {
		snap := wgstat.Snapshot{Devices: map[string]wgstat.Device{}, Taken: time.Now(), Available: !*noRoot}
		if *noRoot {
			return snap
		}
		el := time.Since(start).Seconds()
		for _, n := range names {
			m := byName[n]
			if m == nil || !m.active {
				continue
			}
			dev := wgstat.Device{Name: n, ListenPort: 51820 + len(n)}
			for _, p := range m.peers {
				// Counters grow at a wobbling rate so the chart has a shape.
				wobble := 1 + 0.45*math.Sin(el/3+float64(len(p.name)))
				rx := int64(5e8 + p.rate*el*wobble)
				tx := int64(4e7 + p.rate*0.3*el*wobble)
				peer := wgstat.Peer{PublicKey: p.key.String(), Endpoint: p.endpoint, RX: rx, TX: tx}
				if p.online {
					peer.Handshake = time.Now().Add(-time.Duration(10+len(p.name)) * time.Second)
				} else if p.rate == 0 && p.name != "" {
					peer.Handshake = time.Now().Add(-3 * time.Hour)
				}
				dev.Peers = append(dev.Peers, peer)
				dev.RX += rx
				dev.TX += tx
			}
			snap.Devices[n] = dev
		}
		return snap
	}

	a := app.NewWithID("com.wireguard.manager.preview")
	cfg := settings.Load(a.Preferences())
	cfg.WGConfigPath = dir
	cfg.ClientConfigDir = filepath.Join(dir, "clients")
	cfg.ThemeVariant = *variant
	cfg.PrivilegeEscalation = "none"
	cfg.AutoRefreshEnabled = true
	cfg.AutoRefreshSecs = 2
	cfg.Density = settings.DensityComfortable
	if *compact {
		cfg.Density = settings.DensityCompact
	}
	a.Settings().SetTheme(wgtheme.NewWGAdminTheme(cfg))

	ctrl := wg.New(dir)
	language := lang.New(cfg.Language)
	w := a.NewWindow("wgAdmin · preview")
	w.Resize(fyne.NewSize(float32(*width), float32(*height)))

	mv := ui.NewMainView(w, &ctrl, cfg, language)
	w.SetContent(mv.Build(fyne.AppMetadata{Name: "wgAdmin", Version: "0.4.0"}))
	mv.Select(*sel)
	mv.Refresh()
	if *openWhat != "" {
		go func() {
			time.Sleep(1800 * time.Millisecond)
			fyne.Do(func() { mv.Open(*openWhat) })
		}()
	}
	if *scroll > 0 {
		go func() {
			time.Sleep(1500 * time.Millisecond)
			fyne.Do(func() { mv.ScrollDetail(float32(*scroll)) })
		}()
	}
	w.ShowAndRun()
}
