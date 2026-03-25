// Package colorpicker provides an interactive HSV color picker dialog for Fyne v2.
//
// Usage:
//
//	colorpicker.ShowColorPickerDialog("Pick Color", currentColor, func(c color.Color) {
//	    myColor = c
//	}, parentWindow)
package wgwidget

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ─────────────────────────────────────────────
// HSV / colour helpers
// ─────────────────────────────────────────────

// HSVToRGB converts hue [0,360), saturation [0,1], value [0,1] → RGBA.
func HSVToRGB(h, s, v float64) color.RGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	s = clamp01(s)
	v = clamp01(v)

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// RGBToHSV converts a color.Color → hue [0,360), saturation [0,1], value [0,1].
func RGBToHSV(c color.Color) (h, s, v float64) {
	r32, g32, b32, _ := c.RGBA()
	rf := float64(r32) / 65535
	gf := float64(g32) / 65535
	bf := float64(b32) / 65535

	mx := math.Max(rf, math.Max(gf, bf))
	mn := math.Min(rf, math.Min(gf, bf))
	delta := mx - mn

	v = mx
	if mx == 0 {
		s = 0
	} else {
		s = delta / mx
	}

	if delta == 0 {
		h = 0
	} else if mx == rf {
		h = 60 * math.Mod((gf-bf)/delta, 6)
	} else if mx == gf {
		h = 60 * ((bf-rf)/delta + 2)
	} else {
		h = 60 * ((rf-gf)/delta + 4)
	}
	if h < 0 {
		h += 360
	}
	return
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ColorToHex converts a color.Color to a 6-character hex string (without #).
func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%02x%02x%02x", r>>8, g>>8, b>>8)
}

// HexToColor parses a 6-character hex string (with optional #) to color.RGBA.
func HexToColor(hex string) (color.RGBA, error) {
	hex = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(hex)), "#")
	if len(hex) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex length")
	}
	rv, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	gv, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	bv, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: uint8(rv), G: uint8(gv), B: uint8(bv), A: 255}, nil
}

// drawCrosshair draws a small crosshair at pixel (cx,cy) in white with black outline.
func drawCrosshair(img *image.RGBA, cx, cy, w, h int) {
	set := func(x, y int, c color.Color) {
		if x >= 0 && x < w && y >= 0 && y < h {
			img.Set(x, y, c)
		}
	}
	for i := -6; i <= 6; i++ {
		if i == 0 {
			continue
		}
		set(cx+i, cy, color.White)
		set(cx, cy+i, color.White)
		// thin black shadow
		set(cx+i, cy+1, color.Black)
		set(cx+i+1, cy, color.Black)
	}
}

// ─────────────────────────────────────────────
// svCanvas – the 2-D saturation / value picker
// ─────────────────────────────────────────────

type svCanvas struct {
	widget.BaseWidget
	hue      float64 // 0–360
	sat      float64 // X axis 0–1
	val      float64 // Y axis 0–1 (1 = top/brightest)
	onChange func(s, v float64)
	raster   *canvas.Raster
}

func newSVCanvas(hue, sat, val float64, onChange func(s, v float64)) *svCanvas {
	c := &svCanvas{hue: hue, sat: sat, val: val, onChange: onChange}
	c.raster = canvas.NewRaster(c.draw)
	c.ExtendBaseWidget(c)
	return c
}

func (c *svCanvas) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if w <= 1 || h <= 1 {
		return img
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			s := float64(x) / float64(w-1)
			v := 1 - float64(y)/float64(h-1)
			img.Set(x, y, HSVToRGB(c.hue, s, v))
		}
	}
	cx := int(c.sat * float64(w-1))
	cy := int((1 - c.val) * float64(h-1))
	drawCrosshair(img, cx, cy, w, h)
	return img
}

func (c *svCanvas) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.raster)
}

func (c *svCanvas) MinSize() fyne.Size { return fyne.NewSize(220, 220) }

func (c *svCanvas) Tapped(ev *fyne.PointEvent) { c.pick(ev.Position) }
func (c *svCanvas) Dragged(ev *fyne.DragEvent) { c.pick(ev.Position) }
func (c *svCanvas) DragEnd()                   {}

func (c *svCanvas) pick(pos fyne.Position) {
	sz := c.Size()
	s := clamp01(float64(pos.X) / float64(sz.Width))
	v := clamp01(1 - float64(pos.Y)/float64(sz.Height))
	c.sat = s
	c.val = v
	c.raster.Refresh()
	if c.onChange != nil {
		c.onChange(s, v)
	}
}

func (c *svCanvas) setHue(h float64) {
	c.hue = h
	c.raster.Refresh()
}

func (c *svCanvas) setSV(s, v float64) {
	c.sat = s
	c.val = v
	c.raster.Refresh()
}

// ─────────────────────────────────────────────
// hueBar – vertical rainbow bar
// ─────────────────────────────────────────────

type hueBar struct {
	widget.BaseWidget
	hue      float64 // 0–360
	onChange func(h float64)
	raster   *canvas.Raster
}

func newHueBar(hue float64, onChange func(h float64)) *hueBar {
	b := &hueBar{hue: hue, onChange: onChange}
	b.raster = canvas.NewRaster(b.draw)
	b.ExtendBaseWidget(b)
	return b
}

func (b *hueBar) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if h <= 1 {
		return img
	}
	for y := 0; y < h; y++ {
		hue := float64(y) / float64(h-1) * 360
		col := HSVToRGB(hue, 1, 1)
		for x := 0; x < w; x++ {
			img.Set(x, y, col)
		}
	}
	// Draw indicator line at current hue
	indicatorY := int(b.hue / 360 * float64(h-1))
	for x := 0; x < w; x++ {
		img.Set(x, indicatorY, color.White)
		if indicatorY+1 < h {
			img.Set(x, indicatorY+1, color.Black)
		}
	}
	return img
}

func (b *hueBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.raster)
}

func (b *hueBar) MinSize() fyne.Size { return fyne.NewSize(28, 220) }

func (b *hueBar) Tapped(ev *fyne.PointEvent) { b.pick(ev.Position) }
func (b *hueBar) Dragged(ev *fyne.DragEvent) { b.pick(ev.Position) }
func (b *hueBar) DragEnd()                   {}

func (b *hueBar) pick(pos fyne.Position) {
	sz := b.Size()
	h := clampF(float64(pos.Y)/float64(sz.Height)*360, 0, 359.99)
	b.hue = h
	b.raster.Refresh()
	if b.onChange != nil {
		b.onChange(h)
	}
}

func (b *hueBar) setHue(h float64) {
	b.hue = h
	b.raster.Refresh()
}

// ─────────────────────────────────────────────
// alphaBar – vertical alpha bar
// ─────────────────────────────────────────────

type alphaBar struct {
	widget.BaseWidget
	alpha    float64 // 0–1
	baseCol  color.RGBA
	onChange func(a float64)
	raster   *canvas.Raster
}

func newAlphaBar(col color.RGBA, alpha float64, onChange func(a float64)) *alphaBar {
	b := &alphaBar{alpha: alpha, baseCol: col, onChange: onChange}
	b.raster = canvas.NewRaster(b.draw)
	b.ExtendBaseWidget(b)
	return b
}

func (b *alphaBar) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	if h <= 1 {
		return img
	}
	// checkerboard background + gradient
	checker := 4
	for y := 0; y < h; y++ {
		a := 1 - float64(y)/float64(h-1)
		col := color.RGBA{
			R: b.baseCol.R,
			G: b.baseCol.G,
			B: b.baseCol.B,
			A: uint8(a * 255),
		}
		for x := 0; x < w; x++ {
			// checkerboard
			if (x/checker+y/checker)%2 == 0 {
				img.Set(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
			} else {
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
			// blend
			bg := img.RGBAAt(x, y)
			fa := float64(col.A) / 255
			img.Set(x, y, color.RGBA{
				R: uint8(float64(col.R)*fa + float64(bg.R)*(1-fa)),
				G: uint8(float64(col.G)*fa + float64(bg.G)*(1-fa)),
				B: uint8(float64(col.B)*fa + float64(bg.B)*(1-fa)),
				A: 255,
			})
		}
	}
	// indicator
	indicatorY := int((1 - b.alpha) * float64(h-1))
	for x := 0; x < w; x++ {
		img.Set(x, indicatorY, color.White)
		if indicatorY+1 < h {
			img.Set(x, indicatorY+1, color.Black)
		}
	}
	return img
}

func (b *alphaBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.raster)
}

func (b *alphaBar) MinSize() fyne.Size { return fyne.NewSize(28, 220) }

func (b *alphaBar) Tapped(ev *fyne.PointEvent) { b.pick(ev.Position) }
func (b *alphaBar) Dragged(ev *fyne.DragEvent) { b.pick(ev.Position) }
func (b *alphaBar) DragEnd()                   {}

func (b *alphaBar) pick(pos fyne.Position) {
	sz := b.Size()
	a := clamp01(1 - float64(pos.Y)/float64(sz.Height))
	b.alpha = a
	b.raster.Refresh()
	if b.onChange != nil {
		b.onChange(a)
	}
}

func (b *alphaBar) setColor(col color.RGBA) {
	b.baseCol = col
	b.raster.Refresh()
}

func (b *alphaBar) setAlpha(a float64) {
	b.alpha = a
	b.raster.Refresh()
}

// ─────────────────────────────────────────────
// colorSwatch – the preview box
// ─────────────────────────────────────────────

type colorSwatch struct {
	widget.BaseWidget
	col    color.RGBA
	rect   *canvas.Rectangle
	border *canvas.Rectangle
}

func newColorSwatch(col color.RGBA) *colorSwatch {
	s := &colorSwatch{col: col}
	s.rect = canvas.NewRectangle(col)
	s.border = canvas.NewRectangle(color.Transparent)
	s.border.StrokeColor = color.RGBA{R: 80, G: 80, B: 80, A: 255}
	s.border.StrokeWidth = 1
	s.ExtendBaseWidget(s)
	return s
}

func (s *colorSwatch) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(s.rect, s.border))
}

func (s *colorSwatch) MinSize() fyne.Size { return fyne.NewSize(60, 28) }

func (s *colorSwatch) setColor(c color.RGBA) {
	s.col = c
	s.rect.FillColor = c
	s.rect.Refresh()
}

// ─────────────────────────────────────────────
// pickerState – all mutable state
// ─────────────────────────────────────────────

type pickerState struct {
	h, s, v float64 // HSV 0–360 / 0–1 / 0–1
	r, g, b uint8
	alpha   float64 // 0–1

	svC    *svCanvas
	hueB   *hueBar
	alphaB *alphaBar

	hexEntry *widget.Entry
	preview  *colorSwatch

	rSlider, gSlider, bSlider *widget.Slider
	aSlider                   *widget.Slider
	rLabel, gLabel, bLabel    *widget.Label
	aLabel                    *widget.Label
	hLabel                    *widget.Label

	updating bool
}

func (st *pickerState) currentColor() color.RGBA {
	return color.RGBA{R: st.r, G: st.g, B: st.b, A: uint8(st.alpha * 255)}
}

func (st *pickerState) fromHSV() {
	col := HSVToRGB(st.h, st.s, st.v)
	st.r = col.R
	st.g = col.G
	st.b = col.B
	st.syncUI()
}

func (st *pickerState) fromRGB() {
	h, s, v := RGBToHSV(color.RGBA{R: st.r, G: st.g, B: st.b, A: 255})
	st.h = h
	st.s = s
	st.v = v
	st.syncUI()
}

// syncUI updates every UI element from the current h/s/v/r/g/b/alpha.
func (st *pickerState) syncUI() {
	if st.updating {
		return
	}
	st.updating = true
	defer func() { st.updating = false }()

	col := st.currentColor()

	st.preview.setColor(col)
	st.hexEntry.SetText(ColorToHex(col))

	st.rSlider.SetValue(float64(st.r))
	st.gSlider.SetValue(float64(st.g))
	st.bSlider.SetValue(float64(st.b))
	st.aSlider.SetValue(st.alpha * 255)

	st.rLabel.SetText(fmt.Sprintf("%3d", st.r))
	st.gLabel.SetText(fmt.Sprintf("%3d", st.g))
	st.bLabel.SetText(fmt.Sprintf("%3d", st.b))
	st.aLabel.SetText(fmt.Sprintf("%3d%%", int(st.alpha*100)))
	st.hLabel.SetText(fmt.Sprintf("H %.0f°  S %.0f%%  V %.0f%%", st.h, st.s*100, st.v*100))

	st.svC.setHue(st.h)
	st.svC.setSV(st.s, st.v)
	st.hueB.setHue(st.h)
	st.alphaB.setColor(color.RGBA{R: st.r, G: st.g, B: st.b, A: 255})
	st.alphaB.setAlpha(st.alpha)
}

// ─────────────────────────────────────────────
// ShowColorPickerDialog – the public API
// ─────────────────────────────────────────────

// ShowColorPickerDialog opens a modal HSV colour picker.
//
//   - title     – dialog title
//   - current   – starting colour
//   - withAlpha – show alpha / opacity control
//   - onSelected – called with the chosen colour when OK is pressed
//   - parent    – parent fyne.Window
func ShowColorPickerDialog(
	title string,
	current color.Color,
	withAlpha bool,
	onSelected func(color.Color),
	parent fyne.Window,
) {
	r32, g32, b32, a32 := current.RGBA()
	initR := uint8(r32 >> 8)
	initG := uint8(g32 >> 8)
	initB := uint8(b32 >> 8)
	initA := float64(a32>>8) / 255.0
	initH, initS, initV := RGBToHSV(current)

	st := &pickerState{
		h: initH, s: initS, v: initV,
		r: initR, g: initG, b: initB,
		alpha: initA,
	}

	// ── SV canvas ──
	st.svC = newSVCanvas(initH, initS, initV, func(s, v float64) {
		if st.updating {
			return
		}
		st.s = s
		st.v = v
		st.fromHSV()
	})

	// ── Hue bar ──
	st.hueB = newHueBar(initH, func(h float64) {
		if st.updating {
			return
		}
		st.h = h
		st.fromHSV()
	})

	// ── Alpha bar ──
	st.alphaB = newAlphaBar(
		color.RGBA{R: initR, G: initG, B: initB, A: 255},
		initA,
		func(a float64) {
			if st.updating {
				return
			}
			st.alpha = a
			st.syncUI()
		},
	)

	// ── Preview swatch ──
	st.preview = newColorSwatch(color.RGBA{R: initR, G: initG, B: initB, A: uint8(initA * 255)})

	// ── Hex entry ──
	st.hexEntry = widget.NewEntry()
	st.hexEntry.SetText(ColorToHex(current))
	st.hexEntry.OnChanged = func(txt string) {
		if st.updating {
			return
		}
		col, err := HexToColor(txt)
		if err != nil {
			return
		}
		st.r = col.R
		st.g = col.G
		st.b = col.B
		st.fromRGB()
	}

	// ── RGB sliders ──
	mkSlider := func(max float64) *widget.Slider {
		sl := widget.NewSlider(0, max)
		sl.Step = 1
		return sl
	}
	mkLabel := func(txt string) *widget.Label {
		l := widget.NewLabel(txt)
		l.TextStyle = fyne.TextStyle{Monospace: true}
		return l
	}

	st.rSlider = mkSlider(255)
	st.gSlider = mkSlider(255)
	st.bSlider = mkSlider(255)
	st.aSlider = mkSlider(255)

	st.rLabel = mkLabel(fmt.Sprintf("%3d", initR))
	st.gLabel = mkLabel(fmt.Sprintf("%3d", initG))
	st.bLabel = mkLabel(fmt.Sprintf("%3d", initB))
	st.aLabel = mkLabel(fmt.Sprintf("%3d%%", int(initA*100)))
	st.hLabel = widget.NewLabelWithStyle(
		fmt.Sprintf("H %.0f°  S %.0f%%  V %.0f%%", initH, initS*100, initV*100),
		fyne.TextAlignCenter,
		fyne.TextStyle{Monospace: true},
	)

	st.rSlider.SetValue(float64(initR))
	st.gSlider.SetValue(float64(initG))
	st.bSlider.SetValue(float64(initB))
	st.aSlider.SetValue(initA * 255)

	st.rSlider.OnChanged = func(v float64) {
		if st.updating {
			return
		}
		st.r = uint8(v)
		st.fromRGB()
	}
	st.gSlider.OnChanged = func(v float64) {
		if st.updating {
			return
		}
		st.g = uint8(v)
		st.fromRGB()
	}
	st.bSlider.OnChanged = func(v float64) {
		if st.updating {
			return
		}
		st.b = uint8(v)
		st.fromRGB()
	}
	st.aSlider.OnChanged = func(v float64) {
		if st.updating {
			return
		}
		st.alpha = v / 255
		st.syncUI()
	}

	// ── Layout ──

	// Top: SV canvas + hue bar (+ optional alpha bar)
	var topRow *fyne.Container
	if withAlpha {
		topRow = container.NewBorder(nil, nil, nil,
			container.NewHBox(st.hueB, st.alphaB),
			st.svC,
		)
	} else {
		topRow = container.NewBorder(nil, nil, nil, st.hueB, st.svC)
	}

	// Slider rows: label | slider | value
	sliderRow := func(label string, sl *widget.Slider, lbl *widget.Label) *fyne.Container {
		lw := canvas.NewText(label, theme.ForegroundColor())
		lw.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewBorder(nil, nil,
			container.NewHBox(lw), lbl,
			sl,
		)
	}

	sliders := container.NewVBox(
		sliderRow("R", st.rSlider, st.rLabel),
		sliderRow("G", st.gSlider, st.gLabel),
		sliderRow("B", st.bSlider, st.bLabel),
	)
	if withAlpha {
		sliders.Add(sliderRow("A", st.aSlider, st.aLabel))
	}

	hexRow := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Hex #", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil,
		st.hexEntry,
	)

	previewRow := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Color:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil,
		st.preview,
	)

	content := container.NewVBox(
		topRow,
		st.hLabel,
		widget.NewSeparator(),
		sliders,
		widget.NewSeparator(),
		hexRow,
		previewRow,
	)

	d := dialog.NewCustomConfirm(title, "OK", "Cancel", content,
		func(ok bool) {
			if ok && onSelected != nil {
				onSelected(st.currentColor())
			}
		}, parent)
	d.Show()
}
