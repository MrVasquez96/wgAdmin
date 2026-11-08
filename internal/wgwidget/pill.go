package wgwidget

import (
	"image/color"
	"math"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// ------------------------------------------------------------------ pulse clock

// Live things pulse. The detail pane is rebuilt on every refresh tick, so an
// animation per dot would pile up forever; instead one shared clock drives
// every pulsing dot and forgets dots that are no longer on any canvas.
var pulse struct {
	sync.Mutex
	dots  map[*PulseDot]struct{}
	anim  *fyne.Animation
	phase float32
}

const pulsePeriod = 1800 * time.Millisecond

func registerPulse(d *PulseDot) {
	pulse.Lock()
	defer pulse.Unlock()
	if pulse.dots == nil {
		pulse.dots = map[*PulseDot]struct{}{}
	}
	pulse.dots[d] = struct{}{}
	if pulse.anim == nil && !ReduceMotion() && fyne.CurrentApp() != nil {
		pulse.anim = fyne.NewAnimation(pulsePeriod, tickPulse)
		pulse.anim.RepeatCount = fyne.AnimationRepeatForever
		pulse.anim.Curve = fyne.AnimationLinear
		pulse.anim.Start()
	}
}

func tickPulse(t float32) {
	pulse.Lock()
	pulse.phase = t
	dots := make([]*PulseDot, 0, len(pulse.dots))
	for d := range pulse.dots {
		dots = append(dots, d)
	}
	pulse.Unlock()

	drv := fyne.CurrentApp().Driver()
	alive := 0
	for _, d := range dots {
		if drv.CanvasForObject(d) == nil {
			pulse.Lock()
			delete(pulse.dots, d)
			pulse.Unlock()
			continue
		}
		alive++
		d.Refresh()
	}

	if alive == 0 {
		pulse.Lock()
		if len(pulse.dots) == 0 && pulse.anim != nil {
			pulse.anim.Stop()
			pulse.anim = nil
		}
		pulse.Unlock()
	}
}

// PulseDot is a status dot. When live it emits a soft expanding ring, so up
// tunnels and online peers read as alive even in peripheral vision.
type PulseDot struct {
	widget.BaseWidget
	Tone color.NRGBA
	Live bool
	Diam float32
}

// NewPulseDot creates a dot of diameter 8.
func NewPulseDot(tone color.NRGBA, live bool) *PulseDot {
	d := &PulseDot{Tone: tone, Live: live, Diam: 8}
	d.ExtendBaseWidget(d)
	return d
}

// Set changes the dot's tone and liveness.
func (d *PulseDot) Set(tone color.NRGBA, live bool) {
	d.Tone, d.Live = tone, live
	d.Refresh()
}

func (d *PulseDot) MinSize() fyne.Size {
	return fyne.NewSize(d.Diam*2+2, d.Diam*2+2)
}

func (d *PulseDot) CreateRenderer() fyne.WidgetRenderer {
	r := &pulseRenderer{d: d, dot: canvas.NewCircle(d.Tone), ring: canvas.NewCircle(color.Transparent)}
	if d.Live {
		registerPulse(d)
	}
	r.Refresh()
	return r
}

type pulseRenderer struct {
	d         *PulseDot
	dot, ring *canvas.Circle
}

func (r *pulseRenderer) Layout(size fyne.Size) {
	c := fyne.NewPos(size.Width/2, size.Height/2)
	rd := r.d.Diam / 2
	r.dot.Move(fyne.NewPos(c.X-rd, c.Y-rd))
	r.dot.Resize(fyne.NewSize(r.d.Diam, r.d.Diam))

	pulse.Lock()
	t := pulse.phase
	pulse.Unlock()
	if ReduceMotion() {
		t = 0.35
	}
	// The ring grows from the dot's edge to twice its size while fading out.
	rr := rd * (1 + float32(t))
	r.ring.Move(fyne.NewPos(c.X-rr, c.Y-rr))
	r.ring.Resize(fyne.NewSize(rr*2, rr*2))
}

func (r *pulseRenderer) MinSize() fyne.Size { return r.d.MinSize() }

func (r *pulseRenderer) Refresh() {
	r.dot.FillColor = r.d.Tone
	if r.d.Live {
		pulse.Lock()
		t := pulse.phase
		pulse.Unlock()
		if ReduceMotion() {
			t = 0.35
		}
		a := uint8(110 * math.Pow(float64(1-t), 1.6))
		r.ring.FillColor = wgtheme.Alpha(r.d.Tone, a)
		registerPulse(r.d)
	} else {
		r.ring.FillColor = color.Transparent
	}
	r.Layout(r.d.Size())
	r.dot.Refresh()
	r.ring.Refresh()
}

func (r *pulseRenderer) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.ring, r.dot} }
func (r *pulseRenderer) Destroy()                     {}

// --------------------------------------------------------------------- pills

// StatusPill is the UP / DOWN badge: a pulsing dot and a label on a tinted,
// fully rounded background.
func StatusPill(up bool, upText, downText string) fyne.CanvasObject {
	p := wgtheme.Current()
	tone, text := p.TextDim, downText
	if up {
		tone, text = p.Success, upText
	}
	return pill(text, tone, NewPulseDot(tone, up))
}

// Tag is a small tinted label, optionally with an icon, for states and kinds.
func Tag(text string, tone color.NRGBA, icon *wgtheme.Icon) fyne.CanvasObject {
	var lead fyne.CanvasObject
	if icon != nil {
		lead = iconImage(icon, tone, 13)
	}
	return pill(text, tone, lead)
}

func pill(text string, tone color.NRGBA, lead fyne.CanvasObject) fyne.CanvasObject {
	p := wgtheme.Current()
	label := newText(text, wgtheme.TextXS, wgtheme.WeightSemiBold, tone)

	bg := canvas.NewRectangle(p.Tint(tone, 0.14))
	bg.StrokeColor = p.Tint(tone, 0.34)
	bg.StrokeWidth = 1

	items := []fyne.CanvasObject{label}
	if lead != nil {
		items = []fyne.CanvasObject{container.NewCenter(lead), label}
	}
	row := container.New(layout.NewCustomPaddedHBoxLayout(4), items...)
	padded := container.New(layout.NewCustomPaddedLayout(2, 2, 7, 9), row)
	bg.CornerRadius = padded.MinSize().Height / 2

	return container.NewStack(bg, padded)
}

// Badge is a compact count, e.g. peers per tunnel.
func Badge(text string, emphasis bool) fyne.CanvasObject {
	p := wgtheme.Current()
	fg, fill := p.TextSecondary, p.SurfaceRaised
	if emphasis {
		fg, fill = p.Accent, p.Tint(p.Accent, 0.16)
	}
	label := newText(text, wgtheme.TextXS, wgtheme.WeightSemiBold, fg)
	bg := canvas.NewRectangle(fill)
	padded := container.New(layout.NewCustomPaddedLayout(1, 1, 6, 6), container.NewCenter(label))
	bg.CornerRadius = padded.MinSize().Height / 2
	return container.NewStack(bg, padded)
}
