package wgwidget

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// CodeView renders a WireGuard config as tinted monospace text. It uses
// TextGrid rather than a Label so each token can carry its own colour, which is
// what makes a raw .conf readable at a glance: sections, keys, values and
// comments all separate visually.
type CodeView struct {
	grid *widget.TextGrid
}

// NewCodeView creates a config viewer showing text.
func NewCodeView(text string) *CodeView {
	c := &CodeView{grid: widget.NewTextGrid()}
	c.SetText(text)
	return c
}

// Object returns the widget for placement in a container.
func (c *CodeView) Object() fyne.CanvasObject { return c.grid }

// Text returns the current contents.
func (c *CodeView) Text() string { return c.grid.Text() }

// SetText replaces the contents and re-applies highlighting.
func (c *CodeView) SetText(text string) {
	c.grid.SetText(text)
	c.highlight()
}

// highlight colours each line according to WireGuard config syntax.
func (c *CodeView) highlight() {
	p := wgtheme.Current()

	sectionStyle := &widget.CustomTextGridStyle{FGColor: p.Accent, TextStyle: fyne.TextStyle{Bold: true}}
	keyStyle := &widget.CustomTextGridStyle{FGColor: p.Info}
	valueStyle := &widget.CustomTextGridStyle{FGColor: p.TextPrimary}
	commentStyle := &widget.CustomTextGridStyle{FGColor: p.TextDim, TextStyle: fyne.TextStyle{Italic: true}}

	for row, line := range strings.Split(c.grid.Text(), "\n") {
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			continue

		case strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";"):
			c.styleRange(row, 0, len(line), commentStyle)

		case strings.HasPrefix(trimmed, "["):
			c.styleRange(row, 0, len(line), sectionStyle)

		default:
			// key = value — colour the key and the value differently so the
			// eye can scan down the keys.
			if eq := strings.Index(line, "="); eq >= 0 {
				c.styleRange(row, 0, eq, keyStyle)
				c.styleRange(row, eq, eq+1, commentStyle)
				c.styleRange(row, eq+1, len(line), valueStyle)
			} else {
				c.styleRange(row, 0, len(line), valueStyle)
			}
		}
	}

	c.grid.Refresh()
}

// styleRange applies a style to a half-open column range of a row.
func (c *CodeView) styleRange(row, from, to int, style widget.TextGridStyle) {
	for col := from; col < to; col++ {
		c.grid.SetStyle(row, col, style)
	}
}
