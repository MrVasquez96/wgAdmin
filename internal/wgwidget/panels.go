package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
)

// StatTile is one live metric: icon, label, a large value and an optional
// unit. Tiles line up in a grid above the traffic chart.
func StatTile(icon *wgtheme.Icon, label, value, unit string, tone color.NRGBA) fyne.CanvasObject {
	p := wgtheme.Current()

	head := container.New(layout.NewCustomPaddedHBoxLayout(6),
		container.NewCenter(iconImage(icon, p.TextDim, 15)),
		newText(label, wgtheme.TextSM, wgtheme.WeightMedium, p.TextSecondary),
	)

	val := newText(value, wgtheme.TextXL+2, wgtheme.WeightSemiBold, tone)
	line := fyne.CanvasObject(val)
	if unit != "" {
		u := newText(unit, wgtheme.TextSM, wgtheme.WeightMedium, p.TextDim)
		// Baseline-align the unit with the value: both texts are laid out
		// from their top, so offset the smaller one by the difference in
		// ascent (Inter's ascent is ~0.97 of the font size).
		top := (val.TextSize - u.TextSize) * 0.97
		line = container.New(layout.NewCustomPaddedHBoxLayout(4), val,
			container.New(layout.NewCustomPaddedLayout(top, 0, 0, 0), u))
	}

	return cardWith(container.New(layout.NewCustomPaddedVBoxLayout(6), head, line),
		0, p.Surface, wgtheme.SpaceLG)
}

// BannerTone selects a Banner's colour.
type BannerTone int

const (
	BannerInfo BannerTone = iota
	BannerWarning
	BannerError
	BannerSuccess
)

// Banner is an inline notice: tinted surface, status icon, title, body and an
// optional action button on the right.
func Banner(tone BannerTone, title, body string, action *Button) fyne.CanvasObject {
	p := wgtheme.Current()
	c, icon := p.Info, wgtheme.IconInfo
	switch tone {
	case BannerWarning:
		c, icon = p.Warning, wgtheme.IconAlert
	case BannerError:
		c, icon = p.Error, wgtheme.IconError
	case BannerSuccess:
		c, icon = p.Success, wgtheme.IconSuccess
	}

	bg := canvas.NewRectangle(p.Tint(c, 0.09))
	bg.CornerRadius = wgtheme.RadiusLG
	bg.StrokeColor = p.Tint(c, 0.30)
	bg.StrokeWidth = 1

	t := newText(title, wgtheme.TextMD, wgtheme.WeightSemiBold, p.TextPrimary)
	b := widget.NewLabel(body)
	b.Wrapping = fyne.TextWrapWord
	b.Importance = widget.LowImportance
	text := container.New(layout.NewCustomPaddedVBoxLayout(0),
		container.New(layout.NewCustomPaddedLayout(4, 0, 4, 0), t),
		b)

	lead := container.New(layout.NewCustomPaddedLayout(4, 0, 0, 0),
		container.NewVBox(iconImage(icon, c, 20)))

	var right fyne.CanvasObject
	if action != nil {
		right = container.NewCenter(action)
	}
	row := container.NewBorder(nil, nil, lead, right, text)

	return container.NewStack(bg, container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceLG, wgtheme.SpaceLG), row))
}

// EmptyState is the centred placeholder for an empty screen or section: an
// illustration in a soft disc, a title, a line of guidance and actions.
func EmptyState(illustration *wgtheme.Icon, title, body string, actions ...fyne.CanvasObject) fyne.CanvasObject {
	p := wgtheme.Current()

	disc := canvas.NewCircle(p.Tint(p.Accent, 0.08))
	art := container.NewStack(
		container.New(layout.NewGridWrapLayout(fyne.NewSize(112, 112)), disc),
		container.NewCenter(iconImage(illustration, p.Accent, 72)),
	)

	t := newText(title, wgtheme.TextLG+1, wgtheme.WeightSemiBold, p.TextPrimary)
	t.Alignment = fyne.TextAlignCenter

	b := widget.NewLabel(body)
	b.Alignment = fyne.TextAlignCenter
	b.Wrapping = fyne.TextWrapWord
	b.Importance = widget.LowImportance
	// Wrapping labels need a fixed width to wrap against; centring alone would
	// shrink them to one glyph per line.
	bodyBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(380, b.MinSize().Height*2.2)), b)

	items := []fyne.CanvasObject{
		container.NewCenter(art),
		spacerH(4),
		container.NewCenter(t),
		container.NewCenter(bodyBox),
	}
	if len(actions) > 0 {
		items = append(items, container.NewCenter(
			container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD), actions...)))
	}
	return container.NewCenter(container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM), items...))
}

// SectionTitle is a section heading with an optional icon and trailing
// content (counts, actions), replacing the old uppercase micro-header in the
// redesigned screens.
func SectionTitle(icon *wgtheme.Icon, title string, trailing ...fyne.CanvasObject) fyne.CanvasObject {
	p := wgtheme.Current()
	items := []fyne.CanvasObject{}
	if icon != nil {
		items = append(items, container.NewCenter(iconImage(icon, p.TextDim, 16)))
	}
	items = append(items, newText(title, wgtheme.TextMD+1, wgtheme.WeightSemiBold, p.TextPrimary))
	left := container.New(layout.NewCustomPaddedHBoxLayout(7), items...)
	var right fyne.CanvasObject
	if len(trailing) > 0 {
		right = container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM), trailing...)
	}
	return container.NewBorder(nil, nil, left, right)
}

// spacerH is a fixed vertical gap.
func spacerH(h float32) fyne.CanvasObject {
	r := canvas.NewRectangle(color.Transparent)
	r.SetMinSize(fyne.NewSize(0, h))
	return r
}
