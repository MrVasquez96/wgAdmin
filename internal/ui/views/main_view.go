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

	statusBar     *wgwidget.StatusBar
	busyDialog    *wgwidget.BusyDialog
	filterEntry   *widget.Entry
	autoRefresh   *widget.Check
	fullscreenBtn *widget.Button

	sidebar     *widget.List
	sidebarHead *canvas.Text
	detailBox   *fyne.Container
	split       *container.Split
	brand       *canvas.Text
	headerBg    *canvas.Rectangle

	// interfaces is everything on disk; filtered is what the sidebar shows.
	interfaces []config.Interface
	filtered   []config.Interface
	selected   string

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
		autoRefresh: widget.NewCheck(language.Tf(lang.Main_auto_refresh, cfg.AutoRefreshSecs), nil),
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
			v.statusBar.Flash(label, true)
		},
		OnPeerQR: v.showPeerQR,
		OnNew:    v.showAddTunnelForm,
		OnImport: v.importConfig,
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
	content := container.NewBorder(
		v.buildToolbar(meta),
		v.statusBar,
		nil, nil,
		v.buildBody(),
	)

	v.registerShortcuts()
	v.updateStatusChips()

	// The floating layer is where tooltips and toasts appear.
	return wgwidget.WithLayer(v.window, content)
}

// buildToolbar is the top strip: brand, filter, and the global actions.
func (v *MainView) buildToolbar(meta fyne.AppMetadata) fyne.CanvasObject {
	p := wgtheme.Current()

	name := meta.Name
	if name == "" {
		name = v.lang.T(lang.Main_title)
	}
	v.brand = canvas.NewText(name, p.Accent)
	v.brand.TextStyle = fyne.TextStyle{Bold: true}
	v.brand.TextSize = 18

	version := canvas.NewText(meta.Version, p.TextDim)
	version.TextSize = 10
	version.TextStyle = fyne.TextStyle{Monospace: true}

	brandBox := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(18, 18)),
			widget.NewIcon(wgtheme.IconShield)),
		v.brand,
		container.NewVBox(layout.NewSpacer(), version, layout.NewSpacer()),
	)

	v.filterEntry.SetPlaceHolder(v.lang.T(lang.Main_filter_placeholder))
	v.filterEntry.OnChanged = func(string) { v.applyFilter() }
	v.filterEntry.ActionItem = widget.NewIcon(wgtheme.IconFilter)
	filterBox := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(210, v.filterEntry.MinSize().Height)),
		v.filterEntry)

	newBtn := widget.NewButtonWithIcon(v.lang.T(lang.Main_new_tunnel), wgtheme.IconAdd, v.showAddTunnelForm)
	newBtn.Importance = widget.HighImportance

	importBtn := widget.NewButtonWithIcon(v.lang.T(lang.Main_import), wgtheme.IconImport, v.importConfig)
	backupsBtn := widget.NewButtonWithIcon(v.lang.T(lang.Main_backups), wgtheme.IconFolder, v.showBackupsDialog)
	refreshBtn := widget.NewButtonWithIcon("", wgtheme.IconRefresh, v.Refresh)
	settingsBtn := widget.NewButtonWithIcon("", wgtheme.IconSettings, v.showSettings)

	v.fullscreenBtn = widget.NewButtonWithIcon("", wgtheme.IconFullscreen, v.toggleFullscreen)
	v.syncFullscreenButton()

	v.autoRefresh.OnChanged = func(on bool) {
		if on {
			v.startAutoRefresh()
		} else {
			v.stopAutoRefresh()
		}
	}
	if v.settings.AutoRefreshEnabled {
		v.autoRefresh.SetChecked(true)
	}

	right := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		filterBox,
		wgwidget.VHairline(),
		newBtn, importBtn, backupsBtn,
		wgwidget.VHairline(),
		container.NewCenter(v.autoRefresh),
		refreshBtn, v.fullscreenBtn, settingsBtn,
	)

	bar := container.NewBorder(nil, nil, brandBox, right)

	v.headerBg = canvas.NewRectangle(p.Header)
	rule := canvas.NewRectangle(p.Border)
	rule.SetMinSize(fyne.NewSize(0, 1))

	return container.NewVBox(
		container.NewStack(v.headerBg,
			container.New(layout.NewCustomPaddedLayout(
				wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceLG, wgtheme.SpaceLG), bar)),
		rule,
	)
}

// buildBody is the sidebar/detail split.
func (v *MainView) buildBody() fyne.CanvasObject {
	p := wgtheme.Current()

	v.sidebar = widget.NewList(
		func() int { return len(v.filtered) },
		func() fyne.CanvasObject { return wgwidget.NewTunnelRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row, ok := obj.(*wgwidget.TunnelRow)
			if !ok || id >= len(v.filtered) {
				return
			}
			iface := v.filtered[id]
			dev, _ := v.snapshot.Device(iface.Name)
			live := v.snapshot.Available && iface.Active

			row.Set(iface.Name, iface.IP,
				summaryLine(iface, v.configs[iface.Name], dev, live),
				iface.Active, iface.Name == v.selected)
		},
	)
	v.sidebar.OnSelected = func(id widget.ListItemID) {
		if id < len(v.filtered) {
			v.selectTunnel(v.filtered[id].Name)
		}
	}

	v.sidebarHead = wgwidget.Caption(v.lang.T(lang.Main_sidebar_heading))

	sidebarBg := canvas.NewRectangle(p.SurfaceRaised)
	sidebarPane := container.NewStack(sidebarBg,
		container.NewBorder(
			container.New(layout.NewCustomPaddedLayout(
				wgtheme.SpaceMD, wgtheme.SpaceSM, wgtheme.SpaceLG, wgtheme.SpaceMD), v.sidebarHead),
			nil, nil, nil,
			v.sidebar,
		),
	)

	v.detailBox = container.NewStack()

	v.split = container.NewHSplit(sidebarPane, v.detailBox)
	v.split.SetOffset(0.24)
	return v.split
}

// Refresh reloads interfaces and live state.
func (v *MainView) Refresh() {
	go func() {
		interfaces, err := v.ctrl.ListInterfaces()

		names := make([]string, 0, len(interfaces))
		for _, i := range interfaces {
			names = append(names, i.Name)
		}
		snap := wgstat.Read(names...)
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

	v.sidebar.Refresh()
	v.syncSidebarSelection()
	v.renderDetail()
}

func (v *MainView) isVisible(name string) bool {
	for _, i := range v.filtered {
		if i.Name == name {
			return true
		}
	}
	return false
}

// syncSidebarSelection mirrors v.selected onto the list widget without
// re-triggering OnSelected.
func (v *MainView) syncSidebarSelection() {
	for i, iface := range v.filtered {
		if iface.Name == v.selected {
			v.sidebar.Select(i)
			return
		}
	}
	v.sidebar.UnselectAll()
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
func (v *MainView) renderDetail() {
	var content fyne.CanvasObject

	switch {
	case len(v.interfaces) == 0:
		content = v.detail.emptyState(lang.Main_no_tunnels_title, lang.Main_no_tunnels_body, true)
	case len(v.filtered) == 0:
		content = v.detail.emptyState(lang.Main_no_match_title, lang.Main_no_match_body, false)
	case v.selected == "":
		content = v.detail.emptyState(lang.Main_empty_title, lang.Main_empty_body, false)
	default:
		iface, ok := v.findInterface(v.selected)
		if !ok {
			content = v.detail.emptyState(lang.Main_empty_title, lang.Main_empty_body, false)
			break
		}
		rate, _ := v.tracker.Rate(iface.Name)
		content = v.detail.build(
			*iface,
			v.configs[iface.Name],
			v.snapshot,
			rate,
			v.tracker.History(iface.Name),
			v.ctrl.GetConfigPath(iface.Name),
		)
	}

	v.detailBox.Objects = []fyne.CanvasObject{content}
	v.detailBox.Refresh()
}

// updateStatusChips refreshes the environment chips on the status strip.
func (v *MainView) updateStatusChips() {
	p := wgtheme.Current()

	chips := []fyne.CanvasObject{
		wgwidget.NeutralChip(v.settings.WGConfigPath),
		wgwidget.NeutralChip(v.lang.T(lang.Main_chip_backend)),
	}

	if os.Geteuid() == 0 {
		chips = append(chips, wgwidget.Chip(v.lang.T(lang.Main_chip_root), p.Success, true))
	} else {
		chips = append(chips, wgwidget.Chip(v.lang.T(lang.Main_chip_no_root), p.Warning, true))
	}

	v.statusBar.SetChips(chips...)
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
	v.statusBar.Flash(v.lang.Tf(lang.Main_import_success_msg, name), true)
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
				v.statusBar.Flash(v.lang.Tf(statusKey, name), true)
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
	NewScanView(name, ip, plan, v.settings.ScanWorkers, v.settings.ScanTimeoutSecs, v.lang).Show()
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
					v.statusBar.Flash(v.lang.Tf(lang.Main_deleted_status, name), true)
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

	// Rebuild the theme before repainting so new colours are picked up by every
	// widget that reads the palette at construction time.
	fyne.CurrentApp().Settings().SetTheme(wgtheme.NewWGAdminTheme(v.settings))
	p := wgtheme.Current()

	v.autoRefresh.Text = v.lang.Tf(lang.Main_auto_refresh, updated.AutoRefreshSecs)
	v.autoRefresh.Refresh()

	v.brand.Color = p.Accent
	v.brand.Refresh()
	v.headerBg.FillColor = p.Header
	v.headerBg.Refresh()

	v.stopAutoRefresh()
	if v.autoRefresh.Checked {
		v.startAutoRefresh()
	}

	v.updateStatusChips()
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
		v.fullscreenBtn.SetIcon(wgtheme.IconFullscreenExit)
		v.fullscreenBtn.Importance = widget.HighImportance
	} else {
		v.fullscreenBtn.SetIcon(wgtheme.IconFullscreen)
		v.fullscreenBtn.Importance = widget.MediumImportance
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
	if iface, ok := v.findInterface(v.selected); ok && iface.IP != "unknown" {
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
	v.sidebar.ScrollTo(idx)
	v.syncSidebarSelection()
}
