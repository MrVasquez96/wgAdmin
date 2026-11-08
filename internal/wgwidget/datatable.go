package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// Column describes one DataTable column. Widths are MinWidth plus a share of
// the remaining space proportional to Weight, so the table always fills its
// pane instead of leaving a gap or forcing a horizontal scrollbar.
type Column struct {
	Title    string
	MinWidth float32
	Weight   float32
	Mono     bool
	Sortable bool
	// Dot makes the column a status-dot column; Cell.Dot supplies the tone.
	Dot bool
}

// Cell is the content of one table cell.
type Cell struct {
	Text   string
	Strong bool        // semibold, primary tone: the row's identity
	Dim    bool        // secondary data and placeholders
	Dot    color.NRGBA // for Dot columns
	Live   bool        // for Dot columns: pulse
}

// RowAction is an icon button shown at the end of every row.
type RowAction struct {
	Icon    *wgtheme.Icon
	Tooltip string
	OnTap   func(row int)
}

// DataTable is the application's table: sortable header, hoverable rows,
// clickable cells and per-row icon actions.
type DataTable struct {
	widget.BaseWidget

	Columns []Column
	Actions []RowAction
	Length  func() int
	CellAt  func(row, col int) Cell

	// OnCellTap fires for a click on a data cell. Typically: copy its value.
	OnCellTap func(row, col int)
	// OnSort fires when a sortable header is clicked. The owner reorders its
	// data and calls Refresh.
	OnSort func(col int, ascending bool)

	SortCol int
	SortAsc bool

	list   *widget.List
	header *tableHeader
}

const (
	tableRowH    = 36
	tableHeaderH = 34
	tablePadX    = 12
	actionSize   = 30
)

// NewDataTable creates a table. SortCol starts at -1 (unsorted).
func NewDataTable(cols []Column, length func() int, cell func(row, col int) Cell) *DataTable {
	t := &DataTable{Columns: cols, Length: length, CellAt: cell, SortCol: -1, SortAsc: true}
	t.ExtendBaseWidget(t)
	return t
}

// Height returns the natural height for n rows, capped at maxRows. Fyne's
// list spaces items by the theme padding, so each row takes its own height
// plus that gap.
func (t *DataTable) Height(n, maxRows int) float32 {
	if n > maxRows {
		n = maxRows
	}
	if n < 1 {
		n = 1
	}
	stride := float32(tableRowH) + theme.Padding()
	return tableHeaderH + 1 + float32(n)*stride + 2*theme.Padding()
}

func (t *DataTable) actionsWidth() float32 {
	if len(t.Actions) == 0 {
		return 0
	}
	return float32(len(t.Actions))*(actionSize+2) + tablePadX
}

// columnX returns the left edge and width of every column at total width w.
func (t *DataTable) columnX(w float32) (xs, ws []float32) {
	avail := w - 2*tablePadX - t.actionsWidth()
	fixed, weight := float32(0), float32(0)
	for _, c := range t.Columns {
		fixed += c.MinWidth
		weight += c.Weight
	}
	extra := avail - fixed
	if extra < 0 {
		extra = 0
	}
	x := float32(tablePadX)
	for _, c := range t.Columns {
		cw := c.MinWidth
		if weight > 0 {
			cw += extra * c.Weight / weight
		}
		xs = append(xs, x)
		ws = append(ws, cw)
		x += cw
	}
	return xs, ws
}

// Refresh redraws header and rows.
func (t *DataTable) Refresh() {
	if t.list != nil {
		t.list.Refresh()
	}
	if t.header != nil {
		t.header.Refresh()
	}
	t.BaseWidget.Refresh()
}

func (t *DataTable) CreateRenderer() fyne.WidgetRenderer {
	p := wgtheme.Current()
	t.header = &tableHeader{t: t, hover: -1}
	t.header.ExtendBaseWidget(t.header)

	t.list = widget.NewList(
		func() int { return t.Length() },
		func() fyne.CanvasObject { return newTableRow(t) },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if r, ok := obj.(*tableRow); ok {
				r.setRow(id)
			}
		},
	)
	// Selection is meaningless here — cells act, rows do not select — so the
	// list's own highlight is cleared the moment it appears.
	t.list.OnSelected = func(id widget.ListItemID) { t.list.UnselectAll() }
	t.list.HideSeparators = true

	bg := canvas.NewRectangle(p.Surface)
	bg.CornerRadius = wgtheme.RadiusLG
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1

	body := container.NewBorder(container.NewVBox(t.header, Hairline()), nil, nil, nil, t.list)
	return widget.NewSimpleRenderer(container.NewStack(bg,
		container.New(&insetLayout{1}, body)))
}

// insetLayout keeps content inside a bordered background.
type insetLayout struct{ px float32 }

func (l *insetLayout) Layout(objs []fyne.CanvasObject, s fyne.Size) {
	for _, o := range objs {
		o.Move(fyne.NewPos(l.px, l.px))
		o.Resize(fyne.NewSize(s.Width-2*l.px, s.Height-2*l.px))
	}
}
func (l *insetLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	var m fyne.Size
	for _, o := range objs {
		m = m.Max(o.MinSize())
	}
	return m.Add(fyne.NewSize(2*l.px, 2*l.px))
}

// ---------------------------------------------------------------- header

type tableHeader struct {
	widget.BaseWidget
	t     *DataTable
	hover int
}

var (
	_ fyne.Tappable      = (*tableHeader)(nil)
	_ desktop.Hoverable  = (*tableHeader)(nil)
	_ desktop.Cursorable = (*tableHeader)(nil)
)

func (h *tableHeader) colAt(x float32) int {
	xs, ws := h.t.columnX(h.Size().Width)
	for i := range xs {
		if x >= xs[i] && x < xs[i]+ws[i] {
			return i
		}
	}
	return -1
}

func (h *tableHeader) Tapped(ev *fyne.PointEvent) {
	i := h.colAt(ev.Position.X)
	if i < 0 || !h.t.Columns[i].Sortable {
		return
	}
	if h.t.SortCol == i {
		h.t.SortAsc = !h.t.SortAsc
	} else {
		h.t.SortCol, h.t.SortAsc = i, true
	}
	if h.t.OnSort != nil {
		h.t.OnSort(i, h.t.SortAsc)
	}
	h.t.Refresh()
}

func (h *tableHeader) MouseIn(ev *desktop.MouseEvent) { h.MouseMoved(ev) }
func (h *tableHeader) MouseMoved(ev *desktop.MouseEvent) {
	if i := h.colAt(ev.Position.X); i != h.hover {
		h.hover = i
		h.Refresh()
	}
}
func (h *tableHeader) MouseOut() { h.hover = -1; h.Refresh() }
func (h *tableHeader) Cursor() desktop.Cursor {
	if h.hover >= 0 && h.t.Columns[h.hover].Sortable {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (h *tableHeader) CreateRenderer() fyne.WidgetRenderer {
	r := &headerRenderer{h: h}
	for _, c := range h.t.Columns {
		r.titles = append(r.titles, newText(c.Title, wgtheme.TextSM, wgtheme.WeightSemiBold, color.Transparent))
		r.arrows = append(r.arrows, &canvas.Image{FillMode: canvas.ImageFillContain})
	}
	r.Refresh()
	return r
}

type headerRenderer struct {
	h      *tableHeader
	titles []*canvas.Text
	arrows []*canvas.Image
}

func (r *headerRenderer) MinSize() fyne.Size { return fyne.NewSize(100, tableHeaderH) }

func (r *headerRenderer) Layout(size fyne.Size) {
	xs, _ := r.h.t.columnX(size.Width)
	for i, t := range r.titles {
		ts := t.MinSize()
		t.Move(fyne.NewPos(xs[i], (size.Height-ts.Height)/2))
		t.Resize(ts)
		r.arrows[i].Resize(fyne.NewSize(12, 12))
		r.arrows[i].Move(fyne.NewPos(xs[i]+ts.Width+4, (size.Height-12)/2))
	}
}

func (r *headerRenderer) Refresh() {
	p := wgtheme.Current()
	t := r.h.t
	for i, title := range r.titles {
		title.Text = t.Columns[i].Title
		styleText(title, wgtheme.TextSM, wgtheme.WeightSemiBold)
		c := p.TextDim
		if i == t.SortCol || (i == r.h.hover && t.Columns[i].Sortable) {
			c = p.TextPrimary
		}
		title.Color = c
		title.Refresh()

		switch {
		case i == t.SortCol && t.SortAsc:
			setIconColor(r.arrows[i], wgtheme.IconSortAsc, p.Accent)
			r.arrows[i].Show()
		case i == t.SortCol:
			setIconColor(r.arrows[i], wgtheme.IconSortDesc, p.Accent)
			r.arrows[i].Show()
		case i == r.h.hover && t.Columns[i].Sortable:
			setIconColor(r.arrows[i], wgtheme.IconSortAsc, p.TextDim)
			r.arrows[i].Show()
		default:
			r.arrows[i].Hide()
		}
	}
	r.Layout(r.h.Size())
}

func (r *headerRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, 0, len(r.titles)*2)
	for i := range r.titles {
		objs = append(objs, r.titles[i], r.arrows[i])
	}
	return objs
}
func (r *headerRenderer) Destroy() {}

// ------------------------------------------------------------------- rows

type tableRow struct {
	widget.BaseWidget
	t        *DataTable
	row      int
	hovered  bool
	hoverCol int

	bg      *canvas.Rectangle
	cellBg  *canvas.Rectangle
	labels  []*widget.Label
	dots    []*PulseDot
	actions []*Button
}

var (
	_ fyne.Tappable      = (*tableRow)(nil)
	_ desktop.Hoverable  = (*tableRow)(nil)
	_ desktop.Cursorable = (*tableRow)(nil)
)

func newTableRow(t *DataTable) *tableRow {
	r := &tableRow{t: t, row: -1, hoverCol: -1}
	r.bg = canvas.NewRectangle(color.Transparent)
	r.bg.CornerRadius = wgtheme.RadiusSM + 2
	r.cellBg = canvas.NewRectangle(color.Transparent)
	r.cellBg.CornerRadius = wgtheme.RadiusSM
	for _, c := range t.Columns {
		l := widget.NewLabel("")
		l.Truncation = fyne.TextTruncateEllipsis
		l.TextStyle = fyne.TextStyle{Monospace: c.Mono}
		r.labels = append(r.labels, l)
		r.dots = append(r.dots, NewPulseDot(color.NRGBA{}, false))
	}
	for _, a := range t.Actions {
		act := a
		b := NewIconButton(act.Icon, act.Tooltip, nil)
		b.Scale = ScaleSM
		b.OnTapped = func() {
			if r.row >= 0 && act.OnTap != nil {
				act.OnTap(r.row)
			}
		}
		r.actions = append(r.actions, b)
	}
	r.ExtendBaseWidget(r)
	return r
}

func (r *tableRow) setRow(i int) {
	r.row = i
	r.Refresh()
}

func (r *tableRow) colAt(x float32) int {
	xs, ws := r.t.columnX(r.Size().Width)
	for i := range xs {
		if x >= xs[i] && x < xs[i]+ws[i] {
			return i
		}
	}
	return -1
}

func (r *tableRow) tappable(col int) bool {
	return col >= 0 && r.t.OnCellTap != nil && !r.t.Columns[col].Dot
}

func (r *tableRow) Tapped(ev *fyne.PointEvent) {
	if col := r.colAt(ev.Position.X); r.tappable(col) && r.row >= 0 {
		r.t.OnCellTap(r.row, col)
	}
}

func (r *tableRow) MouseIn(ev *desktop.MouseEvent) { r.hovered = true; r.MouseMoved(ev) }
func (r *tableRow) MouseMoved(ev *desktop.MouseEvent) {
	c := r.colAt(ev.Position.X)
	if c != r.hoverCol || !r.hovered {
		r.hovered, r.hoverCol = true, c
		r.Refresh()
	}
}
func (r *tableRow) MouseOut() { r.hovered, r.hoverCol = false, -1; r.Refresh() }
func (r *tableRow) Cursor() desktop.Cursor {
	if r.tappable(r.hoverCol) {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

func (r *tableRow) MinSize() fyne.Size { return fyne.NewSize(100, tableRowH) }

func (r *tableRow) CreateRenderer() fyne.WidgetRenderer {
	objs := []fyne.CanvasObject{r.bg, r.cellBg}
	for i := range r.labels {
		objs = append(objs, r.dots[i], r.labels[i])
	}
	for _, a := range r.actions {
		objs = append(objs, a)
	}
	return &rowRenderer{r: r, objs: objs}
}

type rowRenderer struct {
	r    *tableRow
	objs []fyne.CanvasObject
}

func (rr *rowRenderer) MinSize() fyne.Size           { return rr.r.MinSize() }
func (rr *rowRenderer) Objects() []fyne.CanvasObject { return rr.objs }
func (rr *rowRenderer) Destroy()                     {}

func (rr *rowRenderer) Layout(size fyne.Size) {
	r := rr.r
	r.bg.Move(fyne.NewPos(4, 2))
	r.bg.Resize(fyne.NewSize(size.Width-8, size.Height-4))

	xs, ws := r.t.columnX(size.Width)
	const inner = 6 // widget.Label's own padding, pulled back to align with headers
	for i := range r.labels {
		lh := r.labels[i].MinSize().Height
		r.labels[i].Move(fyne.NewPos(xs[i]-inner, (size.Height-lh)/2))
		r.labels[i].Resize(fyne.NewSize(ws[i]-4+inner, lh))
		ds := r.dots[i].MinSize()
		r.dots[i].Move(fyne.NewPos(xs[i], (size.Height-ds.Height)/2))
		r.dots[i].Resize(ds)
	}
	if c := r.hoverCol; r.tappable(c) {
		r.cellBg.Move(fyne.NewPos(xs[c]-6, 7))
		r.cellBg.Resize(fyne.NewSize(ws[c], size.Height-14))
	}

	x := size.Width - tablePadX - float32(len(r.actions))*(actionSize+2)
	for _, a := range r.actions {
		a.Resize(fyne.NewSize(actionSize, actionSize))
		a.Move(fyne.NewPos(x, (size.Height-actionSize)/2))
		x += actionSize + 2
	}
}

func (rr *rowRenderer) Refresh() {
	r := rr.r
	p := wgtheme.Current()

	r.bg.FillColor = color.Transparent
	if r.hovered {
		r.bg.FillColor = p.HoverOverlay()
	}
	r.cellBg.FillColor = color.Transparent
	if r.hovered && r.tappable(r.hoverCol) {
		r.cellBg.FillColor = p.HoverOverlay()
	}

	n := 0
	if r.t.Length != nil {
		n = r.t.Length()
	}
	for i, col := range r.t.Columns {
		if r.row < 0 || r.row >= n {
			r.labels[i].SetText("")
			r.dots[i].Hide()
			continue
		}
		c := r.t.CellAt(r.row, i)
		if col.Dot {
			r.labels[i].Hide()
			r.dots[i].Show()
			r.dots[i].Set(c.Dot, c.Live)
			continue
		}
		r.dots[i].Hide()
		r.labels[i].Show()
		l := r.labels[i]
		l.TextStyle = fyne.TextStyle{Monospace: col.Mono, Bold: c.Strong && !col.Mono}
		l.Importance = widget.MediumImportance
		if c.Dim {
			l.Importance = widget.LowImportance
		}
		l.SetText(c.Text)
	}

	rr.Layout(r.Size())
	r.bg.Refresh()
	r.cellBg.Refresh()
}
