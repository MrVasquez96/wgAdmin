package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
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
	OnToggle func(name string, up bool)
	OnScan   func(name, ip string)
	OnPeers  func(name string)
	OnEdit   func(name string)
	OnConfig func(name string)
	OnDelete func(name string)
	OnCopy   func(label, value string)
	OnPeerQR func(tunnel, publicKey string)
	OnNew    func()
	OnImport func()
}

// tunnelDetail renders the right-hand pane for one tunnel.
type tunnelDetail struct {
	lang    *lang.Language
	actions detailActions

	// peerRows backs the table; the table reads it through a pointer so rows
	// can be refreshed without rebuilding the widget.
	peerRows []peerRow
}

func newTunnelDetail(l *lang.Language, actions detailActions) *tunnelDetail {
	return &tunnelDetail{lang: l, actions: actions}
}

// emptyState is shown when nothing is selected, there are no tunnels at all, or
// the filter excludes everything. Each case gets its own wording and, where
// useful, its own call to action.
func (d *tunnelDetail) emptyState(titleKey, bodyKey string, withActions bool) fyne.CanvasObject {
	title := wgwidget.Title(d.lang.T(titleKey), 17)
	title.Alignment = fyne.TextAlignCenter

	body := widget.NewLabel(d.lang.T(bodyKey))
	body.Alignment = fyne.TextAlignCenter
	body.Wrapping = fyne.TextWrapWord
	body.Importance = widget.LowImportance

	items := []fyne.CanvasObject{
		container.NewCenter(title),
		container.NewCenter(container.New(
			layout.NewGridWrapLayout(fyne.NewSize(360, body.MinSize().Height*2)), body)),
	}

	if withActions {
		newBtn := widget.NewButtonWithIcon(d.lang.T(lang.Main_add_tunnel), wgtheme.IconAdd, d.actions.OnNew)
		newBtn.Importance = widget.HighImportance
		importBtn := widget.NewButtonWithIcon(d.lang.T(lang.Main_import), wgtheme.IconImport, d.actions.OnImport)
		items = append(items, container.NewCenter(
			container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD), newBtn, importBtn)))
	}

	return container.NewCenter(container.New(
		layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG), items...))
}

// build renders the full detail pane for iface. cfg may be nil when the config
// could not be parsed; live carries the kernel's view when it is available.
func (d *tunnelDetail) build(iface config.Interface, cfg *config.Config,
	snap wgstat.Snapshot, rate wgstat.Sample, history []wgstat.Sample, configPath string) fyne.CanvasObject {

	dev, hasDev := snap.Device(iface.Name)
	live := snap.Available && hasDev && iface.Active

	sections := []fyne.CanvasObject{
		d.header(iface),
		d.metrics(iface, dev, live, snap, rate, history),
		wgwidget.Hairline(),
		wgwidget.SectionHeader(d.lang.T(lang.Detail_section_interface)),
		d.interfaceGrid(iface, cfg, dev, live, configPath),
	}

	sections = append(sections,
		wgwidget.SectionHeader(d.lang.T(lang.Detail_section_peers)),
		d.peers(iface, cfg, dev, live),
	)

	if routes := d.routes(cfg); routes != nil {
		sections = append(sections,
			wgwidget.SectionHeader(d.lang.T(lang.Detail_section_routes)),
			routes,
		)
	}

	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG), sections...)
	return container.NewVScroll(container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceLG, wgtheme.SpaceXL, wgtheme.SpaceLG, wgtheme.SpaceLG), body))
}

// header is the tunnel name, its state pill and the action bar. The primary
// up/down action is visually separated from the secondary operations so the
// five equal-weight buttons of the old card become a clear hierarchy.
func (d *tunnelDetail) header(iface config.Interface) fyne.CanvasObject {
	name := wgwidget.Title(iface.Name, 21)

	pill := wgwidget.StatePill(iface.Active,
		d.lang.T(lang.Card_state_up), d.lang.T(lang.Card_state_down))

	titleRow := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceLG),
		name, container.NewCenter(pill))

	var toggle *widget.Button
	if iface.Active {
		toggle = widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_deactivate), wgtheme.IconStop, func() {
			d.actions.OnToggle(iface.Name, false)
		})
		toggle.Importance = widget.DangerImportance
	} else {
		toggle = widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_activate), wgtheme.IconPlay, func() {
			d.actions.OnToggle(iface.Name, true)
		})
		toggle.Importance = widget.SuccessImportance
	}

	scan := widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_scan), wgtheme.IconScan, func() {
		d.actions.OnScan(iface.Name, iface.IP)
	})
	if iface.IP == "unknown" {
		scan.Disable()
	}
	peers := widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_peers), wgtheme.IconPeers, func() {
		d.actions.OnPeers(iface.Name)
	})
	edit := widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_edit), wgtheme.IconEdit, func() {
		d.actions.OnEdit(iface.Name)
	})
	cfgBtn := widget.NewButtonWithIcon(d.lang.T(lang.Detail_btn_config), wgtheme.IconDoc, func() {
		d.actions.OnConfig(iface.Name)
	})
	del := widget.NewButtonWithIcon(d.lang.T(lang.Card_btn_delete), wgtheme.IconTrash, func() {
		d.actions.OnDelete(iface.Name)
	})
	del.Importance = widget.DangerImportance

	for _, b := range []*widget.Button{scan, peers, edit, cfgBtn} {
		b.Importance = widget.MediumImportance
	}

	actions := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		toggle,
		wgwidget.VHairline(),
		scan, peers, edit, cfgBtn,
		layout.NewSpacer(),
		del,
	)

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD), titleRow, actions)
}

// metrics is the row of live counters. When telemetry is unavailable it is
// replaced by an explanation, because a row of dashes with no reason is worse
// than no row at all.
func (d *tunnelDetail) metrics(iface config.Interface, dev wgstat.Device, live bool,
	snap wgstat.Snapshot, rate wgstat.Sample, history []wgstat.Sample) fyne.CanvasObject {

	p := wgtheme.Current()

	if !snap.Available {
		return d.telemetryNotice()
	}

	peerCount, online := 0, 0
	rx, tx, throughput := wgstat.Dash, wgstat.Dash, wgstat.Dash

	if live {
		peerCount = len(dev.Peers)
		online = dev.OnlinePeers(time.Now())
		rx = wgstat.FormatBytes(dev.RX)
		tx = wgstat.FormatBytes(dev.TX)
		if rate.RX > 0 || rate.TX > 0 {
			throughput = wgstat.FormatRate(rate.RX + rate.TX)
		}
	}

	onlineTone := p.TextPrimary
	if live && online > 0 {
		onlineTone = p.Success
	}

	tiles := []fyne.CanvasObject{
		wgwidget.Metric(d.lang.T(lang.Detail_metric_peers), strconv.Itoa(peerCount), ""),
		wgwidget.ToneMetric(d.lang.T(lang.Detail_metric_online), strconv.Itoa(online), "", onlineTone),
		wgwidget.Metric(d.lang.T(lang.Detail_metric_rx), rx, ""),
		wgwidget.Metric(d.lang.T(lang.Detail_metric_tx), tx, ""),
		wgwidget.Metric(d.lang.T(lang.Detail_metric_rate), throughput, ""),
	}

	if !live {
		// An inactive tunnel has no counters at all; show dashes rather than
		// zeroes so it never reads as "up but silent".
		tiles = []fyne.CanvasObject{
			wgwidget.Metric(d.lang.T(lang.Detail_metric_peers), wgstat.Dash, ""),
			wgwidget.Metric(d.lang.T(lang.Detail_metric_online), wgstat.Dash, ""),
			wgwidget.Metric(d.lang.T(lang.Detail_metric_rx), wgstat.Dash, ""),
			wgwidget.Metric(d.lang.T(lang.Detail_metric_tx), wgstat.Dash, ""),
			wgwidget.Metric(d.lang.T(lang.Detail_metric_rate), wgstat.Dash, ""),
		}
	}

	row := container.New(layout.NewGridLayoutWithColumns(len(tiles)), tiles...)

	if live && len(history) > 1 {
		values := make([]float64, 0, len(history))
		for _, s := range history {
			values = append(values, s.RX+s.TX)
		}
		spark := wgwidget.NewSparkline(p.Accent, fyne.NewSize(0, 26))
		spark.SetSamples(values)
		row = container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD), row, spark.Build())
	}

	return wgwidget.RaisedPanel(row)
}

// telemetryNotice explains why live numbers are missing instead of silently
// showing nothing.
func (d *tunnelDetail) telemetryNotice() fyne.CanvasObject {
	p := wgtheme.Current()

	title := wgwidget.Caption(d.lang.T(lang.Detail_telemetry_off))
	title.Color = p.Warning

	body := widget.NewLabel(d.lang.T(lang.Detail_telemetry_off_detail))
	body.Wrapping = fyne.TextWrapWord
	body.Importance = widget.LowImportance

	return wgwidget.RaisedPanel(container.New(
		layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM), title, body))
}

// interfaceGrid lists the interface's own settings, preferring live values
// where the kernel and the file can disagree (notably the listen port, which
// wg-quick assigns dynamically when the config leaves it unset).
func (d *tunnelDetail) interfaceGrid(iface config.Interface, cfg *config.Config,
	dev wgstat.Device, live bool, configPath string) fyne.CanvasObject {

	const labelWidth = 96

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

	rows := []fyne.CanvasObject{
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_address), address, labelWidth),
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_endpoint), endpoint, labelWidth),
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_port), port, labelWidth),
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_mtu), mtu, labelWidth),
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_dns), dns, labelWidth),
		wgwidget.KeyValue(d.lang.T(lang.Detail_label_table), table, labelWidth),
	}

	if pubKey != wgstat.Dash {
		key := pubKey
		rows = append(rows, wgwidget.KeyValueCopy(
			d.lang.T(lang.Detail_label_public_key), key, labelWidth, func() {
				d.actions.OnCopy(d.lang.T(lang.Status_copied_pubkey), key)
			}))
	}

	rows = append(rows, wgwidget.KeyValueCopy(
		d.lang.T(lang.Detail_label_config), configPath, labelWidth, func() {
			d.actions.OnCopy(d.lang.T(lang.Common_copied), configPath)
		}))

	return wgwidget.Panel(container.New(
		layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM), rows...), nil)
}

// peers renders the live peer table, or an explanation when there are none.
//
// Selecting a row offers that peer's client config, so a QR code can be shown
// whenever it is needed rather than only in the session that created the peer.
func (d *tunnelDetail) peers(iface config.Interface, cfg *config.Config, dev wgstat.Device, live bool) fyne.CanvasObject {
	var configured []config.PeerConfig
	if cfg != nil {
		configured = cfg.Peers
	}

	if len(configured) == 0 {
		return emptyNote(d.lang.T(lang.Detail_no_peers))
	}

	d.peerRows = buildPeerRows(configured, dev, live, d.lang)

	table := newPeerTable(&d.peerRows, d.lang, func(publicKey string) {
		d.actions.OnPeerQR(iface.Name, publicKey)
	})

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXS),
		fixedHeight(table, peerTableHeight(len(d.peerRows))),
		formHint(d.lang.T(lang.Detail_peers_hint)),
	)
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
		chips = append(chips, wgwidget.NeutralChip(n))
	}

	return container.NewGridWrap(fyne.NewSize(150, 22), chips...)
}

// summaryLine is the one-line description used in the sidebar row's meta slot.
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
