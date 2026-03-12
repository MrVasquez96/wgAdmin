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
	KeyAccentColor         = "accent_color"
	KeyUseCustomFont       = "use_custom_font"
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
	DefaultAccentColor         = "#1a73e8"
	DefaultUseCustomFont       = true
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
	AccentColor         string
	UseCustomFont       bool
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
}
