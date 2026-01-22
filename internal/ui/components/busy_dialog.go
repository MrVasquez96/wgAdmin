package components

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// BusyDialog shows a loading spinner dialog
type BusyDialog struct {
	mu     sync.Mutex
	dialog dialog.Dialog
	window fyne.Window
}

// NewBusyDialog creates a new busy dialog
func NewBusyDialog(window fyne.Window) *BusyDialog {
	return &BusyDialog{
		window: window,
	}
}

// Show displays the busy dialog with a message
func (b *BusyDialog) Show(msg string) {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.dialog != nil {
			b.dialog.Hide()
		}

		spinner := widget.NewProgressBarInfinite()
		label := widget.NewLabel(msg)
		content := container.NewVBox(label, spinner)

		b.dialog = dialog.NewCustomWithoutButtons("Working...", content, b.window)
		b.dialog.Show()
	})
}

// Hide hides the busy dialog
func (b *BusyDialog) Hide() {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.dialog != nil {
			b.dialog.Hide()
			b.dialog = nil
		}
	})
}
