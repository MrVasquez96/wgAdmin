package wgwidget

import (
	"image/color"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgstat"
)

// StatusDot is the small coloured disc that carries tunnel and peer state.
// An "up" dot gets a soft halo so live things read differently from dead ones
// even in peripheral vision, and colour is never the only signal — every dot
// is paired with a text label by its caller.
func StatusDot(tone color.NRGBA, halo bool) fyne.CanvasObject {
	dot := canvas.NewCircle(tone)
	sized := container.New(layout.NewGridWrapLayout(fyne.NewSize(8, 8)), dot)
	if !halo {
		return container.NewCenter(sized)
	}

	ring := canvas.NewCircle(color.Transparent)
	ring.StrokeColor = wgtheme.Alpha(tone, 70)
	ring.StrokeWidth = 2
	ringBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(14, 14)), ring)

	return container.NewCenter(container.NewStack(ringBox, container.NewCenter(sized)))
}

// StateTone maps a peer state onto its palette colour.
func StateTone(s wgstat.PeerState) color.NRGBA {
	p := wgtheme.Current()
	switch s {
	case wgstat.StateOnline:
		return p.Success
	case wgstat.StateIdle:
		return p.Warning
	case wgstat.StateStale:
		return p.Error
	default:
		return p.TextDim
	}
}

// PeerStateDot renders the dot for a peer's handshake state.
func PeerStateDot(s wgstat.PeerState) fyne.CanvasObject {
	return StatusDot(StateTone(s), s == wgstat.StateOnline)
}

// StatePill is the UP / DOWN badge. It combines a dot, an uppercase mono label
// and a tinted background so state is legible three ways.
func StatePill(up bool, upText, downText string) fyne.CanvasObject {
	p := wgtheme.Current()

	tone, text := p.TextDim, downText
	if up {
		tone, text = p.Success, upText
	}

	label := canvas.NewText(text, tone)
	label.TextSize = 10
	label.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	bg := canvas.NewRectangle(p.Tint(tone, 0.16))
	bg.CornerRadius = wgtheme.RadiusChip
	bg.StrokeColor = p.Tint(tone, 0.45)
	bg.StrokeWidth = 1

	row := container.New(layout.NewCustomPaddedHBoxLayout(5),
		StatusDot(tone, up),
		label,
	)
	pad := container.New(layout.NewCustomPaddedLayout(3, 3, 7, 7), row)

	return container.NewStack(bg, pad)
}

// Metric is a stat tile: a dim uppercase caption above a monospace value, with
// an optional unit in the dim tone. These line up into the metrics row of the
// detail pane.
func Metric(label, value, unit string) fyne.CanvasObject {
	return ToneMetric(label, value, unit, wgtheme.Current().TextPrimary)
}

// ToneMetric is Metric with the value painted in a specific tone, used to make
// live throughput or a peer count stand out.
func ToneMetric(label, value, unit string, tone color.NRGBA) fyne.CanvasObject {
	val := Value(value)
	val.TextSize = 16
	val.Color = tone

	line := fyne.CanvasObject(val)
	if unit != "" {
		u := canvas.NewText(unit, wgtheme.Current().TextDim)
		u.TextSize = 10
		u.TextStyle = fyne.TextStyle{Monospace: true}
		// Bottom-align the unit against the value's baseline.
		line = container.New(layout.NewCustomPaddedHBoxLayout(3), val,
			container.NewVBox(layout.NewSpacer(), u))
	}

	return container.New(layout.NewCustomPaddedVBoxLayout(1), Caption(label), line)
}

// Sparkline draws a filled area chart of recent throughput. It is deliberately
// unlabelled and small: its job is to show the shape of traffic next to the
// exact numbers, not to replace them.
type Sparkline struct {
	samples []float64
	tone    color.NRGBA
	size    fyne.Size
}

// NewSparkline creates a sparkline of the given size.
func NewSparkline(tone color.NRGBA, size fyne.Size) *Sparkline {
	return &Sparkline{tone: tone, size: size}
}

// SetSamples replaces the plotted series.
func (s *Sparkline) SetSamples(values []float64) {
	s.samples = values
}

// Build renders the sparkline. It returns a flat baseline when there is not
// enough history to draw a line, so the layout never jumps around.
func (s *Sparkline) Build() fyne.CanvasObject {
	p := wgtheme.Current()

	base := canvas.NewRectangle(color.Transparent)
	base.SetMinSize(s.size)

	if len(s.samples) < 2 {
		flat := canvas.NewRectangle(wgtheme.Alpha(p.TextDim, 60))
		flat.SetMinSize(fyne.NewSize(s.size.Width, 1))
		return container.NewStack(base, container.NewVBox(layout.NewSpacer(), flat))
	}

	max := 0.0
	for _, v := range s.samples {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		flat := canvas.NewRectangle(wgtheme.Alpha(p.TextDim, 60))
		flat.SetMinSize(fyne.NewSize(s.size.Width, 1))
		return container.NewStack(base, container.NewVBox(layout.NewSpacer(), flat))
	}

	// One vertical bar per sample: cheaper and crisper at this size than a
	// polyline, and it degrades gracefully at any width.
	bars := make([]fyne.CanvasObject, 0, len(s.samples))
	for _, v := range s.samples {
		h := float32(v/max) * s.size.Height
		if h < 1 {
			h = 1
		}
		bar := canvas.NewRectangle(wgtheme.Alpha(s.tone, 190))
		bar.SetMinSize(fyne.NewSize(2, h))
		bars = append(bars, container.NewVBox(layout.NewSpacer(), bar))
	}

	return container.NewStack(base,
		container.New(layout.NewCustomPaddedHBoxLayout(1), bars...),
	)
}

// RelativeTime renders a timestamp as "just now" / "12m ago" / "3d ago" for
// human-facing lists such as backups, where an exact clock time is noise.
func RelativeTime(t time.Time, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return plural(int(d.Minutes()), "minute") + " ago"
	case d < 24*time.Hour:
		return plural(int(d.Hours()), "hour") + " ago"
	case d < 30*24*time.Hour:
		return plural(int(d.Hours())/24, "day") + " ago"
	default:
		return t.Format("2006-01-02")
	}
}

func plural(n int, unit string) string {
	s := ""
	if n != 1 {
		s = "s"
	}
	return strconv.Itoa(n) + " " + unit + s
}
