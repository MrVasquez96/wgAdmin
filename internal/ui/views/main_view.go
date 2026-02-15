package views

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	"wgAdmin/internal/ui/components"
	"wgAdmin/internal/wireguard"
)

// MainView is the main application view
type MainView struct {
	window        fyne.Window
	listContainer *fyne.Container
	statusBar     *components.StatusBar
	busyDialog    *components.BusyDialog
	filterEntry   *widget.Entry
	autoRefresh   *widget.Check

	interfaces  []models.Interface
	stopAuto    chan struct{}
	lastRefresh time.Time
}

// NewMainView creates a new main view
func NewMainView(window fyne.Window) *MainView {
	return &MainView{
		window:        window,
		listContainer: container.NewVBox(),
		statusBar:     components.NewStatusBar(),
		busyDialog:    components.NewBusyDialog(window),
		filterEntry:   widget.NewEntry(),
		autoRefresh:   widget.NewCheck("Auto refresh (5s)", nil),
		stopAuto:      make(chan struct{}),
	}
}

// Build creates the main view content
func (v *MainView) Build() fyne.CanvasObject {
	// Title
	title := canvas.NewText("WireGuard Manager", theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 24

	// Add tunnel button
	addBtn := widget.NewButtonWithIcon("Add Tunnel", theme.ContentAddIcon(), func() {
		v.showAddTunnelForm()
	})
	addBtn.Importance = widget.HighImportance

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

	// Header layout
	leftHeader := container.NewHBox(title)
	rightHeader := container.NewHBox(addBtn, filterContainer, v.autoRefresh, refreshBtn)
	header := container.NewBorder(nil, nil, leftHeader, rightHeader)

	// Scrollable list
	scroll := container.NewVScroll(v.listContainer)
	scroll.SetMinSize(fyne.NewSize(860, 480))

	// Footer hint
	hint := widget.NewRichTextFromMarkdown("Configs: `/etc/wireguard` | Native WireGuard | Requires root privileges")
	hint.Wrapping = fyne.TextWrapWord

	// Main content
	content := container.NewBorder(
		header,
		container.NewVBox(hint, v.statusBar),
		nil, nil,
		scroll,
	)

	return container.NewPadded(content)
}

// Refresh reloads the interface list
func (v *MainView) Refresh() {
	v.busyDialog.Show("Refreshing interfaces...")

	go func() {
		interfaces, err := wireguard.ListInterfaces()

		fyne.DoAndWait(func() {
			v.busyDialog.Hide()

			if err != nil {
				v.statusBar.SetStatus(fmt.Sprintf("Error: %v", err), false)
				return
			}

			v.interfaces = interfaces
			v.lastRefresh = time.Now()
			v.statusBar.SetStatus(
				fmt.Sprintf("Loaded %d interface(s) at %s", len(interfaces), v.lastRefresh.Format(time.Kitchen)),
				true,
			)
			v.rebuild()
		})
	}()
}

func (v *MainView) rebuild() {
	v.listContainer.Objects = nil

	filter := strings.TrimSpace(strings.ToLower(v.filterEntry.Text))

	for _, iface := range v.interfaces {
		if filter != "" && !strings.Contains(strings.ToLower(iface.Name), filter) {
			continue
		}

		ifaceCopy := iface
		card := components.NewInterfaceCard(ifaceCopy, components.InterfaceCardCallbacks{
			OnToggle: func(name string, activate bool) {
				v.toggleInterface(name, activate)
			},
			OnScan: func(name, ip string) {
				scanView := NewScanView(name, ip)
				scanView.Show()
			},
			OnEdit: func(name string) {
				v.showEditTunnelForm(name)
			},
			OnDelete: func(name string) {
				v.confirmDeleteTunnel(name)
			},
		})

		v.listContainer.Add(card)
		v.listContainer.Add(widget.NewSeparator())
	}

	v.listContainer.Refresh()
}

func (v *MainView) toggleInterface(name string, activate bool) {
	action := "Deactivating"
	if activate {
		action = "Activating"
	}

	v.busyDialog.Show(fmt.Sprintf("%s %s...", action, name))

	go func() {
		err := wireguard.ToggleInterface(name, activate)

		fyne.DoAndWait(func() {
			v.busyDialog.Hide()

			if err != nil { 
				v.statusBar.SetStatus(fmt.Sprintf("Error: %v", err), false)
			} else {
				action := "deactivated"
				if activate {
					action = "activated"
				}
				v.statusBar.SetStatus(fmt.Sprintf("%s %s successfully", name, action), true)
			}

			v.Refresh()
		})
	}()
}

func (v *MainView) showAddTunnelForm() {
	form := NewTunnelForm(v.window, "", nil, func(name string, config *models.Config) error {
		err := wireguard.WriteConfig(name, config)
		if err == nil {
			v.Refresh()
		}
		return err
	}, nil)
	form.Show()
}

func (v *MainView) showEditTunnelForm(name string) {
	path := wireguard.GetConfigPath(name)
	config, err := wireguard.ParseConfig(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load config: %w", err), v.window)
		return
	}

	// Check if active
	iface := v.findInterface(name)
	if iface != nil && iface.Active {
		dialog.ShowConfirm("Tunnel Active",
			fmt.Sprintf("Tunnel '%s' is currently active. It's recommended to deactivate before editing.\n\nContinue anyway?", name),
			func(yes bool) {
				if yes {
					v.openEditForm(name, config)
				}
			}, v.window)
		return
	}

	v.openEditForm(name, config)
}

func (v *MainView) openEditForm(name string, config *models.Config) {
	form := NewTunnelForm(v.window, name, config, func(_ string, newConfig *models.Config) error {
		err := wireguard.WriteConfig(name, newConfig)
		if err == nil {
			v.Refresh()
		}
		return err
	}, nil)
	form.Show()
}

func (v *MainView) confirmDeleteTunnel(name string) {
	iface := v.findInterface(name)

	msg := fmt.Sprintf("Delete tunnel '%s'?\n\nThis will remove:\n%s", name, wireguard.GetConfigPath(name))
	if iface != nil && iface.Active {
		msg += "\n\nWarning: This tunnel is currently active and will be deactivated."
	}

	dialog.ShowConfirm("Delete Tunnel", msg, func(yes bool) {
		if !yes {
			return
		}

		v.busyDialog.Show(fmt.Sprintf("Deleting %s...", name))

		go func() {
			// Deactivate if active
			if iface != nil && iface.Active {
				_ = wireguard.ToggleInterface(name, false)
			}

			err := wireguard.DeleteInterface(name, true)

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
	}, v.window)
}

func (v *MainView) findInterface(name string) *models.Interface {
	for _, iface := range v.interfaces {
		if iface.Name == name {
			return &iface
		}
	}
	return nil
}

func (v *MainView) startAutoRefresh() {
	ticker := time.NewTicker(5 * time.Second)
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
