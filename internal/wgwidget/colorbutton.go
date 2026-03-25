package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ColorButton is a Fyne widget: a coloured rectangle that opens the
// colour picker dialog when tapped.
//
// Typical use:
//
//	btn := colorpicker.NewColorButton(color.RGBA{R:255,G:0,B:0,A:255}, window)
//	btn.OnChanged = func(c color.Color) { myColor = c }
type ColorButton struct {
	widget.BaseWidget

	Color     color.Color       // current colour
	OnChanged func(color.Color) // called after the user picks a new colour
	WithAlpha bool              // show alpha slider in the dialog

	label  string
	parent fyne.Window

	rect   *canvas.Rectangle
	border *canvas.Rectangle
	lbl    *canvas.Text
}

// NewColorButton creates a ColorButton with the given initial colour.
// label is optional descriptive text shown on the button.
func NewColorButton(label string, col color.Color, parent fyne.Window) *ColorButton {
	b := &ColorButton{
		Color:  col,
		label:  label,
		parent: parent,
	}
	b.rect = canvas.NewRectangle(col)
	b.border = canvas.NewRectangle(color.Transparent)
	b.border.StrokeColor = color.RGBA{R: 60, G: 60, B: 60, A: 200}
	b.border.StrokeWidth = 1.5

	b.lbl = canvas.NewText(label, contrastColor(col))
	b.lbl.TextStyle = fyne.TextStyle{Bold: true}
	b.lbl.Alignment = fyne.TextAlignCenter

	b.ExtendBaseWidget(b)
	return b
}

// SetColor programmatically updates the button's displayed colour.
func (b *ColorButton) SetColor(c color.Color) {
	b.Color = c
	b.rect.FillColor = c
	b.rect.Refresh()
	b.lbl.Color = contrastColor(c)
	b.lbl.Refresh()
}

// Tapped opens the colour picker dialog.
func (b *ColorButton) Tapped(_ *fyne.PointEvent) {
	title := "Pick Colour"
	if b.label != "" {
		title = b.label
	}
	ShowColorPickerDialog(title, b.Color, b.WithAlpha, func(c color.Color) {
		b.SetColor(c)
		if b.OnChanged != nil {
			b.OnChanged(c)
		}
	}, b.parent)
}

func (b *ColorButton) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewStack(b.rect, b.border, b.lbl),
	)
}

func (b *ColorButton) MinSize() fyne.Size {
	return fyne.NewSize(72, 32)
}

// contrastColor returns black or white depending on luminance of c.
func contrastColor(c color.Color) color.Color {
	r, g, bv, _ := c.RGBA()
	lum := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bv>>8)
	if lum > 140 {
		return color.Black
	}
	return color.White
}
