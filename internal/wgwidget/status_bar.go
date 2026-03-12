package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	customTheme "wgAdmin/internal/ui/theme"
)

// StatusBar displays status messages with icons and colored backgrounds
type StatusBar struct {
	widget.BaseWidget

	container  *fyne.Container
	background *canvas.Rectangle
	icon       *widget.Icon
	message    *widget.Label
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = 6

	icon := widget.NewIcon(theme.InfoIcon())
	msg := widget.NewLabel("Ready")

	content := container.NewHBox(icon, msg)

	s := &StatusBar{
		icon:       icon,
		message:    msg,
		background: bg,
		container:  container.NewStack(bg, container.NewPadded(content)),
	}

	s.ExtendBaseWidget(s)
	return s
}

// SetStatus sets a status message with success/error styling
func (s *StatusBar) SetStatus(msg string, success bool) {
	variant := customTheme.CurrentVariant()

	if success {
		s.icon.SetResource(theme.ConfirmIcon())
		s.background.FillColor = customTheme.AppColors.StatusSuccessBg(variant)
	} else {
		s.icon.SetResource(theme.ErrorIcon())
		s.background.FillColor = customTheme.AppColors.StatusErrorBg(variant)
	}
	s.message.SetText(msg)
	s.background.Refresh()
	s.Refresh()
}

// SetInfo sets an informational message
func (s *StatusBar) SetInfo(msg string) {
	variant := customTheme.CurrentVariant()

	s.icon.SetResource(theme.InfoIcon())
	s.background.FillColor = customTheme.AppColors.StatusInfoBg(variant)
	s.message.SetText(msg)
	s.background.Refresh()
	s.Refresh()
}

// CreateRenderer implements fyne.Widget
func (s *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}
