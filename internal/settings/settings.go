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
	KeyAccentColor         = "accent_color" // Deprecated: use KeyLightAccentColor and KeyDarkAccentColor

	// Color settings - Light mode
	KeyLightAccentColor         = "light_accent_color"
	KeyLightBackgroundColor     = "light_background_color"
	KeyLightSurfaceColor        = "light_surface_color"
	KeyLightTextPrimaryColor    = "light_text_primary_color"
	KeyLightTextSecondaryColor  = "light_text_secondary_color"
	KeyLightTextDisabledColor   = "light_text_disabled_color"
	KeyLightBorderColor         = "light_border_color"
	KeyLightSuccessColor        = "light_success_color"
	KeyLightErrorColor          = "light_error_color"
	KeyLightWarningColor        = "light_warning_color"
	KeyLightInfoColor           = "light_info_color"
	KeyLightInputBackgroundColor = "light_input_background_color"
	KeyLightInputBorderColor    = "light_input_border_color"
	KeyLightHoverColor          = "light_hover_color"
	KeyLightPressedColor        = "light_pressed_color"
	KeyLightSelectionColor      = "light_selection_color"
	KeyLightHeaderBackgroundColor = "light_header_background_color"
	KeyLightMenuBackgroundColor = "light_menu_background_color"
	KeyLightScrollbarColor      = "light_scrollbar_color"

	// Color settings - Dark mode
	KeyDarkAccentColor         = "dark_accent_color"
	KeyDarkBackgroundColor     = "dark_background_color"
	KeyDarkSurfaceColor        = "dark_surface_color"
	KeyDarkTextPrimaryColor    = "dark_text_primary_color"
	KeyDarkTextSecondaryColor  = "dark_text_secondary_color"
	KeyDarkTextDisabledColor   = "dark_text_disabled_color"
	KeyDarkBorderColor         = "dark_border_color"
	KeyDarkSuccessColor        = "dark_success_color"
	KeyDarkErrorColor          = "dark_error_color"
	KeyDarkWarningColor        = "dark_warning_color"
	KeyDarkInfoColor           = "dark_info_color"
	KeyDarkInputBackgroundColor = "dark_input_background_color"
	KeyDarkInputBorderColor    = "dark_input_border_color"
	KeyDarkHoverColor          = "dark_hover_color"
	KeyDarkPressedColor        = "dark_pressed_color"
	KeyDarkSelectionColor      = "dark_selection_color"
	KeyDarkHeaderBackgroundColor = "dark_header_background_color"
	KeyDarkMenuBackgroundColor = "dark_menu_background_color"
	KeyDarkScrollbarColor      = "dark_scrollbar_color"
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
	DefaultThemeVariant        = "system"
	DefaultScanWorkers         = 100
	DefaultScanTimeoutSecs     = 2
	DefaultPrivilegeEscalation = "none"
	DefaultFontSize            = "normal"
	DefaultUseCustomFont       = true
	DefaultAccentColor         = "#1a73e8" // Deprecated: use DefaultLightAccentColor

	// Light mode color defaults - Material Design inspired
	DefaultLightAccentColor         = "#1a73e8" // Google Blue
	DefaultLightBackgroundColor     = "#f8f9fa" // Very light gray
	DefaultLightSurfaceColor        = "#ffffff" // White
	DefaultLightTextPrimaryColor    = "#202124" // Very dark gray
	DefaultLightTextSecondaryColor  = "#5f6368" // Medium gray
	DefaultLightTextDisabledColor   = "#9e9e9e" // Light gray
	DefaultLightBorderColor         = "#dadce0" // Border gray
	DefaultLightSuccessColor        = "#34a853" // Green
	DefaultLightErrorColor          = "#ea4335" // Red
	DefaultLightWarningColor        = "#fbbc05" // Amber
	DefaultLightInfoColor           = "#4285f4" // Light blue
	DefaultLightInputBackgroundColor = "#ffffff" // White
	DefaultLightInputBorderColor    = "#dadce0" // Border gray
	DefaultLightHoverColor          = "#f1f5fb" // Soft blue tint
	DefaultLightPressedColor        = "#e2edfa" // Deeper blue tint
	DefaultLightSelectionColor      = "#cce0ff" // Light blue selection
	DefaultLightHeaderBackgroundColor = "#ffffff" // White
	DefaultLightMenuBackgroundColor = "#ffffff" // White
	DefaultLightScrollbarColor      = "#a0a0a0" // Medium gray

	// Dark mode color defaults - Material Design inspired
	DefaultDarkAccentColor         = "#8ab4f8" // Lighter blue
	DefaultDarkBackgroundColor     = "#121212" // Very dark gray
	DefaultDarkSurfaceColor        = "#1e1e1e" // Dark gray
	DefaultDarkTextPrimaryColor    = "#e8eaed" // Light gray
	DefaultDarkTextSecondaryColor  = "#9aa0a6" // Medium gray
	DefaultDarkTextDisabledColor   = "#757575" // Darker gray
	DefaultDarkBorderColor         = "#3c4043" // Dark border
	DefaultDarkSuccessColor        = "#81c995" // Light green
	DefaultDarkErrorColor          = "#f28b82" // Light red
	DefaultDarkWarningColor        = "#fdd663" // Light yellow
	DefaultDarkInfoColor           = "#8ab4f8" // Light blue
	DefaultDarkInputBackgroundColor = "#1e1e1e" // Dark gray
	DefaultDarkInputBorderColor    = "#3c4043" // Dark border
	DefaultDarkHoverColor          = "#373737" // Lighter dark gray
	DefaultDarkPressedColor        = "#464646" // Even lighter gray
	DefaultDarkSelectionColor      = "#2d4a6f" // Dark blue selection
	DefaultDarkHeaderBackgroundColor = "#1e1e1e" // Dark gray
	DefaultDarkMenuBackgroundColor = "#282828" // Slightly lighter dark
	DefaultDarkScrollbarColor      = "#5a5a5a" // Medium gray
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
	AccentColor         string // Deprecated: use LightAccentColor and DarkAccentColor
	UseCustomFont       bool

	// Light mode colors
	LightAccentColor         string
	LightBackgroundColor     string
	LightSurfaceColor        string
	LightTextPrimaryColor    string
	LightTextSecondaryColor  string
	LightTextDisabledColor   string
	LightBorderColor         string
	LightSuccessColor        string
	LightErrorColor          string
	LightWarningColor        string
	LightInfoColor           string
	LightInputBackgroundColor string
	LightInputBorderColor    string
	LightHoverColor          string
	LightPressedColor        string
	LightSelectionColor      string
	LightHeaderBackgroundColor string
	LightMenuBackgroundColor string
	LightScrollbarColor      string

	// Dark mode colors
	DarkAccentColor         string
	DarkBackgroundColor     string
	DarkSurfaceColor        string
	DarkTextPrimaryColor    string
	DarkTextSecondaryColor  string
	DarkTextDisabledColor   string
	DarkBorderColor         string
	DarkSuccessColor        string
	DarkErrorColor          string
	DarkWarningColor        string
	DarkInfoColor           string
	DarkInputBackgroundColor string
	DarkInputBorderColor    string
	DarkHoverColor          string
	DarkPressedColor        string
	DarkSelectionColor      string
	DarkHeaderBackgroundColor string
	DarkMenuBackgroundColor string
	DarkScrollbarColor      string
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
		AccentColor:         prefs.StringWithFallback(KeyAccentColor, DefaultAccentColor),
		UseCustomFont:       prefs.BoolWithFallback(KeyUseCustomFont, DefaultUseCustomFont),

		// Light mode colors
		LightAccentColor:         prefs.StringWithFallback(KeyLightAccentColor, DefaultLightAccentColor),
		LightBackgroundColor:     prefs.StringWithFallback(KeyLightBackgroundColor, DefaultLightBackgroundColor),
		LightSurfaceColor:        prefs.StringWithFallback(KeyLightSurfaceColor, DefaultLightSurfaceColor),
		LightTextPrimaryColor:    prefs.StringWithFallback(KeyLightTextPrimaryColor, DefaultLightTextPrimaryColor),
		LightTextSecondaryColor:  prefs.StringWithFallback(KeyLightTextSecondaryColor, DefaultLightTextSecondaryColor),
		LightTextDisabledColor:   prefs.StringWithFallback(KeyLightTextDisabledColor, DefaultLightTextDisabledColor),
		LightBorderColor:         prefs.StringWithFallback(KeyLightBorderColor, DefaultLightBorderColor),
		LightSuccessColor:        prefs.StringWithFallback(KeyLightSuccessColor, DefaultLightSuccessColor),
		LightErrorColor:          prefs.StringWithFallback(KeyLightErrorColor, DefaultLightErrorColor),
		LightWarningColor:        prefs.StringWithFallback(KeyLightWarningColor, DefaultLightWarningColor),
		LightInfoColor:           prefs.StringWithFallback(KeyLightInfoColor, DefaultLightInfoColor),
		LightInputBackgroundColor: prefs.StringWithFallback(KeyLightInputBackgroundColor, DefaultLightInputBackgroundColor),
		LightInputBorderColor:    prefs.StringWithFallback(KeyLightInputBorderColor, DefaultLightInputBorderColor),
		LightHoverColor:          prefs.StringWithFallback(KeyLightHoverColor, DefaultLightHoverColor),
		LightPressedColor:        prefs.StringWithFallback(KeyLightPressedColor, DefaultLightPressedColor),
		LightSelectionColor:      prefs.StringWithFallback(KeyLightSelectionColor, DefaultLightSelectionColor),
		LightHeaderBackgroundColor: prefs.StringWithFallback(KeyLightHeaderBackgroundColor, DefaultLightHeaderBackgroundColor),
		LightMenuBackgroundColor: prefs.StringWithFallback(KeyLightMenuBackgroundColor, DefaultLightMenuBackgroundColor),
		LightScrollbarColor:      prefs.StringWithFallback(KeyLightScrollbarColor, DefaultLightScrollbarColor),

		// Dark mode colors
		DarkAccentColor:         prefs.StringWithFallback(KeyDarkAccentColor, DefaultDarkAccentColor),
		DarkBackgroundColor:     prefs.StringWithFallback(KeyDarkBackgroundColor, DefaultDarkBackgroundColor),
		DarkSurfaceColor:        prefs.StringWithFallback(KeyDarkSurfaceColor, DefaultDarkSurfaceColor),
		DarkTextPrimaryColor:    prefs.StringWithFallback(KeyDarkTextPrimaryColor, DefaultDarkTextPrimaryColor),
		DarkTextSecondaryColor:  prefs.StringWithFallback(KeyDarkTextSecondaryColor, DefaultDarkTextSecondaryColor),
		DarkTextDisabledColor:   prefs.StringWithFallback(KeyDarkTextDisabledColor, DefaultDarkTextDisabledColor),
		DarkBorderColor:         prefs.StringWithFallback(KeyDarkBorderColor, DefaultDarkBorderColor),
		DarkSuccessColor:        prefs.StringWithFallback(KeyDarkSuccessColor, DefaultDarkSuccessColor),
		DarkErrorColor:          prefs.StringWithFallback(KeyDarkErrorColor, DefaultDarkErrorColor),
		DarkWarningColor:        prefs.StringWithFallback(KeyDarkWarningColor, DefaultDarkWarningColor),
		DarkInfoColor:           prefs.StringWithFallback(KeyDarkInfoColor, DefaultDarkInfoColor),
		DarkInputBackgroundColor: prefs.StringWithFallback(KeyDarkInputBackgroundColor, DefaultDarkInputBackgroundColor),
		DarkInputBorderColor:    prefs.StringWithFallback(KeyDarkInputBorderColor, DefaultDarkInputBorderColor),
		DarkHoverColor:          prefs.StringWithFallback(KeyDarkHoverColor, DefaultDarkHoverColor),
		DarkPressedColor:        prefs.StringWithFallback(KeyDarkPressedColor, DefaultDarkPressedColor),
		DarkSelectionColor:      prefs.StringWithFallback(KeyDarkSelectionColor, DefaultDarkSelectionColor),
		DarkHeaderBackgroundColor: prefs.StringWithFallback(KeyDarkHeaderBackgroundColor, DefaultDarkHeaderBackgroundColor),
		DarkMenuBackgroundColor: prefs.StringWithFallback(KeyDarkMenuBackgroundColor, DefaultDarkMenuBackgroundColor),
		DarkScrollbarColor:      prefs.StringWithFallback(KeyDarkScrollbarColor, DefaultDarkScrollbarColor),
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
	prefs.SetString(KeyAccentColor, s.AccentColor)
	prefs.SetBool(KeyUseCustomFont, s.UseCustomFont)

	// Light mode colors
	prefs.SetString(KeyLightAccentColor, s.LightAccentColor)
	prefs.SetString(KeyLightBackgroundColor, s.LightBackgroundColor)
	prefs.SetString(KeyLightSurfaceColor, s.LightSurfaceColor)
	prefs.SetString(KeyLightTextPrimaryColor, s.LightTextPrimaryColor)
	prefs.SetString(KeyLightTextSecondaryColor, s.LightTextSecondaryColor)
	prefs.SetString(KeyLightTextDisabledColor, s.LightTextDisabledColor)
	prefs.SetString(KeyLightBorderColor, s.LightBorderColor)
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
	prefs.SetString(KeyDarkTextPrimaryColor, s.DarkTextPrimaryColor)
	prefs.SetString(KeyDarkTextSecondaryColor, s.DarkTextSecondaryColor)
	prefs.SetString(KeyDarkTextDisabledColor, s.DarkTextDisabledColor)
	prefs.SetString(KeyDarkBorderColor, s.DarkBorderColor)
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
