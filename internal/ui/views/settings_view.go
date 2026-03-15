package ui

import (
	"fmt"
	"strconv"
	"strings"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"

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
	lang    *lang.Language
	onApply func(updated *settings.AppSettings)
}

// createColorEntry creates a new entry widget for hex color input
func (sv *SettingsView) createColorEntry(currentValue, placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetText(currentValue)
	entry.SetPlaceHolder(placeholder)
	return entry
}

// NewSettingsView creates a new settings view
func NewSettingsView(parent fyne.Window, current *settings.AppSettings, onApply func(*settings.AppSettings), language *lang.Language) *SettingsView {
	return &SettingsView{
		parent:  parent,
		current: current,
		lang:    language,
		onApply: onApply,
	}
}

// Show displays the settings window
func (sv *SettingsView) Show() {
	win := fyne.CurrentApp().NewWindow(sv.lang.T(lang.Settings_title))
	win.Resize(fyne.NewSize(900, 900))

	// --- Paths section ---
	wgPathEntry := widget.NewEntry()
	wgPathEntry.SetText(sv.current.WGConfigPath)
	wgPathEntry.SetPlaceHolder(sv.lang.T(lang.Settings_placeholder_wg_path))

	clientDirEntry := widget.NewEntry()
	clientDirEntry.SetText(sv.current.ClientConfigDir)
	clientDirEntry.SetPlaceHolder(sv.lang.T(lang.Settings_placeholder_client_dir))

	pathsForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_wg_path), wgPathEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_client_dir), clientDirEntry),
	)
	pathsCard := widget.NewCard(sv.lang.T(lang.Settings_card_paths), sv.lang.T(lang.Settings_card_paths_desc), pathsForm)

	// --- Window section ---
	widthEntry := widget.NewEntry()
	widthEntry.SetText(strconv.Itoa(sv.current.WindowWidth))

	heightEntry := widget.NewEntry()
	heightEntry.SetText(strconv.Itoa(sv.current.WindowHeight))

	fullscreenCheck := widget.NewCheck(sv.lang.T(lang.Settings_label_fullscreen), nil)
	fullscreenCheck.Checked = sv.current.StartFullscreen

	windowForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_width), widthEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_height), heightEntry),
		widget.NewFormItem("", fullscreenCheck),
	)
	windowNote := widget.NewLabel(sv.lang.T(lang.Settings_note_window))
	windowNote.TextStyle = fyne.TextStyle{Italic: true}
	windowCard := widget.NewCard(sv.lang.T(lang.Settings_card_window), sv.lang.T(lang.Settings_card_window_desc),
		container.NewVBox(windowForm, windowNote))

	// --- Behavior section ---
	autoRefreshCheck := widget.NewCheck(sv.lang.T(lang.Settings_label_auto_refresh), nil)
	autoRefreshCheck.Checked = sv.current.AutoRefreshEnabled

	refreshSecsEntry := widget.NewEntry()
	refreshSecsEntry.SetText(strconv.Itoa(sv.current.AutoRefreshSecs))

	confirmDeleteCheck := widget.NewCheck(sv.lang.T(lang.Settings_label_confirm_delete), nil)
	confirmDeleteCheck.Checked = sv.current.ConfirmBeforeDelete

	themeSelect := widget.NewSelect([]string{"system", "light", "dark"}, nil)
	themeSelect.SetSelected(sv.current.ThemeVariant)

	// Language selector
	langOptions := make([]string, 0, len(lang.AvailableLanguages))
	langCodeToDisplay := make(map[string]string)
	langDisplayToCode := make(map[string]string)
	for code, display := range lang.AvailableLanguages {
		langOptions = append(langOptions, display)
		langCodeToDisplay[code] = display
		langDisplayToCode[display] = code
	}
	langSelect := widget.NewSelect(langOptions, nil)
	langSelect.SetSelected(langCodeToDisplay[sv.current.Language])

	behaviorForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_theme), themeSelect),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_language), langSelect),
		widget.NewFormItem("", autoRefreshCheck),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_refresh_interval), refreshSecsEntry),
		widget.NewFormItem("", confirmDeleteCheck),
	)
	behaviorCard := widget.NewCard(sv.lang.T(lang.Settings_card_behavior), "", behaviorForm)

	// --- Scanner section ---
	workersEntry := widget.NewEntry()
	workersEntry.SetText(strconv.Itoa(sv.current.ScanWorkers))

	scanTimeoutEntry := widget.NewEntry()
	scanTimeoutEntry.SetText(strconv.Itoa(sv.current.ScanTimeoutSecs))

	scanForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_scan_workers), workersEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_scan_timeout), scanTimeoutEntry),
	)
	scanCard := widget.NewCard(sv.lang.T(lang.Settings_card_scanner), "", scanForm)

	// --- Privilege escalation section ---
	privOptions := []string{"none", "pkexec", "sudo"}
	pkexecNotInstalled := sv.lang.T(lang.Settings_privilege_pkexec_unavailable)
	if !settings.PkexecAvailable() {
		privOptions = []string{"none", pkexecNotInstalled, "sudo"}
	}
	privSelect := widget.NewSelect(privOptions, nil)
	privSelected := sv.current.PrivilegeEscalation
	if !settings.PkexecAvailable() && privSelected == "pkexec" {
		privSelected = pkexecNotInstalled
	}
	privSelect.SetSelected(privSelected)

	privNote := widget.NewRichTextFromMarkdown(sv.lang.T(lang.Settings_privilege_desc))
	privNote.Wrapping = fyne.TextWrapWord

	privCard := widget.NewCard(sv.lang.T(lang.Settings_card_privilege), sv.lang.T(lang.Settings_card_privilege_desc),
		container.NewVBox(
			widget.NewForm(widget.NewFormItem(sv.lang.T(lang.Settings_label_privilege_method), privSelect)),
			privNote,
		),
	)

	// --- Appearance section ---
	fontSizeSelect := widget.NewSelect([]string{"small", "normal", "large"}, nil)
	fontSizeSelect.SetSelected(sv.current.FontSize)

	useCustomFontCheck := widget.NewCheck(sv.lang.T(lang.Settings_label_custom_font), nil)
	useCustomFontCheck.Checked = sv.current.UseCustomFont

	appearanceNote := widget.NewLabel(sv.lang.T(lang.Settings_note_appearance))
	appearanceNote.TextStyle = fyne.TextStyle{Italic: true}

	// Light Mode Colors
	lightAccentEntry := sv.createColorEntry(sv.current.LightAccentColor, settings.DefaultLightAccentColor)
	lightBackgroundEntry := sv.createColorEntry(sv.current.LightBackgroundColor, settings.DefaultLightBackgroundColor)
	lightSurfaceEntry := sv.createColorEntry(sv.current.LightSurfaceColor, settings.DefaultLightSurfaceColor)
	lightTextPrimaryEntry := sv.createColorEntry(sv.current.LightTextPrimaryColor, settings.DefaultLightTextPrimaryColor)
	lightTextSecondaryEntry := sv.createColorEntry(sv.current.LightTextSecondaryColor, settings.DefaultLightTextSecondaryColor)
	lightTextDisabledEntry := sv.createColorEntry(sv.current.LightTextDisabledColor, settings.DefaultLightTextDisabledColor)
	lightBorderEntry := sv.createColorEntry(sv.current.LightBorderColor, settings.DefaultLightBorderColor)
	lightSuccessEntry := sv.createColorEntry(sv.current.LightSuccessColor, settings.DefaultLightSuccessColor)
	lightErrorEntry := sv.createColorEntry(sv.current.LightErrorColor, settings.DefaultLightErrorColor)
	lightWarningEntry := sv.createColorEntry(sv.current.LightWarningColor, settings.DefaultLightWarningColor)
	lightInfoEntry := sv.createColorEntry(sv.current.LightInfoColor, settings.DefaultLightInfoColor)
	lightInputBgEntry := sv.createColorEntry(sv.current.LightInputBackgroundColor, settings.DefaultLightInputBackgroundColor)
	lightInputBorderEntry := sv.createColorEntry(sv.current.LightInputBorderColor, settings.DefaultLightInputBorderColor)
	lightHoverEntry := sv.createColorEntry(sv.current.LightHoverColor, settings.DefaultLightHoverColor)
	lightPressedEntry := sv.createColorEntry(sv.current.LightPressedColor, settings.DefaultLightPressedColor)
	lightSelectionEntry := sv.createColorEntry(sv.current.LightSelectionColor, settings.DefaultLightSelectionColor)
	lightHeaderBgEntry := sv.createColorEntry(sv.current.LightHeaderBackgroundColor, settings.DefaultLightHeaderBackgroundColor)
	lightMenuBgEntry := sv.createColorEntry(sv.current.LightMenuBackgroundColor, settings.DefaultLightMenuBackgroundColor)
	lightScrollbarEntry := sv.createColorEntry(sv.current.LightScrollbarColor, settings.DefaultLightScrollbarColor)

	lightColorsForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_accent), lightAccentEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_background), lightBackgroundEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_surface), lightSurfaceEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_primary), lightTextPrimaryEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_secondary), lightTextSecondaryEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_disabled), lightTextDisabledEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_border), lightBorderEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_success), lightSuccessEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_error), lightErrorEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_warning), lightWarningEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_info), lightInfoEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_input_bg), lightInputBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_input_border), lightInputBorderEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_hover), lightHoverEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_pressed), lightPressedEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_selection), lightSelectionEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_header_bg), lightHeaderBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_menu_bg), lightMenuBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_scrollbar), lightScrollbarEntry),
	)

	// Dark Mode Colors
	darkAccentEntry := sv.createColorEntry(sv.current.DarkAccentColor, settings.DefaultDarkAccentColor)
	darkBackgroundEntry := sv.createColorEntry(sv.current.DarkBackgroundColor, settings.DefaultDarkBackgroundColor)
	darkSurfaceEntry := sv.createColorEntry(sv.current.DarkSurfaceColor, settings.DefaultDarkSurfaceColor)
	darkTextPrimaryEntry := sv.createColorEntry(sv.current.DarkTextPrimaryColor, settings.DefaultDarkTextPrimaryColor)
	darkTextSecondaryEntry := sv.createColorEntry(sv.current.DarkTextSecondaryColor, settings.DefaultDarkTextSecondaryColor)
	darkTextDisabledEntry := sv.createColorEntry(sv.current.DarkTextDisabledColor, settings.DefaultDarkTextDisabledColor)
	darkBorderEntry := sv.createColorEntry(sv.current.DarkBorderColor, settings.DefaultDarkBorderColor)
	darkSuccessEntry := sv.createColorEntry(sv.current.DarkSuccessColor, settings.DefaultDarkSuccessColor)
	darkErrorEntry := sv.createColorEntry(sv.current.DarkErrorColor, settings.DefaultDarkErrorColor)
	darkWarningEntry := sv.createColorEntry(sv.current.DarkWarningColor, settings.DefaultDarkWarningColor)
	darkInfoEntry := sv.createColorEntry(sv.current.DarkInfoColor, settings.DefaultDarkInfoColor)
	darkInputBgEntry := sv.createColorEntry(sv.current.DarkInputBackgroundColor, settings.DefaultDarkInputBackgroundColor)
	darkInputBorderEntry := sv.createColorEntry(sv.current.DarkInputBorderColor, settings.DefaultDarkInputBorderColor)
	darkHoverEntry := sv.createColorEntry(sv.current.DarkHoverColor, settings.DefaultDarkHoverColor)
	darkPressedEntry := sv.createColorEntry(sv.current.DarkPressedColor, settings.DefaultDarkPressedColor)
	darkSelectionEntry := sv.createColorEntry(sv.current.DarkSelectionColor, settings.DefaultDarkSelectionColor)
	darkHeaderBgEntry := sv.createColorEntry(sv.current.DarkHeaderBackgroundColor, settings.DefaultDarkHeaderBackgroundColor)
	darkMenuBgEntry := sv.createColorEntry(sv.current.DarkMenuBackgroundColor, settings.DefaultDarkMenuBackgroundColor)
	darkScrollbarEntry := sv.createColorEntry(sv.current.DarkScrollbarColor, settings.DefaultDarkScrollbarColor)

	darkColorsForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_accent), darkAccentEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_background), darkBackgroundEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_surface), darkSurfaceEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_primary), darkTextPrimaryEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_secondary), darkTextSecondaryEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_disabled), darkTextDisabledEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_border), darkBorderEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_success), darkSuccessEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_error), darkErrorEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_warning), darkWarningEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_info), darkInfoEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_input_bg), darkInputBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_input_border), darkInputBorderEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_hover), darkHoverEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_pressed), darkPressedEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_selection), darkSelectionEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_header_bg), darkHeaderBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_menu_bg), darkMenuBgEntry),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_scrollbar), darkScrollbarEntry),
	)

	// Create accordion for color settings with better sizing
	lightColorContainer := container.NewVBox(
		widget.NewLabel(sv.lang.T(lang.Settings_light_colors_label)),
		lightColorsForm,
	)
	darkColorContainer := container.NewVBox(
		widget.NewLabel(sv.lang.T(lang.Settings_dark_colors_label)),
		darkColorsForm,
	)

	colorAccordion := widget.NewAccordion(
		widget.NewAccordionItem(sv.lang.T(lang.Settings_accordion_light_colors), lightColorContainer),
		widget.NewAccordionItem(sv.lang.T(lang.Settings_accordion_dark_colors), darkColorContainer),
	)

	appearanceForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_font_size), fontSizeSelect),
		widget.NewFormItem("", useCustomFontCheck),
	)

	appearanceCard := widget.NewCard(sv.lang.T(lang.Settings_card_appearance), sv.lang.T(lang.Settings_card_appearance_desc),
		container.NewVBox(appearanceForm, colorAccordion, appearanceNote))

	// --- Buttons ---
	saveBtn := widget.NewButtonWithIcon(sv.lang.T(lang.Common_save), theme.DocumentSaveIcon(), func() {
		updated, err := sv.validate(
			wgPathEntry, clientDirEntry,
			widthEntry, heightEntry, fullscreenCheck,
			autoRefreshCheck, refreshSecsEntry, confirmDeleteCheck, themeSelect, langSelect,
			workersEntry, scanTimeoutEntry,
			privSelect,
			fontSizeSelect, useCustomFontCheck,
			// Light mode colors
			lightAccentEntry, lightBackgroundEntry, lightSurfaceEntry,
			lightTextPrimaryEntry, lightTextSecondaryEntry, lightTextDisabledEntry,
			lightBorderEntry, lightSuccessEntry, lightErrorEntry, lightWarningEntry, lightInfoEntry,
			lightInputBgEntry, lightInputBorderEntry, lightHoverEntry, lightPressedEntry,
			lightSelectionEntry, lightHeaderBgEntry, lightMenuBgEntry, lightScrollbarEntry,
			// Dark mode colors
			darkAccentEntry, darkBackgroundEntry, darkSurfaceEntry,
			darkTextPrimaryEntry, darkTextSecondaryEntry, darkTextDisabledEntry,
			darkBorderEntry, darkSuccessEntry, darkErrorEntry, darkWarningEntry, darkInfoEntry,
			darkInputBgEntry, darkInputBorderEntry, darkHoverEntry, darkPressedEntry,
			darkSelectionEntry, darkHeaderBgEntry, darkMenuBgEntry, darkScrollbarEntry,
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

	resetBtn := widget.NewButtonWithIcon(sv.lang.T(lang.Settings_btn_reset), theme.ViewRefreshIcon(), func() {
		helpers.ShowConfirm(sv.lang.T(lang.Settings_reset_confirm_title), sv.lang.T(lang.Settings_reset_confirm_msg), func(yes bool) {
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
			langSelect.SetSelected(langCodeToDisplay[settings.DefaultLanguage])
			workersEntry.SetText(strconv.Itoa(settings.DefaultScanWorkers))
			scanTimeoutEntry.SetText(strconv.Itoa(settings.DefaultScanTimeoutSecs))
			privSelect.SetSelected(settings.DefaultPrivilegeEscalation)
			fontSizeSelect.SetSelected(settings.DefaultFontSize)
			useCustomFontCheck.SetChecked(settings.DefaultUseCustomFont)

			// Light mode colors
			lightAccentEntry.SetText(settings.DefaultLightAccentColor)
			lightBackgroundEntry.SetText(settings.DefaultLightBackgroundColor)
			lightSurfaceEntry.SetText(settings.DefaultLightSurfaceColor)
			lightTextPrimaryEntry.SetText(settings.DefaultLightTextPrimaryColor)
			lightTextSecondaryEntry.SetText(settings.DefaultLightTextSecondaryColor)
			lightTextDisabledEntry.SetText(settings.DefaultLightTextDisabledColor)
			lightBorderEntry.SetText(settings.DefaultLightBorderColor)
			lightSuccessEntry.SetText(settings.DefaultLightSuccessColor)
			lightErrorEntry.SetText(settings.DefaultLightErrorColor)
			lightWarningEntry.SetText(settings.DefaultLightWarningColor)
			lightInfoEntry.SetText(settings.DefaultLightInfoColor)
			lightInputBgEntry.SetText(settings.DefaultLightInputBackgroundColor)
			lightInputBorderEntry.SetText(settings.DefaultLightInputBorderColor)
			lightHoverEntry.SetText(settings.DefaultLightHoverColor)
			lightPressedEntry.SetText(settings.DefaultLightPressedColor)
			lightSelectionEntry.SetText(settings.DefaultLightSelectionColor)
			lightHeaderBgEntry.SetText(settings.DefaultLightHeaderBackgroundColor)
			lightMenuBgEntry.SetText(settings.DefaultLightMenuBackgroundColor)
			lightScrollbarEntry.SetText(settings.DefaultLightScrollbarColor)

			// Dark mode colors
			darkAccentEntry.SetText(settings.DefaultDarkAccentColor)
			darkBackgroundEntry.SetText(settings.DefaultDarkBackgroundColor)
			darkSurfaceEntry.SetText(settings.DefaultDarkSurfaceColor)
			darkTextPrimaryEntry.SetText(settings.DefaultDarkTextPrimaryColor)
			darkTextSecondaryEntry.SetText(settings.DefaultDarkTextSecondaryColor)
			darkTextDisabledEntry.SetText(settings.DefaultDarkTextDisabledColor)
			darkBorderEntry.SetText(settings.DefaultDarkBorderColor)
			darkSuccessEntry.SetText(settings.DefaultDarkSuccessColor)
			darkErrorEntry.SetText(settings.DefaultDarkErrorColor)
			darkWarningEntry.SetText(settings.DefaultDarkWarningColor)
			darkInfoEntry.SetText(settings.DefaultDarkInfoColor)
			darkInputBgEntry.SetText(settings.DefaultDarkInputBackgroundColor)
			darkInputBorderEntry.SetText(settings.DefaultDarkInputBorderColor)
			darkHoverEntry.SetText(settings.DefaultDarkHoverColor)
			darkPressedEntry.SetText(settings.DefaultDarkPressedColor)
			darkSelectionEntry.SetText(settings.DefaultDarkSelectionColor)
			darkHeaderBgEntry.SetText(settings.DefaultDarkHeaderBackgroundColor)
			darkMenuBgEntry.SetText(settings.DefaultDarkMenuBackgroundColor)
			darkScrollbarEntry.SetText(settings.DefaultDarkScrollbarColor)
		}, win)
	})

	cancelBtn := widget.NewButton(sv.lang.T(lang.Common_cancel), func() { win.Close() })

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
	autoRefreshCheck *widget.Check, refreshSecsEntry *widget.Entry, confirmDeleteCheck *widget.Check, themeSelect, langSelect *widget.Select,
	workersEntry, scanTimeoutEntry *widget.Entry,
	privSelect *widget.Select,
	fontSizeSelect *widget.Select, useCustomFontCheck *widget.Check,
	// Light mode colors
	lightAccentEntry, lightBackgroundEntry, lightSurfaceEntry *widget.Entry,
	lightTextPrimaryEntry, lightTextSecondaryEntry, lightTextDisabledEntry *widget.Entry,
	lightBorderEntry, lightSuccessEntry, lightErrorEntry, lightWarningEntry, lightInfoEntry *widget.Entry,
	lightInputBgEntry, lightInputBorderEntry, lightHoverEntry, lightPressedEntry *widget.Entry,
	lightSelectionEntry, lightHeaderBgEntry, lightMenuBgEntry, lightScrollbarEntry *widget.Entry,
	// Dark mode colors
	darkAccentEntry, darkBackgroundEntry, darkSurfaceEntry *widget.Entry,
	darkTextPrimaryEntry, darkTextSecondaryEntry, darkTextDisabledEntry *widget.Entry,
	darkBorderEntry, darkSuccessEntry, darkErrorEntry, darkWarningEntry, darkInfoEntry *widget.Entry,
	darkInputBgEntry, darkInputBorderEntry, darkHoverEntry, darkPressedEntry *widget.Entry,
	darkSelectionEntry, darkHeaderBgEntry, darkMenuBgEntry, darkScrollbarEntry *widget.Entry,
) (*settings.AppSettings, error) {
	if wgPathEntry.Text == "" {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_wg_path_empty))
	}
	if clientDirEntry.Text == "" {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_client_dir_empty))
	}

	width, err := strconv.Atoi(widthEntry.Text)
	if err != nil || width < 400 {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_width_invalid))
	}
	height, err := strconv.Atoi(heightEntry.Text)
	if err != nil || height < 300 {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_height_invalid))
	}

	refreshSecs, err := strconv.Atoi(refreshSecsEntry.Text)
	if err != nil || refreshSecs < 1 {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_refresh_invalid))
	}

	workers, err := strconv.Atoi(workersEntry.Text)
	if err != nil || workers < 1 {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_workers_invalid))
	}

	scanTimeout, err := strconv.Atoi(scanTimeoutEntry.Text)
	if err != nil || scanTimeout < 1 {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_timeout_invalid))
	}

	privMethod := privSelect.Selected
	if privMethod == "pkexec (not installed)" {
		privMethod = "pkexec"
	}
	if privMethod != "none" && privMethod != "pkexec" && privMethod != "sudo" {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_privilege_invalid))
	}

	themeVariant := themeSelect.Selected
	if themeVariant != "system" && themeVariant != "light" && themeVariant != "dark" {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_theme_invalid))
	}

	fontSize := fontSizeSelect.Selected
	if fontSize != "small" && fontSize != "normal" && fontSize != "large" {
		return nil, fmt.Errorf(sv.lang.T(lang.Error_fontsize_invalid))
	}

	// Convert language display name back to code
	langCode := ""
	for code, display := range lang.AvailableLanguages {
		if display == langSelect.Selected {
			langCode = code
			break
		}
	}
	if langCode == "" {
		langCode = settings.DefaultLanguage
	}

	// Validate all color entries
	colorEntries := map[string]*widget.Entry{
		"light accent color":         lightAccentEntry,
		"light background color":     lightBackgroundEntry,
		"light surface color":        lightSurfaceEntry,
		"light text primary color":   lightTextPrimaryEntry,
		"light text secondary color": lightTextSecondaryEntry,
		"light text disabled color":  lightTextDisabledEntry,
		"light border color":         lightBorderEntry,
		"light success color":        lightSuccessEntry,
		"light error color":          lightErrorEntry,
		"light warning color":        lightWarningEntry,
		"light info color":           lightInfoEntry,
		"light input background":     lightInputBgEntry,
		"light input border":         lightInputBorderEntry,
		"light hover color":          lightHoverEntry,
		"light pressed color":        lightPressedEntry,
		"light selection color":      lightSelectionEntry,
		"light header background":    lightHeaderBgEntry,
		"light menu background":      lightMenuBgEntry,
		"light scrollbar color":      lightScrollbarEntry,
		"dark accent color":          darkAccentEntry,
		"dark background color":      darkBackgroundEntry,
		"dark surface color":         darkSurfaceEntry,
		"dark text primary color":    darkTextPrimaryEntry,
		"dark text secondary color":  darkTextSecondaryEntry,
		"dark text disabled color":   darkTextDisabledEntry,
		"dark border color":          darkBorderEntry,
		"dark success color":         darkSuccessEntry,
		"dark error color":           darkErrorEntry,
		"dark warning color":         darkWarningEntry,
		"dark info color":            darkInfoEntry,
		"dark input background":      darkInputBgEntry,
		"dark input border":          darkInputBorderEntry,
		"dark hover color":           darkHoverEntry,
		"dark pressed color":         darkPressedEntry,
		"dark selection color":       darkSelectionEntry,
		"dark header background":     darkHeaderBgEntry,
		"dark menu background":       darkMenuBgEntry,
		"dark scrollbar color":       darkScrollbarEntry,
	}

	for name, entry := range colorEntries {
		if entry.Text != "" && !isValidHexColor(entry.Text) {
			return nil, fmt.Errorf(sv.lang.Tf(lang.Error_hex_color_invalid, name))
		}
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
		Language:            langCode,
		AccentColor:         lightAccentEntry.Text, // Keep for backward compatibility
		UseCustomFont:       useCustomFontCheck.Checked,

		// Light mode colors
		LightAccentColor:           lightAccentEntry.Text,
		LightBackgroundColor:       lightBackgroundEntry.Text,
		LightSurfaceColor:          lightSurfaceEntry.Text,
		LightTextPrimaryColor:      lightTextPrimaryEntry.Text,
		LightTextSecondaryColor:    lightTextSecondaryEntry.Text,
		LightTextDisabledColor:     lightTextDisabledEntry.Text,
		LightBorderColor:           lightBorderEntry.Text,
		LightSuccessColor:          lightSuccessEntry.Text,
		LightErrorColor:            lightErrorEntry.Text,
		LightWarningColor:          lightWarningEntry.Text,
		LightInfoColor:             lightInfoEntry.Text,
		LightInputBackgroundColor:  lightInputBgEntry.Text,
		LightInputBorderColor:      lightInputBorderEntry.Text,
		LightHoverColor:            lightHoverEntry.Text,
		LightPressedColor:          lightPressedEntry.Text,
		LightSelectionColor:        lightSelectionEntry.Text,
		LightHeaderBackgroundColor: lightHeaderBgEntry.Text,
		LightMenuBackgroundColor:   lightMenuBgEntry.Text,
		LightScrollbarColor:        lightScrollbarEntry.Text,

		// Dark mode colors
		DarkAccentColor:           darkAccentEntry.Text,
		DarkBackgroundColor:       darkBackgroundEntry.Text,
		DarkSurfaceColor:          darkSurfaceEntry.Text,
		DarkTextPrimaryColor:      darkTextPrimaryEntry.Text,
		DarkTextSecondaryColor:    darkTextSecondaryEntry.Text,
		DarkTextDisabledColor:     darkTextDisabledEntry.Text,
		DarkBorderColor:           darkBorderEntry.Text,
		DarkSuccessColor:          darkSuccessEntry.Text,
		DarkErrorColor:            darkErrorEntry.Text,
		DarkWarningColor:          darkWarningEntry.Text,
		DarkInfoColor:             darkInfoEntry.Text,
		DarkInputBackgroundColor:  darkInputBgEntry.Text,
		DarkInputBorderColor:      darkInputBorderEntry.Text,
		DarkHoverColor:            darkHoverEntry.Text,
		DarkPressedColor:          darkPressedEntry.Text,
		DarkSelectionColor:        darkSelectionEntry.Text,
		DarkHeaderBackgroundColor: darkHeaderBgEntry.Text,
		DarkMenuBackgroundColor:   darkMenuBgEntry.Text,
		DarkScrollbarColor:        darkScrollbarEntry.Text,
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
