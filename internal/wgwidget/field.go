package wgwidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// Field is one form row: a label above the input and an optional hint below.
// Stacked labels scan faster than Fyne's two-column form and leave the full
// width to the input, which matters for keys and address lists.
func Field(label string, input fyne.CanvasObject, hint string) fyne.CanvasObject {
	p := wgtheme.Current()
	items := []fyne.CanvasObject{
		newText(label, wgtheme.TextSM, wgtheme.WeightSemiBold, p.TextSecondary),
		input,
	}
	if hint != "" {
		h := widget.NewLabel(hint)
		h.Wrapping = fyne.TextWrapWord
		h.Importance = widget.LowImportance
		h.SizeName = theme.SizeNameCaptionText
		items = append(items, container.New(layout.NewCustomPaddedLayout(-6, -4, -6, 0), h))
	}
	return container.New(layout.NewCustomPaddedVBoxLayout(5), items...)
}

// FieldRow lays fields side by side in equal columns, for short related
// values such as port and MTU.
func FieldRow(fields ...fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewGridLayoutWithColumns(len(fields)), fields...)
}

// FormSection groups fields under a title with an optional description.
func FormSection(icon *wgtheme.Icon, title, description string, fields ...fyne.CanvasObject) fyne.CanvasObject {
	items := []fyne.CanvasObject{SectionTitle(icon, title)}
	if description != "" {
		d := widget.NewLabel(description)
		d.Wrapping = fyne.TextWrapWord
		d.Importance = widget.LowImportance
		items = append(items, container.New(layout.NewCustomPaddedLayout(-8, 0, -6, 0), d))
	}
	items = append(items, fields...)
	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG), items...)
}
