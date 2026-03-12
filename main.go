package main

import (
	"fmt"
	"os"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/theme"
	ui "wgAdmin/internal/ui/views"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/MrVasquez96/go-wg/wg"
)

func main() {
	a := app.NewWithID("com.wireguard.manager")
	//
	meta := a.Metadata()
	cfg := settings.Load(a.Preferences())

	// Apply theme — NewWGAdminTheme reads cfg.ThemeVariant and forces
	// light/dark variant accordingly (or follows system when "system")
	a.Settings().SetTheme(theme.NewWGAdminTheme(cfg))

	// Privilege escalation via pkexec (before creating any windows).
	// We spawn a new root process and then quit this one gracefully.
	if os.Geteuid() != 0 && cfg.PrivilegeEscalation == "pkexec" {
		if settings.PkexecAvailable() {
			if err := settings.RelaunchWithPkexec(); err != nil {
				fmt.Fprintf(os.Stderr, "pkexec failed: %v\n", err)
			} else {
				// New process spawned — exit this one.
				// Cannot use a.Quit() here since ShowAndRun hasn't been called.
				return
			}
		} else {
			fmt.Fprintln(os.Stderr, "pkexec not available, continuing without root.")
		}
	}

	if os.Geteuid() != 0 && cfg.PrivilegeEscalation == "none" {
		fmt.Fprintln(os.Stderr, "Warning: This application typically requires root privileges for WireGuard operations.")
		fmt.Fprintln(os.Stderr, "Some features may not work correctly without elevated permissions.")
		fmt.Fprintln(os.Stderr, "")
	}

	ctrl := wg.New(cfg.WGConfigPath)

	version := "unknown"
	if meta.Version != "" {
		version = meta.Version
	}

	w := a.NewWindow(meta.Name + " - version: " + version)
	w.Resize(fyne.NewSize(float32(cfg.WindowWidth), float32(cfg.WindowHeight)))
	if cfg.StartFullscreen {
		w.SetFullScreen(true)
	}

	mainView := ui.NewMainView(w, &ctrl, cfg).Build(meta)
	w.SetContent(mainView)
	mainView.Refresh()

	// If sudo mode is selected and not root, show password dialog over the main UI.
	// On successful auth, a new root process is spawned and this app quits gracefully.
	if os.Geteuid() != 0 && cfg.PrivilegeEscalation == "sudo" {
		settings.ShowSudoRelaunchDialog(w)
	}

	w.ShowAndRun()
}
