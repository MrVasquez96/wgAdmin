package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"
)

// MainView is the main application view
type MainView struct {
	window        fyne.Window
	ctrl          *wg.WG
	settings      *settings.AppSettings
	lang          *lang.Language
	listContainer *fyne.Container
	statusBar     *wgwidget.StatusBar
	busyDialog    *wgwidget.BusyDialog
	filterEntry   *widget.Entry
	autoRefresh   *widget.Check
	hint          *widget.RichText
	headerTitle   *canvas.Text
	headerBg      *canvas.Rectangle

	interfaces  []config.Interface
	stopAuto    chan struct{}
	lastRefresh time.Time
}

// NewMainView creates a new main view
func NewMainView(window fyne.Window, ctrl *wg.WG, cfg *settings.AppSettings, language *lang.Language) *MainView {
	return &MainView{
		window:        window,
		ctrl:          ctrl,
		settings:      cfg,
		lang:          language,
		listContainer: container.NewVBox(),
		statusBar:     wgwidget.NewStatusBar(language),
		busyDialog:    wgwidget.NewBusyDialog(window),
		filterEntry:   widget.NewEntry(),
		autoRefresh:   widget.NewCheck(language.Tf(lang.Main_auto_refresh, cfg.AutoRefreshSecs), nil),
		stopAuto:      make(chan struct{}),
	}
}

// Build creates the main view content
func (v *MainView) Build(meta fyne.AppMetadata) fyne.CanvasObject {
	// Title — use the custom theme so forced variant is respected
	customT := wgtheme.NewWGAdminTheme(v.settings)
	variant := wgtheme.CurrentVariant()
	v.headerTitle = canvas.NewText(meta.Name, customT.Color(theme.ColorNameForeground, variant))
	v.headerTitle.TextStyle = fyne.TextStyle{Bold: true}
	v.headerTitle.TextSize = 28

	// Import tunnel
	importBtn := v.newImportButton()
	// Add tunnel button
	addBtn := widget.NewButtonWithIcon(v.lang.T(lang.Main_add_tunnel), theme.ContentAddIcon(), func() {
		v.showAddTunnelForm()
	})

	addBtn.Importance = widget.HighImportance

	// Backups button
	backupsBtn := widget.NewButtonWithIcon(v.lang.T(lang.Main_backups), theme.FolderOpenIcon(), func() {
		v.showBackupsDialog()
	})

	// Filter
	v.filterEntry.SetPlaceHolder(v.lang.T(lang.Main_filter_placeholder))
	v.filterEntry.OnChanged = func(s string) {
		v.rebuild()
	}
	filterContainer := container.NewGridWrap(fyne.NewSize(200, v.filterEntry.MinSize().Height), v.filterEntry)

	// Refresh button
	refreshBtn := widget.NewButtonWithIcon(v.lang.T(lang.Common_refresh), theme.ViewRefreshIcon(), func() {
		v.Refresh()
	})

	// Auto refresh
	v.autoRefresh.OnChanged = func(on bool) {
		if on {
			go v.startAutoRefresh()
		} else {
			v.stopAutoRefresh()
		}
	}

	// Settings button
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		sv := NewSettingsView(v.window, v.settings, func(updated *settings.AppSettings) {
			v.applySettings(updated)
		}, v.lang)
		sv.Show()
	})

	// Auto-refresh on startup if configured
	if v.settings.AutoRefreshEnabled {
		v.autoRefresh.SetChecked(true)
	}

	// Header layout with background
	leftHeader := container.NewHBox(v.headerTitle)
	rightHeader := container.NewHBox(filterContainer, importBtn, addBtn, backupsBtn, v.autoRefresh, refreshBtn, settingsBtn)
	headerContent := container.NewBorder(nil, nil, leftHeader, rightHeader)
	v.headerBg = canvas.NewRectangle(customT.Color(theme.ColorNameHeaderBackground, variant))
	header := container.NewVBox(
		container.NewStack(v.headerBg, container.NewPadded(headerContent)),
		widget.NewSeparator(),
	)

	// Scrollable list
	scroll := container.NewVScroll(v.listContainer)
	scroll.SetMinSize(fyne.NewSize(860, 480))

	// Footer hint
	v.hint = widget.NewRichTextFromMarkdown(v.lang.Tf(lang.Main_footer_hint, v.settings.WGConfigPath))
	v.hint.Wrapping = fyne.TextWrapWord

	// Main content
	content := container.NewBorder(
		header,
		container.NewVBox(v.hint, v.statusBar),
		nil, nil,
		scroll,
	)

	return container.NewPadded(content)
}
func (v *MainView) newImportButton() *widget.Button {

	return widget.NewButtonWithIcon(v.lang.T(lang.Main_import_from_file), theme.DocumentSaveIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				helpers.ShowError(err, v.window)
				return
			}
			if reader == nil {
				return
			}
			// 2. Handle the file content
			defer reader.Close()
			// data, _ := io.ReadAll(reader)
			filePath := reader.URI().Path()
			if filepath.Ext(filePath) != ".conf" {
				helpers.ShowError(errors.New(v.lang.T(lang.Main_import_error_ext)), v.window)
				return
			}
			cfg, err := config.ParseConfig(filePath)
			if err != nil {
				helpers.ShowError(errors.New(v.lang.Tf(lang.Main_import_error_invalid, err.Error())), v.window)
				return
			}
			if err = cfg.Validate(); err != nil {
				helpers.ShowInformation(v.lang.T(lang.Main_import_validation_title), v.lang.T(lang.Main_import_validation_msg), v.window)
			}
			name := cfg.Name
			if name == "Unknown" {
				name = reader.URI().Name()
				name = filepath.Base(name)
			}
			_, err = os.Stat(filepath.Join(v.settings.WGConfigPath, reader.URI().Name()))

			if err == nil {
				helpers.ShowConfirm(v.lang.T(lang.Main_import_overwrite_title), v.lang.T(lang.Main_import_overwrite_msg), func(b bool) {
					if b {
						v.save(name, cfg)
					} else {
						helpers.ShowInformation(v.lang.T(lang.Common_aborted), v.lang.T(lang.Main_import_aborted_msg), v.window)
					}
				}, v.window)
				return
			}
			v.save(name, cfg)

		}, v.window)
		fileDialog.Resize(fyne.NewSize(800, 600))
		fileDialog.Show()
	})
}
func (v *MainView) save(name string, cfg *config.Config) {
	if strings.Contains(name, ".conf") {
		name = strings.ReplaceAll(name, ".conf", "")
	}
	err := config.WriteConfig(v.settings.WGConfigPath, name, cfg)
	if err != nil {
		helpers.ShowError(errors.New(v.lang.Tf(lang.Main_error_saving, err.Error())), v.window)
		return
	}
	helpers.ShowInformation(v.lang.T(lang.Main_import_success_title), v.lang.Tf(lang.Main_import_success_msg, name), v.window)
}

// Refresh reloads the interface list
func (v *MainView) Refresh() {
	v.busyDialog.Show(v.lang.T(lang.Status_refresh_title), v.lang.T(lang.Status_refreshing))

	go func() {
		interfaces, err := v.ctrl.ListInterfaces()

		fyne.DoAndWait(func() {
			v.busyDialog.Hide()

			if err != nil {
				v.statusBar.SetStatus(fmt.Sprintf("Error: %v", err), false)
				return
			}

			v.interfaces = interfaces
			v.lastRefresh = time.Now()
			v.rebuild()
		})
	}()
}

func (v *MainView) rebuild() {
	v.listContainer.Objects = nil

	filter := strings.TrimSpace(strings.ToLower(v.filterEntry.Text))

	for i, iface := range v.interfaces {
		if filter != "" && !strings.Contains(strings.ToLower(iface.Name), filter) {
			continue
		}

		ifaceCopy := iface
		card := wgwidget.NewInterfaceCard(ifaceCopy, wgwidget.InterfaceCardCallbacks{
			OnToggle: func(name string, activate bool) {
				v.toggleInterface(name, activate)
			},
			OnScan: func(name, ip string) {
				scanView := NewScanViewWithProgress(name, ip, v.settings.ScanWorkers, v.settings.ScanTimeoutSecs, v.lang)
				scanView.Show()
			},
			OnEdit: func(name string) {
				v.showEditTunnelForm(name)
			},
			OnPeers: func(name string) {
				v.showEditPeersTunnelForm(name)
			},
			OnDelete: func(name string) {
				v.confirmDeleteTunnel(name)
			},
			OnCopyPubKey: func(pubKey string) {
				v.window.Clipboard().SetContent(pubKey)
				v.statusBar.SetStatus(v.lang.T(lang.Status_copied_pubkey), true)
			},
		}, v.lang)

		v.listContainer.Add(card)
		if i != len(v.interfaces)-1 {
			v.listContainer.Add(widget.NewSeparator())
		}
	}

	v.listContainer.Refresh()
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

			if err != nil && !strings.Contains(err.Error(), "invalid endpoint") {
				v.statusBar.SetStatus(fmt.Sprintf("Error: %v", err), false)
			} else {
				if err != nil {
					if strings.Contains(err.Error(), "invalid endpoint") {
						v.statusBar.SetStatus(
							v.lang.Tf(lang.Status_endpoint_unreachable, name, v.lastRefresh.Format(time.Kitchen)),
							false,
						)
					}
				} else {
					statusKey := lang.Main_deactivated_status
					if activate {
						statusKey = lang.Main_activated_status
					}
					v.statusBar.SetStatus(v.lang.Tf(statusKey, name), true)
				}
			}
			v.statusBar.Refresh()
			v.Refresh()
		})
	}()
}

func (v *MainView) showAddTunnelForm() {
	form := NewTunnelForm(v.window, v.ctrl, "", nil, func(name string, cfg *config.Config) error {
		err := v.ctrl.WriteConfig(name, *cfg)
		if err == nil {
			v.Refresh()
		}
		return err
	}, nil, v.settings.ClientConfigDir, v.lang)
	form.Show()
}

func (v *MainView) showEditTunnelForm(name string) {
	path := v.ctrl.GetConfigPath(name)
	cfg, err := config.ParseConfig(path)
	if err != nil {
		helpers.ShowError(fmt.Errorf(v.lang.T(lang.Main_failed_to_load_config), err), v.window)
		return
	}

	opened := v.preCheckActiveDialog(name, cfg)
	if !opened {
		v.openEditForm(name, cfg)
	}
}

func (v *MainView) showEditPeersTunnelForm(name string) {
	path := v.ctrl.GetConfigPath(name)
	cfg, err := config.ParseConfig(path)
	if err != nil {
		helpers.ShowError(fmt.Errorf(v.lang.T(lang.Main_failed_to_load_config), err), v.window)
		return
	}

	opened := v.preCheckActiveDialog(name, cfg)
	if !opened {
		v.openPeersForm(name, cfg)
	}
}

func (v *MainView) preCheckActiveDialog(name string, cfg *config.Config) bool {
	iface := v.findInterface(name)
	if iface != nil && iface.Active {
		helpers.ShowConfirm(v.lang.T(lang.Main_tunnel_active_title),
			v.lang.Tf(lang.Main_tunnel_active_msg, name),
			func(yes bool) {
				if yes {
					v.openEditForm(name, cfg)
				}
			}, v.window)
		return true
	}
	return false
}

func (v *MainView) openEditForm(name string, cfg *config.Config) {
	v.openForm(name, cfg, true)
}

func (v *MainView) openPeersForm(name string, cfg *config.Config) {
	v.openForm(name, cfg, false)
}

func (v *MainView) openForm(name string, cfg *config.Config, isMain bool) {
	form := NewTunnelForm(v.window, v.ctrl, name, cfg, func(_ string, newConfig *config.Config) error {
		err := v.ctrl.WriteConfig(name, *newConfig)
		if err == nil {
			v.Refresh()
		}
		return err
	}, nil, v.settings.ClientConfigDir, v.lang)
	if isMain {
		form.Show()
	} else {
		form.ShowPeers()
	}
}

func (v *MainView) confirmDeleteTunnel(name string) {
	iface := v.findInterface(name)

	deleteFn := func() {
		v.busyDialog.Show(v.lang.T(lang.Main_delete_tunnel_title), v.lang.Tf(lang.Main_deleting, name))

		go func() {
			if iface != nil && iface.Active {
				_ = v.ctrl.ToggleInterface(name, false)
			}

			err := v.ctrl.DeleteInterface(name, true)

			fyne.DoAndWait(func() {
				v.busyDialog.Hide()

				if err != nil {
					v.statusBar.SetStatus(fmt.Sprintf("Error: %v", err), false)
				} else {
					v.statusBar.SetStatus(v.lang.Tf(lang.Main_deleted_status, name), true)
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
		if !yes {
			return
		}
		deleteFn()
	}, v.window)
}

func (v *MainView) showBackupsDialog() {
	bv := NewBackupView(v.window, v.ctrl, func() {
		v.Refresh()
	}, v.lang)
	bv.Show()
}

func (v *MainView) findInterface(name string) *config.Interface {
	for _, iface := range v.interfaces {
		if iface.Name == name {
			return &iface
		}
	}
	return nil
}

func (v *MainView) startAutoRefresh() {
	ticker := time.NewTicker(time.Duration(v.settings.AutoRefreshSecs) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v.Refresh()
		case <-v.stopAuto:
			return
		}
	}
}

func (v *MainView) stopAutoRefresh() {
	select {
	case v.stopAuto <- struct{}{}:
	default:
	}
}

func (v *MainView) applySettings(updated *settings.AppSettings) {
	oldPath := v.settings.WGConfigPath
	v.settings = updated

	// Re-initialize controller if path changed
	if updated.WGConfigPath != oldPath {
		newCtrl := wg.New(updated.WGConfigPath)
		v.ctrl = &newCtrl
	}

	// Restart auto-refresh with new interval
	v.stopAutoRefresh()
	if v.autoRefresh.Checked {
		go v.startAutoRefresh()
	}

	// Update auto-refresh label
	v.autoRefresh.Text = v.lang.Tf(lang.Main_auto_refresh, updated.AutoRefreshSecs)
	v.autoRefresh.Refresh()

	// Update footer hint
	v.hint.ParseMarkdown(v.lang.Tf(lang.Main_footer_hint, updated.WGConfigPath))

	// Apply theme variant — the new theme reads ThemeVariant from settings
	// and forces light/dark/system accordingly
	newTheme := wgtheme.NewWGAdminTheme(v.settings)
	fyne.CurrentApp().Settings().SetTheme(newTheme)

	// Update header colors to match new theme variant
	variant := wgtheme.CurrentVariant()
	v.headerTitle.Color = newTheme.Color(theme.ColorNameForeground, variant)
	v.headerTitle.Refresh()
	v.headerBg.FillColor = newTheme.Color(theme.ColorNameHeaderBackground, variant)
	v.headerBg.Refresh()

	v.Refresh()
}
