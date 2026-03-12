package ui

import (
	"fmt"
	"strconv"
	"strings"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SettingsView handles the application settings UI
type SettingsView struct {
	parent  fyne.Window
	current *settings.AppSettings
	onApply func(updated *settings.AppSettings)
}

// NewSettingsView creates a new settings view
func NewSettingsView(parent fyne.Window, current *settings.AppSettings, onApply func(*settings.AppSettings)) *SettingsView {
	return &SettingsView{
		parent:  parent,
		current: current,
		onApply: onApply,
	}
}

// Show displays the settings window
func (sv *SettingsView) Show() {
	win := fyne.CurrentApp().NewWindow("Settings")
	win.Resize(fyne.NewSize(650, 750))

	// --- Paths section ---
	wgPathEntry := widget.NewEntry()
	wgPathEntry.SetText(sv.current.WGConfigPath)
	wgPathEntry.SetPlaceHolder("/etc/wireguard")

	clientDirEntry := widget.NewEntry()
	clientDirEntry.SetText(sv.current.ClientConfigDir)
	clientDirEntry.SetPlaceHolder("clients")

	pathsForm := widget.NewForm(
		widget.NewFormItem("WireGuard Config Path", wgPathEntry),
		widget.NewFormItem("Client Config Directory", clientDirEntry),
	)
	pathsCard := widget.NewCard("Paths", "Directories for configuration files", pathsForm)

	// --- Window section ---
	widthEntry := widget.NewEntry()
	widthEntry.SetText(strconv.Itoa(sv.current.WindowWidth))

	heightEntry := widget.NewEntry()
	heightEntry.SetText(strconv.Itoa(sv.current.WindowHeight))

	fullscreenCheck := widget.NewCheck("Start in fullscreen mode", nil)
	fullscreenCheck.Checked = sv.current.StartFullscreen

	windowForm := widget.NewForm(
		widget.NewFormItem("Width", widthEntry),
		widget.NewFormItem("Height", heightEntry),
		widget.NewFormItem("", fullscreenCheck),
	)
	windowNote := widget.NewLabel("Window size changes take effect on next launch.")
	windowNote.TextStyle = fyne.TextStyle{Italic: true}
	windowCard := widget.NewCard("Window", "Default window dimensions",
		container.NewVBox(windowForm, windowNote))

	// --- Behavior section ---
	autoRefreshCheck := widget.NewCheck("Enable auto-refresh on startup", nil)
	autoRefreshCheck.Checked = sv.current.AutoRefreshEnabled

	refreshSecsEntry := widget.NewEntry()
	refreshSecsEntry.SetText(strconv.Itoa(sv.current.AutoRefreshSecs))

	confirmDeleteCheck := widget.NewCheck("Confirm before deleting tunnels", nil)
	confirmDeleteCheck.Checked = sv.current.ConfirmBeforeDelete

	themeSelect := widget.NewSelect([]string{"system", "light", "dark"}, nil)
	themeSelect.SetSelected(sv.current.ThemeVariant)

	behaviorForm := widget.NewForm(
		widget.NewFormItem("Theme", themeSelect),
		widget.NewFormItem("", autoRefreshCheck),
		widget.NewFormItem("Refresh Interval (s)", refreshSecsEntry),
		widget.NewFormItem("", confirmDeleteCheck),
	)
	behaviorCard := widget.NewCard("Behavior", "", behaviorForm)

	// --- Scanner section ---
	workersEntry := widget.NewEntry()
	workersEntry.SetText(strconv.Itoa(sv.current.ScanWorkers))

	scanTimeoutEntry := widget.NewEntry()
	scanTimeoutEntry.SetText(strconv.Itoa(sv.current.ScanTimeoutSecs))

	scanForm := widget.NewForm(
		widget.NewFormItem("Concurrent Workers", workersEntry),
		widget.NewFormItem("Timeout (seconds)", scanTimeoutEntry),
	)
	scanCard := widget.NewCard("Network Scanner", "", scanForm)

	// --- Privilege escalation section ---
	privOptions := []string{"none", "pkexec", "sudo"}
	if !settings.PkexecAvailable() {
		privOptions = []string{"none", "pkexec (not installed)", "sudo"}
	}
	privSelect := widget.NewSelect(privOptions, nil)
	privSelected := sv.current.PrivilegeEscalation
	if !settings.PkexecAvailable() && privSelected == "pkexec" {
		privSelected = "pkexec (not installed)"
	}
	privSelect.SetSelected(privSelected)

	privNote := widget.NewRichTextFromMarkdown(
		"**none**: Run as current user (may lack permissions)\n\n" +
			"**pkexec**: Uses Polkit graphical prompt at startup\n\n" +
			"**sudo**: Prompts for password in-app at startup\n\n" +
			"_Changes take effect on next launch._")
	privNote.Wrapping = fyne.TextWrapWord

	privCard := widget.NewCard("Privilege Escalation", "How to obtain root for WireGuard commands",
		container.NewVBox(
			widget.NewForm(widget.NewFormItem("Method", privSelect)),
			privNote,
		),
	)

	// --- Appearance section ---
	fontSizeSelect := widget.NewSelect([]string{"small", "normal", "large"}, nil)
	fontSizeSelect.SetSelected(sv.current.FontSize)

	useCustomFontCheck := widget.NewCheck("Use Inter font (disable for system default)", nil)
	useCustomFontCheck.Checked = sv.current.UseCustomFont

	accentColorEntry := widget.NewEntry()
	accentColorEntry.SetText(sv.current.AccentColor)
	accentColorEntry.SetPlaceHolder("#1a73e8")

	appearanceNote := widget.NewLabel("Appearance changes take effect on next launch.")
	appearanceNote.TextStyle = fyne.TextStyle{Italic: true}

	appearanceForm := widget.NewForm(
		widget.NewFormItem("Font Size", fontSizeSelect),
		widget.NewFormItem("", useCustomFontCheck),
		widget.NewFormItem("Accent Color (hex)", accentColorEntry),
	)
	appearanceCard := widget.NewCard("Appearance", "Customize visual styling",
		container.NewVBox(appearanceForm, appearanceNote))

	// --- Buttons ---
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		updated, err := sv.validate(
			wgPathEntry, clientDirEntry,
			widthEntry, heightEntry, fullscreenCheck,
			autoRefreshCheck, refreshSecsEntry, confirmDeleteCheck, themeSelect,
			workersEntry, scanTimeoutEntry,
			privSelect,
			fontSizeSelect, accentColorEntry, useCustomFontCheck,
		)
		if err != nil {
			helpers.ShowError(err, win)
			return
		}
		updated.Save(fyne.CurrentApp().Preferences())
		if sv.onApply != nil {
			sv.onApply(updated)
		}
		win.Close()
	})
	saveBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButtonWithIcon("Reset to Defaults", theme.ViewRefreshIcon(), func() {
		helpers.ShowConfirm("Reset Settings", "Restore all settings to their default values?", func(yes bool) {
			if !yes {
				return
			}
			wgPathEntry.SetText(settings.DefaultWGConfigPath)
			clientDirEntry.SetText(settings.DefaultClientConfigDir)
			widthEntry.SetText(strconv.Itoa(settings.DefaultWindowWidth))
			heightEntry.SetText(strconv.Itoa(settings.DefaultWindowHeight))
			fullscreenCheck.SetChecked(settings.DefaultStartFullscreen)
			autoRefreshCheck.SetChecked(settings.DefaultAutoRefreshEnabled)
			refreshSecsEntry.SetText(strconv.Itoa(settings.DefaultAutoRefreshSecs))
			confirmDeleteCheck.SetChecked(settings.DefaultConfirmBeforeDelete)
			themeSelect.SetSelected(settings.DefaultThemeVariant)
			workersEntry.SetText(strconv.Itoa(settings.DefaultScanWorkers))
			scanTimeoutEntry.SetText(strconv.Itoa(settings.DefaultScanTimeoutSecs))
			privSelect.SetSelected(settings.DefaultPrivilegeEscalation)
			fontSizeSelect.SetSelected(settings.DefaultFontSize)
			accentColorEntry.SetText(settings.DefaultAccentColor)
			useCustomFontCheck.SetChecked(settings.DefaultUseCustomFont)
		}, win)
	})

	cancelBtn := widget.NewButton("Cancel", func() { win.Close() })

	buttons := container.NewHBox(resetBtn, layout.NewSpacer(), cancelBtn, saveBtn)

	content := container.NewBorder(
		nil, container.NewPadded(buttons), nil, nil,
		container.NewVScroll(container.NewVBox(
			container.NewPadded(pathsCard),
			container.NewPadded(windowCard),
			container.NewPadded(appearanceCard),
			container.NewPadded(behaviorCard),
			container.NewPadded(scanCard),
			container.NewPadded(privCard),
		)),
	)

	win.SetContent(container.NewPadded(content))
	win.Show()
}

func (sv *SettingsView) validate(
	wgPathEntry, clientDirEntry *widget.Entry,
	widthEntry, heightEntry *widget.Entry, fullscreenCheck *widget.Check,
	autoRefreshCheck *widget.Check, refreshSecsEntry *widget.Entry, confirmDeleteCheck *widget.Check, themeSelect *widget.Select,
	workersEntry, scanTimeoutEntry *widget.Entry,
	privSelect *widget.Select,
	fontSizeSelect *widget.Select, accentColorEntry *widget.Entry, useCustomFontCheck *widget.Check,
) (*settings.AppSettings, error) {
	if wgPathEntry.Text == "" {
		return nil, fmt.Errorf("WireGuard config path cannot be empty")
	}
	if clientDirEntry.Text == "" {
		return nil, fmt.Errorf("client config directory cannot be empty")
	}

	width, err := strconv.Atoi(widthEntry.Text)
	if err != nil || width < 400 {
		return nil, fmt.Errorf("window width must be a number >= 400")
	}
	height, err := strconv.Atoi(heightEntry.Text)
	if err != nil || height < 300 {
		return nil, fmt.Errorf("window height must be a number >= 300")
	}

	refreshSecs, err := strconv.Atoi(refreshSecsEntry.Text)
	if err != nil || refreshSecs < 1 {
		return nil, fmt.Errorf("refresh interval must be a number >= 1")
	}

	workers, err := strconv.Atoi(workersEntry.Text)
	if err != nil || workers < 1 {
		return nil, fmt.Errorf("scan workers must be a number >= 1")
	}

	scanTimeout, err := strconv.Atoi(scanTimeoutEntry.Text)
	if err != nil || scanTimeout < 1 {
		return nil, fmt.Errorf("scan timeout must be a number >= 1")
	}

	privMethod := privSelect.Selected
	if privMethod == "pkexec (not installed)" {
		privMethod = "pkexec"
	}
	if privMethod != "none" && privMethod != "pkexec" && privMethod != "sudo" {
		return nil, fmt.Errorf("invalid privilege escalation method")
	}

	themeVariant := themeSelect.Selected
	if themeVariant != "system" && themeVariant != "light" && themeVariant != "dark" {
		return nil, fmt.Errorf("invalid theme variant")
	}

	fontSize := fontSizeSelect.Selected
	if fontSize != "small" && fontSize != "normal" && fontSize != "large" {
		return nil, fmt.Errorf("invalid font size")
	}

	accentColor := accentColorEntry.Text
	if accentColor != "" && !isValidHexColor(accentColor) {
		return nil, fmt.Errorf("accent color must be a valid hex color (e.g. #1a73e8)")
	}

	return &settings.AppSettings{
		WGConfigPath:        wgPathEntry.Text,
		ClientConfigDir:     clientDirEntry.Text,
		WindowWidth:         width,
		WindowHeight:        height,
		StartFullscreen:     fullscreenCheck.Checked,
		AutoRefreshEnabled:  autoRefreshCheck.Checked,
		AutoRefreshSecs:     refreshSecs,
		ConfirmBeforeDelete: confirmDeleteCheck.Checked,
		ThemeVariant:        themeVariant,
		ScanWorkers:         workers,
		ScanTimeoutSecs:     scanTimeout,
		PrivilegeEscalation: privMethod,
		FontSize:            fontSize,
		AccentColor:         accentColor,
		UseCustomFont:       useCustomFontCheck.Checked,
	}, nil
}

func isValidHexColor(s string) bool {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
