package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	"wgAdmin/internal/wgwidget"

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

// createColorButton creates a ColorButton from a hex string value
func (sv *SettingsView) createColorButton(win fyne.Window, currentHex, label string) *wgwidget.ColorButton {
	col, err := wgwidget.HexToColor(currentHex)
	if err != nil {
		col = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}
	return wgwidget.NewColorButton(label, col, win)
}

// resetColorButton resets a ColorButton to a default hex color
func resetColorButton(btn *wgwidget.ColorButton, defaultHex string) {
	col, err := wgwidget.HexToColor(defaultHex)
	if err != nil {
		col = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}
	btn.SetColor(col)
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
	lightAccentBtn := sv.createColorButton(win, sv.current.LightAccentColor, sv.lang.T(lang.Settings_color_light_accent))
	lightBackgroundBtn := sv.createColorButton(win, sv.current.LightBackgroundColor, sv.lang.T(lang.Settings_color_light_background))
	lightSurfaceBtn := sv.createColorButton(win, sv.current.LightSurfaceColor, sv.lang.T(lang.Settings_color_light_surface))
	lightTextPrimaryBtn := sv.createColorButton(win, sv.current.LightTextPrimaryColor, sv.lang.T(lang.Settings_color_light_text_primary))
	lightTextSecondaryBtn := sv.createColorButton(win, sv.current.LightTextSecondaryColor, sv.lang.T(lang.Settings_color_light_text_secondary))
	lightTextDisabledBtn := sv.createColorButton(win, sv.current.LightTextDisabledColor, sv.lang.T(lang.Settings_color_light_text_disabled))
	lightBorderBtn := sv.createColorButton(win, sv.current.LightBorderColor, sv.lang.T(lang.Settings_color_light_border))
	lightSuccessBtn := sv.createColorButton(win, sv.current.LightSuccessColor, sv.lang.T(lang.Settings_color_light_success))
	lightErrorBtn := sv.createColorButton(win, sv.current.LightErrorColor, sv.lang.T(lang.Settings_color_light_error))
	lightWarningBtn := sv.createColorButton(win, sv.current.LightWarningColor, sv.lang.T(lang.Settings_color_light_warning))
	lightInfoBtn := sv.createColorButton(win, sv.current.LightInfoColor, sv.lang.T(lang.Settings_color_light_info))
	lightInputBgBtn := sv.createColorButton(win, sv.current.LightInputBackgroundColor, sv.lang.T(lang.Settings_color_light_input_bg))
	lightInputBorderBtn := sv.createColorButton(win, sv.current.LightInputBorderColor, sv.lang.T(lang.Settings_color_light_input_border))
	lightHoverBtn := sv.createColorButton(win, sv.current.LightHoverColor, sv.lang.T(lang.Settings_color_light_hover))
	lightPressedBtn := sv.createColorButton(win, sv.current.LightPressedColor, sv.lang.T(lang.Settings_color_light_pressed))
	lightSelectionBtn := sv.createColorButton(win, sv.current.LightSelectionColor, sv.lang.T(lang.Settings_color_light_selection))
	lightHeaderBgBtn := sv.createColorButton(win, sv.current.LightHeaderBackgroundColor, sv.lang.T(lang.Settings_color_light_header_bg))
	lightMenuBgBtn := sv.createColorButton(win, sv.current.LightMenuBackgroundColor, sv.lang.T(lang.Settings_color_light_menu_bg))
	lightScrollbarBtn := sv.createColorButton(win, sv.current.LightScrollbarColor, sv.lang.T(lang.Settings_color_light_scrollbar))

	lightColorsForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_accent), lightAccentBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_background), lightBackgroundBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_surface), lightSurfaceBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_primary), lightTextPrimaryBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_secondary), lightTextSecondaryBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_text_disabled), lightTextDisabledBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_border), lightBorderBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_success), lightSuccessBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_error), lightErrorBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_warning), lightWarningBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_info), lightInfoBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_input_bg), lightInputBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_input_border), lightInputBorderBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_hover), lightHoverBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_pressed), lightPressedBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_selection), lightSelectionBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_header_bg), lightHeaderBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_menu_bg), lightMenuBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_light_scrollbar), lightScrollbarBtn),
	)

	// Dark Mode Colors
	darkAccentBtn := sv.createColorButton(win, sv.current.DarkAccentColor, sv.lang.T(lang.Settings_color_dark_accent))
	darkBackgroundBtn := sv.createColorButton(win, sv.current.DarkBackgroundColor, sv.lang.T(lang.Settings_color_dark_background))
	darkSurfaceBtn := sv.createColorButton(win, sv.current.DarkSurfaceColor, sv.lang.T(lang.Settings_color_dark_surface))
	darkTextPrimaryBtn := sv.createColorButton(win, sv.current.DarkTextPrimaryColor, sv.lang.T(lang.Settings_color_dark_text_primary))
	darkTextSecondaryBtn := sv.createColorButton(win, sv.current.DarkTextSecondaryColor, sv.lang.T(lang.Settings_color_dark_text_secondary))
	darkTextDisabledBtn := sv.createColorButton(win, sv.current.DarkTextDisabledColor, sv.lang.T(lang.Settings_color_dark_text_disabled))
	darkBorderBtn := sv.createColorButton(win, sv.current.DarkBorderColor, sv.lang.T(lang.Settings_color_dark_border))
	darkSuccessBtn := sv.createColorButton(win, sv.current.DarkSuccessColor, sv.lang.T(lang.Settings_color_dark_success))
	darkErrorBtn := sv.createColorButton(win, sv.current.DarkErrorColor, sv.lang.T(lang.Settings_color_dark_error))
	darkWarningBtn := sv.createColorButton(win, sv.current.DarkWarningColor, sv.lang.T(lang.Settings_color_dark_warning))
	darkInfoBtn := sv.createColorButton(win, sv.current.DarkInfoColor, sv.lang.T(lang.Settings_color_dark_info))
	darkInputBgBtn := sv.createColorButton(win, sv.current.DarkInputBackgroundColor, sv.lang.T(lang.Settings_color_dark_input_bg))
	darkInputBorderBtn := sv.createColorButton(win, sv.current.DarkInputBorderColor, sv.lang.T(lang.Settings_color_dark_input_border))
	darkHoverBtn := sv.createColorButton(win, sv.current.DarkHoverColor, sv.lang.T(lang.Settings_color_dark_hover))
	darkPressedBtn := sv.createColorButton(win, sv.current.DarkPressedColor, sv.lang.T(lang.Settings_color_dark_pressed))
	darkSelectionBtn := sv.createColorButton(win, sv.current.DarkSelectionColor, sv.lang.T(lang.Settings_color_dark_selection))
	darkHeaderBgBtn := sv.createColorButton(win, sv.current.DarkHeaderBackgroundColor, sv.lang.T(lang.Settings_color_dark_header_bg))
	darkMenuBgBtn := sv.createColorButton(win, sv.current.DarkMenuBackgroundColor, sv.lang.T(lang.Settings_color_dark_menu_bg))
	darkScrollbarBtn := sv.createColorButton(win, sv.current.DarkScrollbarColor, sv.lang.T(lang.Settings_color_dark_scrollbar))

	darkColorsForm := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_accent), darkAccentBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_background), darkBackgroundBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_surface), darkSurfaceBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_primary), darkTextPrimaryBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_secondary), darkTextSecondaryBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_text_disabled), darkTextDisabledBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_border), darkBorderBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_success), darkSuccessBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_error), darkErrorBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_warning), darkWarningBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_info), darkInfoBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_input_bg), darkInputBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_input_border), darkInputBorderBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_hover), darkHoverBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_pressed), darkPressedBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_selection), darkSelectionBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_header_bg), darkHeaderBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_menu_bg), darkMenuBgBtn),
		widget.NewFormItem(sv.lang.T(lang.Settings_color_dark_scrollbar), darkScrollbarBtn),
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
			lightAccentBtn, lightBackgroundBtn, lightSurfaceBtn,
			lightTextPrimaryBtn, lightTextSecondaryBtn, lightTextDisabledBtn,
			lightBorderBtn, lightSuccessBtn, lightErrorBtn, lightWarningBtn, lightInfoBtn,
			lightInputBgBtn, lightInputBorderBtn, lightHoverBtn, lightPressedBtn,
			lightSelectionBtn, lightHeaderBgBtn, lightMenuBgBtn, lightScrollbarBtn,
			// Dark mode colors
			darkAccentBtn, darkBackgroundBtn, darkSurfaceBtn,
			darkTextPrimaryBtn, darkTextSecondaryBtn, darkTextDisabledBtn,
			darkBorderBtn, darkSuccessBtn, darkErrorBtn, darkWarningBtn, darkInfoBtn,
			darkInputBgBtn, darkInputBorderBtn, darkHoverBtn, darkPressedBtn,
			darkSelectionBtn, darkHeaderBgBtn, darkMenuBgBtn, darkScrollbarBtn,
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
			resetColorButton(lightAccentBtn, settings.DefaultLightAccentColor)
			resetColorButton(lightBackgroundBtn, settings.DefaultLightBackgroundColor)
			resetColorButton(lightSurfaceBtn, settings.DefaultLightSurfaceColor)
			resetColorButton(lightTextPrimaryBtn, settings.DefaultLightTextPrimaryColor)
			resetColorButton(lightTextSecondaryBtn, settings.DefaultLightTextSecondaryColor)
			resetColorButton(lightTextDisabledBtn, settings.DefaultLightTextDisabledColor)
			resetColorButton(lightBorderBtn, settings.DefaultLightBorderColor)
			resetColorButton(lightSuccessBtn, settings.DefaultLightSuccessColor)
			resetColorButton(lightErrorBtn, settings.DefaultLightErrorColor)
			resetColorButton(lightWarningBtn, settings.DefaultLightWarningColor)
			resetColorButton(lightInfoBtn, settings.DefaultLightInfoColor)
			resetColorButton(lightInputBgBtn, settings.DefaultLightInputBackgroundColor)
			resetColorButton(lightInputBorderBtn, settings.DefaultLightInputBorderColor)
			resetColorButton(lightHoverBtn, settings.DefaultLightHoverColor)
			resetColorButton(lightPressedBtn, settings.DefaultLightPressedColor)
			resetColorButton(lightSelectionBtn, settings.DefaultLightSelectionColor)
			resetColorButton(lightHeaderBgBtn, settings.DefaultLightHeaderBackgroundColor)
			resetColorButton(lightMenuBgBtn, settings.DefaultLightMenuBackgroundColor)
			resetColorButton(lightScrollbarBtn, settings.DefaultLightScrollbarColor)

			// Dark mode colors
			resetColorButton(darkAccentBtn, settings.DefaultDarkAccentColor)
			resetColorButton(darkBackgroundBtn, settings.DefaultDarkBackgroundColor)
			resetColorButton(darkSurfaceBtn, settings.DefaultDarkSurfaceColor)
			resetColorButton(darkTextPrimaryBtn, settings.DefaultDarkTextPrimaryColor)
			resetColorButton(darkTextSecondaryBtn, settings.DefaultDarkTextSecondaryColor)
			resetColorButton(darkTextDisabledBtn, settings.DefaultDarkTextDisabledColor)
			resetColorButton(darkBorderBtn, settings.DefaultDarkBorderColor)
			resetColorButton(darkSuccessBtn, settings.DefaultDarkSuccessColor)
			resetColorButton(darkErrorBtn, settings.DefaultDarkErrorColor)
			resetColorButton(darkWarningBtn, settings.DefaultDarkWarningColor)
			resetColorButton(darkInfoBtn, settings.DefaultDarkInfoColor)
			resetColorButton(darkInputBgBtn, settings.DefaultDarkInputBackgroundColor)
			resetColorButton(darkInputBorderBtn, settings.DefaultDarkInputBorderColor)
			resetColorButton(darkHoverBtn, settings.DefaultDarkHoverColor)
			resetColorButton(darkPressedBtn, settings.DefaultDarkPressedColor)
			resetColorButton(darkSelectionBtn, settings.DefaultDarkSelectionColor)
			resetColorButton(darkHeaderBgBtn, settings.DefaultDarkHeaderBackgroundColor)
			resetColorButton(darkMenuBgBtn, settings.DefaultDarkMenuBackgroundColor)
			resetColorButton(darkScrollbarBtn, settings.DefaultDarkScrollbarColor)
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
	lightAccentBtn, lightBackgroundBtn, lightSurfaceBtn *wgwidget.ColorButton,
	lightTextPrimaryBtn, lightTextSecondaryBtn, lightTextDisabledBtn *wgwidget.ColorButton,
	lightBorderBtn, lightSuccessBtn, lightErrorBtn, lightWarningBtn, lightInfoBtn *wgwidget.ColorButton,
	lightInputBgBtn, lightInputBorderBtn, lightHoverBtn, lightPressedBtn *wgwidget.ColorButton,
	lightSelectionBtn, lightHeaderBgBtn, lightMenuBgBtn, lightScrollbarBtn *wgwidget.ColorButton,
	// Dark mode colors
	darkAccentBtn, darkBackgroundBtn, darkSurfaceBtn *wgwidget.ColorButton,
	darkTextPrimaryBtn, darkTextSecondaryBtn, darkTextDisabledBtn *wgwidget.ColorButton,
	darkBorderBtn, darkSuccessBtn, darkErrorBtn, darkWarningBtn, darkInfoBtn *wgwidget.ColorButton,
	darkInputBgBtn, darkInputBorderBtn, darkHoverBtn, darkPressedBtn *wgwidget.ColorButton,
	darkSelectionBtn, darkHeaderBgBtn, darkMenuBgBtn, darkScrollbarBtn *wgwidget.ColorButton,
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
		AccentColor:         wgwidget.ColorToHex(lightAccentBtn.Color),
		UseCustomFont:       useCustomFontCheck.Checked,

		// Light mode colors
		LightAccentColor:           wgwidget.ColorToHex(lightAccentBtn.Color),
		LightBackgroundColor:       wgwidget.ColorToHex(lightBackgroundBtn.Color),
		LightSurfaceColor:          wgwidget.ColorToHex(lightSurfaceBtn.Color),
		LightTextPrimaryColor:      wgwidget.ColorToHex(lightTextPrimaryBtn.Color),
		LightTextSecondaryColor:    wgwidget.ColorToHex(lightTextSecondaryBtn.Color),
		LightTextDisabledColor:     wgwidget.ColorToHex(lightTextDisabledBtn.Color),
		LightBorderColor:           wgwidget.ColorToHex(lightBorderBtn.Color),
		LightSuccessColor:          wgwidget.ColorToHex(lightSuccessBtn.Color),
		LightErrorColor:            wgwidget.ColorToHex(lightErrorBtn.Color),
		LightWarningColor:          wgwidget.ColorToHex(lightWarningBtn.Color),
		LightInfoColor:             wgwidget.ColorToHex(lightInfoBtn.Color),
		LightInputBackgroundColor:  wgwidget.ColorToHex(lightInputBgBtn.Color),
		LightInputBorderColor:      wgwidget.ColorToHex(lightInputBorderBtn.Color),
		LightHoverColor:            wgwidget.ColorToHex(lightHoverBtn.Color),
		LightPressedColor:          wgwidget.ColorToHex(lightPressedBtn.Color),
		LightSelectionColor:        wgwidget.ColorToHex(lightSelectionBtn.Color),
		LightHeaderBackgroundColor: wgwidget.ColorToHex(lightHeaderBgBtn.Color),
		LightMenuBackgroundColor:   wgwidget.ColorToHex(lightMenuBgBtn.Color),
		LightScrollbarColor:        wgwidget.ColorToHex(lightScrollbarBtn.Color),

		// Dark mode colors
		DarkAccentColor:           wgwidget.ColorToHex(darkAccentBtn.Color),
		DarkBackgroundColor:       wgwidget.ColorToHex(darkBackgroundBtn.Color),
		DarkSurfaceColor:          wgwidget.ColorToHex(darkSurfaceBtn.Color),
		DarkTextPrimaryColor:      wgwidget.ColorToHex(darkTextPrimaryBtn.Color),
		DarkTextSecondaryColor:    wgwidget.ColorToHex(darkTextSecondaryBtn.Color),
		DarkTextDisabledColor:     wgwidget.ColorToHex(darkTextDisabledBtn.Color),
		DarkBorderColor:           wgwidget.ColorToHex(darkBorderBtn.Color),
		DarkSuccessColor:          wgwidget.ColorToHex(darkSuccessBtn.Color),
		DarkErrorColor:            wgwidget.ColorToHex(darkErrorBtn.Color),
		DarkWarningColor:          wgwidget.ColorToHex(darkWarningBtn.Color),
		DarkInfoColor:             wgwidget.ColorToHex(darkInfoBtn.Color),
		DarkInputBackgroundColor:  wgwidget.ColorToHex(darkInputBgBtn.Color),
		DarkInputBorderColor:      wgwidget.ColorToHex(darkInputBorderBtn.Color),
		DarkHoverColor:            wgwidget.ColorToHex(darkHoverBtn.Color),
		DarkPressedColor:          wgwidget.ColorToHex(darkPressedBtn.Color),
		DarkSelectionColor:        wgwidget.ColorToHex(darkSelectionBtn.Color),
		DarkHeaderBackgroundColor: wgwidget.ColorToHex(darkHeaderBgBtn.Color),
		DarkMenuBackgroundColor:   wgwidget.ColorToHex(darkMenuBgBtn.Color),
		DarkScrollbarColor:        wgwidget.ColorToHex(darkScrollbarBtn.Color),
	}, nil
}
