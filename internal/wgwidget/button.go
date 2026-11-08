package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// Variant is a button's visual role.
type Variant int

const (
	// Secondary is the default: raised surface, hairline border.
	Secondary Variant = iota
	// Primary is the one action a screen is for. Use at most one per area.
	Primary
	// Ghost has no surface until hovered; for toolbars and row actions.
	Ghost
	// Danger is a soft red button for destructive actions offered inline.
	Danger
	// DangerSolid is the filled red button that confirms a destructive action.
	DangerSolid
	// Success is the filled green button, used for "Bring up".
	Success
)

// Scale is a button's height class.
type Scale int

const (
	ScaleMD Scale = iota
	ScaleSM
	ScaleLG
)

// Button is the application's push button: a variant colour scheme, an
// optional duotone icon, hover/press/focus states with short colour
// transitions, a pointer cursor, keyboard activation and an optional tooltip.
type Button struct {
	widget.DisableableWidget

	Text     string
	Icon     *wgtheme.Icon
	Variant  Variant
	Scale    Scale
	Tooltip  string
	OnTapped func()

	hovered, pressed, focused bool
}

var (
	_ fyne.Tappable      = (*Button)(nil)
	_ fyne.Focusable     = (*Button)(nil)
	_ desktop.Hoverable  = (*Button)(nil)
	_ desktop.Mouseable  = (*Button)(nil)
	_ desktop.Cursorable = (*Button)(nil)
)

// NewButton creates a button. icon may be nil.
func NewButton(text string, icon *wgtheme.Icon, variant Variant, onTapped func()) *Button {
	b := &Button{Text: text, Icon: icon, Variant: variant, OnTapped: onTapped}
	b.ExtendBaseWidget(b)
	return b
}

// NewIconButton creates a square, icon-only ghost button. The tooltip is what
// names it, so it is required.
func NewIconButton(icon *wgtheme.Icon, tooltip string, onTapped func()) *Button {
	b := NewButton("", icon, Ghost, onTapped)
	b.Tooltip = tooltip
	return b
}

// SetText changes the label.
func (b *Button) SetText(s string) { b.Text = s; b.Refresh() }

// SetIcon changes the icon.
func (b *Button) SetIcon(ic *wgtheme.Icon) { b.Icon = ic; b.Refresh() }

func (b *Button) Tapped(*fyne.PointEvent) {
	if b.Disabled() {
		return
	}
	HideTooltip()
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

func (b *Button) MouseIn(*desktop.MouseEvent) {
	b.hovered = true
	tooltipHover(b, b.Tooltip, true)
	b.Refresh()
}

func (b *Button) MouseMoved(*desktop.MouseEvent) {}

func (b *Button) MouseOut() {
	b.hovered, b.pressed = false, false
	tooltipHover(b, b.Tooltip, false)
	b.Refresh()
}

func (b *Button) MouseDown(*desktop.MouseEvent) {
	if !b.Disabled() {
		b.pressed = true
		b.Refresh()
	}
}

func (b *Button) MouseUp(*desktop.MouseEvent) {
	b.pressed = false
	b.Refresh()
}

func (b *Button) Cursor() desktop.Cursor {
	if b.Disabled() {
		return desktop.DefaultCursor
	}
	return desktop.PointerCursor
}

func (b *Button) FocusGained() { b.focused = true; b.Refresh() }
func (b *Button) FocusLost()   { b.focused = false; b.Refresh() }
func (b *Button) TypedRune(r rune) {
	if r == ' ' {
		b.Tapped(nil)
	}
}
func (b *Button) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyReturn || ev.Name == fyne.KeyEnter {
		b.Tapped(nil)
	}
}

// metrics returns the height, horizontal padding and icon size for the size.
func (b *Button) metrics() (h, padX, icon float32) {
	switch b.Scale {
	case ScaleSM:
		return wgtheme.ControlHeight(wgtheme.ControlSM), 10, 15
	case ScaleLG:
		return wgtheme.ControlHeight(wgtheme.ControlLG), 18, 19
	}
	return wgtheme.ControlHeight(wgtheme.ControlMD), 13, 17
}

func (b *Button) textSize() float32 {
	if b.Scale == ScaleSM {
		return wgtheme.TextSM
	}
	return wgtheme.TextMD
}

// colors resolves fill, border and content colours for the current state.
func (b *Button) colors() (fill, border, fg color.NRGBA) {
	p := wgtheme.Current()
	clear := color.NRGBA{}

	switch b.Variant {
	case Primary:
		fill, border, fg = p.Accent, clear, p.OnAccent()
	case Success:
		fill, border, fg = p.Success, clear, p.OnColor(p.Success)
	case DangerSolid:
		fill, border, fg = p.Error, clear, p.OnColor(p.Error)
	case Danger:
		fill, border, fg = p.Tint(p.Error, 0.14), p.Tint(p.Error, 0.38), p.Error
	case Ghost:
		fill, border, fg = clear, clear, p.TextSecondary
		if b.hovered || b.focused {
			fg = p.TextPrimary
		}
	default:
		fill, border, fg = p.SurfaceRaised, p.Border, p.TextPrimary
		if b.hovered {
			border = p.BorderStrong
		}
	}

	if b.Disabled() {
		if fill.A > 0 {
			fill = wgtheme.Mix(p.Surface, fill, 0.35)
		}
		border = wgtheme.Mix(p.Surface, border, 0.5)
		return fill, border, p.TextDisabled
	}
	return fill, border, fg
}

func (b *Button) MinSize() fyne.Size {
	b.ExtendBaseWidget(b)
	return b.BaseWidget.MinSize()
}

func (b *Button) CreateRenderer() fyne.WidgetRenderer {
	b.ExtendBaseWidget(b)
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = wgtheme.RadiusMD
	overlay := canvas.NewRectangle(color.Transparent)
	overlay.CornerRadius = wgtheme.RadiusMD
	ring := canvas.NewRectangle(color.Transparent)
	ring.CornerRadius = wgtheme.RadiusMD + 2
	ring.StrokeWidth = 2

	r := &buttonRenderer{b: b, bg: bg, overlay: overlay, ring: ring,
		label: newText("", b.textSize(), wgtheme.WeightSemiBold, color.Transparent),
		icon:  &canvas.Image{FillMode: canvas.ImageFillContain},
	}
	r.Refresh()
	return r
}

type buttonRenderer struct {
	b       *Button
	bg      *canvas.Rectangle
	overlay *canvas.Rectangle
	ring    *canvas.Rectangle
	label   *canvas.Text
	icon    *canvas.Image

	fillTween tween
	lastFill  color.Color
}

func (r *buttonRenderer) MinSize() fyne.Size {
	h, padX, icon := r.b.metrics()
	if r.b.Text == "" {
		return fyne.NewSize(h, h)
	}
	w := padX*2 + r.label.MinSize().Width
	if r.b.Icon != nil {
		w += icon + 7
	}
	return fyne.NewSize(w, h)
}

func (r *buttonRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.overlay.Resize(size)
	r.ring.Move(fyne.NewPos(-3, -3))
	r.ring.Resize(fyne.NewSize(size.Width+6, size.Height+6))

	_, _, icon := r.b.metrics()
	ls := r.label.MinSize()
	content := ls.Width
	if r.b.Icon != nil {
		content += icon
		if r.b.Text != "" {
			content += 7
		}
	}
	x := (size.Width - content) / 2
	if r.b.Icon != nil {
		r.icon.Resize(fyne.NewSize(icon, icon))
		r.icon.Move(fyne.NewPos(x, (size.Height-icon)/2))
		x += icon + 7
	}
	r.label.Move(fyne.NewPos(x, (size.Height-ls.Height)/2))
	r.label.Resize(ls)
}

func (r *buttonRenderer) Refresh() {
	b := r.b
	p := wgtheme.Current()
	fill, border, fg := b.colors()

	if r.lastFill == nil {
		r.bg.FillColor = fill
	} else if r.lastFill != fill {
		r.fillTween.to(r.bg.FillColor, fill, wgtheme.MotionFast, func(c color.Color) {
			r.bg.FillColor = c
			r.bg.Refresh()
		})
	}
	r.lastFill = fill
	r.bg.StrokeColor = border
	r.bg.StrokeWidth = 0
	if border.A > 0 {
		r.bg.StrokeWidth = 1
	}

	// Hover and press are a translucent wash over any fill, so every variant
	// responds the same way without needing its own hover colours.
	switch {
	case b.Disabled():
		r.overlay.FillColor = color.Transparent
	case b.pressed:
		r.overlay.FillColor = p.PressOverlay()
	case b.hovered:
		r.overlay.FillColor = p.HoverOverlay()
	default:
		r.overlay.FillColor = color.Transparent
	}

	r.ring.StrokeColor = color.Transparent
	if b.focused && !b.Disabled() {
		r.ring.StrokeColor = p.Ring()
	}

	r.label.Text = b.Text
	r.label.Color = fg
	styleText(r.label, b.textSize(), wgtheme.WeightSemiBold)
	if b.Icon != nil {
		setIconColor(r.icon, b.Icon, fg)
		r.icon.Show()
	} else {
		r.icon.Hide()
	}

	r.Layout(b.Size())
	r.bg.Refresh()
	r.overlay.Refresh()
	r.ring.Refresh()
	r.label.Refresh()
}

func (r *buttonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.ring, r.bg, r.overlay, r.icon, r.label}
}

func (r *buttonRenderer) Destroy() {}
