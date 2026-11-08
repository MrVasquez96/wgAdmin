package wgwidget

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

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
