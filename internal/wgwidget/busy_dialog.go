package wgwidget

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
func (b *BusyDialog) Show(title, msg string) {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.dialog != nil {
			b.dialog.Hide()
		}

		icon := widget.NewIcon(theme.InfoIcon())
		spinner := widget.NewProgressBarInfinite()
		label := widget.NewLabel(msg)
		label.Alignment = fyne.TextAlignCenter
		label.Wrapping = fyne.TextWrapWord

		content := container.NewVBox(
			container.NewCenter(icon),
			widget.NewSeparator(),
			container.NewPadded(label),
			container.NewPadded(spinner),
		)

		b.dialog = dialog.NewCustomWithoutButtons(title, content, b.window)
		b.dialog.Resize(fyne.NewSize(400, 200))
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
