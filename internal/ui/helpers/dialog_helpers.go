package helpers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// ShowConfirm shows a confirm dialog with consistent sizing
func ShowConfirm(title, message string, callback func(bool), parent fyne.Window) {
	d := dialog.NewConfirm(title, message, callback, parent)
	d.Resize(fyne.NewSize(500, 200))
	d.Show()
}

// ShowError shows an error dialog with consistent sizing
func ShowError(err error, parent fyne.Window) {
	d := dialog.NewError(err, parent)
	d.Resize(fyne.NewSize(500, 200))
	d.Show()
}

// ShowInformation shows an info dialog with consistent sizing
func ShowInformation(title, message string, parent fyne.Window) {
	d := dialog.NewInformation(title, message, parent)
	d.Resize(fyne.NewSize(500, 200))
	d.Show()
}
