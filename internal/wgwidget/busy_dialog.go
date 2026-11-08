package wgwidget

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// BusyDialog is the blocking card shown while a privileged operation runs:
// a spinner, what is happening and to which tunnel. It cannot be dismissed —
// the operation is already under way and closing the card would not stop it.
type BusyDialog struct {
	mu       sync.Mutex
	surface  *Surface
	activity *widget.Activity
	window   fyne.Window
}

// NewBusyDialog creates a busy dialog bound to a window.
func NewBusyDialog(window fyne.Window) *BusyDialog {
	return &BusyDialog{window: window}
}

// Show displays the busy card with a title (the tunnel) and message (the
// operation).
func (b *BusyDialog) Show(title, msg string) {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.surface != nil {
			b.surface.Close()
		}

		p := wgtheme.Current()
		activity := widget.NewActivity()
		b.activity = activity

		head := newText(msg, wgtheme.TextLG, wgtheme.WeightSemiBold, p.TextPrimary)
		sub := newText(title, wgtheme.TextSM, wgtheme.WeightMono, p.TextSecondary)

		bg := canvas.NewRectangle(p.Surface)
		bg.CornerRadius = wgtheme.RadiusXL
		bg.StrokeColor = p.BorderStrong
		bg.StrokeWidth = 1

		row := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceLG+2),
			container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(30, 30)), activity)),
			container.NewVBox(layout.NewSpacer(), head, sub, layout.NewSpacer()))
		card := container.NewStack(shadowFor(3, bg.CornerRadius), bg,
			container.New(layout.NewCustomPaddedLayout(18, 18, 22, 28), row))

		o := &overlayRoot{place: placeCenter, width: 300, dim: true}
		o.panel = guard(card)
		b.surface = open(b.window, o)
		b.surface.Intercept = func(func()) {} // Esc and backdrop clicks do nothing
		activity.Start()
	})
}

// Hide dismisses the busy card.
func (b *BusyDialog) Hide() {
	fyne.Do(func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.activity != nil {
			b.activity.Stop()
			b.activity = nil
		}
		if b.surface != nil {
			b.surface.Close()
			b.surface = nil
		}
	})
}
