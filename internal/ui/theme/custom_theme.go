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

// WGAdminTheme is a modern custom theme for the WireGuard Admin app
type WGAdminTheme struct {
	settings       *settings.AppSettings
	forcedVariant  fyne.ThemeVariant
}

var _ fyne.Theme = (*WGAdminTheme)(nil)

// NewWGAdminTheme creates a new WGAdminTheme that follows the system variant
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
	// Store for use by CurrentVariant()
	activeTheme = t
	return t
}

// activeTheme holds the currently active WGAdminTheme instance
var activeTheme *WGAdminTheme

// CurrentVariant returns the resolved theme variant, respecting the user's
// forced light/dark setting. Use this instead of fyne.CurrentApp().Settings().ThemeVariant()
// when you need colors that match the custom theme.
func CurrentVariant() fyne.ThemeVariant {
	sys := fyne.CurrentApp().Settings().ThemeVariant()
	if activeTheme != nil {
		return activeTheme.resolveVariant(sys)
	}
	return sys
}

// resolveVariant returns the forced variant if set, otherwise the system variant
func (t *WGAdminTheme) resolveVariant(systemVariant fyne.ThemeVariant) fyne.ThemeVariant {
	if t.forcedVariant != VariantNone {
		return t.forcedVariant
	}
	return systemVariant
}

// accentColor returns the user's configured accent color or the default
func (t *WGAdminTheme) accentColor(variant fyne.ThemeVariant) color.NRGBA {
	if t.settings != nil && t.settings.AccentColor != "" {
		if c, err := parseHexColor(t.settings.AccentColor); err == nil {
			if variant == theme.VariantDark {
				// Lighten for dark mode
				return color.NRGBA{
					R: uint8(min(int(c.R)+80, 255)),
					G: uint8(min(int(c.G)+80, 255)),
					B: uint8(min(int(c.B)+80, 255)),
					A: 255,
				}
			}
			return c
		}
	}
	// Default blue
	if variant == theme.VariantLight {
		return color.NRGBA{R: 26, G: 115, B: 232, A: 255} // #1a73e8
	}
	return color.NRGBA{R: 138, G: 180, B: 248, A: 255} // #8ab4f8
}

// fontSizeOffset returns the size adjustment based on font size setting
func (t *WGAdminTheme) fontSizeOffset() float32 {
	if t.settings == nil {
		return 0
	}
	switch t.settings.FontSize {
	case "small":
		return -2
	case "large":
		return 2
	default:
		return 0
	}
}

// Color returns the color for the specified name
func (t *WGAdminTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	variant = t.resolveVariant(variant)
	accent := t.accentColor(variant)

	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 248, G: 249, B: 250, A: 255} // #f8f9fa
		}
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // #121212

	case theme.ColorNameButton:
		return accent

	case theme.ColorNameDisabledButton:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 189, G: 189, B: 189, A: 255}
		}
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255}

	case theme.ColorNameDisabled:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 158, G: 158, B: 158, A: 255}
		}
		return color.NRGBA{R: 117, G: 117, B: 117, A: 255}

	case theme.ColorNameError:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 234, G: 67, B: 53, A: 255} // #ea4335
		}
		return color.NRGBA{R: 242, G: 139, B: 130, A: 255} // #f28b82

	case theme.ColorNameFocus:
		return accent

	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 32, G: 33, B: 36, A: 255} // #202124
		}
		return color.NRGBA{R: 232, G: 234, B: 237, A: 255} // #e8eaed

	case theme.ColorNameHover:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 241, G: 245, B: 251, A: 255} // softer hover
		}
		return color.NRGBA{R: 55, G: 55, B: 55, A: 255}

	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}

	case theme.ColorNameInputBorder:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
		}
		return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043

	case theme.ColorNameMenuBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	case theme.ColorNameOverlayBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	case theme.ColorNamePlaceHolder:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 95, G: 99, B: 104, A: 255} // #5f6368
		}
		return color.NRGBA{R: 154, G: 160, B: 166, A: 255} // #9aa0a6

	case theme.ColorNamePressed:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 226, G: 237, B: 250, A: 255} // deeper than hover
		}
		return color.NRGBA{R: 70, G: 70, B: 70, A: 255}

	case theme.ColorNamePrimary:
		return accent

	case theme.ColorNameScrollBar:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 160, G: 160, B: 160, A: 255} // more visible
		}
		return color.NRGBA{R: 90, G: 90, B: 90, A: 255} // brighter in dark mode

	case theme.ColorNameSelection:
		return color.NRGBA{R: accent.R, G: accent.G, B: accent.B, A: 60}

	case theme.ColorNameSeparator:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
		}
		return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043

	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 40}

	case theme.ColorNameSuccess:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 52, G: 168, B: 83, A: 255} // #34a853
		}
		return color.NRGBA{R: 129, G: 201, B: 149, A: 255} // #81c995

	case theme.ColorNameWarning:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 251, G: 188, B: 5, A: 255} // #fbbc05
		}
		return color.NRGBA{R: 253, G: 214, B: 99, A: 255}

	case theme.ColorNameHeaderBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	}

	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the font for the specified style
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

// Icon returns the icon for the specified name
func (t *WGAdminTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified name
func (t *WGAdminTheme) Size(name fyne.ThemeSizeName) float32 {
	offset := t.fontSizeOffset()

	switch name {
	case theme.SizeNamePadding:
		return 12
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameScrollBar:
		return 18
	case theme.SizeNameScrollBarSmall:
		return 10
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14 + offset
	case theme.SizeNameHeadingText:
		return 24 + offset
	case theme.SizeNameSubHeadingText:
		return 18 + offset
	case theme.SizeNameCaptionText:
		return 12 + offset
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 10
	case theme.SizeNameSelectionRadius:
		return 6
	}
	return theme.DefaultTheme().Size(name)
}

// parseHexColor parses a hex color string like "#1a73e8" or "1a73e8"
func parseHexColor(hex string) (color.NRGBA, error) {
	hex = strings.TrimPrefix(hex, "#")
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

// Colors provides pre-defined colors for use in the app
type Colors struct{}

// Active returns the active/success color
func (c Colors) Active(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 52, G: 168, B: 83, A: 255} // #34a853
	}
	return color.NRGBA{R: 129, G: 201, B: 149, A: 255} // #81c995
}

// Inactive returns the inactive/error color
func (c Colors) Inactive(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 234, G: 67, B: 53, A: 255} // #ea4335
	}
	return color.NRGBA{R: 242, G: 139, B: 130, A: 255} // #f28b82
}

// CardBackground returns the card background color
func (c Colors) CardBackground(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return color.NRGBA{R: 30, G: 30, B: 30, A: 255} // #1e1e1e
}

// CardActiveBackground returns the background for active cards
func (c Colors) CardActiveBackground(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 245, B: 233, A: 255} // light green tint
	}
	return color.NRGBA{R: 30, G: 50, B: 35, A: 255} // dark green tint
}

// Border returns the border color
func (c Colors) Border(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
	}
	return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043
}

// CardElevation returns a subtle shadow color for card depth
func (c Colors) CardElevation(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 12}
	}
	return color.NRGBA{R: 0, G: 0, B: 0, A: 30}
}

// StatusSuccessBg returns a tinted background for success status
func (c Colors) StatusSuccessBg(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 245, B: 233, A: 255}
	}
	return color.NRGBA{R: 30, G: 50, B: 35, A: 255}
}

// StatusErrorBg returns a tinted background for error status
func (c Colors) StatusErrorBg(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 254, G: 235, B: 234, A: 255}
	}
	return color.NRGBA{R: 60, G: 30, B: 30, A: 255}
}

// StatusInfoBg returns a tinted background for info status
func (c Colors) StatusInfoBg(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 240, B: 254, A: 255}
	}
	return color.NRGBA{R: 25, G: 35, B: 50, A: 255}
}

// AppColors provides easy access to theme colors
var AppColors = Colors{}
