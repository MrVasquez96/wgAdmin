package theme

import (
	"image/color"

	"wgAdmin/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Spacing scale. Every gap in the UI should come from here rather than a
// literal, so density stays consistent across views.
const (
	SpaceXS = float32(2)
	SpaceSM = float32(4)
	SpaceMD = float32(8)
	SpaceLG = float32(12)
	SpaceXL = float32(20)
)

// Corner radii. Panels are slightly rounded, chips are pill-like.
const (
	RadiusChip   = float32(3)
	RadiusPanel  = float32(6)
	RadiusDialog = float32(10)
)

// Palette is the full set of resolved colours for one theme variant. Views
// should read colours from here instead of reaching for theme.XxxColor(), so
// that the user's customised tokens are always respected.
type Palette struct {
	Variant fyne.ThemeVariant

	Accent        color.NRGBA
	Background    color.NRGBA
	Surface       color.NRGBA
	SurfaceRaised color.NRGBA
	Header        color.NRGBA

	TextPrimary   color.NRGBA
	TextSecondary color.NRGBA
	TextDim       color.NRGBA
	TextDisabled  color.NRGBA

	Border       color.NRGBA
	BorderStrong color.NRGBA

	Success color.NRGBA
	Error   color.NRGBA
	Warning color.NRGBA
	Info    color.NRGBA

	InputBackground color.NRGBA
	Hover           color.NRGBA
	Pressed         color.NRGBA
	Selection       color.NRGBA
}

// Current returns the palette for the active theme and resolved variant.
// Safe to call before a theme exists — it falls back to the dark defaults.
func Current() Palette {
	variant := CurrentVariant()
	cfg := (*settings.AppSettings)(nil)
	if activeTheme != nil {
		cfg = activeTheme.settings
	}
	return paletteFor(cfg, variant)
}

// paletteFor resolves every token for the given settings and variant.
func paletteFor(cfg *settings.AppSettings, variant fyne.ThemeVariant) Palette {
	light := variant == theme.VariantLight

	// pick returns the user's value for this variant, falling back to the
	// matching default when unset or unparseable.
	pick := func(get func(*settings.AppSettings) string, lightDefault, darkDefault string) color.NRGBA {
		def := darkDefault
		if light {
			def = lightDefault
		}
		if cfg != nil {
			if c, err := parseHexColor(get(cfg)); err == nil {
				return c
			}
		}
		c, err := parseHexColor(def)
		if err != nil {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return c
	}

	// sel chooses the light or dark field of a settings pair.
	sel := func(lightField, darkField func(*settings.AppSettings) string) func(*settings.AppSettings) string {
		if light {
			return lightField
		}
		return darkField
	}

	p := Palette{Variant: variant}

	p.Accent = pick(sel(
		func(s *settings.AppSettings) string { return s.LightAccentColor },
		func(s *settings.AppSettings) string { return s.DarkAccentColor },
	), settings.DefaultLightAccentColor, settings.DefaultDarkAccentColor)

	p.Background = pick(sel(
		func(s *settings.AppSettings) string { return s.LightBackgroundColor },
		func(s *settings.AppSettings) string { return s.DarkBackgroundColor },
	), settings.DefaultLightBackgroundColor, settings.DefaultDarkBackgroundColor)

	p.Surface = pick(sel(
		func(s *settings.AppSettings) string { return s.LightSurfaceColor },
		func(s *settings.AppSettings) string { return s.DarkSurfaceColor },
	), settings.DefaultLightSurfaceColor, settings.DefaultDarkSurfaceColor)

	p.SurfaceRaised = pick(sel(
		func(s *settings.AppSettings) string { return s.LightSurfaceRaisedColor },
		func(s *settings.AppSettings) string { return s.DarkSurfaceRaisedColor },
	), settings.DefaultLightSurfaceRaisedColor, settings.DefaultDarkSurfaceRaisedColor)

	p.Header = pick(sel(
		func(s *settings.AppSettings) string { return s.LightHeaderBackgroundColor },
		func(s *settings.AppSettings) string { return s.DarkHeaderBackgroundColor },
	), settings.DefaultLightHeaderBackgroundColor, settings.DefaultDarkHeaderBackgroundColor)

	p.TextPrimary = pick(sel(
		func(s *settings.AppSettings) string { return s.LightTextPrimaryColor },
		func(s *settings.AppSettings) string { return s.DarkTextPrimaryColor },
	), settings.DefaultLightTextPrimaryColor, settings.DefaultDarkTextPrimaryColor)

	p.TextSecondary = pick(sel(
		func(s *settings.AppSettings) string { return s.LightTextSecondaryColor },
		func(s *settings.AppSettings) string { return s.DarkTextSecondaryColor },
	), settings.DefaultLightTextSecondaryColor, settings.DefaultDarkTextSecondaryColor)

	p.TextDim = pick(sel(
		func(s *settings.AppSettings) string { return s.LightTextDimColor },
		func(s *settings.AppSettings) string { return s.DarkTextDimColor },
	), settings.DefaultLightTextDimColor, settings.DefaultDarkTextDimColor)

	p.TextDisabled = pick(sel(
		func(s *settings.AppSettings) string { return s.LightTextDisabledColor },
		func(s *settings.AppSettings) string { return s.DarkTextDisabledColor },
	), settings.DefaultLightTextDisabledColor, settings.DefaultDarkTextDisabledColor)

	p.Border = pick(sel(
		func(s *settings.AppSettings) string { return s.LightBorderColor },
		func(s *settings.AppSettings) string { return s.DarkBorderColor },
	), settings.DefaultLightBorderColor, settings.DefaultDarkBorderColor)

	p.BorderStrong = pick(sel(
		func(s *settings.AppSettings) string { return s.LightBorderStrongColor },
		func(s *settings.AppSettings) string { return s.DarkBorderStrongColor },
	), settings.DefaultLightBorderStrongColor, settings.DefaultDarkBorderStrongColor)

	p.Success = pick(sel(
		func(s *settings.AppSettings) string { return s.LightSuccessColor },
		func(s *settings.AppSettings) string { return s.DarkSuccessColor },
	), settings.DefaultLightSuccessColor, settings.DefaultDarkSuccessColor)

	p.Error = pick(sel(
		func(s *settings.AppSettings) string { return s.LightErrorColor },
		func(s *settings.AppSettings) string { return s.DarkErrorColor },
	), settings.DefaultLightErrorColor, settings.DefaultDarkErrorColor)

	p.Warning = pick(sel(
		func(s *settings.AppSettings) string { return s.LightWarningColor },
		func(s *settings.AppSettings) string { return s.DarkWarningColor },
	), settings.DefaultLightWarningColor, settings.DefaultDarkWarningColor)

	p.Info = pick(sel(
		func(s *settings.AppSettings) string { return s.LightInfoColor },
		func(s *settings.AppSettings) string { return s.DarkInfoColor },
	), settings.DefaultLightInfoColor, settings.DefaultDarkInfoColor)

	p.InputBackground = pick(sel(
		func(s *settings.AppSettings) string { return s.LightInputBackgroundColor },
		func(s *settings.AppSettings) string { return s.DarkInputBackgroundColor },
	), settings.DefaultLightInputBackgroundColor, settings.DefaultDarkInputBackgroundColor)

	p.Hover = pick(sel(
		func(s *settings.AppSettings) string { return s.LightHoverColor },
		func(s *settings.AppSettings) string { return s.DarkHoverColor },
	), settings.DefaultLightHoverColor, settings.DefaultDarkHoverColor)

	p.Pressed = pick(sel(
		func(s *settings.AppSettings) string { return s.LightPressedColor },
		func(s *settings.AppSettings) string { return s.DarkPressedColor },
	), settings.DefaultLightPressedColor, settings.DefaultDarkPressedColor)

	p.Selection = pick(sel(
		func(s *settings.AppSettings) string { return s.LightSelectionColor },
		func(s *settings.AppSettings) string { return s.DarkSelectionColor },
	), settings.DefaultLightSelectionColor, settings.DefaultDarkSelectionColor)

	return p
}

// Tint blends a status colour into the surface so status backgrounds follow the
// user's customised palette instead of hardcoded values. strength is 0..1.
func (p Palette) Tint(c color.NRGBA, strength float32) color.NRGBA {
	return Mix(p.Surface, c, strength)
}

// SuccessTint, ErrorTint, WarningTint and InfoTint are the standard subtle
// status backgrounds used by pills, rows and the status strip.
func (p Palette) SuccessTint() color.NRGBA { return p.Tint(p.Success, 0.16) }
func (p Palette) ErrorTint() color.NRGBA   { return p.Tint(p.Error, 0.16) }
func (p Palette) WarningTint() color.NRGBA { return p.Tint(p.Warning, 0.16) }
func (p Palette) InfoTint() color.NRGBA    { return p.Tint(p.Info, 0.16) }

// AccentTint is the selected-row background in the sidebar and tables.
func (p Palette) AccentTint() color.NRGBA { return p.Tint(p.Accent, 0.18) }

// Mix linearly interpolates between two colours. t=0 returns a, t=1 returns b.
func Mix(a, b color.NRGBA, t float32) color.NRGBA {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(float32(x) + (float32(y)-float32(x))*t)
	}
	return color.NRGBA{
		R: lerp(a.R, b.R),
		G: lerp(a.G, b.G),
		B: lerp(a.B, b.B),
		A: lerp(a.A, b.A),
	}
}

// Alpha returns c with its opacity replaced.
func Alpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
