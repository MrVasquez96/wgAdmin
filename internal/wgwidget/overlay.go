package wgwidget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// Modals, sheets and menus live in Fyne's canvas overlay stack. While an
// overlay is open Fyne routes all pointer and keyboard input to it alone,
// which is exactly the blocking behaviour these need. Each overlay also
// registers its own floating Layer, so tooltips inside a sheet appear above
// it rather than underneath on the window's layer.

// placement decides where the panel sits inside the full-window overlay.
type placement int

const (
	placeCenter placement = iota
	placeRight            // full-height sheet on the right edge
	placeAt               // explicit position (menus)
)

// Surface is an open modal, sheet or menu.
type Surface struct {
	root    *overlayRoot
	OnClose func()
	// Intercept, when set, is asked before the close button, Esc or a
	// backdrop click close the surface, e.g. to confirm discarding edits.
	// It calls close to go ahead. Explicit calls to Close are not intercepted.
	Intercept func(close func())
}

// RequestClose closes the surface unless Intercept holds it open.
func (s *Surface) RequestClose() {
	if s == nil {
		return
	}
	if s.Intercept != nil {
		s.Intercept(s.Close)
		return
	}
	s.Close()
}

// Close removes the surface and runs OnClose once.
func (s *Surface) Close() {
	if s == nil || s.root == nil || s.root.closed {
		return
	}
	s.root.closed = true
	HideTooltip()
	c := s.root.canvas
	popLayer(c, s.root.layer)
	c.Overlays().Remove(s.root)
	if s.OnClose != nil {
		s.OnClose()
	}
}

type overlayRoot struct {
	widget.BaseWidget

	canvas      fyne.Canvas
	surface     *Surface
	panel       fyne.CanvasObject
	place       placement
	width       float32 // sheets: requested width; menus: unused
	at          fyne.Position
	dim         bool // draw a dimmed backdrop
	dismissable bool // a tap on the backdrop closes it

	backdrop *canvas.Rectangle
	layer    *Layer
	closed   bool

	// enter animation state: 0 → 1
	progress float32
}

var (
	_ fyne.Tappable  = (*overlayRoot)(nil)
	_ fyne.Focusable = (*overlayRoot)(nil)
)

// Taps that reach the root missed the panel: they landed on the backdrop.
func (o *overlayRoot) Tapped(*fyne.PointEvent) {
	if o.dismissable {
		o.surface.RequestClose()
	}
}

func (o *overlayRoot) FocusGained()     {}
func (o *overlayRoot) FocusLost()       {}
func (o *overlayRoot) TypedRune(r rune) {}
func (o *overlayRoot) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyEscape {
		o.surface.RequestClose()
	}
}

func (o *overlayRoot) CreateRenderer() fyne.WidgetRenderer {
	return &overlayRenderer{o: o}
}

type overlayRenderer struct{ o *overlayRoot }

func (r *overlayRenderer) panelRect(size fyne.Size) (fyne.Position, fyne.Size) {
	o := r.o
	ms := o.panel.MinSize()
	switch o.place {
	case placeRight:
		w := o.width
		if max := size.Width - 48; w > max {
			w = max
		}
		// Slides in from the right edge as progress goes 0 → 1.
		x := size.Width - w + (1-o.progress)*(w+24)
		return fyne.NewPos(x, 0), fyne.NewSize(w, size.Height)
	case placeAt:
		x, y := o.at.X, o.at.Y
		if x+ms.Width > size.Width-6 {
			x = size.Width - ms.Width - 6
		}
		if y+ms.Height > size.Height-6 {
			y = o.at.Y - ms.Height - 40 // flip above the anchor
		}
		return fyne.NewPos(x, y), ms
	default:
		w := fyne.Max(ms.Width, o.width)
		if max := size.Width - 48; w > max {
			w = max
		}
		h := ms.Height
		if max := size.Height - 64; h > max {
			h = max
		}
		// Rises a few pixels into place as it appears.
		y := (size.Height-h)/2 + (1-o.progress)*14
		return fyne.NewPos((size.Width-w)/2, y), fyne.NewSize(w, h)
	}
}

func (r *overlayRenderer) Layout(size fyne.Size) {
	o := r.o
	o.backdrop.Resize(size)
	o.layer.Resize(size)
	pos, ps := r.panelRect(size)
	o.panel.Move(pos)
	o.panel.Resize(ps)
}

func (r *overlayRenderer) MinSize() fyne.Size { return fyne.Size{} }

func (r *overlayRenderer) Refresh() {
	o := r.o
	if o.dim {
		o.backdrop.FillColor = color.NRGBA{A: uint8(120 * o.progress)}
	} else {
		o.backdrop.FillColor = color.Transparent
	}
	o.backdrop.Refresh()
	r.Layout(o.Size())
}

func (r *overlayRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.o.backdrop, r.o.panel, r.o.layer.Container}
}
func (r *overlayRenderer) Destroy() {}

// open adds the root to win's overlay stack and plays the enter animation.
func open(win fyne.Window, o *overlayRoot) *Surface {
	s := &Surface{root: o}
	o.surface = s
	o.canvas = win.Canvas()
	o.backdrop = canvas.NewRectangle(color.Transparent)
	o.layer = &Layer{Container: container.NewWithoutLayout()}
	o.ExtendBaseWidget(o)

	pushLayer(o.canvas, o.layer)
	o.canvas.Overlays().Add(o)
	o.Resize(o.canvas.Size())
	o.canvas.Focus(o)

	if ReduceMotion() || o.place == placeAt {
		o.progress = 1
		o.Refresh()
		return s
	}
	anim := fyne.NewAnimation(wgtheme.MotionSlow, func(t float32) {
		o.progress = t
		o.Refresh()
	})
	anim.Curve = fyne.AnimationEaseOut
	anim.Start()
	return s
}

// ------------------------------------------------------------------- frame

// frame is the shared chrome of modals and sheets: a title bar with a close
// button, a scrolling body and a footer of actions.
// bodyMode is how a frame's body handles content taller than the frame.
type bodyMode int

const (
	bodyFit    bodyMode = iota // modal: size to content, scroll when capped
	bodyScroll                 // sheet: fill the height, scroll the content
	bodyFill                   // sheet: fill the height, content scrolls itself
)

func frame(title string, body fyne.CanvasObject, footer []fyne.CanvasObject, close func(), radius float32, mode bodyMode) fyne.CanvasObject {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(p.Surface)
	bg.CornerRadius = radius
	bg.StrokeColor = p.BorderStrong
	bg.StrokeWidth = 1

	closeBtn := NewIconButton(wgtheme.IconClose, "Close", close)
	closeBtn.Scale = ScaleSM
	head := container.New(layout.NewCustomPaddedLayout(12, 10, 18, 12),
		container.NewBorder(nil, nil, nil, closeBtn,
			container.NewVBox(layout.NewSpacer(),
				newText(title, wgtheme.TextLG, wgtheme.WeightSemiBold, p.TextPrimary),
				layout.NewSpacer())))

	padded := container.New(layout.NewCustomPaddedLayout(16, 16, 18, 18), body)
	var scroller fyne.CanvasObject
	switch mode {
	case bodyScroll:
		scroller = container.NewVScroll(padded)
	case bodyFill:
		scroller = padded
	default:
		// A modal sizes to its content, but a long form must still scroll
		// once the window is too short for it.
		scroller = newFitScroll(padded)
	}

	var foot fyne.CanvasObject
	if len(footer) > 0 {
		row := append([]fyne.CanvasObject{layout.NewSpacer()}, footer...)
		foot = container.NewVBox(Hairline(),
			container.New(layout.NewCustomPaddedLayout(12, 12, 18, 18),
				container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD), row...)))
	}

	content := container.NewBorder(container.NewVBox(head, Hairline()), foot, nil, nil, scroller)
	return container.NewStack(shadowFor(3, radius), bg, content)
}

// panelGuard swallows taps on the panel's empty areas so they never fall
// through to the root and close it.
type panelGuard struct {
	widget.BaseWidget
	content fyne.CanvasObject
}

func (g *panelGuard) Tapped(*fyne.PointEvent) {}
func (g *panelGuard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(g.content)
}

func guard(obj fyne.CanvasObject) fyne.CanvasObject {
	g := &panelGuard{content: obj}
	g.ExtendBaseWidget(g)
	return g
}

// ShowModal opens a centred dialog. width is a minimum; the dialog grows to
// fit its content up to the window size. Footer buttons sit right-aligned.
func ShowModal(win fyne.Window, title string, width float32, body fyne.CanvasObject, footer ...fyne.CanvasObject) *Surface {
	o := &overlayRoot{place: placeCenter, width: width, dim: true, dismissable: true}
	var s *Surface
	o.panel = guard(frame(title, body, footer, func() { s.RequestClose() }, wgtheme.RadiusXL, bodyFit))
	s = open(win, o)
	return s
}

// SheetWidth is a slide-over sheet's width class.
type SheetWidth float32

const (
	SheetNarrow SheetWidth = 440
	SheetWide   SheetWidth = 680
	SheetFull   SheetWidth = 980
)

// ShowSheet opens a full-height panel that slides in from the right. Its body
// scrolls. Unlike a modal it does not close on a backdrop click, because
// sheets hold forms and a stray click must not discard an edit.
func ShowSheet(win fyne.Window, title string, width SheetWidth, body fyne.CanvasObject, footer ...fyne.CanvasObject) *Surface {
	return showSheet(win, title, width, body, bodyScroll, footer)
}

// ShowSheetFill is ShowSheet for content that fills the height and scrolls
// itself, such as an editor or a table.
func ShowSheetFill(win fyne.Window, title string, width SheetWidth, body fyne.CanvasObject, footer ...fyne.CanvasObject) *Surface {
	return showSheet(win, title, width, body, bodyFill, footer)
}

func showSheet(win fyne.Window, title string, width SheetWidth, body fyne.CanvasObject, mode bodyMode, footer []fyne.CanvasObject) *Surface {
	o := &overlayRoot{place: placeRight, width: float32(width), dim: true}
	var s *Surface
	o.panel = guard(frame(title, body, footer, func() { s.RequestClose() }, 0, mode))
	s = open(win, o)
	return s
}

// ------------------------------------------------------------------- menus

// MenuItem is one entry of a popover menu. A zero item with Separator set
// draws a divider.
type MenuItem struct {
	Label     string
	Icon      *wgtheme.Icon
	Danger    bool
	Disabled  bool
	Separator bool
	Action    func()
}

// ShowMenu opens a popover menu under anchor, right-aligned with it. A click
// anywhere else closes it.
func ShowMenu(anchor fyne.CanvasObject, items []MenuItem) *Surface {
	var s *Surface
	rows := make([]fyne.CanvasObject, 0, len(items))
	for _, it := range items {
		if it.Separator {
			rows = append(rows, container.New(layout.NewCustomPaddedLayout(4, 4, 0, 0), Hairline()))
			continue
		}
		item := it
		rows = append(rows, newMenuRow(item, func() {
			s.Close()
			if item.Action != nil {
				item.Action()
			}
		}))
	}
	list := container.New(layout.NewCustomPaddedVBoxLayout(1), rows...)
	s = ShowPopover(anchor, list, 180)
	return s
}

// ShowPopover floats arbitrary content in a raised panel under anchor,
// right-aligned with it, at least minWidth wide. A click elsewhere or Esc
// closes it.
func ShowPopover(anchor fyne.CanvasObject, content fyne.CanvasObject, minWidth float32) *Surface {
	app := fyne.CurrentApp()
	c := app.Driver().CanvasForObject(anchor)
	if c == nil {
		return nil
	}
	win := windowForCanvas(c)
	if win == nil {
		return nil
	}

	p := wgtheme.Current()
	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusLG
	bg.StrokeColor = p.BorderStrong
	bg.StrokeWidth = 1
	panel := container.NewStack(shadowFor(2, bg.CornerRadius), bg,
		container.New(layout.NewCustomPaddedLayout(5, 5, 5, 5), content))

	ms := panel.MinSize()
	w := fyne.Max(ms.Width, minWidth)
	pos := app.Driver().AbsolutePositionForObject(anchor)
	at := fyne.NewPos(pos.X+anchor.Size().Width-w, pos.Y+anchor.Size().Height+4)
	if at.X < 6 {
		at.X = 6
	}

	o := &overlayRoot{place: placeAt, at: at, dismissable: true}
	o.panel = guard(container.New(layout.NewGridWrapLayout(fyne.NewSize(w, ms.Height)), panel))
	return open(win, o)
}

func windowForCanvas(c fyne.Canvas) fyne.Window {
	for _, w := range fyne.CurrentApp().Driver().AllWindows() {
		if w.Canvas() == c {
			return w
		}
	}
	return nil
}

type menuRow struct {
	widget.BaseWidget
	item    MenuItem
	tap     func()
	hovered bool
}

func newMenuRow(item MenuItem, tap func()) *menuRow {
	r := &menuRow{item: item, tap: tap}
	r.ExtendBaseWidget(r)
	return r
}

func (r *menuRow) Tapped(*fyne.PointEvent) {
	if !r.item.Disabled {
		r.tap()
	}
}
func (r *menuRow) MouseIn(*desktop.MouseEvent)    { r.hovered = true; r.Refresh() }
func (r *menuRow) MouseMoved(*desktop.MouseEvent) {}
func (r *menuRow) MouseOut()                      { r.hovered = false; r.Refresh() }
func (r *menuRow) Cursor() desktop.Cursor         { return desktop.PointerCursor }

func (r *menuRow) CreateRenderer() fyne.WidgetRenderer {
	p := wgtheme.Current()
	fg := p.TextPrimary
	if r.item.Danger {
		fg = p.Error
	}
	if r.item.Disabled {
		fg = p.TextDisabled
	}
	bg := canvas.NewRectangle(color.Transparent)
	bg.CornerRadius = wgtheme.RadiusMD - 2

	items := []fyne.CanvasObject{}
	if r.item.Icon != nil {
		items = append(items, container.NewCenter(iconImage(r.item.Icon, fg, 16)))
	}
	items = append(items, newText(r.item.Label, wgtheme.TextMD, wgtheme.WeightMedium, fg))
	row := container.New(layout.NewCustomPaddedLayout(7, 7, 10, 16),
		container.New(layout.NewCustomPaddedHBoxLayout(10), items...))

	return &menuRowRenderer{r: r, bg: bg, objs: []fyne.CanvasObject{bg, row}, row: row}
}

type menuRowRenderer struct {
	r    *menuRow
	bg   *canvas.Rectangle
	row  fyne.CanvasObject
	objs []fyne.CanvasObject
}

func (m *menuRowRenderer) Layout(s fyne.Size)           { m.bg.Resize(s); m.row.Resize(s) }
func (m *menuRowRenderer) MinSize() fyne.Size           { return m.row.MinSize() }
func (m *menuRowRenderer) Objects() []fyne.CanvasObject { return m.objs }
func (m *menuRowRenderer) Destroy()                     {}
func (m *menuRowRenderer) Refresh() {
	p := wgtheme.Current()
	m.bg.FillColor = color.Transparent
	if m.r.hovered && !m.r.item.Disabled {
		m.bg.FillColor = p.HoverOverlay()
		if m.r.item.Danger {
			m.bg.FillColor = p.Tint(p.Error, 0.14)
		}
	}
	m.bg.Refresh()
}

// ------------------------------------------------------------------- toasts

// ToastTone selects a toast's icon and accent.
type ToastTone int

const (
	ToastSuccess ToastTone = iota
	ToastInfo
	ToastError
)

const toastLife = 2800 * time.Millisecond

// ShowToast floats a short message in the bottom-right corner of win (or of
// the open sheet/modal) and removes it after a few seconds. Toasts stack
// upward and never block input.
func ShowToast(win fyne.Window, tone ToastTone, text string) {
	l, _ := topLayer(win.Content())
	if l == nil {
		return
	}
	p := wgtheme.Current()
	c, icon := p.Success, wgtheme.IconSuccess
	switch tone {
	case ToastInfo:
		c, icon = p.Info, wgtheme.IconInfo
	case ToastError:
		c, icon = p.Error, wgtheme.IconError
	}

	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusLG
	bg.StrokeColor = p.BorderStrong
	bg.StrokeWidth = 1

	label := newText(ellipsize(text, 72), wgtheme.TextMD, wgtheme.WeightMedium, p.TextPrimary)
	row := container.New(layout.NewCustomPaddedHBoxLayout(9),
		container.NewCenter(iconImage(icon, c, 18)), container.NewCenter(label))
	toast := container.NewStack(shadowFor(2, bg.CornerRadius), bg,
		container.New(layout.NewCustomPaddedLayout(9, 9, 12, 16), row))

	origin := fyne.CurrentApp().Driver().AbsolutePositionForObject(l.Container)
	ms := toast.MinSize()
	b := l.bounds()

	// Stack above any toasts already showing.
	y := b.Height - ms.Height - 18
	for _, o := range l.Objects {
		if o.Position().Y-ms.Height-8 < y && o.Position().X > b.Width/2 {
			y = o.Position().Y - ms.Height - 8
		}
	}
	l.place(toast, fyne.NewPos(origin.X+b.Width-ms.Width-18, origin.Y+y))

	time.AfterFunc(toastLife, func() {
		fyne.Do(func() { l.Remove(toast) })
	})
}

func ellipsize(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// fitScroll is a vertical scroller whose minimum size is its content's, so a
// modal can size itself to the content and only scroll when the window caps
// its height. container.Scroll alone reports a tiny minimum, which would
// collapse the modal.
type fitScroll struct {
	widget.BaseWidget
	content fyne.CanvasObject
	scroll  *container.Scroll
}

func newFitScroll(content fyne.CanvasObject) *fitScroll {
	f := &fitScroll{content: content, scroll: container.NewVScroll(content)}
	f.ExtendBaseWidget(f)
	return f
}

func (f *fitScroll) MinSize() fyne.Size {
	ms := f.content.MinSize()
	// Leave room for the scrollbar so content never sits under it.
	return fyne.NewSize(ms.Width+4, ms.Height)
}

func (f *fitScroll) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(f.scroll)
}

// ShowModalSized opens a modal with a fixed body size, for content with no
// natural height such as lists.
func ShowModalSized(win fyne.Window, title string, size fyne.Size, body fyne.CanvasObject, footer ...fyne.CanvasObject) *Surface {
	sized := container.NewStack(sizedSpacer(size), body)
	return ShowModal(win, title, size.Width, sized, footer...)
}

func sizedSpacer(s fyne.Size) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(s)
	return r
}

// Kbd renders a key or chord as a keycap, for shortcut hints.
func Kbd(keys string) fyne.CanvasObject {
	p := wgtheme.Current()
	label := newText(keys, wgtheme.TextXS+0.5, wgtheme.WeightMono, p.TextSecondary)
	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusSM
	bg.StrokeColor = p.BorderStrong
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.New(layout.NewCustomPaddedLayout(2, 3, 6, 6), label))
}
