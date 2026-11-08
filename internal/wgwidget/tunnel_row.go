package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// TunnelRow is one entry in the sidebar: a pulsing state dot, the name, the
// address and a peer badge, with its own hover and selected states.
//
// It handles taps itself rather than relying on widget.List selection, so the
// list's own highlight never doubles up with the row's. Rows are reused by
// the list, hence the Set method instead of rebuilding.
type TunnelRow struct {
	widget.BaseWidget

	OnTap func()

	name, address, meta string
	active, selected    bool
	hovered             bool

	bg     *canvas.Rectangle
	bar    *canvas.Rectangle
	dot    *PulseDot
	nameT  *canvas.Text
	addrT  *canvas.Text
	badge  *canvas.Text
	badgeB *canvas.Rectangle
}

var (
	_ fyne.Tappable      = (*TunnelRow)(nil)
	_ desktop.Hoverable  = (*TunnelRow)(nil)
	_ desktop.Cursorable = (*TunnelRow)(nil)
)

// NewTunnelRow creates an empty row for widget.List to populate.
func NewTunnelRow() *TunnelRow {
	r := &TunnelRow{}
	r.bg = canvas.NewRectangle(color.Transparent)
	r.bg.CornerRadius = wgtheme.RadiusMD
	r.bar = canvas.NewRectangle(color.Transparent)
	r.bar.CornerRadius = 1.5
	r.dot = NewPulseDot(color.NRGBA{}, false)
	r.nameT = newText("", wgtheme.TextMD, wgtheme.WeightSemiBold, color.Transparent)
	r.addrT = newText("", wgtheme.TextSM, wgtheme.WeightMono, color.Transparent)
	r.badge = newText("", wgtheme.TextXS, wgtheme.WeightSemiBold, color.Transparent)
	r.badgeB = canvas.NewRectangle(color.Transparent)
	r.ExtendBaseWidget(r)
	return r
}

// Set updates the row's content. active drives the state colours, selected
// the highlight.
func (r *TunnelRow) Set(name, address, meta string, active, selected bool) {
	r.name, r.address, r.meta = name, address, meta
	r.active, r.selected = active, selected
	r.Refresh()
}

func (r *TunnelRow) Tapped(*fyne.PointEvent) {
	if r.OnTap != nil {
		r.OnTap()
	}
}
func (r *TunnelRow) MouseIn(*desktop.MouseEvent)    { r.hovered = true; r.Refresh() }
func (r *TunnelRow) MouseMoved(*desktop.MouseEvent) {}
func (r *TunnelRow) MouseOut()                      { r.hovered = false; r.Refresh() }
func (r *TunnelRow) Cursor() desktop.Cursor         { return desktop.PointerCursor }

// MinSize keeps every row the same height so the list scrolls predictably.
func (r *TunnelRow) MinSize() fyne.Size { return fyne.NewSize(180, 50) }

func (r *TunnelRow) CreateRenderer() fyne.WidgetRenderer {
	return &tunnelRowRenderer{r: r}
}

type tunnelRowRenderer struct{ r *TunnelRow }

func (rr *tunnelRowRenderer) MinSize() fyne.Size { return rr.r.MinSize() }

func (rr *tunnelRowRenderer) Layout(size fyne.Size) {
	r := rr.r
	r.bg.Move(fyne.NewPos(6, 2))
	r.bg.Resize(fyne.NewSize(size.Width-12, size.Height-4))
	r.bar.Move(fyne.NewPos(6, 12))
	r.bar.Resize(fyne.NewSize(3, size.Height-24))

	ds := r.dot.MinSize()
	r.dot.Move(fyne.NewPos(16, (size.Height-ds.Height)/2))
	r.dot.Resize(ds)

	x := 16 + ds.Width + 6
	nh, ah := r.nameT.MinSize().Height, r.addrT.MinSize().Height
	top := (size.Height - nh - ah) / 2
	r.nameT.Move(fyne.NewPos(x, top))
	r.addrT.Move(fyne.NewPos(x, top+nh))

	bs := r.badge.MinSize()
	bw, bh := bs.Width+12, bs.Height+2
	bx := size.Width - 16 - bw
	r.badgeB.Move(fyne.NewPos(bx, (size.Height-bh)/2))
	r.badgeB.Resize(fyne.NewSize(bw, bh))
	r.badgeB.CornerRadius = bh / 2
	r.badge.Move(fyne.NewPos(bx+6, (size.Height-bs.Height)/2))

	// canvas.Text does not truncate; clip long names before the badge.
	maxW := bx - x - 8
	r.nameT.Text = fitText(r.nameT, r.name, maxW)
	r.nameT.Resize(fyne.NewSize(maxW, nh))
	r.addrT.Resize(fyne.NewSize(maxW, ah))
}

func (rr *tunnelRowRenderer) Refresh() {
	r := rr.r
	p := wgtheme.Current()

	switch {
	case r.selected:
		r.bg.FillColor = p.AccentTint()
		r.bar.FillColor = p.Accent
	case r.hovered:
		r.bg.FillColor = p.HoverOverlay()
		r.bar.FillColor = color.Transparent
	default:
		r.bg.FillColor = color.Transparent
		r.bar.FillColor = color.Transparent
	}

	tone := p.TextDim
	if r.active {
		tone = p.Success
	}
	r.dot.Set(tone, r.active)

	styleText(r.nameT, wgtheme.TextMD, wgtheme.WeightSemiBold)
	r.nameT.Color = p.TextPrimary
	if !r.active && !r.selected {
		r.nameT.Color = p.TextSecondary
	}
	styleText(r.addrT, wgtheme.TextSM, wgtheme.WeightMono)
	r.addrT.Text = r.address
	r.addrT.Color = p.TextDim

	styleText(r.badge, wgtheme.TextXS, wgtheme.WeightSemiBold)
	r.badge.Text = r.meta
	r.badge.Color = p.TextSecondary
	r.badgeB.FillColor = p.Tint(p.TextDim, 0.12)
	if r.active {
		r.badge.Color = p.Success
		r.badgeB.FillColor = p.Tint(p.Success, 0.14)
	}
	if r.meta == "" {
		r.badgeB.FillColor = color.Transparent
	}

	rr.Layout(r.Size())
	for _, o := range rr.Objects() {
		o.Refresh()
	}
}

func (rr *tunnelRowRenderer) Objects() []fyne.CanvasObject {
	r := rr.r
	return []fyne.CanvasObject{r.bg, r.bar, r.dot, r.nameT, r.addrT, r.badgeB, r.badge}
}
func (rr *tunnelRowRenderer) Destroy() {}

// fitText shortens s with an ellipsis until it fits w at t's size and font.
func fitText(t *canvas.Text, s string, w float32) string {
	if w <= 0 {
		return ""
	}
	probe := canvas.NewText(s, nil)
	probe.TextSize, probe.TextStyle, probe.FontSource = t.TextSize, t.TextStyle, t.FontSource
	if probe.MinSize().Width <= w {
		return s
	}
	r := []rune(s)
	for n := len(r) - 1; n > 0; n-- {
		probe.Text = string(r[:n]) + "…"
		if probe.MinSize().Width <= w {
			return probe.Text
		}
	}
	return "…"
}

// SidebarHeader is a group heading in the sidebar ("Active 3").
func SidebarHeader(title string, count int) fyne.CanvasObject {
	p := wgtheme.Current()
	t := newText(title, wgtheme.TextXS+0.5, wgtheme.WeightSemiBold, p.TextDim)
	n := newText(itoa(count), wgtheme.TextXS+0.5, wgtheme.WeightSemiBold, p.TextDim)
	return container.New(layout.NewCustomPaddedLayout(10, 2, 18, 18),
		container.NewBorder(nil, nil, t, n))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// =================================================================== ChipButton

// ChipButton is a pill that acts as a button, for compact status-and-control
// elements such as the "Live · 5s" refresh indicator.
type ChipButton struct {
	widget.BaseWidget
	Lead     fyne.CanvasObject
	Text     string
	Tone     color.NRGBA
	Tooltip  string
	OnTapped func()

	hovered bool
}

var (
	_ fyne.Tappable      = (*ChipButton)(nil)
	_ desktop.Hoverable  = (*ChipButton)(nil)
	_ desktop.Cursorable = (*ChipButton)(nil)
)

// NewChipButton creates a chip button.
func NewChipButton(lead fyne.CanvasObject, text string, tone color.NRGBA, onTapped func()) *ChipButton {
	c := &ChipButton{Lead: lead, Text: text, Tone: tone, OnTapped: onTapped}
	c.ExtendBaseWidget(c)
	return c
}

// Set changes the lead object, text and tone.
func (c *ChipButton) Set(lead fyne.CanvasObject, text string, tone color.NRGBA) {
	c.Lead, c.Text, c.Tone = lead, text, tone
	c.Refresh()
}

func (c *ChipButton) Tapped(*fyne.PointEvent) {
	HideTooltip()
	if c.OnTapped != nil {
		c.OnTapped()
	}
}
func (c *ChipButton) MouseIn(*desktop.MouseEvent) {
	c.hovered = true
	tooltipHover(c, c.Tooltip, true)
	c.Refresh()
}
func (c *ChipButton) MouseMoved(*desktop.MouseEvent) {}
func (c *ChipButton) MouseOut() {
	c.hovered = false
	tooltipHover(c, c.Tooltip, false)
	c.Refresh()
}
func (c *ChipButton) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (c *ChipButton) CreateRenderer() fyne.WidgetRenderer {
	r := &chipRenderer{c: c, bg: canvas.NewRectangle(color.Transparent),
		label: newText("", wgtheme.TextSM, wgtheme.WeightSemiBold, color.Transparent)}
	r.Refresh()
	return r
}

type chipRenderer struct {
	c     *ChipButton
	bg    *canvas.Rectangle
	label *canvas.Text
	box   *fyne.Container
}

func (r *chipRenderer) build() {
	items := []fyne.CanvasObject{}
	if r.c.Lead != nil {
		items = append(items, container.NewCenter(r.c.Lead))
	}
	items = append(items, container.NewCenter(r.label))
	r.box = container.New(layout.NewCustomPaddedLayout(0, 0, 10, 12),
		container.New(layout.NewCustomPaddedHBoxLayout(6), items...))
}

func (r *chipRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.box.MinSize().Width, wgtheme.ControlHeight(wgtheme.ControlSM))
}

func (r *chipRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.CornerRadius = size.Height / 2
	r.box.Resize(size)
}

func (r *chipRenderer) Refresh() {
	p := wgtheme.Current()
	r.label.Text = r.c.Text
	styleText(r.label, wgtheme.TextSM, wgtheme.WeightSemiBold)
	r.label.Color = r.c.Tone
	r.bg.FillColor = p.Tint(r.c.Tone, 0.12)
	r.bg.StrokeColor = p.Tint(r.c.Tone, 0.32)
	if r.c.hovered {
		r.bg.FillColor = p.Tint(r.c.Tone, 0.2)
	}
	r.bg.StrokeWidth = 1
	r.build()
	r.Layout(r.c.Size())
	r.bg.Refresh()
	r.label.Refresh()
}

func (r *chipRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.bg, r.box} }
func (r *chipRenderer) Destroy()                     {}

// CodeChip is a monospace chip for machine values such as CIDRs.
func CodeChip(text string) fyne.CanvasObject {
	p := wgtheme.Current()
	label := newText(text, wgtheme.TextSM, wgtheme.WeightMono, p.TextSecondary)
	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusSM + 1
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(3, 3, 8, 8), label))
}

// CopyChip is a CodeChip that copies its text when clicked, with a hover
// state so it reads as clickable.
type CopyChip struct {
	widget.BaseWidget
	Text   string
	OnCopy func()

	hovered bool
	bg      *canvas.Rectangle
	label   *canvas.Text
}

var (
	_ fyne.Tappable      = (*CopyChip)(nil)
	_ desktop.Hoverable  = (*CopyChip)(nil)
	_ desktop.Cursorable = (*CopyChip)(nil)
)

// NewCopyChip creates a copyable chip.
func NewCopyChip(text string, onCopy func()) *CopyChip {
	c := &CopyChip{Text: text, OnCopy: onCopy}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CopyChip) Tapped(*fyne.PointEvent) {
	if c.OnCopy != nil {
		c.OnCopy()
	}
}
func (c *CopyChip) MouseIn(*desktop.MouseEvent)    { c.hovered = true; c.Refresh() }
func (c *CopyChip) MouseMoved(*desktop.MouseEvent) {}
func (c *CopyChip) MouseOut()                      { c.hovered = false; c.Refresh() }
func (c *CopyChip) Cursor() desktop.Cursor         { return desktop.PointerCursor }

func (c *CopyChip) CreateRenderer() fyne.WidgetRenderer {
	c.bg = canvas.NewRectangle(color.Transparent)
	c.bg.CornerRadius = wgtheme.RadiusSM + 1
	c.bg.StrokeWidth = 1
	c.label = newText(c.Text, wgtheme.TextSM, wgtheme.WeightMono, color.Transparent)
	box := container.NewStack(c.bg, container.New(layout.NewCustomPaddedLayout(3, 3, 8, 8), c.label))
	r := &copyChipRenderer{c: c, box: box}
	r.Refresh()
	return r
}

type copyChipRenderer struct {
	c   *CopyChip
	box *fyne.Container
}

func (r *copyChipRenderer) Layout(s fyne.Size)           { r.box.Resize(s) }
func (r *copyChipRenderer) MinSize() fyne.Size           { return r.box.MinSize() }
func (r *copyChipRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.box} }
func (r *copyChipRenderer) Destroy()                     {}
func (r *copyChipRenderer) Refresh() {
	p := wgtheme.Current()
	c := r.c
	c.bg.FillColor, c.bg.StrokeColor, c.label.Color = p.SurfaceRaised, p.Border, p.TextSecondary
	if c.hovered {
		c.bg.FillColor, c.bg.StrokeColor, c.label.Color = p.AccentSoft(), p.Tint(p.Accent, 0.4), p.TextPrimary
	}
	c.label.Text = c.Text
	c.bg.Refresh()
	c.label.Refresh()
}
