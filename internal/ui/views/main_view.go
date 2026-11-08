package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgcfg"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"
)

// MainView is the application console: a toolbar, a tunnel sidebar, a detail
// pane for the selected tunnel and a status strip.
type MainView struct {
	window   fyne.Window
	ctrl     *wg.WG
	settings *settings.AppSettings
	lang     *lang.Language

	meta          fyne.AppMetadata
	statusBar     *wgwidget.StatusBar
	busyDialog    *wgwidget.BusyDialog
	filterEntry   *widget.Entry
	liveChip      *wgwidget.ChipButton
	fullscreenBtn *wgwidget.Button

	sidebar       *widget.List
	sidebarFooter *fyne.Container
	detailBox     *fyne.Container
	split         *wgwidget.SideSplit

	// interfaces is everything on disk; filtered is what passes the filter, in
	// display order; entries is filtered grouped under sidebar headings.
	interfaces []config.Interface
	filtered   []config.Interface
	entries    []sidebarEntry
	selected   string

	// The detail pane is rebuilt from scratch, so its scroll position and the
	// inputs it was last built from are kept to carry the one over and to skip
	// rebuilds when nothing visible changed.
	detailScroll *container.Scroll
	detailName   string
	detailSig    string
	shortcutsSet bool

	// configs caches parsed configs so the detail pane and sidebar do not
	// re-read and re-parse every file on each refresh tick.
	configs map[string]*config.Config

	snapshot wgstat.Snapshot
	tracker  *wgstat.Tracker
	detail   *tunnelDetail

	// autoMu guards the auto-refresh goroutine's lifecycle. The previous
	// implementation could leak tickers when settings changed while running.
	autoMu   sync.Mutex
	autoStop chan struct{}

	lastRefresh time.Time
}

// NewMainView creates the console.
func NewMainView(window fyne.Window, ctrl *wg.WG, cfg *settings.AppSettings, language *lang.Language) *MainView {
	v := &MainView{
		window:      window,
		ctrl:        ctrl,
		settings:    cfg,
		lang:        language,
		statusBar:   wgwidget.NewStatusBar(language),
		busyDialog:  wgwidget.NewBusyDialog(window),
		filterEntry: widget.NewEntry(),
		configs:     map[string]*config.Config{},
		tracker:     wgstat.NewTracker(),
	}
	v.detail = newTunnelDetail(language, v.detailActions())

	helpers.SetLabels(language.T(lang.Common_ok), language.T(lang.Common_cancel),
		language.T(lang.Common_yes), language.T(lang.Common_no))
	helpers.SetErrorTitle(language.T(lang.Common_error))
	return v
}

// detailActions wires the detail pane's buttons back to the view's operations.
func (v *MainView) detailActions() detailActions {
	return detailActions{
		OnToggle: v.toggleInterface,
		OnScan:   v.showScan,
		OnPeers:  v.showEditPeersTunnelForm,
		OnEdit:   v.showEditTunnelForm,
		OnConfig: v.showConfigView,
		OnDelete: v.confirmDeleteTunnel,
		OnCopy: func(label, value string) {
			v.window.Clipboard().SetContent(value)
			v.toast(label)
		},
		OnPeerQR:   v.showPeerQR,
		OnNew:      v.showAddTunnelForm,
		OnImport:   v.importConfig,
		OnRelaunch: v.relaunchAsRoot,
	}
}

// showPeerQR opens the client config for one peer of a tunnel, identified by
// its public key — the only stable handle a peer has.
func (v *MainView) showPeerQR(tunnel, publicKey string) {
	v.withConfig(tunnel, func(cfg *config.Config) {
		for _, peer := range cfg.Peers {
			if peer.PublicKey.String() != publicKey {
				continue
			}
			ShowPeerQR(v.window, qrRequest{
				tunnel:    tunnel,
				serverCfg: cfg,
				peer:      peer,
				clientDir: v.settings.ClientConfigDir,
			}, v.lang)
			return
		}
	})
}

// Build assembles the console layout.
func (v *MainView) Build(meta fyne.AppMetadata) fyne.CanvasObject {
	v.meta = meta
	content := container.NewBorder(
		v.buildToolbar(meta),
		v.statusBar,
		nil, nil,
		v.buildBody(),
	)

	if !v.shortcutsSet {
		v.registerShortcuts()
		v.shortcutsSet = true
	}
	v.updateStatusChips()

	// The floating layer is where tooltips and toasts appear.
	return wgwidget.WithLayer(v.window, content)
}

// buildToolbar is the top strip: brand, filter, and the global actions. The
// one primary action (a new tunnel) is a filled button; everything else is an
// icon button with a tooltip naming it and its shortcut.
func (v *MainView) buildToolbar(meta fyne.AppMetadata) fyne.CanvasObject {
	p := wgtheme.Current()

	name := meta.Name
	if name == "" {
		name = v.lang.T(lang.Main_title)
	}
	mark := canvas.NewImageFromResource(wgtheme.IconBrand.For(p.Accent))
	mark.FillMode = canvas.ImageFillContain
	mark.SetMinSize(fyne.NewSize(24, 24))

	title := wgwidget.Title(name, wgtheme.TextSize(wgtheme.TextLG+1))
	title.FontSource = wgtheme.Font(wgtheme.WeightSemiBold)
	version := wgwidget.DimValue(meta.Version)
	version.TextSize = wgtheme.TextSize(wgtheme.TextXS)
	version.Color = p.TextDim

	brandBox := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD+1),
		container.NewCenter(mark), container.NewCenter(title), container.NewCenter(version))

	v.filterEntry.SetPlaceHolder(v.lang.T(lang.Main_filter_placeholder))
	v.filterEntry.OnChanged = func(string) { v.applyFilter() }
	v.filterEntry.ActionItem = widget.NewIcon(wgtheme.IconSearch)
	filterBox := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(260, v.filterEntry.MinSize().Height)),
		v.filterEntry)

	newBtn := wgwidget.NewButton(v.lang.T(lang.Main_add_tunnel), wgtheme.IconAdd, wgwidget.Primary, v.showAddTunnelForm)
	newBtn.Tooltip = "Ctrl+N"
	importBtn := wgwidget.NewIconButton(wgtheme.IconImport, v.lang.T(lang.Main_tip_import), v.importConfig)
	backupsBtn := wgwidget.NewIconButton(wgtheme.IconArchive, v.lang.T(lang.Main_tip_backups), v.showBackupsDialog)
	refreshBtn := wgwidget.NewIconButton(wgtheme.IconRefresh, v.lang.T(lang.Main_tip_refresh), v.Refresh)
	paletteBtn := wgwidget.NewIconButton(wgtheme.IconCommand, v.lang.T(lang.Main_tip_palette), v.showCommandPalette)
	settingsBtn := wgwidget.NewIconButton(wgtheme.IconSettings, v.lang.T(lang.Main_tip_settings), v.showSettings)

	v.fullscreenBtn = wgwidget.NewIconButton(wgtheme.IconFullscreen, "", v.toggleFullscreen)
	v.syncFullscreenButton()

	v.liveChip = wgwidget.NewChipButton(nil, "", p.TextDim, v.showRefreshPopover)
	v.liveChip.Tooltip = v.lang.T(lang.Main_tip_live)
	v.syncLiveChip()
	if v.settings.AutoRefreshEnabled {
		v.startAutoRefresh()
	}

	sep := func() fyne.CanvasObject {
		return container.New(layout.NewCustomPaddedLayout(8, 8, 4, 4), wgwidget.VHairline())
	}

	right := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		filterBox,
		sep(),
		newBtn, importBtn, backupsBtn,
		sep(),
		container.NewCenter(v.liveChip),
		refreshBtn, paletteBtn, v.fullscreenBtn, settingsBtn,
	)

	bar := container.NewBorder(nil, nil, brandBox, right)

	return container.NewVBox(
		container.NewStack(canvas.NewRectangle(p.Header),
			container.New(layout.NewCustomPaddedLayout(
				wgtheme.SpaceMD+2, wgtheme.SpaceMD+2, wgtheme.SpaceXL, wgtheme.SpaceLG), bar)),
		wgwidget.Hairline(),
	)
}

// syncLiveChip shows whether auto refresh is running and at what interval.
func (v *MainView) syncLiveChip() {
	if v.liveChip == nil {
		return
	}
	p := wgtheme.Current()
	if v.settings.AutoRefreshEnabled {
		v.liveChip.Set(wgwidget.NewPulseDot(p.Success, true),
			v.lang.Tf(lang.Main_live, v.settings.AutoRefreshSecs), p.Success)
		return
	}
	v.liveChip.Set(wgwidget.NewPulseDot(p.TextDim, false), v.lang.T(lang.Main_paused), p.TextSecondary)
}

// refreshIntervals are the choices offered by the live chip, in seconds;
// zero means off.
var refreshIntervals = []int{0, 2, 5, 10, 30}

// showRefreshPopover lets the auto refresh interval be changed in place,
// without a trip to Settings.
func (v *MainView) showRefreshPopover() {
	p := wgtheme.Current()
	labels := make([]string, len(refreshIntervals))
	selected := 0
	for i, secs := range refreshIntervals {
		if secs == 0 {
			labels[i] = v.lang.T(lang.Main_refresh_off)
		} else {
			labels[i] = fmt.Sprintf("%ds", secs)
		}
		if v.settings.AutoRefreshEnabled && secs == v.settings.AutoRefreshSecs {
			selected = i
		}
	}

	var pop *wgwidget.Surface
	seg := wgwidget.NewSegmented(labels, func(i int) {
		secs := refreshIntervals[i]
		v.settings.AutoRefreshEnabled = secs > 0
		if secs > 0 {
			v.settings.AutoRefreshSecs = secs
		}
		v.settings.Save(fyne.CurrentApp().Preferences())
		v.stopAutoRefresh()
		if secs > 0 {
			v.startAutoRefresh()
		}
		v.syncLiveChip()
		pop.Close()
	})
	seg.SetSelected(selected)

	head := wgwidget.Title(v.lang.T(lang.Main_refresh_title), wgtheme.TextSize(wgtheme.TextSM))
	head.FontSource = wgtheme.Font(wgtheme.WeightSemiBold)
	head.Color = p.TextSecondary
	pop = wgwidget.ShowPopover(v.liveChip, container.New(layout.NewCustomPaddedLayout(6, 6, 8, 8),
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD), head, seg)), 0)
}

// sidebarEntry is one sidebar item: a group heading or a tunnel.
type sidebarEntry struct {
	header string
	count  int
	iface  config.Interface
}

const (
	sidebarHeaderH = 30
	sidebarRowH    = 50
)

// sidebarCell renders either kind of sidebar entry, so widget.List can reuse
// one template for both.
type sidebarCell struct {
	widget.BaseWidget
	row  *wgwidget.TunnelRow
	head *fyne.Container
}

func newSidebarCell() *sidebarCell {
	c := &sidebarCell{row: wgwidget.NewTunnelRow(), head: container.NewStack()}
	c.ExtendBaseWidget(c)
	return c
}

func (c *sidebarCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(c.row, c.head))
}

// buildBody is the sidebar/detail split.
func (v *MainView) buildBody() fyne.CanvasObject {
	p := wgtheme.Current()

	v.sidebar = widget.NewList(
		func() int { return len(v.entries) },
		func() fyne.CanvasObject { return newSidebarCell() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			cell, ok := obj.(*sidebarCell)
			if !ok || id >= len(v.entries) {
				return
			}
			e := v.entries[id]
			if e.header != "" {
				cell.row.Hide()
				cell.head.Objects = []fyne.CanvasObject{wgwidget.SidebarHeader(e.header, e.count)}
				cell.head.Show()
				cell.head.Refresh()
				return
			}
			cell.head.Hide()
			cell.row.Show()
			iface := e.iface
			dev, _ := v.snapshot.Device(iface.Name)
			live := v.snapshot.Available && iface.Active
			cell.row.OnTap = func() { v.selectTunnel(iface.Name) }
			cell.row.Set(iface.Name, iface.IP,
				summaryLine(iface, v.configs[iface.Name], dev, live),
				iface.Active, iface.Name == v.selected)
		},
	)
	v.sidebar.HideSeparators = true

	v.sidebarFooter = container.NewStack()
	// The sidebar shares the window background and is set apart by a hairline.
	// On a tone of its own, the split divider's gap showed as a dark band.
	sidebarPane := container.NewStack(canvas.NewRectangle(p.Background),
		container.NewBorder(nil,
			container.NewVBox(wgwidget.Hairline(), v.sidebarFooter),
			nil, nil,
			container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceSM, 0, 0, 0), v.sidebar),
		),
	)

	v.detailBox = container.NewStack()

	v.split = wgwidget.NewSideSplit(sidebarPane, v.detailBox, 270)
	return v.split
}

// updateSidebarFooter shows the tunnel count and how many are up.
func (v *MainView) updateSidebarFooter() {
	if v.sidebarFooter == nil {
		return
	}
	up := 0
	for _, i := range v.interfaces {
		if i.Active {
			up++
		}
	}
	p := wgtheme.Current()
	text := wgwidget.DimValue(v.lang.Tf(lang.Main_tunnel_count, len(v.interfaces), up))
	text.TextStyle = fyne.TextStyle{}
	text.FontSource = wgtheme.Font(wgtheme.WeightMedium)
	text.TextSize = wgtheme.TextSize(wgtheme.TextSM)
	text.Color = p.TextDim
	v.sidebarFooter.Objects = []fyne.CanvasObject{
		container.New(layout.NewCustomPaddedLayout(10, 10, 18, 14), text),
	}
	v.sidebarFooter.Refresh()
}

// Refresh reloads interfaces and live state.
func (v *MainView) Refresh() {
	go func() {
		interfaces, err := ListInterfaces(v.ctrl)

		names := make([]string, 0, len(interfaces))
		for _, i := range interfaces {
			names = append(names, i.Name)
		}
		snap := ReadSnapshot(names...)
		v.tracker.Observe(snap)

		// Parse each config once per refresh, off the UI thread.
		configs := make(map[string]*config.Config, len(interfaces))
		for _, iface := range interfaces {
			if cfg, cfgErr := wgcfg.Load(v.ctrl.GetConfigPath(iface.Name)); cfgErr == nil {
				configs[iface.Name] = cfg
			}
		}

		fyne.Do(func() {
			if err != nil {
				v.statusBar.SetStatus(fmt.Sprintf("%v", err), false)
				return
			}

			v.interfaces = interfaces
			v.configs = configs
			v.snapshot = snap
			v.lastRefresh = time.Now()

			v.applyFilter()
			v.updateStatusChips()
		})
	}()
}

// applyFilter recomputes the visible list, keeps the selection valid and
// repaints both panes. Selection is preserved across refreshes so an
// auto-refresh tick never yanks the detail pane out from under the user.
func (v *MainView) applyFilter() {
	filter := strings.TrimSpace(strings.ToLower(v.filterEntry.Text))

	v.filtered = v.filtered[:0]
	for _, iface := range v.interfaces {
		if filter == "" || strings.Contains(strings.ToLower(iface.Name), filter) {
			v.filtered = append(v.filtered, iface)
		}
	}

	// Keep the current selection if it survived the filter; otherwise fall back
	// to the first visible tunnel so the detail pane is rarely empty.
	if v.selected != "" && !v.isVisible(v.selected) {
		v.selected = ""
	}
	if v.selected == "" && len(v.filtered) > 0 {
		v.selected = v.filtered[0].Name
	}

	v.buildEntries()
	v.sidebar.Refresh()
	v.syncSidebarSelection()
	v.updateSidebarFooter()
	v.renderDetail()
}

// buildEntries groups the filtered tunnels under Active and Inactive headings,
// and orders filtered the same way so arrow-key navigation follows what is on
// screen.
func (v *MainView) buildEntries() {
	var up, down []config.Interface
	for _, i := range v.filtered {
		if i.Active {
			up = append(up, i)
		} else {
			down = append(down, i)
		}
	}
	v.filtered = append(append(v.filtered[:0:0], up...), down...)

	v.entries = v.entries[:0]
	add := func(title string, group []config.Interface) {
		if len(group) == 0 {
			return
		}
		v.entries = append(v.entries, sidebarEntry{header: title, count: len(group)})
		for _, i := range group {
			v.entries = append(v.entries, sidebarEntry{iface: i})
		}
	}
	add(v.lang.T(lang.Main_section_active), up)
	add(v.lang.T(lang.Main_section_inactive), down)

	for i, e := range v.entries {
		if e.header != "" {
			v.sidebar.SetItemHeight(i, sidebarHeaderH)
		} else {
			v.sidebar.SetItemHeight(i, sidebarRowH)
		}
	}
}

func (v *MainView) isVisible(name string) bool {
	for _, i := range v.filtered {
		if i.Name == name {
			return true
		}
	}
	return false
}

// syncSidebarSelection keeps the selected tunnel scrolled into view. Rows
// draw their own selected state, so the list's selection is never used.
func (v *MainView) syncSidebarSelection() {
	for i, e := range v.entries {
		if e.header == "" && e.iface.Name == v.selected {
			v.sidebar.ScrollTo(i)
			return
		}
	}
}

// selectTunnel switches the detail pane to a tunnel.
func (v *MainView) selectTunnel(name string) {
	if v.selected == name {
		return
	}
	v.selected = name
	v.sidebar.Refresh()
	v.renderDetail()
}

// renderDetail repaints the right-hand pane for the current selection.
//
// The pane is rebuilt from scratch, which used to reset its scroll position on
// every refresh tick and would drop hover states on its widgets. So a rebuild
// is skipped when nothing it shows has changed, and when it does happen for
// the same tunnel the scroll offset is carried over.
func (v *MainView) renderDetail() {
	var content fyne.CanvasObject
	var scroll *container.Scroll
	sig := ""

	switch {
	case len(v.interfaces) == 0:
		content = v.detail.emptyState(lang.Main_no_tunnels_title, lang.Main_no_tunnels_body, true)
		sig = "empty"
	case len(v.filtered) == 0:
		content = v.detail.emptyState(lang.Main_no_match_title, lang.Main_no_match_body, false)
		sig = "nomatch"
	case v.selected == "":
		content = v.detail.emptyState(lang.Main_empty_title, lang.Main_empty_body, false)
		sig = "none"
	default:
		iface, ok := v.findInterface(v.selected)
		if !ok {
			content = v.detail.emptyState(lang.Main_empty_title, lang.Main_empty_body, false)
			sig = "none"
			break
		}
		cfg := v.configs[iface.Name]
		rate, _ := v.tracker.Rate(iface.Name)
		history := v.tracker.History(iface.Name)
		dev, hasDev := v.snapshot.Device(iface.Name)
		live := v.snapshot.Available && hasDev && iface.Active

		cfgText := ""
		if cfg != nil {
			cfgText = config.GenerateConfigString(cfg)
		}
		sig = fmt.Sprintf("%s|%v|%s|%s|%v|%s", iface.Name, iface.Active, iface.IP, iface.PublicKey,
			v.snapshot.Available, cfgText) + fmt.Sprintf("|%v", v.snapshot.Taken.IsZero())
		if live {
			// Live data changes every tick (counters, handshake ages), so a
			// live tunnel always redraws.
			sig += fmt.Sprintf("|%v|%v|%d|%d", time.Now().UnixNano(), dev, len(history), int(rate.RX+rate.TX))
		}
		if sig == v.detailSig && v.detailName == iface.Name {
			return
		}

		scroll = v.detail.build(*iface, cfg, v.snapshot, rate, history, v.ctrl.GetConfigPath(iface.Name))
		content = scroll
	}

	if sig == v.detailSig && sig != "" && scroll == nil {
		return
	}

	var offset fyne.Position
	if scroll != nil && v.detailScroll != nil && v.detailName == v.selected {
		offset = v.detailScroll.Offset
	}

	v.detailBox.Objects = []fyne.CanvasObject{content}
	v.detailBox.Refresh()

	v.detailSig = sig
	v.detailScroll = scroll
	v.detailName = v.selected
	if scroll != nil && offset.Y > 0 {
		scroll.Offset = offset
		scroll.Refresh()
	}
}

// updateStatusChips refreshes the environment chips on the status strip.
func (v *MainView) updateStatusChips() {
	p := wgtheme.Current()

	chips := []fyne.CanvasObject{
		wgwidget.Tag(v.settings.WGConfigPath, p.TextSecondary, wgtheme.IconFolder),
		wgwidget.Tag(v.lang.T(lang.Main_chip_backend), p.TextSecondary, nil),
	}
	if IsRoot() {
		chips = append(chips, wgwidget.Tag(v.lang.T(lang.Main_chip_root), p.Success, wgtheme.IconLock))
	} else {
		chips = append(chips, wgwidget.Tag(v.lang.T(lang.Main_chip_no_root), p.Warning, wgtheme.IconAlert))
	}
	v.statusBar.SetChips(chips...)
}

// toast confirms a completed action without taking up the status strip,
// which is kept for state that stays true (errors, warnings).
func (v *MainView) toast(msg string) {
	wgwidget.ShowToast(v.window, wgwidget.ToastSuccess, msg)
}

// relaunchAsRoot restarts the app with root privileges so live telemetry can
// be read, preferring pkexec's graphical prompt over the in-app sudo dialog.
func (v *MainView) relaunchAsRoot() {
	if settings.PkexecAvailable() {
		if err := settings.RelaunchWithPkexec(); err != nil {
			helpers.ShowError(err, v.window)
			return
		}
		fyne.CurrentApp().Quit()
		return
	}
	settings.ShowSudoRelaunchDialog(v.window)
}

// importConfig opens a .conf and copies it into the config directory.
func (v *MainView) importConfig() {
	helpers.OpenFile(v.window, helpers.FileOpts{
		Title:       v.lang.T(lang.Main_import),
		FilterLabel: "WireGuard",
		Extensions:  []string{".conf"},
	}, func(filePath string) {
		if filepath.Ext(filePath) != ".conf" {
			helpers.ShowError(errors.New(v.lang.T(lang.Main_import_error_ext)), v.window)
			return
		}
		cfg, err := wgcfg.Load(filePath)
		if err != nil {
			helpers.ShowError(errors.New(v.lang.Tf(lang.Main_import_error_invalid, err.Error())), v.window)
			return
		}
		if err = cfg.Validate(); err != nil {
			helpers.ShowInformation(v.lang.T(lang.Main_import_validation_title),
				v.lang.T(lang.Main_import_validation_msg), v.window)
		}

		// The file name is the interface name — that is how wg-quick reads a
		// config directory. The config's own Name field is a display string
		// ("[wg0] Public key: …") and using it produced files named after it.
		base := filepath.Base(filePath)
		name := strings.TrimSuffix(base, ".conf")

		if _, statErr := os.Stat(filepath.Join(v.settings.WGConfigPath, base)); statErr == nil {
			helpers.ShowConfirm(v.lang.T(lang.Main_import_overwrite_title),
				v.lang.T(lang.Main_import_overwrite_msg), func(b bool) {
					if b {
						v.save(name, cfg)
					} else {
						helpers.ShowInformation(v.lang.T(lang.Common_aborted),
							v.lang.T(lang.Main_import_aborted_msg), v.window)
					}
				}, v.window)
			return
		}
		v.save(name, cfg)
	})
}

func (v *MainView) save(name string, cfg *config.Config) {
	name = strings.TrimSuffix(name, ".conf")
	if err := config.WriteConfig(v.settings.WGConfigPath, name, cfg); err != nil {
		helpers.ShowError(errors.New(v.lang.Tf(lang.Main_error_saving, err.Error())), v.window)
		return
	}
	v.toast(v.lang.Tf(lang.Main_import_success_msg, name))
	v.Refresh()
}

func (v *MainView) toggleInterface(name string, activate bool) {
	actionKey := lang.Status_deactivating
	if activate {
		actionKey = lang.Status_activating
	}

	v.busyDialog.Show(name, v.lang.T(actionKey))

	go func() {
		err := v.ctrl.ToggleInterface(name, activate)

		fyne.Do(func() {
			v.busyDialog.Hide()

			// Counters restart with the interface; drop the history so the
			// next reading is not rendered as one enormous burst.
			v.tracker.Reset(name)

			switch {
			case err != nil && strings.Contains(err.Error(), "invalid endpoint"):
				v.statusBar.SetWarning(v.lang.Tf(lang.Status_endpoint_unreachable,
					name, v.lastRefresh.Format(time.Kitchen)))
			case err != nil:
				v.statusBar.SetStatus(fmt.Sprintf("%v", err), false)
			default:
				statusKey := lang.Main_deactivated_status
				if activate {
					statusKey = lang.Main_activated_status
				}
				v.statusBar.Reset()
				v.toast(v.lang.Tf(statusKey, name))
			}

			v.Refresh()
		})
	}()
}

func (v *MainView) showScan(name, ip string) {
	var plan ScanPlan
	if cfg, err := wgcfg.Load(v.ctrl.GetConfigPath(name)); err == nil {
		plan = PlanScan(cfg, ip)
	}
	NewScanView(name, ip, plan, v.settings.ScanWorkers, v.settings.ScanTimeoutSecs, v.lang).Show(v.window)
}

func (v *MainView) showAddTunnelForm() {
	form := NewTunnelForm(v.window, v.ctrl, "", nil, func(name string, cfg *config.Config) error {
		err := v.ctrl.WriteConfig(name, *cfg)
		if err == nil {
			v.selected = name
			v.Refresh()
		}
		return err
	}, nil, v.settings.ClientConfigDir, v.lang)
	form.Show()
}

func (v *MainView) showEditTunnelForm(name string) {
	v.withConfig(name, func(cfg *config.Config) {
		v.confirmIfActive(name, func() { v.openForm(name, cfg, true) })
	})
}

func (v *MainView) showEditPeersTunnelForm(name string) {
	v.withConfig(name, func(cfg *config.Config) {
		v.confirmIfActive(name, func() { v.openForm(name, cfg, false) })
	})
}

// withConfig loads a tunnel's config and reports failures once, in one place.
func (v *MainView) withConfig(name string, fn func(*config.Config)) {
	cfg, err := wgcfg.Load(v.ctrl.GetConfigPath(name))
	if err != nil {
		helpers.ShowError(fmt.Errorf(v.lang.T(lang.Main_failed_to_load_config), err), v.window)
		return
	}
	fn(cfg)
}

// confirmIfActive warns before editing a running tunnel, then runs proceed.
//
// The follow-up action is a parameter rather than hardcoded: the previous
// version always reopened the *interface* form, so choosing "Peers" on an
// active tunnel and confirming silently opened the wrong editor.
func (v *MainView) confirmIfActive(name string, proceed func()) {
	iface, ok := v.findInterface(name)
	if !ok || !iface.Active {
		proceed()
		return
	}

	helpers.ShowConfirm(v.lang.T(lang.Main_tunnel_active_title),
		v.lang.Tf(lang.Main_tunnel_active_msg, name),
		func(yes bool) {
			if yes {
				proceed()
			}
		}, v.window)
}

func (v *MainView) openForm(name string, cfg *config.Config, isInterface bool) {
	form := NewTunnelForm(v.window, v.ctrl, name, cfg, func(_ string, newConfig *config.Config) error {
		err := v.ctrl.WriteConfig(name, *newConfig)
		if err == nil {
			v.Refresh()
		}
		return err
	}, nil, v.settings.ClientConfigDir, v.lang)

	if isInterface {
		form.Show()
	} else {
		form.ShowPeers()
	}
}

func (v *MainView) showConfigView(name string) {
	NewConfigView(v.window, v.ctrl, name, v.lang, func() {
		v.Refresh()
	}).Show()
}

func (v *MainView) confirmDeleteTunnel(name string) {
	iface, _ := v.findInterface(name)

	deleteFn := func() {
		v.busyDialog.Show(v.lang.T(lang.Main_delete_tunnel_title), v.lang.Tf(lang.Main_deleting, name))

		go func() {
			if iface != nil && iface.Active {
				_ = v.ctrl.ToggleInterface(name, false)
			}
			err := v.ctrl.DeleteInterface(name, true)

			fyne.Do(func() {
				v.busyDialog.Hide()
				v.tracker.Reset(name)

				if err != nil {
					v.statusBar.SetStatus(fmt.Sprintf("%v", err), false)
				} else {
					v.toast(v.lang.Tf(lang.Main_deleted_status, name))
					if v.selected == name {
						v.selected = ""
					}
				}
				v.Refresh()
			})
		}()
	}

	if !v.settings.ConfirmBeforeDelete {
		deleteFn()
		return
	}

	msg := v.lang.Tf(lang.Main_delete_tunnel_msg, name, v.ctrl.GetConfigPath(name))
	if iface != nil && iface.Active {
		msg += v.lang.T(lang.Main_delete_tunnel_active_warn)
	}

	helpers.ShowConfirm(v.lang.T(lang.Main_delete_tunnel_title), msg, func(yes bool) {
		if yes {
			deleteFn()
		}
	}, v.window)
}

func (v *MainView) showBackupsDialog() {
	NewBackupView(v.window, v.ctrl, v.Refresh, v.lang).Show()
}

func (v *MainView) showSettings() {
	NewSettingsView(v.window, v.settings, v.applySettings, v.lang).Show()
}

func (v *MainView) findInterface(name string) (*config.Interface, bool) {
	for i := range v.interfaces {
		if v.interfaces[i].Name == name {
			return &v.interfaces[i], true
		}
	}
	return nil, false
}

// startAutoRefresh runs the periodic refresh. Any previous loop is stopped
// first so changing the interval cannot leave two tickers running.
func (v *MainView) startAutoRefresh() {
	v.stopAutoRefresh()

	v.autoMu.Lock()
	stop := make(chan struct{})
	v.autoStop = stop
	interval := time.Duration(v.settings.AutoRefreshSecs) * time.Second
	v.autoMu.Unlock()

	if interval <= 0 {
		interval = time.Duration(settings.DefaultAutoRefreshSecs) * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				v.Refresh()
			case <-stop:
				return
			}
		}
	}()
}

func (v *MainView) stopAutoRefresh() {
	v.autoMu.Lock()
	defer v.autoMu.Unlock()

	if v.autoStop != nil {
		close(v.autoStop)
		v.autoStop = nil
	}
}

// applySettings adopts changed settings and repaints everything that depends
// on them.
func (v *MainView) applySettings(updated *settings.AppSettings) {
	oldPath := v.settings.WGConfigPath
	v.settings = updated

	if updated.WGConfigPath != oldPath {
		newCtrl := wg.New(updated.WGConfigPath)
		v.ctrl = &newCtrl
	}

	// A theme change touches every widget, so rebuild the whole window rather
	// than patching colours in place.
	fyne.CurrentApp().Settings().SetTheme(wgtheme.NewWGAdminTheme(v.settings))

	v.stopAutoRefresh()
	v.detailSig = ""
	v.window.SetContent(v.Build(v.meta))
	v.Refresh()
}

// toggleFullscreen switches the window in and out of fullscreen.
//
// Starting fullscreen is a setting, but leaving it used to depend on knowing
// F11 — with no window chrome and no button, anyone who did not would be stuck
// in the app. The toolbar button is the way out that needs no prior knowledge;
// F11 and Escape are there for people who expect them.
func (v *MainView) toggleFullscreen() {
	v.setFullscreen(!v.window.FullScreen())
}

// exitFullscreen leaves fullscreen and reports whether it did anything, so Esc
// can fall through to clearing the filter when there was nothing to exit.
func (v *MainView) exitFullscreen() bool {
	if !v.window.FullScreen() {
		return false
	}
	v.setFullscreen(false)
	return true
}

func (v *MainView) setFullscreen(on bool) {
	v.window.SetFullScreen(on)
	v.syncFullscreenButton()
}

// syncFullscreenButton keeps the icon and its tooltip describing what the
// button will do next rather than what the window currently is.
func (v *MainView) syncFullscreenButton() {
	if v.fullscreenBtn == nil {
		return
	}
	if v.window.FullScreen() {
		v.fullscreenBtn.Icon = wgtheme.IconFullscreenExit
		v.fullscreenBtn.Tooltip = v.lang.T(lang.Main_exit_fullscreen) + " (F11)"
		v.fullscreenBtn.Variant = wgwidget.Secondary
	} else {
		v.fullscreenBtn.Icon = wgtheme.IconFullscreen
		v.fullscreenBtn.Tooltip = v.lang.T(lang.Main_enter_fullscreen) + " (F11)"
		v.fullscreenBtn.Variant = wgwidget.Ghost
	}
	v.fullscreenBtn.Refresh()
}

// tunnelNames lists the loaded tunnels, used by the command palette.
func (v *MainView) tunnelNames() []string {
	names := make([]string, 0, len(v.interfaces))
	for _, i := range v.interfaces {
		names = append(names, i.Name)
	}
	return names
}

// selectedName reports the currently selected tunnel, if any.
func (v *MainView) selectedName() string { return v.selected }

// focusFilter puts the caret in the filter box.
func (v *MainView) focusFilter() {
	v.window.Canvas().Focus(v.filterEntry)
}

// clearFilter empties the filter, returning the full list.
func (v *MainView) clearFilter() {
	if v.filterEntry.Text == "" {
		return
	}
	v.filterEntry.SetText("")
}

// toggleSelected brings the selected tunnel up or down.
func (v *MainView) toggleSelected() {
	if v.selected == "" {
		return
	}
	if iface, ok := v.findInterface(v.selected); ok {
		v.toggleInterface(iface.Name, !iface.Active)
	}
}

// scanSelected opens the scanner for the selected tunnel.
func (v *MainView) scanSelected() {
	if iface, ok := v.findInterface(v.selected); ok && canScan(*iface) {
		v.showScan(iface.Name, iface.IP)
	}
}

// moveSelection steps the sidebar selection by delta.
func (v *MainView) moveSelection(delta int) {
	if len(v.filtered) == 0 {
		return
	}
	idx := 0
	for i, iface := range v.filtered {
		if iface.Name == v.selected {
			idx = i
			break
		}
	}
	idx += delta
	if idx < 0 {
		idx = 0
	}
	if idx >= len(v.filtered) {
		idx = len(v.filtered) - 1
	}
	v.selectTunnel(v.filtered[idx].Name)
	v.syncSidebarSelection()
}

// Select shows the named tunnel once the next refresh has loaded it. Used by
// tooling that drives the console (cmd/preview); the UI selects through the
// sidebar.
func (v *MainView) Select(name string) {
	v.selected = name
	v.renderDetail()
}

// ScrollDetail scrolls the detail pane, for tooling (cmd/preview). The offset
// is then carried across refreshes like a user's own scrolling.
func (v *MainView) ScrollDetail(y float32) {
	if v.detailScroll != nil {
		v.detailScroll.Offset = fyne.NewPos(0, y)
		v.detailScroll.Refresh()
	}
}

// Open shows one of the console's sub-screens for the selected tunnel, for
// tooling (cmd/preview) that needs to screenshot them without clicking.
func (v *MainView) Open(what string) {
	name := v.selected
	switch what {
	case "new":
		v.showAddTunnelForm()
	case "edit":
		v.showEditTunnelForm(name)
	case "peers":
		v.showEditPeersTunnelForm(name)
	case "config":
		v.showConfigView(name)
	case "backups":
		v.showBackupsDialog()
	case "settings":
		v.showSettings()
	case "scan":
		v.scanSelected()
	case "palette":
		v.showCommandPalette()
	case "shortcuts":
		v.showShortcuts()
	case "delete":
		v.confirmDeleteTunnel(name)
	case "busy":
		v.busyDialog.Show(name, v.lang.T(lang.Status_activating))
	case "qr":
		if cfg := v.configs[name]; cfg != nil && len(cfg.Peers) > 0 {
			v.showPeerQR(name, cfg.Peers[0].PublicKey.String())
		}
	}
}
