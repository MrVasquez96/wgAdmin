package wgwidget

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// BusyDialog is the modal shown while a privileged operation runs. It uses the
// Activity spinner rather than an indeterminate progress bar, which reads as a
// process running rather than as progress that never advances.
type BusyDialog struct {
	mu       sync.Mutex
	dialog   dialog.Dialog
	activity *widget.Activity
	window   fyne.Window
}

// NewBusyDialog creates a busy dialog bound to a window.
func NewBusyDialog(window fyne.Window) *BusyDialog {
	return &BusyDialog{window: window}
}

// Show displays the busy dialog with a title and message.
func (b *BusyDialog) Show(title, msg string) {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.dialog != nil {
			b.dialog.Hide()
		}

		activity := widget.NewActivity()
		b.activity = activity

		label := Value(msg)
		label.Alignment = fyne.TextAlignCenter

		caption := Caption(title)
		caption.Alignment = fyne.TextAlignCenter

		content := container.New(layout.NewCustomPaddedVBoxLayout(10),
			container.NewCenter(container.New(
				layout.NewGridWrapLayout(fyne.NewSize(36, 36)), activity)),
			container.NewCenter(label),
			container.NewCenter(caption),
		)

		b.dialog = dialog.NewCustomWithoutButtons(title,
			container.New(layout.NewCustomPaddedLayout(8, 8, 20, 20), content), b.window)
		b.dialog.Resize(fyne.NewSize(340, 170))
		b.dialog.Show()
		activity.Start()
	})
}

// Hide dismisses the busy dialog.
func (b *BusyDialog) Hide() {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.activity != nil {
			b.activity.Stop()
			b.activity = nil
		}
		if b.dialog != nil {
			b.dialog.Hide()
			b.dialog = nil
		}
	})
}
