package theme

import (
	"image/color"
	"testing"

	"wgAdmin/internal/settings"

	ftheme "fyne.io/fyne/v2/theme"
)

// TestForcedLightVariant checks that choosing "light" in Settings actually
// produces light colours even though the desktop reports a dark system theme.
// Fyne passes the *system* variant into Color(), so the theme has to override
// it rather than pass it through.
func TestForcedLightVariant(t *testing.T) {
	cfg := &settings.AppSettings{ThemeVariant: "light"}
	th := NewWGAdminTheme(cfg)

	// Ask using the dark system variant, as Fyne does on a dark desktop.
	got := th.Color(ftheme.ColorNameBackground, ftheme.VariantDark)
	want, _ := parseHexColor(settings.DefaultLightBackgroundColor)

	if !sameColor(got, want) {
		t.Errorf("forced light background = %v, want %v (light must win over the system variant)", got, want)
	}
}

func TestForcedDarkVariant(t *testing.T) {
	cfg := &settings.AppSettings{ThemeVariant: "dark"}
	th := NewWGAdminTheme(cfg)

	got := th.Color(ftheme.ColorNameBackground, ftheme.VariantLight)
	want, _ := parseHexColor(settings.DefaultDarkBackgroundColor)

	if !sameColor(got, want) {
		t.Errorf("forced dark background = %v, want %v", got, want)
	}
}

// TestSystemVariantPassesThrough checks "system" defers to the desktop.
func TestSystemVariantPassesThrough(t *testing.T) {
	cfg := &settings.AppSettings{ThemeVariant: "system"}
	th := NewWGAdminTheme(cfg)

	light := th.Color(ftheme.ColorNameBackground, ftheme.VariantLight)
	wantLight, _ := parseHexColor(settings.DefaultLightBackgroundColor)
	if !sameColor(light, wantLight) {
		t.Errorf("system variant with light system = %v, want %v", light, wantLight)
	}

	dark := th.Color(ftheme.ColorNameBackground, ftheme.VariantDark)
	wantDark, _ := parseHexColor(settings.DefaultDarkBackgroundColor)
	if !sameColor(dark, wantDark) {
		t.Errorf("system variant with dark system = %v, want %v", dark, wantDark)
	}
}

// TestUserColorOverridesDefault checks a customised token reaches the palette.
func TestUserColorOverridesDefault(t *testing.T) {
	cfg := &settings.AppSettings{
		ThemeVariant:        "dark",
		DarkBackgroundColor: "#123456",
	}
	th := NewWGAdminTheme(cfg)

	got := th.Color(ftheme.ColorNameBackground, ftheme.VariantDark)
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 255}
	if !sameColor(got, want) {
		t.Errorf("customised background = %v, want %v", got, want)
	}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
