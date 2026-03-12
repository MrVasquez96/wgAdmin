package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
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
func NewMainView(window fyne.Window, ctrl *wg.WG, cfg *settings.AppSettings) *MainView {
	return &MainView{
		window:        window,
		ctrl:          ctrl,
		settings:      cfg,
		listContainer: container.NewVBox(),
		statusBar:     wgwidget.NewStatusBar(),
		busyDialog:    wgwidget.NewBusyDialog(window),
		filterEntry:   widget.NewEntry(),
		autoRefresh:   widget.NewCheck(fmt.Sprintf("Auto refresh (%ds)", cfg.AutoRefreshSecs), nil),
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
	addBtn := widget.NewButtonWithIcon("Add Tunnel", theme.ContentAddIcon(), func() {
		v.showAddTunnelForm()
	})

	addBtn.Importance = widget.HighImportance

	// Backups button
	backupsBtn := widget.NewButtonWithIcon("Backups", theme.FolderOpenIcon(), func() {
		v.showBackupsDialog()
	})

	// Filter
	v.filterEntry.SetPlaceHolder("Filter by name...")
	v.filterEntry.OnChanged = func(s string) {
		v.rebuild()
	}
	filterContainer := container.NewGridWrap(fyne.NewSize(200, v.filterEntry.MinSize().Height), v.filterEntry)

	// Refresh button
	refreshBtn := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
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
		})
		sv.Show()
	})

	// Auto-refresh on startup if configured
	if v.settings.AutoRefreshEnabled {
		v.autoRefresh.SetChecked(true)
	}

	// Header layout with background
	leftHeader := container.NewHBox(v.headerTitle)
	rightHeader := container.NewHBox(importBtn, addBtn, backupsBtn, filterContainer, v.autoRefresh, refreshBtn, settingsBtn)
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
	v.hint = widget.NewRichTextFromMarkdown(
		fmt.Sprintf("Configs: `%s` | Native WireGuard | Requires root privileges", v.settings.WGConfigPath))
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

	return widget.NewButtonWithIcon("Import from file", theme.DocumentSaveIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				helpers.ShowError(err, v.window)
				return
			}
			if reader == nil {
				log.Println("Cancelled")
				return
			}
			// 2. Handle the file content
			defer reader.Close()
			// data, _ := io.ReadAll(reader)
			filePath := reader.URI().Path()
			if filepath.Ext(filePath) != ".conf" {
				helpers.ShowError(errors.New("Wireguard config must end with '.conf'"), v.window)
				return
			}
			cfg, err := config.ParseConfig(filePath)
			if err != nil {
				helpers.ShowError(errors.New("invalid config\n"+err.Error()), v.window)
				return
			}
			if err = cfg.Validate(); err != nil {
				helpers.ShowInformation("Validation", "Failed to validate config!", v.window)
			}
			name := cfg.Name
			if name == "Unknown" {
				name = reader.URI().Name()
				name = filepath.Base(name)
			}
			_, err = os.Stat(filepath.Join(v.settings.WGConfigPath, reader.URI().Name()))

			if err == nil {
				helpers.ShowConfirm("File exists", "File already exists. Overwrite?", func(b bool) {
					if b {
						v.save(name, cfg)
					} else {
						helpers.ShowInformation("Aborted", "Nothing saved", v.window)
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
		helpers.ShowError(errors.New("Error saving:\n"+err.Error()), v.window)
		return
	}
	helpers.ShowInformation("Success", "Imported:"+name, v.window)
}

// Refresh reloads the interface list
func (v *MainView) Refresh() {
	v.busyDialog.Show("Refresh", "Refreshing interfaces...")

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
				scanView := NewScanViewWithProgress(name, ip, v.settings.ScanWorkers, v.settings.ScanTimeoutSecs)
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
				v.statusBar.SetStatus("Public key copied to clipboard", true)
			},
		})

		v.listContainer.Add(card)
		if i != len(v.interfaces)-1 {
			v.listContainer.Add(widget.NewSeparator())
		}
	}

	v.listContainer.Refresh()
}

func (v *MainView) toggleInterface(name string, activate bool) {
	action := "Deactivating"
	if activate {
		action = "Activating"
	}

	v.busyDialog.Show(name, fmt.Sprintf("%s...", action))

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
							fmt.Sprintf("Failed to reach endpoint for '%s' at %s", name, v.lastRefresh.Format(time.Kitchen)),
							false,
						)
					}
				} else {
					action := "deactivated"
					if activate {
						action = "activated"
					}
					v.statusBar.SetStatus(fmt.Sprintf("%s %s successfully", name, action), true)
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
	}, nil, v.settings.ClientConfigDir)
	form.Show()
}

func (v *MainView) showEditTunnelForm(name string) {
	path := v.ctrl.GetConfigPath(name)
	cfg, err := config.ParseConfig(path)
	if err != nil {
		helpers.ShowError(fmt.Errorf("failed to load config: %w", err), v.window)
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
		helpers.ShowError(fmt.Errorf("failed to load config: %w", err), v.window)
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
		helpers.ShowConfirm("Tunnel Active",
			fmt.Sprintf("Tunnel '%s' is currently active. It's recommended to deactivate before editing.\n\nContinue anyway?", name),
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
	}, nil, v.settings.ClientConfigDir)
	if isMain {
		form.Show()
	} else {
		form.ShowPeers()
	}
}

func (v *MainView) confirmDeleteTunnel(name string) {
	iface := v.findInterface(name)

	deleteFn := func() {
		v.busyDialog.Show("Delete Tunnel", fmt.Sprintf("Deleting %s...", name))

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
					v.statusBar.SetStatus(fmt.Sprintf("Deleted %s (backup created)", name), true)
				}

				v.Refresh()
			})
		}()
	}

	if !v.settings.ConfirmBeforeDelete {
		deleteFn()
		return
	}

	msg := fmt.Sprintf("Delete tunnel '%s'?\n\nThis will remove:\n%s", name, v.ctrl.GetConfigPath(name))
	if iface != nil && iface.Active {
		msg += "\n\nWarning: This tunnel is currently active and will be deactivated."
	}

	helpers.ShowConfirm("Delete Tunnel", msg, func(yes bool) {
		if !yes {
			return
		}
		deleteFn()
	}, v.window)
}

func (v *MainView) showBackupsDialog() {
	bv := NewBackupView(v.window, v.ctrl, func() {
		v.Refresh()
	})
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
	v.autoRefresh.Text = fmt.Sprintf("Auto refresh (%ds)", updated.AutoRefreshSecs)
	v.autoRefresh.Refresh()

	// Update footer hint
	v.hint.ParseMarkdown(
		fmt.Sprintf("Configs: `%s` | Native WireGuard | Requires root privileges", updated.WGConfigPath))

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
