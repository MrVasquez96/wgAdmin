package wgwidget

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// tooltipDelay is how long the pointer must rest before a tooltip appears.
// Long enough that sweeping across a toolbar does not flash a row of them.
const tooltipDelay = 450 * time.Millisecond

// tooltips shows at most one tooltip at a time, app-wide.
var tooltips struct {
	sync.Mutex
	timer   *time.Timer
	owner   fyne.CanvasObject
	shown   fyne.CanvasObject
	layer   *Layer
	pending int // generation counter: a stale timer must not show anything
}

// TooltipsEnabled turns tooltips off when false; cmd/preview uses it so a
// pointer that happens to rest over the window does not end up in screenshots.
var TooltipsEnabled = true

// tooltipHover is called by components on hover changes. text may be empty,
// in which case nothing is shown.
func tooltipHover(owner fyne.CanvasObject, text string, in bool) {
	tooltips.Lock()
	defer tooltips.Unlock()

	if tooltips.timer != nil {
		tooltips.timer.Stop()
		tooltips.timer = nil
	}
	tooltips.pending++

	if !in || text == "" || !TooltipsEnabled {
		if tooltips.owner == owner || !in {
			hideTooltipLocked()
		}
		return
	}

	gen := tooltips.pending
	tooltips.timer = time.AfterFunc(tooltipDelay, func() {
		fyne.Do(func() {
			tooltips.Lock()
			defer tooltips.Unlock()
			if gen != tooltips.pending {
				return
			}
			showTooltipLocked(owner, text)
		})
	})
}

// HideTooltip dismisses any visible tooltip, e.g. when its owner is tapped.
func HideTooltip() {
	tooltips.Lock()
	defer tooltips.Unlock()
	tooltips.pending++
	hideTooltipLocked()
}

func hideTooltipLocked() {
	if tooltips.shown != nil && tooltips.layer != nil {
		tooltips.layer.Remove(tooltips.shown)
	}
	tooltips.shown, tooltips.layer, tooltips.owner = nil, nil, nil
}

func showTooltipLocked(owner fyne.CanvasObject, text string) {
	hideTooltipLocked()

	l, _ := topLayer(owner)
	if l == nil || !owner.Visible() {
		return
	}

	tip := tooltipBubble(text)
	size := tip.MinSize()

	d := fyne.CurrentApp().Driver()
	ownerPos := d.AbsolutePositionForObject(owner)
	ownerSize := owner.Size()
	origin := d.AbsolutePositionForObject(l.Container)
	bounds := l.bounds()

	// Centred below the owner; flipped above when it would leave the window,
	// and clamped so it never runs off either side.
	x := ownerPos.X + (ownerSize.Width-size.Width)/2
	y := ownerPos.Y + ownerSize.Height + 6
	if y+size.Height > origin.Y+bounds.Height-4 {
		y = ownerPos.Y - size.Height - 6
	}
	minX, maxX := origin.X+4, origin.X+bounds.Width-size.Width-4
	if x < minX {
		x = minX
	}
	if x > maxX {
		x = maxX
	}

	l.place(tip, fyne.NewPos(x, y))
	tooltips.shown, tooltips.layer, tooltips.owner = tip, l, owner
}

func tooltipBubble(text string) fyne.CanvasObject {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusSM + 2
	bg.StrokeColor = p.BorderStrong
	bg.StrokeWidth = 1

	label := newText(text, wgtheme.TextSM, wgtheme.WeightMedium, p.TextPrimary)

	return container.NewStack(
		shadowFor(1, bg.CornerRadius),
		bg,
		container.New(layout.NewCustomPaddedLayout(5, 5, 9, 9), label),
	)
}

// TooltipArea gives any object a tooltip. Components that already handle
// hover (Button, IconButton) take a tooltip directly instead.
type TooltipArea struct {
	widget.BaseWidget
	Content fyne.CanvasObject
	Text    string
}

var _ desktop.Hoverable = (*TooltipArea)(nil)

// WithTooltip wraps obj so hovering it shows text.
func WithTooltip(obj fyne.CanvasObject, text string) *TooltipArea {
	t := &TooltipArea{Content: obj, Text: text}
	t.ExtendBaseWidget(t)
	return t
}

func (t *TooltipArea) MouseIn(*desktop.MouseEvent)    { tooltipHover(t, t.Text, true) }
func (t *TooltipArea) MouseMoved(*desktop.MouseEvent) {}
func (t *TooltipArea) MouseOut()                      { tooltipHover(t, t.Text, false) }

func (t *TooltipArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.Content)
}
