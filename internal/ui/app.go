package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	customTheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/ui/views"
)

// App represents the WireGuard Admin application
type App struct {
	fyneApp fyne.App
	window  fyne.Window
	view    *views.MainView
}

// NewApp creates a new application instance
func NewApp() *App {
	a := app.NewWithID("com.wireguard.manager")
	a.Settings().SetTheme(customTheme.NewWGAdminTheme())

	w := a.NewWindow("WireGuard Manager")
	w.Resize(fyne.NewSize(950, 800))

	return &App{
		fyneApp: a,
		window:  w,
	}
}

// Run starts the application
func (a *App) Run() {
	a.view = views.NewMainView(a.window)
	a.window.SetContent(a.view.Build())

	// Initial refresh
	a.view.Refresh()

	a.window.ShowAndRun()
}
