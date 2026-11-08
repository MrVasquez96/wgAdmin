package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgcfg"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"github.com/MrVasquez96/go-wg/wg/config"
)

// detailActions are the operations the detail pane can trigger. They are
// supplied by MainView so the pane stays a pure renderer.
type detailActions struct {
	OnToggle   func(name string, up bool)
	OnScan     func(name, ip string)
	OnPeers    func(name string)
	OnEdit     func(name string)
	OnConfig   func(name string)
	OnDelete   func(name string)
	OnCopy     func(label, value string)
	OnPeerQR   func(tunnel, publicKey string)
	OnNew      func()
	OnImport   func()
	OnRelaunch func()
}

// tunnelDetail renders the right-hand pane for one tunnel.
type tunnelDetail struct {
	lang    *lang.Language
	actions detailActions

	// root is whether this process can read the kernel at all. Without it the
	// fix for missing telemetry is relaunching; with it, missing telemetry
	// just means no tunnel is up (or a real error worth showing).
	root bool

	// peerRows backs the table; the table reads it through a pointer so rows
	// can be refreshed without rebuilding the widget.
	peerRows []peerRow
	// peers keeps the table's sort order across rebuilds.
	peers peerTable
}

func newTunnelDetail(l *lang.Language, actions detailActions) *tunnelDetail {
	return &tunnelDetail{lang: l, actions: actions, root: IsRoot()}
}

// emptyState is shown when nothing is selected, there are no tunnels at all, or
// the filter excludes everything. Each case gets its own wording, illustration
// and, where useful, its own call to action.
func (d *tunnelDetail) emptyState(titleKey, bodyKey string, withActions bool) fyne.CanvasObject {
	art := wgtheme.IconEmptyTunnels
	if titleKey == lang.Main_no_match_title {
		art = wgtheme.IconEmptySearch
	}
	var actions []fyne.CanvasObject
	if withActions {
		actions = []fyne.CanvasObject{
			wgwidget.NewButton(d.lang.T(lang.Main_add_tunnel), wgtheme.IconAdd, wgwidget.Primary, d.actions.OnNew),
			wgwidget.NewButton(d.lang.T(lang.Main_import), wgtheme.IconImport, wgwidget.Secondary, d.actions.OnImport),
		}
	}
	return wgwidget.EmptyState(art, d.lang.T(titleKey), d.lang.T(bodyKey), actions...)
}

// build renders the full detail pane for iface. cfg may be nil when the config
// could not be parsed; snap carries the kernel's view when it is available.
func (d *tunnelDetail) build(iface config.Interface, cfg *config.Config,
	snap wgstat.Snapshot, rate wgstat.Sample, history []wgstat.Sample, configPath string) *container.Scroll {

	dev, hasDev := snap.Device(iface.Name)
	live := snap.Available && hasDev && iface.Active

	sections := []fyne.CanvasObject{
		d.header(iface, cfg),
		d.telemetry(iface, dev, live, snap, rate, history),
		d.section(wgtheme.IconRoute, d.lang.T(lang.Detail_section_interface), nil,
			d.interfaceGrid(iface, cfg, dev, live, configPath)),
		d.peersSection(iface, cfg, dev, live),
	}
	if routes := d.routes(cfg); routes != nil {
		sections = append(sections, routes)
	}

	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXL+4), sections...)
	return container.NewVScroll(container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceXL, wgtheme.SpaceXL+8, wgtheme.SpaceXL+4, wgtheme.SpaceXL+4), body))
}

// section is a titled block: heading row, then content.
func (d *tunnelDetail) section(icon *wgtheme.Icon, title string, trailing []fyne.CanvasObject, content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD+2),
		wgwidget.SectionTitle(icon, title, trailing...),
		content)
}

// header is the tunnel's identity and its actions. Bringing the tunnel up or
// down is the one primary action and sits alone on the right; the routine
// operations sit under the name, and the rare or destructive ones (config,
// delete) live in the overflow menu instead of a loud red button.
func (d *tunnelDetail) header(iface config.Interface, cfg *config.Config) fyne.CanvasObject {
	p := wgtheme.Current()

	// Identity tile: the tunnel glyph on a tinted square.
	tone, glyph := p.TextDim, wgtheme.IconTunnelDown
	if iface.Active {
		tone, glyph = p.Success, wgtheme.IconTunnelUp
	}
	tileBg := canvas.NewRectangle(p.Tint(tone, 0.14))
	tileBg.CornerRadius = wgtheme.RadiusLG
	tileBg.StrokeColor = p.Tint(tone, 0.30)
	tileBg.StrokeWidth = 1
	glyphImg := canvas.NewImageFromResource(glyph.For(tone))
	glyphImg.FillMode = canvas.ImageFillContain
	glyphImg.SetMinSize(fyne.NewSize(26, 26))
	tile := container.NewStack(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(48, 48)), tileBg),
		container.NewCenter(glyphImg))

	name := wgwidget.Title(iface.Name, wgtheme.TextSize(wgtheme.TextDisplay))
	name.FontSource = wgtheme.Font(wgtheme.WeightSemiBold)
	pill := wgwidget.StatusPill(iface.Active, d.lang.T(lang.Card_state_up), d.lang.T(lang.Card_state_down))
	titleRow := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD+2),
		name, container.NewCenter(pill))

	sub := wgwidget.DimValue(d.subtitle(iface, cfg))
	identity := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceLG+2),
		container.NewCenter(tile),
		container.NewVBox(layout.NewSpacer(), titleRow, sub, layout.NewSpacer()))

	var toggle *wgwidget.Button
	if iface.Active {
		toggle = wgwidget.NewButton(d.lang.T(lang.Card_btn_deactivate), wgtheme.IconStop, wgwidget.Danger, func() {
			d.actions.OnToggle(iface.Name, false)
		})
	} else {
		toggle = wgwidget.NewButton(d.lang.T(lang.Card_btn_activate), wgtheme.IconPlay, wgwidget.Success, func() {
			d.actions.OnToggle(iface.Name, true)
		})
	}
	toggle.Scale = wgwidget.ScaleLG
	toggle.Tooltip = "Space"

	top := container.NewBorder(nil, nil, identity, container.NewCenter(toggle))

	scan := wgwidget.NewButton(d.lang.T(lang.Card_btn_scan), wgtheme.IconScan, wgwidget.Secondary, func() {
		d.actions.OnScan(iface.Name, iface.IP)
	})
	scan.Tooltip = "Ctrl+S"
	// A scan goes out through the tunnel, so it only makes sense while the
	// tunnel is up; the tooltip says so rather than leaving a dead button.
	if !canScan(iface) {
		scan.Disable()
		scan.Tooltip = d.lang.T(lang.Detail_scan_needs_up)
	}
	peers := wgwidget.NewButton(d.lang.T(lang.Card_btn_peers), wgtheme.IconPeers, wgwidget.Secondary, func() {
		d.actions.OnPeers(iface.Name)
	})
	peers.Tooltip = "Ctrl+P"
	edit := wgwidget.NewButton(d.lang.T(lang.Card_btn_edit), wgtheme.IconEdit, wgwidget.Secondary, func() {
		d.actions.OnEdit(iface.Name)
	})
	edit.Tooltip = "Ctrl+E"

	var more *wgwidget.Button
	more = wgwidget.NewIconButton(wgtheme.IconMoreVertical, d.lang.T(lang.Detail_more), func() {
		wgwidget.ShowMenu(more, []wgwidget.MenuItem{
			{Label: d.lang.T(lang.Detail_menu_config), Icon: wgtheme.IconDoc,
				Action: func() { d.actions.OnConfig(iface.Name) }},
			{Separator: true},
			{Label: d.lang.T(lang.Detail_menu_delete), Icon: wgtheme.IconTrash, Danger: true,
				Action: func() { d.actions.OnDelete(iface.Name) }},
		})
	})

	actions := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM+2),
		scan, peers, edit, more)

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG+2), top, actions)
}

// canScan reports whether a tunnel can be scanned: it must be up, and its
// address known so the networks to probe can be worked out.
func canScan(iface config.Interface) bool {
	return iface.Active && iface.IP != "" && iface.IP != "unknown"
}

// subtitle is the line under the tunnel name: its address and peer count.
func (d *tunnelDetail) subtitle(iface config.Interface, cfg *config.Config) string {
	parts := []string{}
	if iface.IP != "" && iface.IP != "unknown" {
		parts = append(parts, iface.IP)
	}
	if cfg != nil {
		parts = append(parts, fmt.Sprintf("%d %s", len(cfg.Peers), strings.ToLower(d.lang.T(lang.Detail_metric_peers))))
	}
	return strings.Join(parts, "  ·  ")
}

// telemetry is the live picture: stat tiles and the traffic chart when the
// kernel can be read and the tunnel is up, otherwise an explanation of why
// there is nothing to show — a row of dashes with no reason is worse than no
// row at all.
func (d *tunnelDetail) telemetry(iface config.Interface, dev wgstat.Device, live bool,
	snap wgstat.Snapshot, rate wgstat.Sample, history []wgstat.Sample) fyne.CanvasObject {

	p := wgtheme.Current()

	if snap.Taken.IsZero() {
		// The first refresh has not landed yet; say nothing rather than guess.
		return container.NewStack()
	}

	if !snap.Available && !d.root {
		var action *wgwidget.Button
		if d.actions.OnRelaunch != nil {
			action = wgwidget.NewButton(d.lang.T(lang.Detail_relaunch_root), wgtheme.IconLock,
				wgwidget.Secondary, d.actions.OnRelaunch)
		}
		return wgwidget.Banner(wgwidget.BannerWarning,
			d.lang.T(lang.Detail_telemetry_off), d.lang.T(lang.Detail_telemetry_off_detail), action)
	}

	if !snap.Available && iface.Active && snap.Err != nil {
		// Root, tunnel up, and still nothing: that is a real failure, so
		// show it instead of a hint that would not help.
		return wgwidget.Banner(wgwidget.BannerError,
			d.lang.T(lang.Detail_telemetry_off), snap.Err.Error(), nil)
	}

	if !live {
		return wgwidget.Banner(wgwidget.BannerInfo,
			d.lang.T(lang.Detail_telemetry_off), d.lang.T(lang.Detail_tunnel_down_hint), nil)
	}

	online := dev.OnlinePeers(time.Now())
	onlineTone := p.TextPrimary
	if online > 0 {
		onlineTone = p.Success
	}
	rxV, rxU := splitUnit(wgstat.FormatBytes(dev.RX))
	txV, txU := splitUnit(wgstat.FormatBytes(dev.TX))
	rateV, rateU := wgstat.Dash, ""
	if rate.RX > 0 || rate.TX > 0 {
		rateV, rateU = splitUnit(wgstat.FormatRate(rate.RX + rate.TX))
	}

	tiles := container.New(layout.NewGridLayoutWithColumns(5),
		wgwidget.StatTile(wgtheme.IconPeers, d.lang.T(lang.Detail_metric_peers),
			strconv.Itoa(len(dev.Peers)), "", p.TextPrimary),
		wgwidget.StatTile(wgtheme.IconSignal, d.lang.T(lang.Detail_metric_online),
			strconv.Itoa(online), d.lang.Tf(lang.Detail_metric_of, len(dev.Peers)), onlineTone),
		wgwidget.StatTile(wgtheme.IconArrowDown, d.lang.T(lang.Detail_metric_rx), rxV, rxU, p.TextPrimary),
		wgwidget.StatTile(wgtheme.IconArrowUp, d.lang.T(lang.Detail_metric_tx), txV, txU, p.TextPrimary),
		wgwidget.StatTile(wgtheme.IconActivity, d.lang.T(lang.Detail_metric_rate), rateV, rateU, p.Accent),
	)

	if len(history) < 2 {
		return tiles
	}
	rx := make([]float64, 0, len(history))
	tx := make([]float64, 0, len(history))
	for _, s := range history {
		rx = append(rx, s.RX)
		tx = append(tx, s.TX)
	}
	chart := wgwidget.NewTrafficChart(110, d.lang.T(lang.Detail_metric_rx), d.lang.T(lang.Detail_metric_tx), wgstat.FormatRate)
	chart.SetSeries(rx, tx)

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD+2),
		tiles, wgwidget.Card(chart, 0))
}

// splitUnit separates "1.4 GiB" into "1.4" and "GiB" for the stat tiles.
func splitUnit(s string) (string, string) {
	if i := strings.LastIndexByte(s, ' '); i > 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// interfaceGrid lists the interface's own settings, preferring live values
// where the kernel and the file can disagree (notably the listen port, which
// wg-quick assigns dynamically when the config leaves it unset). Every value
// copies itself when clicked.
func (d *tunnelDetail) interfaceGrid(iface config.Interface, cfg *config.Config,
	dev wgstat.Device, live bool, configPath string) fyne.CanvasObject {

	address := iface.IP
	endpoint, mtu, dns, table := wgstat.Dash, wgstat.Dash, wgstat.Dash, wgstat.Dash
	port := wgstat.Dash

	if cfg != nil {
		if addrs := cfg.Interface.Address; len(addrs) > 0 {
			parts := make([]string, 0, len(addrs))
			for _, a := range addrs {
				parts = append(parts, a.String())
			}
			address = strings.Join(parts, ", ")
		}
		// The tunnel's own endpoint when it has one, plus any peer that dials
		// somewhere else — a config with a second gateway or a per-peer port
		// has more than one answer, and showing only the first would hide it.
		if eps := wgcfg.Endpoints(cfg); len(eps) > 0 {
			endpoint = strings.Join(eps, ", ")
		}
		if cfg.Interface.MTU > 0 {
			mtu = strconv.Itoa(cfg.Interface.MTU)
		}
		if len(cfg.Interface.DNS) > 0 {
			parts := make([]string, 0, len(cfg.Interface.DNS))
			for _, ip := range cfg.Interface.DNS {
				parts = append(parts, ip.String())
			}
			dns = strings.Join(parts, ", ")
		}
		if cfg.Interface.Table != "" {
			table = cfg.Interface.Table
		}
		if cfg.Interface.ListenPort != nil && *cfg.Interface.ListenPort > 0 {
			port = strconv.Itoa(*cfg.Interface.ListenPort)
		}
	}

	// The running port is authoritative when the tunnel is up.
	if live && dev.ListenPort > 0 {
		port = strconv.Itoa(dev.ListenPort)
	}
	// wg-quick assigns a port when the config leaves it unset, so a bare dash
	// reads as missing data when the truth is "chosen at start".
	if port == wgstat.Dash && cfg != nil {
		port = d.lang.T(lang.Detail_port_auto)
	}

	pubKey := iface.PublicKey
	if live && dev.PublicKey != "" {
		pubKey = dev.PublicKey
	}
	if pubKey == "" {
		pubKey = wgstat.Dash
	}

	// A dash, or the "auto" placeholder standing in for an unset port, has
	// nothing worth copying, so those fields are read-only.
	field := func(icon *wgtheme.Icon, labelKey, value string) fyne.CanvasObject {
		label := d.lang.T(labelKey)
		var onCopy func(string)
		if value != wgstat.Dash && value != d.lang.T(lang.Detail_port_auto) {
			onCopy = func(v string) { d.copyField(label, v) }
		}
		return wgwidget.NewCopyField(icon, label, value, onCopy)
	}

	grid := container.New(layout.NewGridLayoutWithColumns(2),
		field(wgtheme.IconRoute, lang.Detail_label_address, address),
		field(wgtheme.IconGlobe, lang.Detail_label_endpoint, endpoint),
		field(wgtheme.IconPlug, lang.Detail_label_port, port),
		field(wgtheme.IconRuler, lang.Detail_label_mtu, mtu),
		field(wgtheme.IconDNS, lang.Detail_label_dns, dns),
		field(wgtheme.IconTable, lang.Detail_label_table, table),
		field(wgtheme.IconKey, lang.Detail_label_public_key, pubKey),
		field(wgtheme.IconDoc, lang.Detail_label_config, configPath),
	)
	return wgwidget.Card(container.New(layout.NewCustomPaddedLayout(-4, -4, -4, -4), grid), 0)
}

// copyField copies a value and names it in a toast, which doubles as the way
// to read a value the layout had to truncate.
func (d *tunnelDetail) copyField(label, value string) {
	d.actions.OnCopy(d.lang.Tf(lang.Status_copied_field, label, value), value)
}

// peersSection is the live peer table, or an invitation to add the first
// peer. Each row carries a QR action, so a client config can be shown
// whenever it is needed rather than only in the session that created it.
func (d *tunnelDetail) peersSection(iface config.Interface, cfg *config.Config, dev wgstat.Device, live bool) fyne.CanvasObject {
	var configured []config.PeerConfig
	if cfg != nil {
		configured = cfg.Peers
	}
	title := d.lang.T(lang.Detail_section_peers)

	if len(configured) == 0 {
		add := wgwidget.NewButton(d.lang.T(lang.Detail_add_peer), wgtheme.IconAdd, wgwidget.Primary, func() {
			d.actions.OnPeers(iface.Name)
		})
		return d.section(wgtheme.IconPeers, title, nil, wgwidget.Card(
			wgwidget.EmptyState(wgtheme.IconEmptyPeers, d.lang.T(lang.Detail_no_peers), "", add), 0))
	}

	d.peerRows = buildPeerRows(configured, dev, live, d.lang)

	manage := wgwidget.NewButton(d.lang.T(lang.Detail_manage_peers), wgtheme.IconEdit, wgwidget.Ghost, func() {
		d.actions.OnPeers(iface.Name)
	})
	manage.Scale = wgwidget.ScaleSM

	table := d.peers.build(&d.peerRows, d.lang,
		func(publicKey string) { d.actions.OnPeerQR(iface.Name, publicKey) },
		func(column, value string) { d.copyField(column, value) },
	)

	hint := widget.NewLabel(d.lang.T(lang.Detail_peers_hint))
	hint.Importance = widget.LowImportance

	return d.section(wgtheme.IconPeers, title,
		[]fyne.CanvasObject{wgwidget.Badge(strconv.Itoa(len(d.peerRows)), false), manage},
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXS),
			fixedHeight(table, table.Height(len(d.peerRows), 10)),
			container.New(layout.NewCustomPaddedLayout(-2, 0, -4, 0), hint),
		))
}

// routes summarises every network reachable through the tunnel, which is the
// question an operator actually asks of a hub-and-spoke config.
func (d *tunnelDetail) routes(cfg *config.Config) fyne.CanvasObject {
	if cfg == nil {
		return nil
	}

	seen := map[string]bool{}
	var nets []string
	for _, p := range cfg.Peers {
		for _, n := range p.AllowedIPs {
			s := n.String()
			if !seen[s] {
				seen[s] = true
				nets = append(nets, s)
			}
		}
	}
	if len(nets) == 0 {
		return nil
	}
	sort.Strings(nets)

	chips := make([]fyne.CanvasObject, 0, len(nets))
	for _, n := range nets {
		cidr := n
		chips = append(chips, wgwidget.NewCopyChip(cidr, func() {
			d.copyField(d.lang.T(lang.Detail_section_routes), cidr)
		}))
	}

	return d.section(wgtheme.IconRoute, d.lang.T(lang.Detail_section_routes),
		[]fyne.CanvasObject{wgwidget.Badge(strconv.Itoa(len(nets)), false)},
		container.NewGridWrap(fyne.NewSize(156, 28), chips...))
}

// summaryLine is the sidebar row's badge: online/total peers when live,
// otherwise the configured peer count.
func summaryLine(iface config.Interface, cfg *config.Config, dev wgstat.Device, live bool) string {
	if live {
		online := dev.OnlinePeers(time.Now())
		return fmt.Sprintf("%d/%d", online, len(dev.Peers))
	}
	if cfg != nil && len(cfg.Peers) > 0 {
		return strconv.Itoa(len(cfg.Peers))
	}
	return ""
}
