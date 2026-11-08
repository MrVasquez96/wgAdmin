package wgwidget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// =================================================================== CopyField

// CopyField shows a labelled value that copies itself when clicked. Hovering
// reveals a soft surface and a copy icon, so it is discoverable without
// putting a button on every row; after a copy the icon briefly becomes a
// check, so the click visibly did something.
type CopyField struct {
	widget.BaseWidget

	Label  string
	Value  string
	Icon   *wgtheme.Icon // optional leading field icon
	OnCopy func(value string)

	hovered bool
	copied  bool
	timer   *time.Timer
}

var (
	_ fyne.Tappable      = (*CopyField)(nil)
	_ desktop.Hoverable  = (*CopyField)(nil)
	_ desktop.Cursorable = (*CopyField)(nil)
)

// NewCopyField creates a field. A nil onCopy, or a value that is only a
// placeholder, makes it read-only.
func NewCopyField(icon *wgtheme.Icon, label, value string, onCopy func(string)) *CopyField {
	f := &CopyField{Label: label, Value: value, Icon: icon, OnCopy: onCopy}
	f.ExtendBaseWidget(f)
	return f
}

func (f *CopyField) copyable() bool { return f.OnCopy != nil }

func (f *CopyField) Tapped(*fyne.PointEvent) {
	if !f.copyable() {
		return
	}
	f.OnCopy(f.Value)
	f.copied = true
	f.Refresh()
	if f.timer != nil {
		f.timer.Stop()
	}
	f.timer = time.AfterFunc(1400*time.Millisecond, func() {
		fyne.Do(func() { f.copied = false; f.Refresh() })
	})
}

func (f *CopyField) MouseIn(*desktop.MouseEvent)    { f.hovered = true; f.Refresh() }
func (f *CopyField) MouseMoved(*desktop.MouseEvent) {}
func (f *CopyField) MouseOut()                      { f.hovered = false; f.Refresh() }

func (f *CopyField) Cursor() desktop.Cursor {
	if f.copyable() {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (f *CopyField) CreateRenderer() fyne.WidgetRenderer {
	p := wgtheme.Current()
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = wgtheme.RadiusMD

	value := widget.NewLabel("")
	value.TextStyle = fyne.TextStyle{Monospace: true}
	value.Truncation = fyne.TextTruncateEllipsis

	r := &copyFieldRenderer{
		f:      f,
		bg:     bg,
		label:  newText("", wgtheme.TextXS, wgtheme.WeightSemiBold, p.TextDim),
		value:  value,
		action: &canvas.Image{FillMode: canvas.ImageFillContain},
		lead:   &canvas.Image{FillMode: canvas.ImageFillContain},
	}
	r.Refresh()
	return r
}

type copyFieldRenderer struct {
	f      *CopyField
	bg     *canvas.Rectangle
	label  *canvas.Text
	value  *widget.Label
	action *canvas.Image
	lead   *canvas.Image
	tw     tween
}

const (
	cfPadX   = 10
	cfPadY   = 6
	cfIcon   = 16
	cfAction = 16
)

func (r *copyFieldRenderer) MinSize() fyne.Size {
	lh := r.label.MinSize().Height
	vh := r.value.MinSize().Height
	w := float32(120)
	return fyne.NewSize(w, cfPadY*2+lh+vh-8)
}

func (r *copyFieldRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	x := float32(cfPadX)
	if r.f.Icon != nil {
		r.lead.Resize(fyne.NewSize(cfIcon, cfIcon))
		r.lead.Move(fyne.NewPos(x, (size.Height-cfIcon)/2))
		x += cfIcon + 10
	}
	lh := r.label.MinSize().Height
	r.label.Move(fyne.NewPos(x, cfPadY))
	r.label.Resize(r.label.MinSize())

	// widget.Label carries its own inner padding; pull it back so the value's
	// text lines up with the caption above it.
	inner := float32(6)
	right := size.Width - cfPadX - cfAction - 6
	r.value.Move(fyne.NewPos(x-inner, cfPadY+lh-inner-1))
	r.value.Resize(fyne.NewSize(right-x+inner, r.value.MinSize().Height))

	r.action.Resize(fyne.NewSize(cfAction, cfAction))
	r.action.Move(fyne.NewPos(size.Width-cfPadX-cfAction, (size.Height-cfAction)/2))
}

func (r *copyFieldRenderer) Refresh() {
	p := wgtheme.Current()
	f := r.f

	r.label.Text = f.Label
	styleText(r.label, wgtheme.TextXS, wgtheme.WeightSemiBold)
	r.label.Color = p.TextDim
	r.value.SetText(f.Value)

	target := color.Color(color.Transparent)
	if f.hovered && f.copyable() {
		target = p.HoverOverlay()
	}
	r.tw.to(r.bg.FillColor, target, wgtheme.MotionFast, func(c color.Color) {
		r.bg.FillColor = c
		r.bg.Refresh()
	})

	if f.Icon != nil {
		setIconColor(r.lead, f.Icon, p.TextDim)
		r.lead.Show()
	} else {
		r.lead.Hide()
	}

	switch {
	case f.copied:
		setIconColor(r.action, wgtheme.IconCheck, p.Success)
		r.action.Show()
	case f.hovered && f.copyable():
		setIconColor(r.action, wgtheme.IconCopy, p.TextSecondary)
		r.action.Show()
	default:
		r.action.Hide()
	}

	r.Layout(f.Size())
	r.label.Refresh()
}

func (r *copyFieldRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.lead, r.label, r.value, r.action}
}
func (r *copyFieldRenderer) Destroy() {}

// =================================================================== Segmented

// Segmented is a row of mutually exclusive options on a shared track, with
// the selected one raised. It replaces checkboxes and selects for small
// choice sets (refresh interval, theme, density).
type Segmented struct {
	widget.DisableableWidget

	Options   []string
	Icons     []*wgtheme.Icon // optional, parallel to Options
	Selected  int
	OnChanged func(index int)

	hover int
}

var (
	_ fyne.Tappable      = (*Segmented)(nil)
	_ desktop.Hoverable  = (*Segmented)(nil)
	_ desktop.Cursorable = (*Segmented)(nil)
)

// NewSegmented creates a segmented control with the first option selected.
func NewSegmented(options []string, onChanged func(int)) *Segmented {
	s := &Segmented{Options: options, OnChanged: onChanged, hover: -1}
	s.ExtendBaseWidget(s)
	return s
}

// SetSelected changes the selection without calling OnChanged.
func (s *Segmented) SetSelected(i int) {
	if i >= 0 && i < len(s.Options) {
		s.Selected = i
		s.Refresh()
	}
}

func (s *Segmented) indexAt(x float32) int {
	n := len(s.Options)
	if n == 0 {
		return -1
	}
	i := int(x / (s.Size().Width / float32(n)))
	if i < 0 || i >= n {
		return -1
	}
	return i
}

func (s *Segmented) Tapped(ev *fyne.PointEvent) {
	if s.Disabled() {
		return
	}
	if i := s.indexAt(ev.Position.X); i >= 0 && i != s.Selected {
		s.Selected = i
		s.Refresh()
		if s.OnChanged != nil {
			s.OnChanged(i)
		}
	}
}

func (s *Segmented) MouseIn(ev *desktop.MouseEvent) { s.MouseMoved(ev) }
func (s *Segmented) MouseMoved(ev *desktop.MouseEvent) {
	if i := s.indexAt(ev.Position.X); i != s.hover {
		s.hover = i
		s.Refresh()
	}
}
func (s *Segmented) MouseOut()              { s.hover = -1; s.Refresh() }
func (s *Segmented) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (s *Segmented) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(color.Transparent)
	track.CornerRadius = wgtheme.RadiusMD
	thumb := canvas.NewRectangle(color.Transparent)
	thumb.CornerRadius = wgtheme.RadiusMD - 2
	r := &segRenderer{s: s, track: track, thumb: thumb}
	r.rebuild()
	r.Refresh()
	return r
}

type segRenderer struct {
	s       *Segmented
	track   *canvas.Rectangle
	thumb   *canvas.Rectangle
	labels  []*canvas.Text
	icons   []*canvas.Image
	anim    *fyne.Animation
	animEnd time.Time
	placed  bool
}

func (r *segRenderer) rebuild() {
	r.labels = r.labels[:0]
	r.icons = r.icons[:0]
	for i, o := range r.s.Options {
		r.labels = append(r.labels, newText(o, wgtheme.TextSM, wgtheme.WeightMedium, color.Transparent))
		img := &canvas.Image{FillMode: canvas.ImageFillContain}
		if i >= len(r.s.Icons) || r.s.Icons[i] == nil {
			img.Hide()
		}
		r.icons = append(r.icons, img)
	}
}

func (r *segRenderer) segWidth() float32 {
	w := float32(0)
	for i, l := range r.labels {
		lw := l.MinSize().Width
		if r.icons[i].Visible() {
			lw += 20
		}
		if lw > w {
			w = lw
		}
	}
	return w + 22
}

func (r *segRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.segWidth()*float32(len(r.labels))+4,
		wgtheme.ControlHeight(wgtheme.ControlSM+2))
}

func (r *segRenderer) thumbRect(size fyne.Size) (fyne.Position, fyne.Size) {
	n := float32(len(r.labels))
	if n == 0 {
		return fyne.Position{}, fyne.Size{}
	}
	w := (size.Width - 4) / n
	return fyne.NewPos(2+w*float32(r.s.Selected), 2), fyne.NewSize(w, size.Height-4)
}

func (r *segRenderer) Layout(size fyne.Size) {
	r.track.Resize(size)
	pos, ts := r.thumbRect(size)
	r.thumb.Resize(ts)
	if !r.placed || time.Now().After(r.animEnd) {
		r.thumb.Move(pos)
	}
	r.placed = true

	w := (size.Width - 4) / float32(len(r.labels))
	for i, l := range r.labels {
		ls := l.MinSize()
		cw := ls.Width
		if r.icons[i].Visible() {
			cw += 20
		}
		x := 2 + w*float32(i) + (w-cw)/2
		if r.icons[i].Visible() {
			r.icons[i].Resize(fyne.NewSize(15, 15))
			r.icons[i].Move(fyne.NewPos(x, (size.Height-15)/2))
			x += 20
		}
		l.Move(fyne.NewPos(x, (size.Height-ls.Height)/2))
		l.Resize(ls)
	}
}

func (r *segRenderer) Refresh() {
	p := wgtheme.Current()
	if len(r.labels) != len(r.s.Options) {
		r.rebuild()
	}
	r.track.FillColor = p.InputBackground
	r.track.StrokeColor = p.Border
	r.track.StrokeWidth = 1
	r.thumb.FillColor = p.Lift(p.SurfaceRaised, 0.06)
	r.thumb.StrokeColor = p.BorderStrong
	r.thumb.StrokeWidth = 1

	for i, l := range r.labels {
		l.Text = r.s.Options[i]
		styleText(l, wgtheme.TextSM, wgtheme.WeightMedium)
		c := p.TextSecondary
		switch {
		case r.s.Disabled():
			c = p.TextDisabled
		case i == r.s.Selected:
			c = p.TextPrimary
		case i == r.s.hover:
			c = p.TextPrimary
		}
		l.Color = c
		if i < len(r.s.Icons) && r.s.Icons[i] != nil {
			ic := c
			if i == r.s.Selected {
				ic = p.Accent
			}
			setIconColor(r.icons[i], r.s.Icons[i], ic)
		}
		l.Refresh()
	}

	// Slide the thumb to the new selection.
	target, _ := r.thumbRect(r.s.Size())
	if r.placed && r.thumb.Position() != target && !ReduceMotion() && fyne.CurrentApp() != nil {
		if r.anim != nil {
			r.anim.Stop()
		}
		r.anim = canvas.NewPositionAnimation(r.thumb.Position(), target, wgtheme.MotionBase, func(pos fyne.Position) {
			r.thumb.Move(pos)
			canvas.Refresh(r.thumb)
		})
		r.anim.Curve = fyne.AnimationEaseOut
		r.animEnd = time.Now().Add(wgtheme.MotionBase)
		r.anim.Start()
	}
	r.Layout(r.s.Size())
	r.track.Refresh()
	r.thumb.Refresh()
}

func (r *segRenderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{r.track, r.thumb}
	for i := range r.labels {
		objs = append(objs, r.icons[i], r.labels[i])
	}
	return objs
}
func (r *segRenderer) Destroy() {}

// ====================================================================== Switch

// Switch is an on/off toggle with a sliding knob, for settings that take
// effect immediately. It replaces checkboxes for boolean preferences.
type Switch struct {
	widget.DisableableWidget
	On        bool
	OnChanged func(on bool)

	hovered, focused bool
}

var (
	_ fyne.Tappable      = (*Switch)(nil)
	_ fyne.Focusable     = (*Switch)(nil)
	_ desktop.Hoverable  = (*Switch)(nil)
	_ desktop.Cursorable = (*Switch)(nil)
)

// NewSwitch creates a switch.
func NewSwitch(on bool, onChanged func(bool)) *Switch {
	s := &Switch{On: on, OnChanged: onChanged}
	s.ExtendBaseWidget(s)
	return s
}

// SetOn changes the state without calling OnChanged.
func (s *Switch) SetOn(on bool) { s.On = on; s.Refresh() }

func (s *Switch) toggle() {
	if s.Disabled() {
		return
	}
	s.On = !s.On
	s.Refresh()
	if s.OnChanged != nil {
		s.OnChanged(s.On)
	}
}

func (s *Switch) Tapped(*fyne.PointEvent)        { s.toggle() }
func (s *Switch) MouseIn(*desktop.MouseEvent)    { s.hovered = true; s.Refresh() }
func (s *Switch) MouseMoved(*desktop.MouseEvent) {}
func (s *Switch) MouseOut()                      { s.hovered = false; s.Refresh() }
func (s *Switch) Cursor() desktop.Cursor         { return desktop.PointerCursor }
func (s *Switch) FocusGained()                   { s.focused = true; s.Refresh() }
func (s *Switch) FocusLost()                     { s.focused = false; s.Refresh() }
func (s *Switch) TypedRune(r rune) {
	if r == ' ' {
		s.toggle()
	}
}
func (s *Switch) TypedKey(*fyne.KeyEvent) {}

const (
	swW, swH = 38, 22
	swKnob   = 16
)

func (s *Switch) CreateRenderer() fyne.WidgetRenderer {
	track := canvas.NewRectangle(color.Transparent)
	track.CornerRadius = swH / 2
	ring := canvas.NewRectangle(color.Transparent)
	ring.CornerRadius = swH/2 + 3
	ring.StrokeWidth = 2
	r := &switchRenderer{s: s, track: track, ring: ring, knob: canvas.NewCircle(color.White)}
	r.Refresh()
	return r
}

type switchRenderer struct {
	s       *Switch
	track   *canvas.Rectangle
	ring    *canvas.Rectangle
	knob    *canvas.Circle
	tw      tween
	anim    *fyne.Animation
	animEnd time.Time
	init    bool
}

func (r *switchRenderer) MinSize() fyne.Size { return fyne.NewSize(swW, swH) }

func (r *switchRenderer) knobPos(size fyne.Size) fyne.Position {
	y := (size.Height - swKnob) / 2
	if r.s.On {
		return fyne.NewPos(swW-swKnob-3, y)
	}
	return fyne.NewPos(3, y)
}

func (r *switchRenderer) Layout(size fyne.Size) {
	y := (size.Height - swH) / 2
	r.track.Move(fyne.NewPos(0, y))
	r.track.Resize(fyne.NewSize(swW, swH))
	r.ring.Move(fyne.NewPos(-3, y-3))
	r.ring.Resize(fyne.NewSize(swW+6, swH+6))
	r.knob.Resize(fyne.NewSize(swKnob, swKnob))
	if time.Now().After(r.animEnd) {
		r.knob.Move(r.knobPos(size))
	}
}

func (r *switchRenderer) Refresh() {
	p := wgtheme.Current()
	s := r.s
	fill := p.BorderStrong
	if s.On {
		fill = p.Accent
	}
	if s.hovered && !s.Disabled() {
		fill = p.Lift(fill, 0.1)
	}
	if s.Disabled() {
		fill = wgtheme.Mix(p.Surface, fill, 0.4)
	}
	if !r.init {
		r.track.FillColor = fill
	} else {
		r.tw.to(r.track.FillColor, fill, wgtheme.MotionBase, func(c color.Color) {
			r.track.FillColor = c
			r.track.Refresh()
		})
	}
	r.knob.FillColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if s.On {
		r.knob.FillColor = p.OnAccent()
	}
	r.ring.StrokeColor = color.Transparent
	if s.focused {
		r.ring.StrokeColor = p.Ring()
	}

	target := r.knobPos(s.Size())
	if r.init && r.knob.Position() != target && !ReduceMotion() && fyne.CurrentApp() != nil {
		if r.anim != nil {
			r.anim.Stop()
		}
		r.anim = canvas.NewPositionAnimation(r.knob.Position(), target, wgtheme.MotionBase, func(pos fyne.Position) {
			r.knob.Move(pos)
			canvas.Refresh(r.knob)
		})
		r.anim.Curve = fyne.AnimationEaseOut
		r.animEnd = time.Now().Add(wgtheme.MotionBase)
		r.anim.Start()
	}
	r.init = true
	r.Layout(s.Size())
	r.knob.Refresh()
	r.ring.Refresh()
}

func (r *switchRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.ring, r.track, r.knob}
}
func (r *switchRenderer) Destroy() {}
