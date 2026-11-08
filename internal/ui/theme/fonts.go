package theme

import (
	"fyne.io/fyne/v2"

	"wgAdmin/assets"
)

// Font faces. fyne.TextStyle only knows Bold and Italic, so the theme's Font()
// serves Regular/Bold/Italic/Mono, and the in-between weights are reached
// directly through canvas.Text.FontSource by the custom widgets.
var (
	fontInterRegular    = fyne.NewStaticResource("Inter-Regular.ttf", assets.InterRegular)
	fontInterItalic     = fyne.NewStaticResource("Inter-Italic.ttf", assets.InterItalic)
	fontInterBold       = fyne.NewStaticResource("Inter-Bold.ttf", assets.InterBold)
	fontInterBoldItalic = fyne.NewStaticResource("Inter-BoldItalic.ttf", assets.InterBoldItalic)
	fontJetBrainsMono   = fyne.NewStaticResource("JetBrainsMono-Regular.ttf", assets.JetBrainsMono)

	fontInterMedium   = fyne.NewStaticResource("Inter-Medium.ttf", assets.InterMedium)
	fontInterSemiBold = fyne.NewStaticResource("Inter-SemiBold.ttf", assets.InterSemiBold)
)

// Weight names a font weight for canvas.Text.FontSource.
type Weight int

const (
	WeightRegular Weight = iota
	WeightMedium
	WeightSemiBold
	WeightBold
	WeightMono
)

// Font returns the face for w, or nil when the user turned custom fonts off —
// a nil FontSource makes canvas.Text fall back to the theme, which is exactly
// the system-font behaviour they asked for.
func Font(w Weight) fyne.Resource {
	if activeTheme != nil && activeTheme.settings != nil && !activeTheme.settings.UseCustomFont {
		return nil
	}
	switch w {
	case WeightMedium:
		return fontInterMedium
	case WeightSemiBold:
		return fontInterSemiBold
	case WeightBold:
		return fontInterBold
	case WeightMono:
		return fontJetBrainsMono
	}
	return fontInterRegular
}
