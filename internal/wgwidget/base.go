package wgwidget

import (
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	wgtheme "wgAdmin/internal/ui/theme"
)

// This file holds what every component shares: weighted text, duotone icon
// images, colour transitions, and the pass-through layer that tooltips and
// toasts float in.

// ReduceMotion, when it returns true, makes every transition jump straight to
// its end state. It is a variable so settings can wire it up without this
// package importing them.
var ReduceMotion = wgtheme.ReduceMotion

// newText builds text at a type-scale size and weight. When the user has
// turned custom fonts off, wgtheme.Font returns nil and the closest TextStyle
// is used instead, so the system font still gets bold where it matters.
func newText(s string, size float32, w wgtheme.Weight, c color.Color) *canvas.Text {
	t := canvas.NewText(s, c)
	styleText(t, size, w)
	return t
}

func styleText(t *canvas.Text, size float32, w wgtheme.Weight) {
	t.TextSize = wgtheme.TextSize(size)
	t.FontSource = wgtheme.Font(w)
	t.TextStyle = fyne.TextStyle{
		Bold:      t.FontSource == nil && w >= wgtheme.WeightSemiBold && w != wgtheme.WeightMono,
		Monospace: w == wgtheme.WeightMono,
	}
}

// iconImage renders a duotone icon in c at a square size.
func iconImage(ic *wgtheme.Icon, c color.Color, size float32) *canvas.Image {
	img := canvas.NewImageFromResource(ic.For(c))
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return img
}

// setIconColor recolours an icon image in place.
func setIconColor(img *canvas.Image, ic *wgtheme.Icon, c color.Color) {
	if ic == nil {
		return
	}
	res := ic.For(c)
	if img.Resource != res {
		img.Resource = res
		img.Refresh()
	}
}

// tween runs one colour transition and cancels whatever it replaced, so a
// quick hover-in/hover-out never leaves two animations fighting.
type tween struct{ anim *fyne.Animation }

func (t *tween) to(from, to color.Color, d time.Duration, apply func(color.Color)) {
	if t.anim != nil {
		t.anim.Stop()
		t.anim = nil
	}
	if ReduceMotion() || from == nil || fyne.CurrentApp() == nil {
		apply(to)
		return
	}
	t.anim = canvas.NewColorRGBAAnimation(from, to, d, apply)
	t.anim.Curve = fyne.AnimationEaseOut
	t.anim.Start()
}

// ---------------------------------------------------------------------- layers

// Layer is a pass-through surface stacked above a window's content. Floating
// things that must not block the app — tooltips and toasts — are positioned in
// it by hand. It has no layout and draws nothing, so clicks fall straight
// through to whatever is underneath.
//
// Fyne's own canvas overlays are the wrong tool here: while any overlay is
// open, hit-testing and keyboard focus go to it exclusively, which is exactly
// right for a modal and exactly wrong for a tooltip.
type Layer struct {
	*fyne.Container
}

var (
	layersMu sync.Mutex
	layers   = map[fyne.Canvas][]*Layer{}
)

// WithLayer wraps a window's content with a floating layer and registers it.
// Call it once, when setting the window content.
func WithLayer(win fyne.Window, content fyne.CanvasObject) fyne.CanvasObject {
	l := &Layer{Container: container.NewWithoutLayout()}
	// The window's layer is always the bottom of its stack; calling this again
	// (e.g. rebuilding the window after a theme change) replaces it.
	layersMu.Lock()
	layers[win.Canvas()] = []*Layer{l}
	layersMu.Unlock()
	return container.NewStack(content, l.Container)
}

func pushLayer(c fyne.Canvas, l *Layer) {
	layersMu.Lock()
	defer layersMu.Unlock()
	layers[c] = append(layers[c], l)
}

func popLayer(c fyne.Canvas, l *Layer) {
	layersMu.Lock()
	defer layersMu.Unlock()
	list := layers[c]
	for i := range list {
		if list[i] == l {
			layers[c] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// topLayer is the floating layer for obj's canvas: the topmost modal's own
// layer when one is open, otherwise the window's.
func topLayer(obj fyne.CanvasObject) (*Layer, fyne.Canvas) {
	app := fyne.CurrentApp()
	if app == nil {
		return nil, nil
	}
	c := app.Driver().CanvasForObject(obj)
	if c == nil {
		return nil, nil
	}
	layersMu.Lock()
	defer layersMu.Unlock()
	list := layers[c]
	if len(list) == 0 {
		return nil, c
	}
	return list[len(list)-1], c
}

// place positions obj in l at the absolute canvas position pos.
func (l *Layer) place(obj fyne.CanvasObject, pos fyne.Position) {
	origin := fyne.CurrentApp().Driver().AbsolutePositionForObject(l.Container)
	obj.Resize(obj.MinSize())
	obj.Move(pos.Subtract(origin))
	l.Add(obj)
}

// bounds is the layer's size, for keeping floating things on screen.
func (l *Layer) bounds() fyne.Size { return l.Container.Size() }
