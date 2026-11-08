package wgwidget

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// TrafficChart plots receive and transmit rates over the recent history as
// two soft gradient areas with crisp lines, a legend and a peak marker. It is
// drawn with a raster rather than vector primitives because Fyne has no path
// fill, and per-pixel drawing gives anti-aliased lines and true gradients.
type TrafficChart struct {
	widget.BaseWidget

	RXLabel, TXLabel string
	Format           func(bytesPerSec float64) string

	rx, tx []float64
	height float32
}

// NewTrafficChart creates a chart of the given plot height.
func NewTrafficChart(height float32, rxLabel, txLabel string, format func(float64) string) *TrafficChart {
	c := &TrafficChart{height: height, RXLabel: rxLabel, TXLabel: txLabel, Format: format}
	c.ExtendBaseWidget(c)
	return c
}

// SetSeries replaces the data. The two series are aligned on their newest
// sample; the oldest is on the left.
func (c *TrafficChart) SetSeries(rx, tx []float64) {
	c.rx, c.tx = rx, tx
	c.Refresh()
}

func (c *TrafficChart) peak() float64 {
	m := 0.0
	for _, v := range c.rx {
		m = math.Max(m, v)
	}
	for _, v := range c.tx {
		m = math.Max(m, v)
	}
	return m
}

func (c *TrafficChart) CreateRenderer() fyne.WidgetRenderer {
	p := wgtheme.Current()
	r := &chartRenderer{c: c}
	r.raster = canvas.NewRaster(r.draw)
	r.raster.ScaleMode = canvas.ImageScaleSmooth

	legendItem := func(tone color.NRGBA, label string) fyne.CanvasObject {
		dot := canvas.NewCircle(tone)
		return container.New(layout.NewCustomPaddedHBoxLayout(6),
			container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(8, 8)), dot)),
			newText(label, wgtheme.TextSM, wgtheme.WeightMedium, p.TextSecondary))
	}
	r.peakText = newText("", wgtheme.TextXS, wgtheme.WeightMono, p.TextDim)
	r.legend = container.NewBorder(nil, nil,
		container.New(layout.NewCustomPaddedHBoxLayout(14),
			legendItem(p.Accent, c.RXLabel), legendItem(p.Info, c.TXLabel)),
		r.peakText)
	r.Refresh()
	return r
}

type chartRenderer struct {
	c        *TrafficChart
	raster   *canvas.Raster
	legend   fyne.CanvasObject
	peakText *canvas.Text
}

func (r *chartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, r.legend.MinSize().Height+8+r.c.height)
}

func (r *chartRenderer) Layout(size fyne.Size) {
	lh := r.legend.MinSize().Height
	r.legend.Resize(fyne.NewSize(size.Width, lh))
	r.raster.Move(fyne.NewPos(0, lh+8))
	r.raster.Resize(fyne.NewSize(size.Width, size.Height-lh-8))
}

func (r *chartRenderer) Refresh() {
	if r.c.Format != nil && r.c.peak() > 0 {
		r.peakText.Text = "peak " + r.c.Format(r.c.peak())
	} else {
		r.peakText.Text = ""
	}
	r.peakText.Refresh()
	r.raster.Refresh()
}

func (r *chartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.legend, r.raster}
}
func (r *chartRenderer) Destroy() {}

// sampleAt interpolates a series at horizontal fraction f (0 = oldest).
func sampleAt(s []float64, f float64) float64 {
	if len(s) == 0 {
		return 0
	}
	if len(s) == 1 {
		return s[0]
	}
	x := f * float64(len(s)-1)
	i := int(x)
	if i >= len(s)-1 {
		return s[len(s)-1]
	}
	t := x - float64(i)
	// Smoothstep between samples keeps the curve soft without overshoot.
	t = t * t * (3 - 2*t)
	return s[i]*(1-t) + s[i+1]*t
}

func (r *chartRenderer) draw(w, h int) image.Image {
	p := wgtheme.Current()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if w < 2 || h < 2 {
		return img
	}

	// Faint horizontal grid: quarters of the plot.
	grid := wgtheme.Alpha(p.Border, 140)
	for q := 1; q < 4; q++ {
		y := h * q / 4
		for x := 0; x < w; x += 2 {
			img.SetNRGBA(x, y, grid)
		}
	}
	for x := 0; x < w; x++ {
		img.SetNRGBA(x, h-1, p.Border)
	}

	peak := r.c.peak()
	if peak <= 0 {
		return img
	}
	top := float64(h) * 0.08 // headroom so the peak is not clipped
	scale := (float64(h-1) - top) / peak

	plot := func(series []float64, tone color.NRGBA) {
		if len(series) < 2 {
			return
		}
		ys := make([]float64, w)
		for x := 0; x < w; x++ {
			ys[x] = float64(h-1) - sampleAt(series, float64(x)/float64(w-1))*scale
		}
		for x := 0; x < w; x++ {
			y0 := ys[x]
			// Gradient area: strongest just under the line, fading to nothing.
			for y := int(math.Ceil(y0)); y < h; y++ {
				t := float64(y-int(y0)) / math.Max(float64(h)-y0, 1)
				a := uint8(70 * (1 - t))
				blend(img, x, y, tone, a)
			}
			// Line: vertical span between neighbours, anti-aliased by distance.
			lo, hi := y0, y0
			if x > 0 {
				lo, hi = math.Min(lo, ys[x-1]), math.Max(hi, ys[x-1])
			}
			for y := int(lo) - 2; y <= int(hi)+2; y++ {
				if y < 0 || y >= h {
					continue
				}
				d := 0.0
				fy := float64(y)
				if fy < lo {
					d = lo - fy
				} else if fy > hi {
					d = fy - hi
				}
				cov := 1.75/2 + 0.5 - d
				if cov > 0 {
					blend(img, x, y, tone, uint8(255*math.Min(cov, 1)))
				}
			}
		}
	}
	plot(r.c.tx, p.Info)
	plot(r.c.rx, p.Accent)
	return img
}

// blend composites tone at alpha a over the pixel at x, y.
func blend(img *image.NRGBA, x, y int, tone color.NRGBA, a uint8) {
	dst := img.NRGBAAt(x, y)
	sa := float64(a) / 255
	da := float64(dst.A) / 255
	oa := sa + da*(1-sa)
	if oa == 0 {
		return
	}
	mix := func(s, d uint8) uint8 {
		return uint8((float64(s)*sa + float64(d)*da*(1-sa)) / oa)
	}
	img.SetNRGBA(x, y, color.NRGBA{
		R: mix(tone.R, dst.R), G: mix(tone.G, dst.G), B: mix(tone.B, dst.B), A: uint8(oa * 255),
	})
}
