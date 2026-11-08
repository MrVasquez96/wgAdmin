package theme

import (
	"image/color"
	"time"

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

// Corner radii. The scale is what new components use; the role names are the
// original API and now resolve onto the scale.
const (
	RadiusSM = float32(4)
	RadiusMD = float32(8)
	RadiusLG = float32(12)
	RadiusXL = float32(16)

	RadiusChip   = RadiusSM
	RadiusPanel  = RadiusMD
	RadiusDialog = RadiusLG
)

// Type scale, before the user's font-size offset (see TextSize).
const (
	TextXS      = float32(11)
	TextSM      = float32(12)
	TextMD      = float32(13)
	TextLG      = float32(15)
	TextXL      = float32(18)
	TextDisplay = float32(24)
)

// Control heights for buttons, inputs and segmented controls, before the
// density adjustment (see ControlHeight).
const (
	ControlSM = float32(28)
	ControlMD = float32(34)
	ControlLG = float32(40)
)

// Motion durations. Short enough that nothing ever feels like it is waiting
// on an animation; long enough that state changes read as movement.
const (
	MotionFast = 120 * time.Millisecond
	MotionBase = 180 * time.Millisecond
	MotionSlow = 260 * time.Millisecond
)

// EaseOut decelerates into the final value, which is what UI motion should do:
// respond immediately, settle gently.
func EaseOut(t float32) float32 { return 1 - (1-t)*(1-t)*(1-t) }

// TextSize applies the user's font-size preference to a type-scale value.
func TextSize(base float32) float32 {
	if activeTheme == nil {
		return base
	}
	return base + activeTheme.fontSizeOffset()
}

// ControlHeight applies the density preference to a control height.
func ControlHeight(base float32) float32 {
	if activeTheme != nil && activeTheme.compact() {
		return base - 4
	}
	return base
}

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

// IsDark reports whether this palette is the dark variant.
func (p Palette) IsDark() bool { return p.Variant != theme.VariantLight }

// OnAccent is the text and icon colour on an accent-filled surface.
func (p Palette) OnAccent() color.NRGBA { return onColor(p.Accent) }

// OnColor is the legible text colour on top of any filled colour.
func (p Palette) OnColor(c color.NRGBA) color.NRGBA { return onColor(c) }

// AccentSoft is a faint accent wash for hovered or highlighted surfaces.
func (p Palette) AccentSoft() color.NRGBA { return p.Tint(p.Accent, 0.10) }

// Ring is the keyboard focus ring colour.
func (p Palette) Ring() color.NRGBA { return Alpha(p.Accent, 150) }

// HoverOverlay and PressOverlay are translucent layers drawn over any surface,
// so hover and press work the same on every colour without per-colour tokens.
func (p Palette) HoverOverlay() color.NRGBA {
	if p.IsDark() {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 14}
	}
	return color.NRGBA{A: 12}
}

func (p Palette) PressOverlay() color.NRGBA {
	if p.IsDark() {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 26}
	}
	return color.NRGBA{A: 24}
}

// Shadow is the colour of one layer of an elevation shadow. Dark themes need a
// much stronger shadow to register against a near-black background.
func (p Palette) Shadow(alpha uint8) color.NRGBA {
	if p.IsDark() {
		return color.NRGBA{A: alpha}
	}
	return color.NRGBA{R: 15, G: 23, B: 42, A: alpha / 2}
}

// Lift is a surface one step brighter (dark) or darker (light) than c, used
// for hover states of filled controls.
func (p Palette) Lift(c color.NRGBA, t float32) color.NRGBA {
	if p.IsDark() {
		return Mix(c, color.NRGBA{R: 255, G: 255, B: 255, A: c.A}, t)
	}
	return Mix(c, color.NRGBA{A: c.A}, t)
}

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
