// Command gallery is a developer window that shows every wgAdmin component in
// its states, in either theme, so the design system can be reviewed without
// clicking through the real app.
//
//	go run ./cmd/gallery            # dark
//	go run ./cmd/gallery -theme light
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/settings"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"
)

func main() {
	variant := flag.String("theme", "dark", "dark or light")
	scroll := flag.Float64("scroll", 0, "initial scroll offset, for screenshots")
	demo := flag.String("demo", "", "open an overlay at start: modal, sheet, menu, toast")
	flag.Parse()

	a := app.NewWithID("com.wireguard.manager.gallery")
	cfg := settings.Load(a.Preferences())
	cfg.ThemeVariant = *variant
	a.Settings().SetTheme(wgtheme.NewWGAdminTheme(cfg))

	w := a.NewWindow("wgAdmin · component gallery")
	w.Resize(fyne.NewSize(1180, 900))

	var rebuild func()
	rebuild = func() {
		sc := build(w, cfg, func(v string) {
			cfg.ThemeVariant = v
			a.Settings().SetTheme(wgtheme.NewWGAdminTheme(cfg))
			rebuild()
		})
		w.SetContent(wgwidget.WithLayer(w, sc))
		sc.Offset = fyne.NewPos(0, float32(*scroll))
		sc.Refresh()
	}
	rebuild()
	if *demo != "" {
		go func() {
			time.Sleep(700 * time.Millisecond)
			fyne.Do(func() {
				if b, ok := demoButtons[*demo]; ok {
					b.OnTapped()
				}
			})
		}()
	}
	w.ShowAndRun()
}

// demoButtons lets -demo open an overlay without a real click.
var demoButtons = map[string]*wgwidget.Button{}

func section(title string, objs ...fyne.CanvasObject) fyne.CanvasObject {
	items := append([]fyne.CanvasObject{wgwidget.SectionTitle(nil, title)}, objs...)
	return wgwidget.Card(container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG), items...), 0)
}

func row(objs ...fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD), objs...)
}

func build(w fyne.Window, cfg *settings.AppSettings, setTheme func(string)) *container.Scroll {
	p := wgtheme.Current()
	toast := func(s string) func() {
		return func() { wgwidget.ShowToast(w, wgwidget.ToastSuccess, s) }
	}

	// Theme switch at the top.
	themes := wgwidget.NewSegmented([]string{"Dark", "Light"}, func(i int) {
		setTheme([]string{"dark", "light"}[i])
	})
	themes.Icons = []*wgtheme.Icon{wgtheme.IconMoon, wgtheme.IconSun}
	if cfg.ThemeVariant == "light" {
		themes.SetSelected(1)
	}
	header := container.NewBorder(nil, nil,
		row(container.NewCenter(widget.NewIcon(wgtheme.IconBrand)),
			wgwidget.Title("Component gallery", wgtheme.TextXL)),
		themes)

	// Buttons.
	disabled := wgwidget.NewButton("Disabled", wgtheme.IconEdit, wgwidget.Secondary, nil)
	disabled.Disable()
	loadingSmall := wgwidget.NewButton("Small", wgtheme.IconAdd, wgwidget.Secondary, toast("small"))
	loadingSmall.Scale = wgwidget.ScaleSM
	large := wgwidget.NewButton("Bring up", wgtheme.IconPlay, wgwidget.Success, toast("Bring up"))
	large.Scale = wgwidget.ScaleLG
	buttons := section("Buttons",
		row(
			wgwidget.NewButton("New tunnel", wgtheme.IconAdd, wgwidget.Primary, toast("Primary")),
			wgwidget.NewButton("Scan", wgtheme.IconScan, wgwidget.Secondary, toast("Secondary")),
			wgwidget.NewButton("Peers", wgtheme.IconPeers, wgwidget.Ghost, toast("Ghost")),
			wgwidget.NewButton("Delete", wgtheme.IconTrash, wgwidget.Danger, toast("Danger")),
			wgwidget.NewButton("Delete tunnel", nil, wgwidget.DangerSolid, toast("DangerSolid")),
			disabled,
		),
		row(large, loadingSmall,
			wgwidget.NewIconButton(wgtheme.IconRefresh, "Refresh now", toast("refresh")),
			wgwidget.NewIconButton(wgtheme.IconFullscreen, "Fullscreen (F11)", toast("fullscreen")),
			wgwidget.NewIconButton(wgtheme.IconSettings, "Settings", toast("settings")),
			wgwidget.NewIconButton(wgtheme.IconQR, "Show QR code", toast("qr")),
		),
	)

	// Controls.
	seg := wgwidget.NewSegmented([]string{"Off", "2s", "5s", "10s", "30s"}, nil)
	seg.SetSelected(2)
	density := wgwidget.NewSegmented([]string{"Comfortable", "Compact"}, nil)
	swOn := wgwidget.NewSwitch(true, nil)
	swOff := wgwidget.NewSwitch(false, nil)
	swDis := wgwidget.NewSwitch(true, nil)
	swDis.Disable()
	controls := section("Controls",
		row(seg, density),
		row(swOn, widget.NewLabel("Auto refresh"), swOff, widget.NewLabel("Reduce motion"), swDis, widget.NewLabel("Disabled")),
	)

	// Status.
	status := section("Status",
		row(
			wgwidget.StatusPill(true, "UP", "DOWN"),
			wgwidget.StatusPill(false, "UP", "DOWN"),
			wgwidget.Tag("root", p.Success, wgtheme.IconLock),
			wgwidget.Tag("no root", p.Warning, wgtheme.IconAlert),
			wgwidget.Tag("wg-quick", p.TextSecondary, nil),
			wgwidget.Badge("4", false),
			wgwidget.Badge("2/4", true),
		),
		wgwidget.Banner(wgwidget.BannerWarning, "Live data unavailable",
			"Run wgAdmin as root to read handshakes and traffic counters from the kernel.",
			wgwidget.NewButton("Relaunch as root", wgtheme.IconLock, wgwidget.Secondary, toast("relaunch"))),
		wgwidget.Banner(wgwidget.BannerInfo, "Backups", "A backup is written before every change.", nil),
	)

	// Stats + chart.
	rx, tx := make([]float64, 60), make([]float64, 60)
	for i := range rx {
		f := float64(i)
		rx[i] = 40e3 + 30e3*math.Sin(f/6) + 18e3*math.Sin(f/2.3) + 50e3
		tx[i] = 15e3 + 10e3*math.Sin(f/4+1) + 20e3
	}
	chart := wgwidget.NewTrafficChart(120, "Received", "Sent", wgstat.FormatRate)
	chart.SetSeries(rx, tx)
	stats := section("Live metrics",
		container.New(layout.NewGridLayoutWithColumns(5),
			wgwidget.StatTile(wgtheme.IconPeers, "Peers", "4", "", p.TextPrimary),
			wgwidget.StatTile(wgtheme.IconSignal, "Online", "3", "of 4", p.Success),
			wgwidget.StatTile(wgtheme.IconArrowDown, "Received", "1.4", "GiB", p.TextPrimary),
			wgwidget.StatTile(wgtheme.IconArrowUp, "Sent", "212", "MiB", p.TextPrimary),
			wgwidget.StatTile(wgtheme.IconActivity, "Rate", "96.2", "KiB/s", p.Accent),
		),
		chart,
	)

	// Copy fields.
	copied := func(v string) {
		w.Clipboard().SetContent(v)
		wgwidget.ShowToast(w, wgwidget.ToastSuccess, "Copied "+v)
	}
	fields := section("Copy fields (hover, click)",
		container.New(layout.NewGridLayoutWithColumns(2),
			wgwidget.NewCopyField(wgtheme.IconRoute, "Address", "10.50.2.1/32", copied),
			wgwidget.NewCopyField(wgtheme.IconGlobe, "Endpoint", "example.com:51832", copied),
			wgwidget.NewCopyField(wgtheme.IconPlug, "Listen port", "auto", nil),
			wgwidget.NewCopyField(wgtheme.IconRuler, "MTU", "1420", copied),
			wgwidget.NewCopyField(wgtheme.IconDNS, "DNS", "1.1.1.1, 9.9.9.9", copied),
			wgwidget.NewCopyField(wgtheme.IconTable, "Table", "auto", copied),
			wgwidget.NewCopyField(wgtheme.IconKey, "Public key", "wdNCKwHAQ/AByQM8XaPPCUrHb7WI8n5N9kn2qBYlNQU=", copied),
			wgwidget.NewCopyField(wgtheme.IconDoc, "Config", "/etc/wireguard/wgMockup.conf", copied),
		),
	)

	// Table.
	type peer struct {
		name, ips, ep, hs, rx, tx string
		state                     int
	}
	peers := []peer{
		{"peer_mock_1", "10.50.2.2/32", "example.com:51832", "12s ago", "1.2 GiB", "88 MiB", 0},
		{"peer_mock_2", "10.50.2.3/32", "example.com:51833", "1m ago", "320 MiB", "12 MiB", 1},
		{"peer_mock_3", "10.50.2.4/32", "—", "3h ago", "4 KiB", "1 KiB", 2},
		{"wvmPZd4//HS…", "10.2.0.0/22, 10.2.4.0/22", "wg1.mohntechnology.no:51834", "—", "—", "—", 3},
	}
	tones := []func() wgwidget.Cell{
		func() wgwidget.Cell { return wgwidget.Cell{Dot: p.Success, Live: true} },
		func() wgwidget.Cell { return wgwidget.Cell{Dot: p.Warning} },
		func() wgwidget.Cell { return wgwidget.Cell{Dot: p.Error} },
		func() wgwidget.Cell { return wgwidget.Cell{Dot: p.TextDim} },
	}
	cols := []wgwidget.Column{
		{Dot: true, MinWidth: 26},
		{Title: "Peer", MinWidth: 110, Weight: 1.2, Sortable: true},
		{Title: "Allowed IPs", MinWidth: 120, Weight: 1.4, Mono: true, Sortable: true},
		{Title: "Endpoint", MinWidth: 120, Weight: 1.6, Mono: true},
		{Title: "Handshake", MinWidth: 90, Sortable: true},
		{Title: "RX", MinWidth: 70, Mono: true, Sortable: true},
		{Title: "TX", MinWidth: 70, Mono: true, Sortable: true},
	}
	var table *wgwidget.DataTable
	table = wgwidget.NewDataTable(cols, func() int { return len(peers) }, func(r, c int) wgwidget.Cell {
		pr := peers[r]
		switch c {
		case 0:
			return tones[pr.state]()
		case 1:
			return wgwidget.Cell{Text: pr.name, Strong: true}
		case 2:
			return wgwidget.Cell{Text: pr.ips}
		case 3:
			return wgwidget.Cell{Text: pr.ep, Dim: pr.ep == "—"}
		case 4:
			return wgwidget.Cell{Text: pr.hs, Dim: true}
		case 5:
			return wgwidget.Cell{Text: pr.rx}
		}
		return wgwidget.Cell{Text: pr.tx}
	})
	table.Actions = []wgwidget.RowAction{
		{Icon: wgtheme.IconQR, Tooltip: "Show QR code", OnTap: func(r int) { toast("QR " + peers[r].name)() }},
		{Icon: wgtheme.IconMoreVertical, Tooltip: "More", OnTap: func(r int) { toast("more")() }},
	}
	table.OnCellTap = func(r, c int) { toast(fmt.Sprintf("Copied %s: row %d", cols[c].Title, r+1))() }
	table.OnSort = func(c int, asc bool) {
		sort.SliceStable(peers, func(i, j int) bool {
			a, b := peers[i].name, peers[j].name
			switch c {
			case 2:
				a, b = peers[i].ips, peers[j].ips
			case 4:
				a, b = peers[i].hs, peers[j].hs
			case 5:
				a, b = peers[i].rx, peers[j].rx
			case 6:
				a, b = peers[i].tx, peers[j].tx
			}
			if asc {
				return a < b
			}
			return a > b
		})
	}
	tableBox := container.NewStack(
		wgwidgetFixed(table.Height(len(peers), 8)),
		table)

	// Overlays.
	var moreBtn *wgwidget.Button
	moreBtn = wgwidget.NewButton("Menu", wgtheme.IconMoreVertical, wgwidget.Secondary, func() {
		wgwidget.ShowMenu(moreBtn, []wgwidget.MenuItem{
			{Label: "Show config", Icon: wgtheme.IconDoc, Action: toast("config")},
			{Label: "Export client configs", Icon: wgtheme.IconExport, Action: toast("export")},
			{Separator: true},
			{Label: "Delete tunnel", Icon: wgtheme.IconTrash, Danger: true, Action: toast("delete")},
		})
	})
	var sheet *wgwidget.Surface
	modalBtn := wgwidget.NewButton("Modal", wgtheme.IconInfo, wgwidget.Secondary, func() {
		var s *wgwidget.Surface
		s = wgwidget.ShowModal(w, "Delete wgMockup?", 440,
			widget.NewLabel("The tunnel's config file is removed.\nA backup is kept."),
			wgwidget.NewButton("Cancel", nil, wgwidget.Secondary, func() { s.Close() }),
			wgwidget.NewButton("Delete tunnel", wgtheme.IconTrash, wgwidget.DangerSolid, func() { s.Close() }))
	})
	sheetBtn := wgwidget.NewButton("Sheet", wgtheme.IconEdit, wgwidget.Secondary, func() {
		name := widget.NewEntry()
		name.SetText("wgMockup")
		addr := widget.NewEntry()
		addr.SetText("10.50.2.1/32")
		port := widget.NewEntry()
		port.SetPlaceHolder("auto")
		mtu := widget.NewEntry()
		mtu.SetText("1420")
		body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXL),
			wgwidget.FormSection(wgtheme.IconRoute, "Interface", "How this machine appears on the tunnel.",
				wgwidget.Field("Name", name, "Becomes the file name under /etc/wireguard."),
				wgwidget.Field("Address", addr, ""),
				wgwidget.FieldRow(wgwidget.Field("Listen port", port, ""), wgwidget.Field("MTU", mtu, "")),
			),
		)
		sheet = wgwidget.ShowSheet(w, "Edit tunnel", wgwidget.SheetWide, body,
			wgwidget.NewButton("Cancel", nil, wgwidget.Secondary, func() { sheet.Close() }),
			wgwidget.NewButton("Save", wgtheme.IconCheck, wgwidget.Primary, func() { sheet.Close() }))
	})
	toastBtn := wgwidget.NewButton("Toast", wgtheme.IconSuccess, wgwidget.Secondary, toast("Copied Endpoint: example.com:51832"))
	demoButtons["modal"], demoButtons["sheet"], demoButtons["menu"], demoButtons["toast"] = modalBtn, sheetBtn, moreBtn, toastBtn
	overlays := section("Overlays",
		row(
			modalBtn, sheetBtn, moreBtn, toastBtn,
			wgwidget.NewButton("Error toast", wgtheme.IconError, wgwidget.Secondary, func() {
				wgwidget.ShowToast(w, wgwidget.ToastError, "wg-quick up failed: exit status 1")
			}),
		),
	)

	empty := section("Empty states",
		container.New(layout.NewGridLayoutWithColumns(3),
			wgwidget.EmptyState(wgtheme.IconEmptyTunnels, "No tunnels yet", "Create a tunnel or import an existing .conf file.",
				wgwidget.NewButton("New tunnel", wgtheme.IconAdd, wgwidget.Primary, nil)),
			wgwidget.EmptyState(wgtheme.IconEmptyPeers, "No peers", "Add a peer to let a device join this tunnel."),
			wgwidget.EmptyState(wgtheme.IconEmptySearch, "No matches", "No tunnel matches the filter."),
		),
	)

	// Every icon, small.
	names := make([]string, 0, len(wgtheme.Icons()))
	for n := range wgtheme.Icons() {
		names = append(names, n)
	}
	sort.Strings(names)
	cells := make([]fyne.CanvasObject, 0, len(names))
	for _, n := range names {
		cells = append(cells, wgwidget.WithTooltip(container.NewCenter(
			container.New(layout.NewGridWrapLayout(fyne.NewSize(22, 22)),
				widget.NewIcon(wgtheme.Icons()[n].For(p.TextPrimary)))), n))
	}
	icons := section("Icons (hover for names)", container.NewGridWrap(fyne.NewSize(34, 34), cells...))

	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		header, buttons, controls, status, stats, fields,
		section("Data table (hover rows, click cells, sort headers)", tableBox),
		overlays, empty, icons)
	return container.NewVScroll(container.New(layout.NewCustomPaddedLayout(20, 28, 24, 24), body))
}

// wgwidgetFixed is a transparent spacer fixing a minimum height. It must be
// visible: Fyne leaves hidden objects out of MinSize, which made it zero.
func wgwidgetFixed(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(0, h))
	return r
}
