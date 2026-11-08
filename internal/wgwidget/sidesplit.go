package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// SideSplit is a sidebar/content split with a hairline divider that can be
// dragged to resize the sidebar. Fyne's own split paints its divider with the
// theme's shadow colour, which shows as a heavy dark band between two panes of
// the same tone.
type SideSplit struct {
	widget.BaseWidget

	Leading, Trailing fyne.CanvasObject
	Width             float32 // sidebar width
	MinWidth, MaxFrac float32 // limits: absolute minimum, fraction of the total

	handle *splitHandle
}

// NewSideSplit creates a split with the sidebar at width.
func NewSideSplit(leading, trailing fyne.CanvasObject, width float32) *SideSplit {
	s := &SideSplit{Leading: leading, Trailing: trailing, Width: width, MinWidth: 200, MaxFrac: 0.45}
	s.handle = &splitHandle{split: s}
	s.handle.ExtendBaseWidget(s.handle)
	s.ExtendBaseWidget(s)
	return s
}

func (s *SideSplit) clamp(w, total float32) float32 {
	if max := total * s.MaxFrac; w > max {
		w = max
	}
	if w < s.MinWidth {
		w = s.MinWidth
	}
	return w
}

func (s *SideSplit) CreateRenderer() fyne.WidgetRenderer {
	return &sideSplitRenderer{s: s}
}

type sideSplitRenderer struct{ s *SideSplit }

const splitGrip = 7 // hit area of the divider, centred on the hairline

func (r *sideSplitRenderer) Layout(size fyne.Size) {
	s := r.s
	w := s.clamp(s.Width, size.Width)
	s.Leading.Move(fyne.NewPos(0, 0))
	s.Leading.Resize(fyne.NewSize(w, size.Height))
	s.Trailing.Move(fyne.NewPos(w+1, 0))
	s.Trailing.Resize(fyne.NewSize(size.Width-w-1, size.Height))
	s.handle.Move(fyne.NewPos(w-splitGrip/2, 0))
	s.handle.Resize(fyne.NewSize(splitGrip+1, size.Height))
}

func (r *sideSplitRenderer) MinSize() fyne.Size {
	l, t := r.s.Leading.MinSize(), r.s.Trailing.MinSize()
	return fyne.NewSize(r.s.MinWidth+1+t.Width, fyne.Max(l.Height, t.Height))
}

func (r *sideSplitRenderer) Refresh() {
	r.Layout(r.s.Size())
	r.s.handle.Refresh()
}

func (r *sideSplitRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.s.Leading, r.s.Trailing, r.s.handle}
}
func (r *sideSplitRenderer) Destroy() {}

// splitHandle is the draggable divider.
type splitHandle struct {
	widget.BaseWidget
	split   *SideSplit
	hovered bool
	drag    bool
}

var (
	_ fyne.Draggable     = (*splitHandle)(nil)
	_ desktop.Hoverable  = (*splitHandle)(nil)
	_ desktop.Cursorable = (*splitHandle)(nil)
)

func (h *splitHandle) Dragged(ev *fyne.DragEvent) {
	h.drag = true
	s := h.split
	s.Width = s.clamp(s.Width+ev.Dragged.DX, s.Size().Width)
	s.Refresh()
}
func (h *splitHandle) DragEnd()                       { h.drag = false; h.Refresh() }
func (h *splitHandle) MouseIn(*desktop.MouseEvent)    { h.hovered = true; h.Refresh() }
func (h *splitHandle) MouseMoved(*desktop.MouseEvent) {}
func (h *splitHandle) MouseOut()                      { h.hovered = false; h.Refresh() }
func (h *splitHandle) Cursor() desktop.Cursor         { return desktop.HResizeCursor }

func (h *splitHandle) CreateRenderer() fyne.WidgetRenderer {
	line := canvas.NewRectangle(color.Transparent)
	r := &handleRenderer{h: h, line: line}
	r.Refresh()
	return r
}

type handleRenderer struct {
	h    *splitHandle
	line *canvas.Rectangle
}

func (r *handleRenderer) Layout(size fyne.Size) {
	w := float32(1)
	if r.h.hovered || r.h.drag {
		w = 2
	}
	r.line.Move(fyne.NewPos(splitGrip/2-w/2+0.5, 0))
	r.line.Resize(fyne.NewSize(w, size.Height))
}
func (r *handleRenderer) MinSize() fyne.Size { return fyne.NewSize(splitGrip, 0) }
func (r *handleRenderer) Refresh() {
	p := wgtheme.Current()
	r.line.FillColor = p.Border
	if r.h.hovered || r.h.drag {
		r.line.FillColor = p.Accent
	}
	r.Layout(r.h.Size())
	r.line.Refresh()
}
func (r *handleRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.line} }
func (r *handleRenderer) Destroy()                     {}
