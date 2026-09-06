package theme

import "wgAdmin/internal/settings"

// Preset is a complete named colour scheme for one theme variant. Applying a
// preset writes every colour token at once, so users get a coherent look
// without touching the individual swatches.
type Preset struct {
	Name    string
	Variant string // "dark" or "light" — which set of tokens it fills
	Colors  map[string]string
}

// darkPreset builds a dark-variant preset from the handful of colours that
// actually differ between schemes, deriving the rest.
func darkPreset(name, accent, bg, surface, raised, border, borderStrong,
	textPrimary, textSecondary, textDim, input, inputBorder, hover, pressed,
	selection, success, errC, warning, info string) Preset {
	return Preset{
		Name:    name,
		Variant: "dark",
		Colors: map[string]string{
			settings.KeyDarkAccentColor:           accent,
			settings.KeyDarkBackgroundColor:       bg,
			settings.KeyDarkSurfaceColor:          surface,
			settings.KeyDarkSurfaceRaisedColor:    raised,
			settings.KeyDarkBorderColor:           border,
			settings.KeyDarkBorderStrongColor:     borderStrong,
			settings.KeyDarkTextPrimaryColor:      textPrimary,
			settings.KeyDarkTextSecondaryColor:    textSecondary,
			settings.KeyDarkTextDimColor:          textDim,
			settings.KeyDarkTextDisabledColor:     textDim,
			settings.KeyDarkInputBackgroundColor:  input,
			settings.KeyDarkInputBorderColor:      inputBorder,
			settings.KeyDarkHoverColor:            hover,
			settings.KeyDarkPressedColor:          pressed,
			settings.KeyDarkSelectionColor:        selection,
			settings.KeyDarkSuccessColor:          success,
			settings.KeyDarkErrorColor:            errC,
			settings.KeyDarkWarningColor:          warning,
			settings.KeyDarkInfoColor:             info,
			settings.KeyDarkHeaderBackgroundColor: bg,
			settings.KeyDarkMenuBackgroundColor:   surface,
			settings.KeyDarkScrollbarColor:        borderStrong,
		},
	}
}

// Presets are the named schemes offered in Settings → Appearance.
var Presets = []Preset{
	// The shipped default: GitHub-dark base with the WireGuard green.
	darkPreset("Midnight",
		"#00d472", "#0b0e14", "#12161f", "#181d28", "#232a36", "#333b4a",
		"#e6edf3", "#8b949e", "#5c6672", "#181d28", "#2a3240", "#1c2230",
		"#242c3a", "#1f3a5f", "#3fb950", "#f85149", "#d29922", "#58a6ff"),

	// Warmer, lower-contrast greys with an amber accent.
	darkPreset("Carbon",
		"#f0883e", "#0d0d0f", "#161618", "#1e1e21", "#2a2a2e", "#3a3a40",
		"#e8e6e3", "#9a9691", "#6b6762", "#1e1e21", "#33333a", "#202024",
		"#2a2a30", "#3d3222", "#5cb85c", "#e5534b", "#d9a13b", "#6cb6ff"),

	// Nord — cool blue-grey, popular with terminal users.
	darkPreset("Nord",
		"#88c0d0", "#2e3440", "#3b4252", "#434c5e", "#4c566a", "#5e6a7d",
		"#eceff4", "#d8dee9", "#9aa4b6", "#3b4252", "#4c566a", "#434c5e",
		"#4c566a", "#4c566a", "#a3be8c", "#bf616a", "#ebcb8b", "#81a1c1"),

	// Solarized Dark — the classic low-glare scheme.
	darkPreset("Solarized",
		"#2aa198", "#002b36", "#073642", "#0a4553", "#0f5666", "#2a7180",
		"#eee8d5", "#93a1a1", "#657b83", "#073642", "#0f5666", "#0a4553",
		"#0f5666", "#0f5666", "#859900", "#dc322f", "#b58900", "#268bd2"),

	// Tokyo Night — deep indigo with a violet accent.
	darkPreset("Tokyo",
		"#7aa2f7", "#1a1b26", "#1f2335", "#24283b", "#2f334d", "#414868",
		"#c0caf5", "#9aa5ce", "#565f89", "#1f2335", "#2f334d", "#24283b",
		"#2f334d", "#283457", "#9ece6a", "#f7768e", "#e0af68", "#7dcfff"),
}

// PresetNames returns the preset names in display order.
func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for _, p := range Presets {
		names = append(names, p.Name)
	}
	return names
}

// FindPreset returns the preset with the given name.
func FindPreset(name string) (Preset, bool) {
	for _, p := range Presets {
		if p.Name == name {
			return p, true
		}
	}
	return Preset{}, false
}
