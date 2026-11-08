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

// Fyne has no shadows, so elevation is drawn as a few stacked translucent
// rounded rectangles, each a little larger and fainter than the last and
// nudged downward — the same falloff a real soft shadow has.

// shadowSpec is one elevation level: layer count, spread per layer, vertical
// offset and the alpha of the innermost layer.
type shadowSpec struct {
	layers int
	spread float32
	offset float32
	alpha  uint8
}

var shadowLevels = [...]shadowSpec{
	{0, 0, 0, 0},
	{3, 1, 1, 34},   // tooltips, menus' resting state
	{4, 2, 2, 40},   // cards on hover, popovers
	{5, 3.5, 4, 48}, // modals and sheets
}

// shadow draws an elevation shadow sized to whatever it is stacked with.
type shadow struct {
	widget.BaseWidget
	level  int
	radius float32
}

// shadowFor returns a shadow to put first in a Stack, underneath a surface
// with the given corner radius.
func shadowFor(level int, radius float32) fyne.CanvasObject {
	if level <= 0 {
		return canvas.NewRectangle(color.Transparent)
	}
	if level >= len(shadowLevels) {
		level = len(shadowLevels) - 1
	}
	s := &shadow{level: level, radius: radius}
	s.ExtendBaseWidget(s)
	return s
}

func (s *shadow) MinSize() fyne.Size { return fyne.Size{} }

func (s *shadow) CreateRenderer() fyne.WidgetRenderer {
	spec := shadowLevels[s.level]
	p := wgtheme.Current()
	rects := make([]fyne.CanvasObject, spec.layers)
	for i := range rects {
		// Outer layers are fainter; the sum reads as one soft edge.
		a := uint8(float32(spec.alpha) / float32(i+2))
		r := canvas.NewRectangle(p.Shadow(a))
		r.CornerRadius = s.radius + spec.spread*float32(i+1)
		rects[i] = r
	}
	return &shadowRenderer{s: s, rects: rects}
}

type shadowRenderer struct {
	s     *shadow
	rects []fyne.CanvasObject
}

func (r *shadowRenderer) Layout(size fyne.Size) {
	spec := shadowLevels[r.s.level]
	for i, o := range r.rects {
		grow := spec.spread * float32(i+1)
		o.Move(fyne.NewPos(-grow, -grow+spec.offset))
		o.Resize(fyne.NewSize(size.Width+2*grow, size.Height+2*grow))
	}
}

func (r *shadowRenderer) MinSize() fyne.Size           { return fyne.Size{} }
func (r *shadowRenderer) Refresh()                     { r.Layout(r.s.Size()) }
func (r *shadowRenderer) Objects() []fyne.CanvasObject { return r.rects }
func (r *shadowRenderer) Destroy()                     {}

// Card is a surface with an optional elevation shadow, the building block for
// panels, popovers and modals.
func Card(content fyne.CanvasObject, level int) fyne.CanvasObject {
	return cardWith(content, level, wgtheme.Current().Surface, wgtheme.SpaceLG)
}

// RaisedCard is Card on the raised surface tone.
func RaisedCard(content fyne.CanvasObject, level int) fyne.CanvasObject {
	return cardWith(content, level, wgtheme.Current().SurfaceRaised, wgtheme.SpaceLG)
}

func cardWith(content fyne.CanvasObject, level int, fill color.NRGBA, pad float32) fyne.CanvasObject {
	p := wgtheme.Current()
	bg := canvas.NewRectangle(fill)
	bg.CornerRadius = wgtheme.RadiusLG
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1

	return container.NewStack(
		shadowFor(level, bg.CornerRadius),
		bg,
		container.New(layout.NewCustomPaddedLayout(pad, pad, pad, pad), content),
	)
}
