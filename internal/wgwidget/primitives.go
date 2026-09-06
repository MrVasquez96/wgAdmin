package wgwidget

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// This file holds the small building blocks the console is assembled from.
// They all read their colours from the theme palette so a change in Settings
// flows through without any of them knowing about settings at all.

// Caption returns an uppercase micro-label — the small dim text that sits above
// a value or names a section. Uppercasing plus the dim tone is what separates
// "label" from "data" at a glance.
func Caption(text string) *canvas.Text {
	p := wgtheme.Current()
	t := canvas.NewText(strings.ToUpper(text), p.TextDim)
	t.TextSize = 10
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

// Value returns a monospace data value. Every machine-generated string in the
// UI — address, key, port, byte count — goes through here.
func Value(text string) *canvas.Text {
	p := wgtheme.Current()
	t := canvas.NewText(text, p.TextPrimary)
	t.TextSize = 13
	t.TextStyle = fyne.TextStyle{Monospace: true}
	return t
}

// DimValue is Value in the secondary tone, for less important data.
func DimValue(text string) *canvas.Text {
	t := Value(text)
	t.Color = wgtheme.Current().TextSecondary
	return t
}

// Title returns a pane or section title in the prose font.
func Title(text string, size float32) *canvas.Text {
	p := wgtheme.Current()
	t := canvas.NewText(text, p.TextPrimary)
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

// Hairline is a one-pixel rule used to separate sections.
func Hairline() *canvas.Rectangle {
	r := canvas.NewRectangle(wgtheme.Current().Border)
	r.SetMinSize(fyne.NewSize(0, 1))
	return r
}

// VHairline is a vertical one-pixel rule.
func VHairline() *canvas.Rectangle {
	r := canvas.NewRectangle(wgtheme.Current().Border)
	r.SetMinSize(fyne.NewSize(1, 0))
	return r
}

// Panel wraps content in a themed surface with a hairline border. When accent
// is non-nil a coloured bar is drawn down the left edge, which is how an active
// tunnel is marked without shouting.
func Panel(content fyne.CanvasObject, accent color.Color) fyne.CanvasObject {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(p.Surface)
	bg.CornerRadius = wgtheme.RadiusPanel
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1

	inner := container.NewPadded(content)
	if accent != nil {
		bar := canvas.NewRectangle(accent)
		barBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(2, 0)), bar)
		inner = container.NewBorder(nil, nil, barBox, nil, container.NewPadded(content))
	}

	return container.NewStack(bg, inner)
}

// RaisedPanel is Panel on the raised surface, for sidebars and table headers.
func RaisedPanel(content fyne.CanvasObject) fyne.CanvasObject {
	p := wgtheme.Current()

	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusPanel
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1

	return container.NewStack(bg, container.NewPadded(content))
}

// SectionHeader is an uppercase micro-header with a rule running to the right,
// used to break long forms and detail panes into scannable groups.
func SectionHeader(text string) fyne.CanvasObject {
	// The rule is vertically centred with spacers rather than container.NewCenter:
	// centring sizes a child to its MinSize, and the rule's minimum width is 0,
	// which collapsed it to nothing.
	rule := container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(), Hairline(), layout.NewSpacer())

	return container.NewBorder(nil, nil, Caption(text), nil, rule)
}

// KeyValue lays out an aligned label/value pair. The label is a dim caption and
// the value is monospace, so a column of these reads like a spec sheet.
// labelWidth keeps values aligned down the column.
func KeyValue(label, value string, labelWidth float32) fyne.CanvasObject {
	cap := Caption(label)
	labelBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(labelWidth, cap.MinSize().Height)), cap)
	return container.NewBorder(nil, nil, labelBox, nil, Value(value))
}

// KeyValueCopy is KeyValue with a copy button on the right.
func KeyValueCopy(label, value string, labelWidth float32, onCopy func()) fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", wgtheme.IconCopy, onCopy)
	btn.Importance = widget.LowImportance
	row := KeyValue(label, value, labelWidth)
	return container.NewBorder(nil, nil, nil, btn, row)
}

// Chip is a small pill used for ports, tags and paths. Passing a nil tone uses
// the neutral raised surface.
func Chip(text string, tone color.NRGBA, filled bool) fyne.CanvasObject {
	p := wgtheme.Current()

	label := canvas.NewText(text, p.TextSecondary)
	label.TextSize = 10
	label.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	bg := canvas.NewRectangle(p.SurfaceRaised)
	bg.CornerRadius = wgtheme.RadiusChip
	bg.StrokeColor = p.Border
	bg.StrokeWidth = 1

	if filled {
		bg.FillColor = p.Tint(tone, 0.18)
		bg.StrokeColor = p.Tint(tone, 0.45)
		label.Color = tone
	}

	pad := container.New(layout.NewCustomPaddedLayout(2, 2, 6, 6), label)
	return container.NewStack(bg, pad)
}

// NeutralChip is a Chip in the default tone.
func NeutralChip(text string) fyne.CanvasObject {
	return Chip(text, wgtheme.Current().TextSecondary, false)
}
