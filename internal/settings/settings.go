package settings

import "fyne.io/fyne/v2"

// Preference key constants
const (
	KeyWGConfigPath        = "wg_config_path"
	KeyClientConfigDir     = "client_config_dir"
	KeyWindowWidth         = "window_width"
	KeyWindowHeight        = "window_height"
	KeyStartFullscreen     = "start_fullscreen"
	KeyAutoRefreshEnabled  = "auto_refresh_enabled"
	KeyAutoRefreshSecs     = "auto_refresh_seconds"
	KeyConfirmBeforeDelete = "confirm_before_delete"
	KeyThemeVariant        = "theme_variant"
	KeyScanWorkers         = "scan_workers"
	KeyScanTimeoutSecs     = "scan_timeout_seconds"
	KeyPrivilegeEscalation = "privilege_escalation"
	KeyFontSize            = "font_size"
	KeyUseCustomFont       = "use_custom_font"
	KeyReduceMotion        = "reduce_motion"
	KeyLanguage            = "language"
	KeyDensity             = "density"
	KeyAccentColor         = "accent_color" // Deprecated: use KeyLightAccentColor and KeyDarkAccentColor

	// Color settings - Light mode
	KeyLightAccentColor           = "light_accent_color"
	KeyLightBackgroundColor       = "light_background_color"
	KeyLightSurfaceColor          = "light_surface_color"
	KeyLightSurfaceRaisedColor    = "light_surface_raised_color"
	KeyLightTextPrimaryColor      = "light_text_primary_color"
	KeyLightTextSecondaryColor    = "light_text_secondary_color"
	KeyLightTextDimColor          = "light_text_dim_color"
	KeyLightTextDisabledColor     = "light_text_disabled_color"
	KeyLightBorderColor           = "light_border_color"
	KeyLightBorderStrongColor     = "light_border_strong_color"
	KeyLightSuccessColor          = "light_success_color"
	KeyLightErrorColor            = "light_error_color"
	KeyLightWarningColor          = "light_warning_color"
	KeyLightInfoColor             = "light_info_color"
	KeyLightInputBackgroundColor  = "light_input_background_color"
	KeyLightInputBorderColor      = "light_input_border_color"
	KeyLightHoverColor            = "light_hover_color"
	KeyLightPressedColor          = "light_pressed_color"
	KeyLightSelectionColor        = "light_selection_color"
	KeyLightHeaderBackgroundColor = "light_header_background_color"
	KeyLightMenuBackgroundColor   = "light_menu_background_color"
	KeyLightScrollbarColor        = "light_scrollbar_color"

	// Color settings - Dark mode
	KeyDarkAccentColor           = "dark_accent_color"
	KeyDarkBackgroundColor       = "dark_background_color"
	KeyDarkSurfaceColor          = "dark_surface_color"
	KeyDarkSurfaceRaisedColor    = "dark_surface_raised_color"
	KeyDarkTextPrimaryColor      = "dark_text_primary_color"
	KeyDarkTextSecondaryColor    = "dark_text_secondary_color"
	KeyDarkTextDimColor          = "dark_text_dim_color"
	KeyDarkTextDisabledColor     = "dark_text_disabled_color"
	KeyDarkBorderColor           = "dark_border_color"
	KeyDarkBorderStrongColor     = "dark_border_strong_color"
	KeyDarkSuccessColor          = "dark_success_color"
	KeyDarkErrorColor            = "dark_error_color"
	KeyDarkWarningColor          = "dark_warning_color"
	KeyDarkInfoColor             = "dark_info_color"
	KeyDarkInputBackgroundColor  = "dark_input_background_color"
	KeyDarkInputBorderColor      = "dark_input_border_color"
	KeyDarkHoverColor            = "dark_hover_color"
	KeyDarkPressedColor          = "dark_pressed_color"
	KeyDarkSelectionColor        = "dark_selection_color"
	KeyDarkHeaderBackgroundColor = "dark_header_background_color"
	KeyDarkMenuBackgroundColor   = "dark_menu_background_color"
	KeyDarkScrollbarColor        = "dark_scrollbar_color"
)

// Default values
const (
	DefaultWGConfigPath        = "/etc/wireguard"
	DefaultClientConfigDir     = "clients"
	DefaultWindowWidth         = 950
	DefaultWindowHeight        = 800
	DefaultStartFullscreen     = false
	DefaultAutoRefreshEnabled  = false
	DefaultAutoRefreshSecs     = 5
	DefaultConfirmBeforeDelete = true
	DefaultThemeVariant        = "dark"
	DefaultScanWorkers         = 100
	DefaultScanTimeoutSecs     = 2
	DefaultPrivilegeEscalation = "none"
	DefaultFontSize            = "normal"
	DefaultUseCustomFont       = true
	DefaultReduceMotion        = false
	DefaultLanguage            = "english"
	DefaultDensity             = DensityComfortable
	DefaultAccentColor         = "#1a73e8" // Deprecated: use DefaultLightAccentColor

	// Light mode color defaults - Clean professional dark-on-light
	DefaultLightAccentColor          = "#0a6640" // Dark forest green
	DefaultLightBackgroundColor      = "#f4f6f8" // Cool light gray
	DefaultLightSurfaceColor         = "#ffffff" // White panels
	DefaultLightSurfaceRaisedColor   = "#f8fafc" // Sidebar, table headers, chips
	DefaultLightTextPrimaryColor     = "#12161f" // Near-black values
	DefaultLightTextSecondaryColor   = "#4d5866" // Labels
	DefaultLightTextDimColor         = "#78828f" // Micro-captions, units
	DefaultLightTextDisabledColor    = "#9e9e9e" // Light gray
	DefaultLightBorderColor          = "#d8dee6" // Hairlines
	DefaultLightBorderStrongColor    = "#c2cbd6" // Focused/selected edges
	DefaultLightSuccessColor         = "#2d9e52" // Darker green
	DefaultLightErrorColor           = "#d73a49" // Accessible red
	DefaultLightWarningColor         = "#9a6700" // Dark amber
	DefaultLightInfoColor            = "#0969da" // Accessible blue
	DefaultLightInputBackgroundColor = "#ffffff" // White
	DefaultLightInputBorderColor     = "#d8dee6" // Matches border
	DefaultLightHoverColor           = "#eaeef2" // Subtle gray hover
	DefaultLightPressedColor         = "#dde1e7" // Deeper hover
	DefaultLightSelectionColor       = "#bfe3d0" // Accent-tinted selection
	// The toolbar must follow the variant. Keeping it dark in light mode left
	// the dark-green brand and the checkbox label unreadable, because those
	// widgets paint with the light foreground colour.
	DefaultLightHeaderBackgroundColor = "#ffffff" // Toolbar
	DefaultLightMenuBackgroundColor   = "#ffffff" // White menus
	DefaultLightScrollbarColor        = "#8c959f" // Accessible gray

	// Dark mode color defaults - Dark Cyber/Tech (WireGuard brand aesthetic)
	DefaultDarkAccentColor           = "#00d472" // Neon green — WireGuard brand
	DefaultDarkBackgroundColor       = "#0b0e14" // Window base
	DefaultDarkSurfaceColor          = "#12161f" // Panels
	DefaultDarkSurfaceRaisedColor    = "#181d28" // Sidebar, table headers, chips
	DefaultDarkTextPrimaryColor      = "#e6edf3" // Values
	DefaultDarkTextSecondaryColor    = "#8b949e" // Labels
	DefaultDarkTextDimColor          = "#5c6672" // Micro-captions, units
	DefaultDarkTextDisabledColor     = "#484f58" // Dark disabled
	DefaultDarkBorderColor           = "#232a36" // Hairlines
	DefaultDarkBorderStrongColor     = "#333b4a" // Focused/selected edges
	DefaultDarkSuccessColor          = "#3fb950" // Bright green
	DefaultDarkErrorColor            = "#f85149" // Bright red
	DefaultDarkWarningColor          = "#d29922" // Amber
	DefaultDarkInfoColor             = "#58a6ff" // Bright blue
	DefaultDarkInputBackgroundColor  = "#181d28" // Input field bg
	DefaultDarkInputBorderColor      = "#2a3240" // Input field border
	DefaultDarkHoverColor            = "#1c2230" // Hover state
	DefaultDarkPressedColor          = "#242c3a" // Pressed state
	DefaultDarkSelectionColor        = "#123a2a" // Accent-tinted selection
	DefaultDarkHeaderBackgroundColor = "#0e121a" // Toolbar
	DefaultDarkMenuBackgroundColor   = "#12161f" // Same as surface
	DefaultDarkScrollbarColor        = "#3a4352" // Visible scrollbar
)

// Density controls how tight the global spacing is. Compact suits dense
// admin work on large displays; comfortable is the default.
const (
	DensityCompact     = "compact"
	DensityComfortable = "comfortable"
)

// AppSettings holds all application settings.
type AppSettings struct {
	WGConfigPath        string
	ClientConfigDir     string
	WindowWidth         int
	WindowHeight        int
	StartFullscreen     bool
	AutoRefreshEnabled  bool
	AutoRefreshSecs     int
	ConfirmBeforeDelete bool
	ThemeVariant        string
	ScanWorkers         int
	ScanTimeoutSecs     int
	PrivilegeEscalation string
	FontSize            string
	Language            string
	Density             string
	AccentColor         string // Deprecated: use LightAccentColor and DarkAccentColor
	UseCustomFont       bool
	ReduceMotion        bool // skip UI transitions and pulsing indicators

	// Light mode colors
	LightAccentColor           string
	LightBackgroundColor       string
	LightSurfaceColor          string
	LightSurfaceRaisedColor    string
	LightTextPrimaryColor      string
	LightTextSecondaryColor    string
	LightTextDimColor          string
	LightTextDisabledColor     string
	LightBorderColor           string
	LightBorderStrongColor     string
	LightSuccessColor          string
	LightErrorColor            string
	LightWarningColor          string
	LightInfoColor             string
	LightInputBackgroundColor  string
	LightInputBorderColor      string
	LightHoverColor            string
	LightPressedColor          string
	LightSelectionColor        string
	LightHeaderBackgroundColor string
	LightMenuBackgroundColor   string
	LightScrollbarColor        string

	// Dark mode colors
	DarkAccentColor           string
	DarkBackgroundColor       string
	DarkSurfaceColor          string
	DarkSurfaceRaisedColor    string
	DarkTextPrimaryColor      string
	DarkTextSecondaryColor    string
	DarkTextDimColor          string
	DarkTextDisabledColor     string
	DarkBorderColor           string
	DarkBorderStrongColor     string
	DarkSuccessColor          string
	DarkErrorColor            string
	DarkWarningColor          string
	DarkInfoColor             string
	DarkInputBackgroundColor  string
	DarkInputBorderColor      string
	DarkHoverColor            string
	DarkPressedColor          string
	DarkSelectionColor        string
	DarkHeaderBackgroundColor string
	DarkMenuBackgroundColor   string
	DarkScrollbarColor        string
}

// Load reads all settings from Fyne preferences, applying defaults for missing values.
func Load(prefs fyne.Preferences) *AppSettings {
	return &AppSettings{
		WGConfigPath:        prefs.StringWithFallback(KeyWGConfigPath, DefaultWGConfigPath),
		ClientConfigDir:     prefs.StringWithFallback(KeyClientConfigDir, DefaultClientConfigDir),
		WindowWidth:         prefs.IntWithFallback(KeyWindowWidth, DefaultWindowWidth),
		WindowHeight:        prefs.IntWithFallback(KeyWindowHeight, DefaultWindowHeight),
		StartFullscreen:     prefs.BoolWithFallback(KeyStartFullscreen, DefaultStartFullscreen),
		AutoRefreshEnabled:  prefs.BoolWithFallback(KeyAutoRefreshEnabled, DefaultAutoRefreshEnabled),
		AutoRefreshSecs:     prefs.IntWithFallback(KeyAutoRefreshSecs, DefaultAutoRefreshSecs),
		ConfirmBeforeDelete: prefs.BoolWithFallback(KeyConfirmBeforeDelete, DefaultConfirmBeforeDelete),
		ThemeVariant:        prefs.StringWithFallback(KeyThemeVariant, DefaultThemeVariant),
		ScanWorkers:         prefs.IntWithFallback(KeyScanWorkers, DefaultScanWorkers),
		ScanTimeoutSecs:     prefs.IntWithFallback(KeyScanTimeoutSecs, DefaultScanTimeoutSecs),
		PrivilegeEscalation: prefs.StringWithFallback(KeyPrivilegeEscalation, DefaultPrivilegeEscalation),
		FontSize:            prefs.StringWithFallback(KeyFontSize, DefaultFontSize),
		Language:            prefs.StringWithFallback(KeyLanguage, DefaultLanguage),
		Density:             prefs.StringWithFallback(KeyDensity, DefaultDensity),
		AccentColor:         prefs.StringWithFallback(KeyAccentColor, DefaultAccentColor),
		UseCustomFont:       prefs.BoolWithFallback(KeyUseCustomFont, DefaultUseCustomFont),
		ReduceMotion:        prefs.BoolWithFallback(KeyReduceMotion, DefaultReduceMotion),

		// Light mode colors
		LightAccentColor:           prefs.StringWithFallback(KeyLightAccentColor, DefaultLightAccentColor),
		LightBackgroundColor:       prefs.StringWithFallback(KeyLightBackgroundColor, DefaultLightBackgroundColor),
		LightSurfaceColor:          prefs.StringWithFallback(KeyLightSurfaceColor, DefaultLightSurfaceColor),
		LightSurfaceRaisedColor:    prefs.StringWithFallback(KeyLightSurfaceRaisedColor, DefaultLightSurfaceRaisedColor),
		LightTextPrimaryColor:      prefs.StringWithFallback(KeyLightTextPrimaryColor, DefaultLightTextPrimaryColor),
		LightTextSecondaryColor:    prefs.StringWithFallback(KeyLightTextSecondaryColor, DefaultLightTextSecondaryColor),
		LightTextDimColor:          prefs.StringWithFallback(KeyLightTextDimColor, DefaultLightTextDimColor),
		LightTextDisabledColor:     prefs.StringWithFallback(KeyLightTextDisabledColor, DefaultLightTextDisabledColor),
		LightBorderColor:           prefs.StringWithFallback(KeyLightBorderColor, DefaultLightBorderColor),
		LightBorderStrongColor:     prefs.StringWithFallback(KeyLightBorderStrongColor, DefaultLightBorderStrongColor),
		LightSuccessColor:          prefs.StringWithFallback(KeyLightSuccessColor, DefaultLightSuccessColor),
		LightErrorColor:            prefs.StringWithFallback(KeyLightErrorColor, DefaultLightErrorColor),
		LightWarningColor:          prefs.StringWithFallback(KeyLightWarningColor, DefaultLightWarningColor),
		LightInfoColor:             prefs.StringWithFallback(KeyLightInfoColor, DefaultLightInfoColor),
		LightInputBackgroundColor:  prefs.StringWithFallback(KeyLightInputBackgroundColor, DefaultLightInputBackgroundColor),
		LightInputBorderColor:      prefs.StringWithFallback(KeyLightInputBorderColor, DefaultLightInputBorderColor),
		LightHoverColor:            prefs.StringWithFallback(KeyLightHoverColor, DefaultLightHoverColor),
		LightPressedColor:          prefs.StringWithFallback(KeyLightPressedColor, DefaultLightPressedColor),
		LightSelectionColor:        prefs.StringWithFallback(KeyLightSelectionColor, DefaultLightSelectionColor),
		LightHeaderBackgroundColor: prefs.StringWithFallback(KeyLightHeaderBackgroundColor, DefaultLightHeaderBackgroundColor),
		LightMenuBackgroundColor:   prefs.StringWithFallback(KeyLightMenuBackgroundColor, DefaultLightMenuBackgroundColor),
		LightScrollbarColor:        prefs.StringWithFallback(KeyLightScrollbarColor, DefaultLightScrollbarColor),

		// Dark mode colors
		DarkAccentColor:           prefs.StringWithFallback(KeyDarkAccentColor, DefaultDarkAccentColor),
		DarkBackgroundColor:       prefs.StringWithFallback(KeyDarkBackgroundColor, DefaultDarkBackgroundColor),
		DarkSurfaceColor:          prefs.StringWithFallback(KeyDarkSurfaceColor, DefaultDarkSurfaceColor),
		DarkSurfaceRaisedColor:    prefs.StringWithFallback(KeyDarkSurfaceRaisedColor, DefaultDarkSurfaceRaisedColor),
		DarkTextPrimaryColor:      prefs.StringWithFallback(KeyDarkTextPrimaryColor, DefaultDarkTextPrimaryColor),
		DarkTextSecondaryColor:    prefs.StringWithFallback(KeyDarkTextSecondaryColor, DefaultDarkTextSecondaryColor),
		DarkTextDimColor:          prefs.StringWithFallback(KeyDarkTextDimColor, DefaultDarkTextDimColor),
		DarkTextDisabledColor:     prefs.StringWithFallback(KeyDarkTextDisabledColor, DefaultDarkTextDisabledColor),
		DarkBorderColor:           prefs.StringWithFallback(KeyDarkBorderColor, DefaultDarkBorderColor),
		DarkBorderStrongColor:     prefs.StringWithFallback(KeyDarkBorderStrongColor, DefaultDarkBorderStrongColor),
		DarkSuccessColor:          prefs.StringWithFallback(KeyDarkSuccessColor, DefaultDarkSuccessColor),
		DarkErrorColor:            prefs.StringWithFallback(KeyDarkErrorColor, DefaultDarkErrorColor),
		DarkWarningColor:          prefs.StringWithFallback(KeyDarkWarningColor, DefaultDarkWarningColor),
		DarkInfoColor:             prefs.StringWithFallback(KeyDarkInfoColor, DefaultDarkInfoColor),
		DarkInputBackgroundColor:  prefs.StringWithFallback(KeyDarkInputBackgroundColor, DefaultDarkInputBackgroundColor),
		DarkInputBorderColor:      prefs.StringWithFallback(KeyDarkInputBorderColor, DefaultDarkInputBorderColor),
		DarkHoverColor:            prefs.StringWithFallback(KeyDarkHoverColor, DefaultDarkHoverColor),
		DarkPressedColor:          prefs.StringWithFallback(KeyDarkPressedColor, DefaultDarkPressedColor),
		DarkSelectionColor:        prefs.StringWithFallback(KeyDarkSelectionColor, DefaultDarkSelectionColor),
		DarkHeaderBackgroundColor: prefs.StringWithFallback(KeyDarkHeaderBackgroundColor, DefaultDarkHeaderBackgroundColor),
		DarkMenuBackgroundColor:   prefs.StringWithFallback(KeyDarkMenuBackgroundColor, DefaultDarkMenuBackgroundColor),
		DarkScrollbarColor:        prefs.StringWithFallback(KeyDarkScrollbarColor, DefaultDarkScrollbarColor),
	}
}

// Save writes all settings to Fyne preferences.
func (s *AppSettings) Save(prefs fyne.Preferences) {
	prefs.SetString(KeyWGConfigPath, s.WGConfigPath)
	prefs.SetString(KeyClientConfigDir, s.ClientConfigDir)
	prefs.SetInt(KeyWindowWidth, s.WindowWidth)
	prefs.SetInt(KeyWindowHeight, s.WindowHeight)
	prefs.SetBool(KeyStartFullscreen, s.StartFullscreen)
	prefs.SetBool(KeyAutoRefreshEnabled, s.AutoRefreshEnabled)
	prefs.SetInt(KeyAutoRefreshSecs, s.AutoRefreshSecs)
	prefs.SetBool(KeyConfirmBeforeDelete, s.ConfirmBeforeDelete)
	prefs.SetString(KeyThemeVariant, s.ThemeVariant)
	prefs.SetInt(KeyScanWorkers, s.ScanWorkers)
	prefs.SetInt(KeyScanTimeoutSecs, s.ScanTimeoutSecs)
	prefs.SetString(KeyPrivilegeEscalation, s.PrivilegeEscalation)
	prefs.SetString(KeyFontSize, s.FontSize)
	prefs.SetString(KeyLanguage, s.Language)
	prefs.SetString(KeyDensity, s.Density)
	prefs.SetString(KeyAccentColor, s.AccentColor)
	prefs.SetBool(KeyUseCustomFont, s.UseCustomFont)
	prefs.SetBool(KeyReduceMotion, s.ReduceMotion)

	// Light mode colors
	prefs.SetString(KeyLightAccentColor, s.LightAccentColor)
	prefs.SetString(KeyLightBackgroundColor, s.LightBackgroundColor)
	prefs.SetString(KeyLightSurfaceColor, s.LightSurfaceColor)
	prefs.SetString(KeyLightSurfaceRaisedColor, s.LightSurfaceRaisedColor)
	prefs.SetString(KeyLightTextPrimaryColor, s.LightTextPrimaryColor)
	prefs.SetString(KeyLightTextSecondaryColor, s.LightTextSecondaryColor)
	prefs.SetString(KeyLightTextDimColor, s.LightTextDimColor)
	prefs.SetString(KeyLightTextDisabledColor, s.LightTextDisabledColor)
	prefs.SetString(KeyLightBorderColor, s.LightBorderColor)
	prefs.SetString(KeyLightBorderStrongColor, s.LightBorderStrongColor)
	prefs.SetString(KeyLightSuccessColor, s.LightSuccessColor)
	prefs.SetString(KeyLightErrorColor, s.LightErrorColor)
	prefs.SetString(KeyLightWarningColor, s.LightWarningColor)
	prefs.SetString(KeyLightInfoColor, s.LightInfoColor)
	prefs.SetString(KeyLightInputBackgroundColor, s.LightInputBackgroundColor)
	prefs.SetString(KeyLightInputBorderColor, s.LightInputBorderColor)
	prefs.SetString(KeyLightHoverColor, s.LightHoverColor)
	prefs.SetString(KeyLightPressedColor, s.LightPressedColor)
	prefs.SetString(KeyLightSelectionColor, s.LightSelectionColor)
	prefs.SetString(KeyLightHeaderBackgroundColor, s.LightHeaderBackgroundColor)
	prefs.SetString(KeyLightMenuBackgroundColor, s.LightMenuBackgroundColor)
	prefs.SetString(KeyLightScrollbarColor, s.LightScrollbarColor)

	// Dark mode colors
	prefs.SetString(KeyDarkAccentColor, s.DarkAccentColor)
	prefs.SetString(KeyDarkBackgroundColor, s.DarkBackgroundColor)
	prefs.SetString(KeyDarkSurfaceColor, s.DarkSurfaceColor)
	prefs.SetString(KeyDarkSurfaceRaisedColor, s.DarkSurfaceRaisedColor)
	prefs.SetString(KeyDarkTextPrimaryColor, s.DarkTextPrimaryColor)
	prefs.SetString(KeyDarkTextSecondaryColor, s.DarkTextSecondaryColor)
	prefs.SetString(KeyDarkTextDimColor, s.DarkTextDimColor)
	prefs.SetString(KeyDarkTextDisabledColor, s.DarkTextDisabledColor)
	prefs.SetString(KeyDarkBorderColor, s.DarkBorderColor)
	prefs.SetString(KeyDarkBorderStrongColor, s.DarkBorderStrongColor)
	prefs.SetString(KeyDarkSuccessColor, s.DarkSuccessColor)
	prefs.SetString(KeyDarkErrorColor, s.DarkErrorColor)
	prefs.SetString(KeyDarkWarningColor, s.DarkWarningColor)
	prefs.SetString(KeyDarkInfoColor, s.DarkInfoColor)
	prefs.SetString(KeyDarkInputBackgroundColor, s.DarkInputBackgroundColor)
	prefs.SetString(KeyDarkInputBorderColor, s.DarkInputBorderColor)
	prefs.SetString(KeyDarkHoverColor, s.DarkHoverColor)
	prefs.SetString(KeyDarkPressedColor, s.DarkPressedColor)
	prefs.SetString(KeyDarkSelectionColor, s.DarkSelectionColor)
	prefs.SetString(KeyDarkHeaderBackgroundColor, s.DarkHeaderBackgroundColor)
	prefs.SetString(KeyDarkMenuBackgroundColor, s.DarkMenuBackgroundColor)
	prefs.SetString(KeyDarkScrollbarColor, s.DarkScrollbarColor)
}
