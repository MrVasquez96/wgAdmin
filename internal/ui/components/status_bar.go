package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// StatusBar displays status messages with icons
type StatusBar struct {
	widget.BaseWidget

	container *fyne.Container
	icon      *widget.Icon
	message   *widget.Label
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	s := &StatusBar{
		icon:    widget.NewIcon(theme.InfoIcon()),
		message: widget.NewLabel("Ready"),
	}

	s.container = container.NewHBox(s.icon, s.message)
	s.ExtendBaseWidget(s)
	return s
}

// SetStatus sets a status message with success/error styling
func (s *StatusBar) SetStatus(msg string, success bool) {
	if success {
		s.icon.SetResource(theme.ConfirmIcon())
	} else {
		s.icon.SetResource(theme.CancelIcon())
	}
	s.message.SetText(msg)
	s.Refresh()
}

// SetInfo sets an informational message
func (s *StatusBar) SetInfo(msg string) {
	s.icon.SetResource(theme.InfoIcon())
	s.message.SetText(msg)
	s.Refresh()
}

// CreateRenderer implements fyne.Widget
func (s *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}
