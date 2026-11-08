package wgwidget

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

func TestButtonTapAndDisabled(t *testing.T) {
	test.NewTempApp(t)
	taps := 0
	b := NewButton("Go", wgtheme.IconPlay, Primary, func() { taps++ })
	test.Tap(b)
	if taps != 1 {
		t.Fatalf("taps = %d, want 1", taps)
	}
	b.Disable()
	test.Tap(b)
	if taps != 1 {
		t.Errorf("disabled button fired; taps = %d", taps)
	}
}

func TestButtonKeyboardActivation(t *testing.T) {
	test.NewTempApp(t)
	taps := 0
	b := NewButton("Go", nil, Secondary, func() { taps++ })
	b.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	b.TypedRune(' ')
	b.TypedRune('x')
	if taps != 2 {
		t.Errorf("Enter + Space should activate, other runes not; taps = %d", taps)
	}
}

func TestIconButtonIsSquare(t *testing.T) {
	test.NewTempApp(t)
	b := NewIconButton(wgtheme.IconRefresh, "Refresh", nil)
	s := b.MinSize()
	if s.Width != s.Height || s.Width == 0 {
		t.Errorf("icon button min size %v, want a non-empty square", s)
	}
}

func TestSwitchToggles(t *testing.T) {
	test.NewTempApp(t)
	var got []bool
	s := NewSwitch(false, func(on bool) { got = append(got, on) })
	w := test.NewWindow(s)
	defer w.Close()
	test.Tap(s)
	test.Tap(s)
	if len(got) != 2 || !got[0] || got[1] {
		t.Errorf("toggle sequence = %v, want [true false]", got)
	}
	s.SetOn(true)
	if len(got) != 2 {
		t.Error("SetOn must not call OnChanged")
	}
	s.Disable()
	test.Tap(s)
	if len(got) != 2 {
		t.Error("disabled switch toggled")
	}
}

func TestSegmentedSelectsByPosition(t *testing.T) {
	test.NewTempApp(t)
	picked := -1
	s := NewSegmented([]string{"A", "B", "C"}, func(i int) { picked = i })
	w := test.NewWindow(s)
	defer w.Close()
	s.Resize(fyne.NewSize(300, 30))

	test.TapAt(s, fyne.NewPos(250, 10))
	if picked != 2 || s.Selected != 2 {
		t.Errorf("tap in the last third picked %d (selected %d), want 2", picked, s.Selected)
	}
	picked = -1
	test.TapAt(s, fyne.NewPos(260, 10))
	if picked != -1 {
		t.Error("tapping the selected segment again must not fire OnChanged")
	}
}

func TestCopyField(t *testing.T) {
	test.NewTempApp(t)
	var copied string
	f := NewCopyField(nil, "Endpoint", "example.com:51820", func(v string) { copied = v })
	w := test.NewWindow(f)
	defer w.Close()
	test.Tap(f)
	if copied != "example.com:51820" {
		t.Errorf("copied %q", copied)
	}

	ro := NewCopyField(nil, "Port", "auto", nil)
	test.Tap(ro) // must not panic
	if ro.copyable() {
		t.Error("a field without OnCopy must be read-only")
	}
}

func TestDataTableColumnsFillWidth(t *testing.T) {
	test.NewTempApp(t)
	tbl := NewDataTable([]Column{
		{MinWidth: 20},
		{MinWidth: 100, Weight: 1},
		{MinWidth: 100, Weight: 3},
	}, func() int { return 0 }, func(int, int) Cell { return Cell{} })
	tbl.Actions = []RowAction{{Icon: wgtheme.IconQR}}

	const w = 1000
	xs, ws := tbl.columnX(w)
	end := xs[len(xs)-1] + ws[len(ws)-1]
	want := float32(w) - tablePadX - tbl.actionsWidth()
	if diff := end - want; diff > 0.5 || diff < -0.5 {
		t.Errorf("columns end at %.1f, want %.1f (fill to the actions column)", end, want)
	}
	// Extra space is shared 1:3 between the weighted columns.
	if r := (ws[2] - 100) / (ws[1] - 100); r < 2.9 || r > 3.1 {
		t.Errorf("weighted share ratio %.2f, want 3", r)
	}
	if ws[0] != 20 {
		t.Errorf("unweighted column width %.1f, want its MinWidth 20", ws[0])
	}
}

func TestDataTableCellTap(t *testing.T) {
	test.NewTempApp(t)
	rows := [][]string{{"a", "b"}, {"c", "d"}}
	tbl := NewDataTable([]Column{{Title: "X", MinWidth: 100}, {Title: "Y", MinWidth: 100}},
		func() int { return len(rows) },
		func(r, c int) Cell { return Cell{Text: rows[r][c]} })
	var gotR, gotC = -1, -1
	tbl.OnCellTap = func(r, c int) { gotR, gotC = r, c }

	row := newTableRow(tbl)
	w := test.NewWindow(row)
	defer w.Close()
	row.Resize(fyne.NewSize(400, tableRowH))
	row.setRow(1)

	test.TapAt(row, fyne.NewPos(tablePadX+150, 10))
	if gotR != 1 || gotC != 1 {
		t.Errorf("tap hit row %d col %d, want row 1 col 1", gotR, gotC)
	}
}

func TestOverlayOpenAndClose(t *testing.T) {
	test.NewTempApp(t)
	w := test.NewWindow(widget.NewLabel("content"))
	defer w.Close()
	w.Resize(fyne.NewSize(800, 600))

	closed := 0
	s := ShowModal(w, "Title", 300, widget.NewLabel("body"))
	s.OnClose = func() { closed++ }
	if n := len(w.Canvas().Overlays().List()); n != 1 {
		t.Fatalf("overlays = %d after ShowModal, want 1", n)
	}

	// Esc closes, exactly once.
	s.root.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEscape})
	s.Close()
	if n := len(w.Canvas().Overlays().List()); n != 0 {
		t.Errorf("overlays = %d after Esc, want 0", n)
	}
	if closed != 1 {
		t.Errorf("OnClose ran %d times, want 1", closed)
	}
}

func TestSheetIgnoresBackdropTap(t *testing.T) {
	test.NewTempApp(t)
	w := test.NewWindow(widget.NewLabel("content"))
	defer w.Close()
	w.Resize(fyne.NewSize(800, 600))

	s := ShowSheet(w, "Edit", SheetNarrow, widget.NewLabel("form"))
	s.root.Tapped(&fyne.PointEvent{})
	if len(w.Canvas().Overlays().List()) != 1 {
		t.Error("a sheet holds a form; a backdrop click must not discard it")
	}
	s.Close()

	m := ShowModal(w, "Info", 300, widget.NewLabel("x"))
	m.root.Tapped(&fyne.PointEvent{})
	if len(w.Canvas().Overlays().List()) != 0 {
		t.Error("a modal should close on a backdrop click")
	}
}
