package theme

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"wgAdmin/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// VariantNone means follow the system theme
const VariantNone fyne.ThemeVariant = 255

// WGAdminTheme is the application theme: a dense, technical dark-first look
// driven entirely by the user's colour tokens in settings.
type WGAdminTheme struct {
	settings      *settings.AppSettings
	forcedVariant fyne.ThemeVariant
}

var _ fyne.Theme = (*WGAdminTheme)(nil)

// activeTheme holds the currently active WGAdminTheme instance
var activeTheme *WGAdminTheme

// NewWGAdminTheme creates a theme honouring the user's light/dark preference.
func NewWGAdminTheme(cfg *settings.AppSettings) *WGAdminTheme {
	t := &WGAdminTheme{settings: cfg, forcedVariant: VariantNone}
	if cfg != nil {
		switch cfg.ThemeVariant {
		case "light":
			t.forcedVariant = theme.VariantLight
		case "dark":
			t.forcedVariant = theme.VariantDark
		}
	}
	activeTheme = t
	return t
}

// ReduceMotion reports whether the user asked for transitions to be skipped.
func ReduceMotion() bool {
	return activeTheme != nil && activeTheme.settings != nil && activeTheme.settings.ReduceMotion
}

// CurrentVariant returns the resolved theme variant, respecting the user's
// forced light/dark setting. Use this instead of
// fyne.CurrentApp().Settings().ThemeVariant() when you need colours that match
// the custom theme.
func CurrentVariant() fyne.ThemeVariant {
	if fyne.CurrentApp() == nil {
		if activeTheme != nil && activeTheme.forcedVariant != VariantNone {
			return activeTheme.forcedVariant
		}
		return theme.VariantDark
	}
	sys := fyne.CurrentApp().Settings().ThemeVariant()
	if activeTheme != nil {
		return activeTheme.resolveVariant(sys)
	}
	return sys
}

// resolveVariant returns the forced variant if set, otherwise the system one.
func (t *WGAdminTheme) resolveVariant(systemVariant fyne.ThemeVariant) fyne.ThemeVariant {
	if t.forcedVariant != VariantNone {
		return t.forcedVariant
	}
	return systemVariant
}

// Color maps Fyne's colour names onto the application palette.
func (t *WGAdminTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	p := paletteFor(t.settings, t.resolveVariant(variant))

	switch name {
	case theme.ColorNameBackground:
		return p.Background
	case theme.ColorNameButton:
		// Plain buttons sit on the raised surface; importance-coloured buttons
		// (High/Danger/Success) are painted by Fyne from Primary/Error/Success.
		return p.SurfaceRaised
	case theme.ColorNameDisabledButton:
		return p.Surface
	case theme.ColorNameDisabled:
		return p.TextDisabled
	case theme.ColorNameError:
		return p.Error
	case theme.ColorNameFocus:
		return Alpha(p.Accent, 90)
	case theme.ColorNameForeground:
		return p.TextPrimary
	case theme.ColorNameForegroundOnPrimary:
		return onColor(p.Accent)
	case theme.ColorNameForegroundOnError:
		return onColor(p.Error)
	case theme.ColorNameForegroundOnSuccess:
		return onColor(p.Success)
	case theme.ColorNameForegroundOnWarning:
		return onColor(p.Warning)
	case theme.ColorNameHover:
		return p.Hover
	case theme.ColorNameInputBackground:
		return p.InputBackground
	case theme.ColorNameInputBorder:
		return p.Border
	case theme.ColorNameMenuBackground:
		return p.SurfaceRaised
	case theme.ColorNameOverlayBackground:
		return p.Surface
	case theme.ColorNamePlaceHolder:
		return p.TextDim
	case theme.ColorNamePressed:
		return p.Pressed
	case theme.ColorNamePrimary:
		return p.Accent
	case theme.ColorNameScrollBar:
		return Alpha(p.TextDim, 120)
	case theme.ColorNameSelection:
		return p.Selection
	case theme.ColorNameSeparator:
		return p.Border
	case theme.ColorNameShadow:
		return color.NRGBA{A: 90}
	case theme.ColorNameSuccess:
		return p.Success
	case theme.ColorNameWarning:
		return p.Warning
	case theme.ColorNameHeaderBackground:
		return p.SurfaceRaised
	}

	return theme.DefaultTheme().Color(name, variant)
}

// onColor returns black or white, whichever reads better on top of c.
func onColor(c color.NRGBA) color.NRGBA {
	// Rec. 601 luma — good enough for picking a legible label colour.
	lum := 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
	if lum > 150 {
		return color.NRGBA{R: 8, G: 10, B: 14, A: 255}
	}
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}

// Font returns Inter for prose and JetBrains Mono for machine data.
func (t *WGAdminTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.settings != nil && !t.settings.UseCustomFont {
		return theme.DefaultTheme().Font(style)
	}
	if style.Monospace {
		return fontJetBrainsMono
	}
	if style.Bold {
		if style.Italic {
			return fontInterBoldItalic
		}
		return fontInterBold
	}
	if style.Italic {
		return fontInterItalic
	}
	return fontInterRegular
}

// Icon returns the application's own icon for name, falling back to Fyne's
// default set for anything not covered.
func (t *WGAdminTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if res, ok := overrideIcon(name); ok {
		return res
	}
	return theme.DefaultTheme().Icon(name)
}

// fontSizeOffset returns the size adjustment based on the font size setting.
func (t *WGAdminTheme) fontSizeOffset() float32 {
	if t.settings == nil {
		return 0
	}
	switch t.settings.FontSize {
	case "small":
		return -1
	case "large":
		return 2
	default:
		return 0
	}
}

// compact reports whether the user asked for the tighter spacing scale.
func (t *WGAdminTheme) compact() bool {
	return t.settings != nil && t.settings.Density == settings.DensityCompact
}

// Size returns a dense, technical type and spacing scale. Text sizes track the
// user's font-size preference; spacing tracks the density preference.
func (t *WGAdminTheme) Size(name fyne.ThemeSizeName) float32 {
	offset := t.fontSizeOffset()

	pad, inner := float32(8), float32(6)
	if t.compact() {
		pad, inner = 6, 4
	}

	switch name {
	case theme.SizeNamePadding:
		return pad
	case theme.SizeNameInnerPadding:
		return inner
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 5
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 13 + offset
	case theme.SizeNameHeadingText:
		return 19 + offset
	case theme.SizeNameSubHeadingText:
		return 15 + offset
	case theme.SizeNameCaptionText:
		return 10 + offset
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameInputRadius:
		return RadiusChip + 2
	case theme.SizeNameSelectionRadius:
		return RadiusChip + 1
	}
	return theme.DefaultTheme().Size(name)
}

// parseHexColor parses a hex colour string like "#1a73e8" or "1a73e8".
func parseHexColor(hex string) (color.NRGBA, error) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return color.NRGBA{}, fmt.Errorf("invalid hex color: %s", hex)
	}
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.NRGBA{}, err
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
