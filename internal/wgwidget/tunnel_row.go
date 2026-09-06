package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// TunnelRow is one entry in the sidebar list: state dot, name, address and a
// peer count. It replaces the old full-width card, which spent ~350px of
// vertical space per tunnel and made a dozen tunnels unnavigable.
//
// Rows are reused by widget.List, so the type exposes Set* methods rather than
// rebuilding itself — that is what keeps selection and scroll position stable
// across the auto-refresh tick.
type TunnelRow struct {
	widget.BaseWidget

	background *canvas.Rectangle
	accent     *canvas.Rectangle
	dot        *canvas.Circle
	halo       *canvas.Circle
	name       *canvas.Text
	address    *canvas.Text
	meta       *canvas.Text
	container  *fyne.Container
}

// NewTunnelRow creates an empty row for widget.List to populate.
func NewTunnelRow() *TunnelRow {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = wgtheme.RadiusPanel

	accent := canvas.NewRectangle(color.Transparent)
	accent.SetMinSize(fyne.NewSize(2, 0))

	dot := canvas.NewCircle(p.TextDim)
	halo := canvas.NewCircle(color.Transparent)
	halo.StrokeWidth = 2

	name := canvas.NewText("", p.TextPrimary)
	name.TextSize = 13
	name.TextStyle = fyne.TextStyle{Bold: true}

	address := canvas.NewText("", p.TextSecondary)
	address.TextSize = 11
	address.TextStyle = fyne.TextStyle{Monospace: true}

	meta := canvas.NewText("", p.TextDim)
	meta.TextSize = 10
	meta.TextStyle = fyne.TextStyle{Monospace: true}

	r := &TunnelRow{
		background: bg, accent: accent,
		dot: dot, halo: halo,
		name: name, address: address, meta: meta,
	}

	dotStack := container.NewCenter(container.NewStack(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(14, 14)), halo),
		container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(8, 8)), dot)),
	))

	text := container.New(layout.NewCustomPaddedVBoxLayout(1),
		container.NewBorder(nil, nil, nil, meta, name),
		address,
	)

	body := container.NewBorder(nil, nil,
		container.New(layout.NewCustomPaddedLayout(0, 0, 2, 4), dotStack),
		nil, text)

	r.container = container.NewStack(bg,
		container.NewBorder(nil, nil, accent, nil,
			container.New(layout.NewCustomPaddedLayout(6, 6, 6, 8), body),
		),
	)

	r.ExtendBaseWidget(r)
	return r
}

// Set updates the row's content. active drives the state colours, selected
// drives the highlight.
func (r *TunnelRow) Set(name, address, meta string, active, selected bool) {
	p := wgtheme.Current()

	r.name.Text = name
	r.address.Text = address
	r.meta.Text = meta

	tone := p.TextDim
	if active {
		tone = p.Success
	}
	r.dot.FillColor = tone
	if active {
		r.halo.StrokeColor = wgtheme.Alpha(tone, 70)
	} else {
		r.halo.StrokeColor = color.Transparent
	}

	switch {
	case selected:
		r.background.FillColor = p.AccentTint()
		r.accent.FillColor = p.Accent
		r.name.Color = p.TextPrimary
	default:
		r.background.FillColor = color.Transparent
		r.accent.FillColor = color.Transparent
		r.name.Color = p.TextPrimary
		if !active {
			r.name.Color = p.TextSecondary
		}
	}

	r.background.Refresh()
	r.accent.Refresh()
	r.dot.Refresh()
	r.halo.Refresh()
	r.name.Refresh()
	r.address.Refresh()
	r.meta.Refresh()
}

// MinSize keeps every row the same height so the list scrolls predictably.
func (r *TunnelRow) MinSize() fyne.Size {
	return fyne.NewSize(180, 48)
}

// CreateRenderer implements fyne.Widget
func (r *TunnelRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.container)
}
