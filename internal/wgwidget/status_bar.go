package wgwidget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
)

// StatusBar is the footer strip. The left side carries the most recent message
// and the right side carries persistent environment facts — config path,
// backend, privilege level — as chips, replacing the pipe-separated sentence
// the footer used to be.
type StatusBar struct {
	widget.BaseWidget

	lang *lang.Language

	background *canvas.Rectangle
	icon       *canvas.Image
	message    *canvas.Text
	chipBox    *fyne.Container
	container  *fyne.Container

	// flashID identifies the most recent transient message so a pending
	// auto-clear does not wipe out a newer, still-relevant status.
	// Only touched on the UI thread.
	flashID uint64
}

// NewStatusBar creates a status strip showing the ready message.
func NewStatusBar(language *lang.Language) *StatusBar {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(p.Header)
	icon := iconImage(wgtheme.IconInfo, p.TextDim, 14)
	msg := newText(language.T(lang.Status_ready), wgtheme.TextSM, wgtheme.WeightMedium, p.TextSecondary)

	chips := container.New(layout.NewCustomPaddedHBoxLayout(6))

	s := &StatusBar{
		lang:       language,
		background: bg,
		icon:       icon,
		message:    msg,
		chipBox:    chips,
	}

	left := container.New(layout.NewCustomPaddedHBoxLayout(7),
		container.NewCenter(icon), container.NewCenter(msg))
	body := container.NewBorder(nil, nil, left, chips)

	s.container = container.NewStack(bg,
		container.NewBorder(Hairline(), nil, nil, nil,
			container.New(layout.NewCustomPaddedLayout(5, 5, 14, 12), body)))

	s.ExtendBaseWidget(s)
	return s
}

// SetChips replaces the environment chips on the right of the strip.
func (s *StatusBar) SetChips(items ...fyne.CanvasObject) {
	s.chipBox.Objects = items
	s.chipBox.Refresh()
}

// SetStatus shows a success or failure message.
func (s *StatusBar) SetStatus(msg string, success bool) {
	p := wgtheme.Current()
	if success {
		s.set(msg, wgtheme.IconCheck, p.Success, p.SuccessTint())
		return
	}
	s.set(msg, wgtheme.IconError, p.Error, p.ErrorTint())
}

// SetInfo shows a neutral informational message.
func (s *StatusBar) SetInfo(msg string) {
	p := wgtheme.Current()
	s.set(msg, wgtheme.IconInfo, p.Info, p.InfoTint())
}

// SetWarning shows a warning, used for degraded states such as telemetry being
// unavailable — conditions that are neither success nor hard failure.
func (s *StatusBar) SetWarning(msg string) {
	p := wgtheme.Current()
	s.set(msg, wgtheme.IconAlert, p.Warning, p.WarningTint())
}

// Reset returns the strip to its neutral resting state.
func (s *StatusBar) Reset() {
	p := wgtheme.Current()
	s.flashID++
	s.set(s.lang.T(lang.Status_ready), wgtheme.IconInfo, p.TextSecondary, p.Header)
}

// FlashDuration is how long a transient confirmation stays on the strip.
const FlashDuration = 4 * time.Second

// Flash shows a confirmation that clears itself, for cheap acknowledgements
// like "copied" that should not linger as if they were current state.
// A later Flash or Reset supersedes an earlier one rather than fighting it.
func (s *StatusBar) Flash(msg string, success bool) {
	s.SetStatus(msg, success)

	s.flashID++
	id := s.flashID

	time.AfterFunc(FlashDuration, func() {
		fyne.Do(func() {
			// Only clear if nothing else has written to the strip since.
			if s.flashID == id {
				s.Reset()
			}
		})
	})
}

func (s *StatusBar) set(msg string, icon *wgtheme.Icon, tone, bg color.NRGBA) {
	setIconColor(s.icon, icon, tone)
	s.message.Text = msg
	s.message.Color = tone
	s.message.Refresh()
	s.background.FillColor = bg
	s.background.Refresh()
	s.Refresh()
}

// CreateRenderer implements fyne.Widget
func (s *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}
