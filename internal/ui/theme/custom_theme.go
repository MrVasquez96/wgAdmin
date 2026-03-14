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

// getColorOrDefault returns a parsed hex color from settings, falling back to the default
func (t *WGAdminTheme) getColorOrDefault(settingColor, defaultColor string) color.NRGBA {
	if settingColor != "" {
		if c, err := parseHexColor(settingColor); err == nil {
			return c
		}
	}
	if c, err := parseHexColor(defaultColor); err == nil {
		return c
	}
	// Fallback to white
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}

// accentColor returns the user's configured accent color or the default
func (t *WGAdminTheme) accentColor(variant fyne.ThemeVariant) color.NRGBA {
	if t.settings == nil {
		if variant == theme.VariantLight {
			return t.getColorOrDefault("", settings.DefaultLightAccentColor)
		}
		return t.getColorOrDefault("", settings.DefaultDarkAccentColor)
	}

	if variant == theme.VariantLight {
		return t.getColorOrDefault(t.settings.LightAccentColor, settings.DefaultLightAccentColor)
	}
	return t.getColorOrDefault(t.settings.DarkAccentColor, settings.DefaultDarkAccentColor)
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

	// Helper to get color based on variant
	getColor := func(lightSetting, darkSetting, lightDefault, darkDefault string) color.NRGBA {
		if variant == theme.VariantLight {
			if t.settings != nil {
				return t.getColorOrDefault(lightSetting, lightDefault)
			}
			return t.getColorOrDefault("", lightDefault)
		}
		if t.settings != nil {
			return t.getColorOrDefault(darkSetting, darkDefault)
		}
		return t.getColorOrDefault("", darkDefault)
	}

	switch name {
	case theme.ColorNameBackground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightBackgroundColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkBackgroundColor }),
			settings.DefaultLightBackgroundColor,
			settings.DefaultDarkBackgroundColor,
		)

	case theme.ColorNameButton:
		return accent

	case theme.ColorNameDisabledButton:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightTextDisabledColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkTextDisabledColor }),
			settings.DefaultLightTextDisabledColor,
			settings.DefaultDarkTextDisabledColor,
		)

	case theme.ColorNameDisabled:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightTextDisabledColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkTextDisabledColor }),
			settings.DefaultLightTextDisabledColor,
			settings.DefaultDarkTextDisabledColor,
		)

	case theme.ColorNameError:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightErrorColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkErrorColor }),
			settings.DefaultLightErrorColor,
			settings.DefaultDarkErrorColor,
		)

	case theme.ColorNameFocus:
		return accent

	case theme.ColorNameForeground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightTextPrimaryColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkTextPrimaryColor }),
			settings.DefaultLightTextPrimaryColor,
			settings.DefaultDarkTextPrimaryColor,
		)

	case theme.ColorNameHover:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightHoverColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkHoverColor }),
			settings.DefaultLightHoverColor,
			settings.DefaultDarkHoverColor,
		)

	case theme.ColorNameInputBackground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightInputBackgroundColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkInputBackgroundColor }),
			settings.DefaultLightInputBackgroundColor,
			settings.DefaultDarkInputBackgroundColor,
		)

	case theme.ColorNameInputBorder:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightInputBorderColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkInputBorderColor }),
			settings.DefaultLightInputBorderColor,
			settings.DefaultDarkInputBorderColor,
		)

	case theme.ColorNameMenuBackground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightMenuBackgroundColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkMenuBackgroundColor }),
			settings.DefaultLightMenuBackgroundColor,
			settings.DefaultDarkMenuBackgroundColor,
		)

	case theme.ColorNameOverlayBackground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightSurfaceColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkSurfaceColor }),
			settings.DefaultLightSurfaceColor,
			settings.DefaultDarkSurfaceColor,
		)

	case theme.ColorNamePlaceHolder:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightTextSecondaryColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkTextSecondaryColor }),
			settings.DefaultLightTextSecondaryColor,
			settings.DefaultDarkTextSecondaryColor,
		)

	case theme.ColorNamePressed:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightPressedColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkPressedColor }),
			settings.DefaultLightPressedColor,
			settings.DefaultDarkPressedColor,
		)

	case theme.ColorNamePrimary:
		return accent

	case theme.ColorNameScrollBar:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightScrollbarColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkScrollbarColor }),
			settings.DefaultLightScrollbarColor,
			settings.DefaultDarkScrollbarColor,
		)

	case theme.ColorNameSelection:
		sel := t.getColorOrDefault(
			t.settingOrEmpty(func(s *settings.AppSettings) string {
				if variant == theme.VariantLight {
					return s.LightSelectionColor
				}
				return s.DarkSelectionColor
			}),
			func() string {
				if variant == theme.VariantLight {
					return settings.DefaultLightSelectionColor
				}
				return settings.DefaultDarkSelectionColor
			}(),
		)
		return sel

	case theme.ColorNameSeparator:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightBorderColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkBorderColor }),
			settings.DefaultLightBorderColor,
			settings.DefaultDarkBorderColor,
		)

	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 40}

	case theme.ColorNameSuccess:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightSuccessColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkSuccessColor }),
			settings.DefaultLightSuccessColor,
			settings.DefaultDarkSuccessColor,
		)

	case theme.ColorNameWarning:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightWarningColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkWarningColor }),
			settings.DefaultLightWarningColor,
			settings.DefaultDarkWarningColor,
		)

	case theme.ColorNameHeaderBackground:
		return getColor(
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.LightHeaderBackgroundColor }),
			t.settingOrEmpty(func(s *settings.AppSettings) string { return s.DarkHeaderBackgroundColor }),
			settings.DefaultLightHeaderBackgroundColor,
			settings.DefaultDarkHeaderBackgroundColor,
		)
	}

	return theme.DefaultTheme().Color(name, variant)
}

// settingOrEmpty is a helper to safely get a setting value or empty string
func (t *WGAdminTheme) settingOrEmpty(fn func(*settings.AppSettings) string) string {
	if t.settings == nil {
		return ""
	}
	return fn(t.settings)
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

// getThemeColor is a helper to get colors from the active theme
func (c Colors) getThemeColor(lightSetting, darkSetting, lightDefault, darkDefault string, variant fyne.ThemeVariant) color.Color {
	var lightColor, darkColor string
	if activeTheme != nil && activeTheme.settings != nil {
		lightColor = lightSetting
		darkColor = darkSetting
	}

	if variant == theme.VariantLight {
		if col, err := parseHexColor(lightColor); err == nil {
			return col
		}
		if col, err := parseHexColor(lightDefault); err == nil {
			return col
		}
	} else {
		if col, err := parseHexColor(darkColor); err == nil {
			return col
		}
		if col, err := parseHexColor(darkDefault); err == nil {
			return col
		}
	}
	return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
}

// Active returns the active/success color
func (c Colors) Active(variant fyne.ThemeVariant) color.Color {
	if activeTheme != nil && activeTheme.settings != nil {
		if variant == theme.VariantLight {
			return c.getThemeColor(activeTheme.settings.LightSuccessColor, "", settings.DefaultLightSuccessColor, settings.DefaultDarkSuccessColor, variant)
		}
		return c.getThemeColor("", activeTheme.settings.DarkSuccessColor, settings.DefaultLightSuccessColor, settings.DefaultDarkSuccessColor, variant)
	}
	return c.getThemeColor("", "", settings.DefaultLightSuccessColor, settings.DefaultDarkSuccessColor, variant)
}

// Inactive returns the inactive/error color
func (c Colors) Inactive(variant fyne.ThemeVariant) color.Color {
	if activeTheme != nil && activeTheme.settings != nil {
		if variant == theme.VariantLight {
			return c.getThemeColor(activeTheme.settings.LightErrorColor, "", settings.DefaultLightErrorColor, settings.DefaultDarkErrorColor, variant)
		}
		return c.getThemeColor("", activeTheme.settings.DarkErrorColor, settings.DefaultLightErrorColor, settings.DefaultDarkErrorColor, variant)
	}
	return c.getThemeColor("", "", settings.DefaultLightErrorColor, settings.DefaultDarkErrorColor, variant)
}

// CardBackground returns the card background color
func (c Colors) CardBackground(variant fyne.ThemeVariant) color.Color {
	if activeTheme != nil && activeTheme.settings != nil {
		if variant == theme.VariantLight {
			return c.getThemeColor(activeTheme.settings.LightSurfaceColor, "", settings.DefaultLightSurfaceColor, settings.DefaultDarkSurfaceColor, variant)
		}
		return c.getThemeColor("", activeTheme.settings.DarkSurfaceColor, settings.DefaultLightSurfaceColor, settings.DefaultDarkSurfaceColor, variant)
	}
	return c.getThemeColor("", "", settings.DefaultLightSurfaceColor, settings.DefaultDarkSurfaceColor, variant)
}

// CardActiveBackground returns the background for active cards (tinted with success color)
func (c Colors) CardActiveBackground(variant fyne.ThemeVariant) color.Color {
	// Light green tint for light mode, dark green tint for dark mode
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 245, B: 233, A: 255}
	}
	return color.NRGBA{R: 30, G: 50, B: 35, A: 255}
}

// Border returns the border color
func (c Colors) Border(variant fyne.ThemeVariant) color.Color {
	if activeTheme != nil && activeTheme.settings != nil {
		if variant == theme.VariantLight {
			return c.getThemeColor(activeTheme.settings.LightBorderColor, "", settings.DefaultLightBorderColor, settings.DefaultDarkBorderColor, variant)
		}
		return c.getThemeColor("", activeTheme.settings.DarkBorderColor, settings.DefaultLightBorderColor, settings.DefaultDarkBorderColor, variant)
	}
	return c.getThemeColor("", "", settings.DefaultLightBorderColor, settings.DefaultDarkBorderColor, variant)
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
	if activeTheme != nil && activeTheme.settings != nil {
		// Use info color with alpha for background tint
		if variant == theme.VariantLight {
			return color.NRGBA{R: 232, G: 240, B: 254, A: 255}
		}
		return color.NRGBA{R: 25, G: 35, B: 50, A: 255}
	}
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 240, B: 254, A: 255}
	}
	return color.NRGBA{R: 25, G: 35, B: 50, A: 255}
}

// AppColors provides easy access to theme colors
var AppColors = Colors{}
