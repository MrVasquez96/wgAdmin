package ui

import (
	"errors"
	"image/color"
	"os"
	"strconv"

	"wgAdmin/internal/settings"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// SettingsView is the preferences window: a category list on the left and one
// page at a time on the right, instead of a single scroll containing every
// control in the application.
type SettingsView struct {
	parent  fyne.Window
	current *settings.AppSettings
	lang    *lang.Language
	onApply func(updated *settings.AppSettings)

	win fyne.Window

	// colorButtons is keyed by settings key, which lets presets, reset and
	// save all address the swatches uniformly. The previous version threaded
	// 38 positional parameters through validate() and repeated the same list
	// again in the reset handler.
	colorButtons map[string]*wgwidget.ColorButton

	// Simple controls, held so validate() can read them.
	wgPathEntry      *widget.Entry
	clientDirEntry   *widget.Entry
	widthEntry       *widget.Entry
	heightEntry      *widget.Entry
	fullscreenCheck  *widget.Check
	autoRefreshCheck *widget.Check
	refreshSecsEntry *widget.Entry
	confirmDelCheck  *widget.Check
	themeSelect      *widget.Select
	langSelect       *widget.Select
	densitySelect    *widget.Select
	fontSizeSelect   *widget.Select
	customFontCheck  *widget.Check
	workersEntry     *widget.Entry
	scanTimeoutEntry *widget.Entry
	privSelect       *widget.Select

	langCodeToDisplay map[string]string
}

// colorToken describes one editable colour, tying the settings key to its
// label and default so creation, reset and save stay in step.
type colorToken struct {
	key      string
	labelKey string
	def      string
}

// colorGroup is a titled set of tokens within one theme variant.
type colorGroup struct {
	titleKey string
	tokens   []colorToken
}

// darkColorGroups and lightColorGroups list every customisable colour, grouped
// so the swatch wall reads as four short lists rather than one list of 19.
func darkColorGroups() []colorGroup {
	return []colorGroup{
		{lang.Settings_group_colors_core, []colorToken{
			{settings.KeyDarkAccentColor, lang.Settings_color_dark_accent, settings.DefaultDarkAccentColor},
			{settings.KeyDarkBackgroundColor, lang.Settings_color_dark_background, settings.DefaultDarkBackgroundColor},
			{settings.KeyDarkSurfaceColor, lang.Settings_color_dark_surface, settings.DefaultDarkSurfaceColor},
			{settings.KeyDarkSurfaceRaisedColor, lang.Settings_color_dark_surface_raised, settings.DefaultDarkSurfaceRaisedColor},
			{settings.KeyDarkHeaderBackgroundColor, lang.Settings_color_dark_header_bg, settings.DefaultDarkHeaderBackgroundColor},
			{settings.KeyDarkMenuBackgroundColor, lang.Settings_color_dark_menu_bg, settings.DefaultDarkMenuBackgroundColor},
		}},
		{lang.Settings_group_colors_text, []colorToken{
			{settings.KeyDarkTextPrimaryColor, lang.Settings_color_dark_text_primary, settings.DefaultDarkTextPrimaryColor},
			{settings.KeyDarkTextSecondaryColor, lang.Settings_color_dark_text_secondary, settings.DefaultDarkTextSecondaryColor},
			{settings.KeyDarkTextDimColor, lang.Settings_color_dark_text_dim, settings.DefaultDarkTextDimColor},
			{settings.KeyDarkTextDisabledColor, lang.Settings_color_dark_text_disabled, settings.DefaultDarkTextDisabledColor},
			{settings.KeyDarkBorderColor, lang.Settings_color_dark_border, settings.DefaultDarkBorderColor},
			{settings.KeyDarkBorderStrongColor, lang.Settings_color_dark_border_strong, settings.DefaultDarkBorderStrongColor},
		}},
		{lang.Settings_group_colors_status, []colorToken{
			{settings.KeyDarkSuccessColor, lang.Settings_color_dark_success, settings.DefaultDarkSuccessColor},
			{settings.KeyDarkErrorColor, lang.Settings_color_dark_error, settings.DefaultDarkErrorColor},
			{settings.KeyDarkWarningColor, lang.Settings_color_dark_warning, settings.DefaultDarkWarningColor},
			{settings.KeyDarkInfoColor, lang.Settings_color_dark_info, settings.DefaultDarkInfoColor},
		}},
		{lang.Settings_group_colors_controls, []colorToken{
			{settings.KeyDarkInputBackgroundColor, lang.Settings_color_dark_input_bg, settings.DefaultDarkInputBackgroundColor},
			{settings.KeyDarkInputBorderColor, lang.Settings_color_dark_input_border, settings.DefaultDarkInputBorderColor},
			{settings.KeyDarkHoverColor, lang.Settings_color_dark_hover, settings.DefaultDarkHoverColor},
			{settings.KeyDarkPressedColor, lang.Settings_color_dark_pressed, settings.DefaultDarkPressedColor},
			{settings.KeyDarkSelectionColor, lang.Settings_color_dark_selection, settings.DefaultDarkSelectionColor},
			{settings.KeyDarkScrollbarColor, lang.Settings_color_dark_scrollbar, settings.DefaultDarkScrollbarColor},
		}},
	}
}

func lightColorGroups() []colorGroup {
	return []colorGroup{
		{lang.Settings_group_colors_core, []colorToken{
			{settings.KeyLightAccentColor, lang.Settings_color_light_accent, settings.DefaultLightAccentColor},
			{settings.KeyLightBackgroundColor, lang.Settings_color_light_background, settings.DefaultLightBackgroundColor},
			{settings.KeyLightSurfaceColor, lang.Settings_color_light_surface, settings.DefaultLightSurfaceColor},
			{settings.KeyLightSurfaceRaisedColor, lang.Settings_color_light_surface_raised, settings.DefaultLightSurfaceRaisedColor},
			{settings.KeyLightHeaderBackgroundColor, lang.Settings_color_light_header_bg, settings.DefaultLightHeaderBackgroundColor},
			{settings.KeyLightMenuBackgroundColor, lang.Settings_color_light_menu_bg, settings.DefaultLightMenuBackgroundColor},
		}},
		{lang.Settings_group_colors_text, []colorToken{
			{settings.KeyLightTextPrimaryColor, lang.Settings_color_light_text_primary, settings.DefaultLightTextPrimaryColor},
			{settings.KeyLightTextSecondaryColor, lang.Settings_color_light_text_secondary, settings.DefaultLightTextSecondaryColor},
			{settings.KeyLightTextDimColor, lang.Settings_color_light_text_dim, settings.DefaultLightTextDimColor},
			{settings.KeyLightTextDisabledColor, lang.Settings_color_light_text_disabled, settings.DefaultLightTextDisabledColor},
			{settings.KeyLightBorderColor, lang.Settings_color_light_border, settings.DefaultLightBorderColor},
			{settings.KeyLightBorderStrongColor, lang.Settings_color_light_border_strong, settings.DefaultLightBorderStrongColor},
		}},
		{lang.Settings_group_colors_status, []colorToken{
			{settings.KeyLightSuccessColor, lang.Settings_color_light_success, settings.DefaultLightSuccessColor},
			{settings.KeyLightErrorColor, lang.Settings_color_light_error, settings.DefaultLightErrorColor},
			{settings.KeyLightWarningColor, lang.Settings_color_light_warning, settings.DefaultLightWarningColor},
			{settings.KeyLightInfoColor, lang.Settings_color_light_info, settings.DefaultLightInfoColor},
		}},
		{lang.Settings_group_colors_controls, []colorToken{
			{settings.KeyLightInputBackgroundColor, lang.Settings_color_light_input_bg, settings.DefaultLightInputBackgroundColor},
			{settings.KeyLightInputBorderColor, lang.Settings_color_light_input_border, settings.DefaultLightInputBorderColor},
			{settings.KeyLightHoverColor, lang.Settings_color_light_hover, settings.DefaultLightHoverColor},
			{settings.KeyLightPressedColor, lang.Settings_color_light_pressed, settings.DefaultLightPressedColor},
			{settings.KeyLightSelectionColor, lang.Settings_color_light_selection, settings.DefaultLightSelectionColor},
			{settings.KeyLightScrollbarColor, lang.Settings_color_light_scrollbar, settings.DefaultLightScrollbarColor},
		}},
	}
}

// currentColor reads the saved value for a token from the live settings.
func (sv *SettingsView) currentColor(key string) string {
	s := sv.current
	switch key {
	case settings.KeyDarkAccentColor:
		return s.DarkAccentColor
	case settings.KeyDarkBackgroundColor:
		return s.DarkBackgroundColor
	case settings.KeyDarkSurfaceColor:
		return s.DarkSurfaceColor
	case settings.KeyDarkSurfaceRaisedColor:
		return s.DarkSurfaceRaisedColor
	case settings.KeyDarkHeaderBackgroundColor:
		return s.DarkHeaderBackgroundColor
	case settings.KeyDarkMenuBackgroundColor:
		return s.DarkMenuBackgroundColor
	case settings.KeyDarkTextPrimaryColor:
		return s.DarkTextPrimaryColor
	case settings.KeyDarkTextSecondaryColor:
		return s.DarkTextSecondaryColor
	case settings.KeyDarkTextDimColor:
		return s.DarkTextDimColor
	case settings.KeyDarkTextDisabledColor:
		return s.DarkTextDisabledColor
	case settings.KeyDarkBorderColor:
		return s.DarkBorderColor
	case settings.KeyDarkBorderStrongColor:
		return s.DarkBorderStrongColor
	case settings.KeyDarkSuccessColor:
		return s.DarkSuccessColor
	case settings.KeyDarkErrorColor:
		return s.DarkErrorColor
	case settings.KeyDarkWarningColor:
		return s.DarkWarningColor
	case settings.KeyDarkInfoColor:
		return s.DarkInfoColor
	case settings.KeyDarkInputBackgroundColor:
		return s.DarkInputBackgroundColor
	case settings.KeyDarkInputBorderColor:
		return s.DarkInputBorderColor
	case settings.KeyDarkHoverColor:
		return s.DarkHoverColor
	case settings.KeyDarkPressedColor:
		return s.DarkPressedColor
	case settings.KeyDarkSelectionColor:
		return s.DarkSelectionColor
	case settings.KeyDarkScrollbarColor:
		return s.DarkScrollbarColor

	case settings.KeyLightAccentColor:
		return s.LightAccentColor
	case settings.KeyLightBackgroundColor:
		return s.LightBackgroundColor
	case settings.KeyLightSurfaceColor:
		return s.LightSurfaceColor
	case settings.KeyLightSurfaceRaisedColor:
		return s.LightSurfaceRaisedColor
	case settings.KeyLightHeaderBackgroundColor:
		return s.LightHeaderBackgroundColor
	case settings.KeyLightMenuBackgroundColor:
		return s.LightMenuBackgroundColor
	case settings.KeyLightTextPrimaryColor:
		return s.LightTextPrimaryColor
	case settings.KeyLightTextSecondaryColor:
		return s.LightTextSecondaryColor
	case settings.KeyLightTextDimColor:
		return s.LightTextDimColor
	case settings.KeyLightTextDisabledColor:
		return s.LightTextDisabledColor
	case settings.KeyLightBorderColor:
		return s.LightBorderColor
	case settings.KeyLightBorderStrongColor:
		return s.LightBorderStrongColor
	case settings.KeyLightSuccessColor:
		return s.LightSuccessColor
	case settings.KeyLightErrorColor:
		return s.LightErrorColor
	case settings.KeyLightWarningColor:
		return s.LightWarningColor
	case settings.KeyLightInfoColor:
		return s.LightInfoColor
	case settings.KeyLightInputBackgroundColor:
		return s.LightInputBackgroundColor
	case settings.KeyLightInputBorderColor:
		return s.LightInputBorderColor
	case settings.KeyLightHoverColor:
		return s.LightHoverColor
	case settings.KeyLightPressedColor:
		return s.LightPressedColor
	case settings.KeyLightSelectionColor:
		return s.LightSelectionColor
	case settings.KeyLightScrollbarColor:
		return s.LightScrollbarColor
	}
	return ""
}

// NewSettingsView creates a new settings view
func NewSettingsView(parent fyne.Window, current *settings.AppSettings, onApply func(*settings.AppSettings), language *lang.Language) *SettingsView {
	return &SettingsView{
		parent:       parent,
		current:      current,
		lang:         language,
		onApply:      onApply,
		colorButtons: map[string]*wgwidget.ColorButton{},
	}
}

// hex returns the chosen colour for a token, falling back to its default.
func (sv *SettingsView) hex(key, def string) string {
	if btn, ok := sv.colorButtons[key]; ok {
		return wgwidget.ColorToHex(btn.Color)
	}
	return def
}

// button creates (once) and returns the swatch for a token.
func (sv *SettingsView) button(t colorToken) *wgwidget.ColorButton {
	if btn, ok := sv.colorButtons[t.key]; ok {
		return btn
	}
	value := sv.currentColor(t.key)
	if value == "" {
		value = t.def
	}
	col, err := wgwidget.HexToColor(value)
	if err != nil {
		col = color.RGBA{R: 128, G: 128, B: 128, A: 255}
	}
	btn := wgwidget.NewColorButton(sv.lang.T(t.labelKey), col, sv.win)
	sv.colorButtons[t.key] = btn
	return btn
}

// Show displays the settings window.
func (sv *SettingsView) Show() {
	sv.win = fyne.CurrentApp().NewWindow(sv.lang.T(lang.Settings_title))
	sv.win.Resize(fyne.NewSize(940, 720))

	pages := []struct {
		titleKey string
		build    func() fyne.CanvasObject
	}{
		{lang.Settings_nav_paths, sv.pagePaths},
		{lang.Settings_nav_appearance, sv.pageAppearance},
		{lang.Settings_nav_behavior, sv.pageBehavior},
		{lang.Settings_nav_scanner, sv.pageScanner},
		{lang.Settings_nav_privileges, sv.pagePrivileges},
		{lang.Settings_nav_about, sv.pageAbout},
	}

	// Build every page up front: the controls must exist for validate() to
	// read them even if the user never visits that page.
	content := container.NewStack()
	built := make([]fyne.CanvasObject, len(pages))
	for i, p := range pages {
		built[i] = container.NewVScroll(container.New(
			layout.NewCustomPaddedLayout(wgtheme.SpaceLG, wgtheme.SpaceXL, wgtheme.SpaceLG, wgtheme.SpaceLG),
			p.build()))
	}

	nav := widget.NewList(
		func() int { return len(pages) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(sv.lang.T(pages[id].titleKey))
		},
	)
	nav.OnSelected = func(id widget.ListItemID) {
		content.Objects = []fyne.CanvasObject{built[id]}
		content.Refresh()
	}

	split := container.NewHSplit(wgwidget.RaisedPanel(nav), content)
	split.SetOffset(0.22)

	saveBtn := widget.NewButtonWithIcon(sv.lang.T(lang.Common_save), wgtheme.IconImport, sv.save)
	saveBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButtonWithIcon(sv.lang.T(lang.Settings_btn_reset), wgtheme.IconRefresh, sv.reset)
	cancelBtn := widget.NewButton(sv.lang.T(lang.Common_cancel), func() { sv.win.Close() })

	buttons := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		resetBtn, layout.NewSpacer(), cancelBtn, saveBtn)

	sv.win.SetContent(container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG),
		container.NewBorder(nil,
			container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceMD, 0, 0, 0), buttons),
			nil, nil, split)))

	nav.Select(0)
	sv.win.Show()
}

func (sv *SettingsView) pagePaths() fyne.CanvasObject {
	sv.wgPathEntry = widget.NewEntry()
	sv.wgPathEntry.SetText(sv.current.WGConfigPath)
	sv.wgPathEntry.SetPlaceHolder(sv.lang.T(lang.Settings_placeholder_wg_path))

	sv.clientDirEntry = widget.NewEntry()
	sv.clientDirEntry.SetText(sv.current.ClientConfigDir)
	sv.clientDirEntry.SetPlaceHolder(sv.lang.T(lang.Settings_placeholder_client_dir))

	sv.widthEntry = widget.NewEntry()
	sv.widthEntry.SetText(strconv.Itoa(sv.current.WindowWidth))

	sv.heightEntry = widget.NewEntry()
	sv.heightEntry.SetText(strconv.Itoa(sv.current.WindowHeight))

	sv.fullscreenCheck = widget.NewCheck(sv.lang.T(lang.Settings_label_fullscreen), nil)
	sv.fullscreenCheck.Checked = sv.current.StartFullscreen

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_paths)),
		formHint(sv.lang.T(lang.Settings_card_paths_desc)),
		widget.NewForm(
			widget.NewFormItem(sv.lang.T(lang.Settings_label_wg_path), sv.wgPathEntry),
			widget.NewFormItem(sv.lang.T(lang.Settings_label_client_dir), sv.clientDirEntry),
		),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_window)),
		formHint(sv.lang.T(lang.Settings_note_window)),
		widget.NewForm(
			widget.NewFormItem(sv.lang.T(lang.Settings_label_width), sv.widthEntry),
			widget.NewFormItem(sv.lang.T(lang.Settings_label_height), sv.heightEntry),
			widget.NewFormItem("", sv.fullscreenCheck),
		),
	)
}

func (sv *SettingsView) pageAppearance() fyne.CanvasObject {
	sv.fontSizeSelect = widget.NewSelect([]string{"small", "normal", "large"}, nil)
	sv.fontSizeSelect.SetSelected(sv.current.FontSize)

	sv.customFontCheck = widget.NewCheck(sv.lang.T(lang.Settings_label_custom_font), nil)
	sv.customFontCheck.Checked = sv.current.UseCustomFont

	sv.densitySelect = widget.NewSelect(
		[]string{settings.DensityComfortable, settings.DensityCompact}, nil)
	sv.densitySelect.SetSelected(sv.current.Density)

	sv.themeSelect = widget.NewSelect([]string{"system", "light", "dark"}, nil)
	sv.themeSelect.SetSelected(sv.current.ThemeVariant)

	// Presets write every dark token at once, which is the only practical way
	// to get a coherent scheme out of 22 individual swatches.
	presetSelect := widget.NewSelect(wgtheme.PresetNames(), nil)
	presetSelect.PlaceHolder = sv.lang.T(lang.Settings_label_preset)

	applyPreset := widget.NewButtonWithIcon(sv.lang.T(lang.Settings_btn_apply_preset), wgtheme.IconCheck, func() {
		preset, ok := wgtheme.FindPreset(presetSelect.Selected)
		if !ok {
			return
		}
		for key, hex := range preset.Colors {
			btn, exists := sv.colorButtons[key]
			if !exists {
				continue
			}
			if col, err := wgwidget.HexToColor(hex); err == nil {
				btn.SetColor(col)
			}
		}
		helpers.ShowInformation(sv.lang.T(lang.Settings_card_appearance),
			sv.lang.Tf(lang.Settings_preset_applied, preset.Name), sv.win)
	})

	basics := widget.NewForm(
		widget.NewFormItem(sv.lang.T(lang.Settings_label_theme), sv.themeSelect),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_font_size), sv.fontSizeSelect),
		widget.NewFormItem(sv.lang.T(lang.Settings_label_density), sv.densitySelect),
		widget.NewFormItem("", sv.customFontCheck),
	)

	presetRow := container.NewBorder(nil, nil, nil, applyPreset, presetSelect)

	advanced := widget.NewAccordion(
		widget.NewAccordionItem(sv.lang.T(lang.Settings_accordion_dark_colors),
			sv.colorGroupsUI(darkColorGroups())),
		widget.NewAccordionItem(sv.lang.T(lang.Settings_accordion_light_colors),
			sv.colorGroupsUI(lightColorGroups())),
	)

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_appearance)),
		formHint(sv.lang.T(lang.Settings_card_appearance_desc)),
		basics,
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_label_preset)),
		formHint(sv.lang.T(lang.Settings_preset_desc)),
		presetRow,
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_advanced_tokens)),
		formHint(sv.lang.T(lang.Settings_advanced_tokens_desc)),
		advanced,
		formHint(sv.lang.T(lang.Settings_note_appearance)),
	)
}

// colorGroupsUI renders grouped swatches for one theme variant.
func (sv *SettingsView) colorGroupsUI(groups []colorGroup) fyne.CanvasObject {
	items := make([]fyne.CanvasObject, 0, len(groups)*2)
	for _, g := range groups {
		formItems := make([]*widget.FormItem, 0, len(g.tokens))
		for _, t := range g.tokens {
			formItems = append(formItems, widget.NewFormItem(sv.lang.T(t.labelKey), sv.button(t)))
		}
		items = append(items,
			wgwidget.SectionHeader(sv.lang.T(g.titleKey)),
			widget.NewForm(formItems...),
		)
	}
	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD), items...)
}

func (sv *SettingsView) pageBehavior() fyne.CanvasObject {
	sv.autoRefreshCheck = widget.NewCheck(sv.lang.T(lang.Settings_label_auto_refresh), nil)
	sv.autoRefreshCheck.Checked = sv.current.AutoRefreshEnabled

	sv.refreshSecsEntry = widget.NewEntry()
	sv.refreshSecsEntry.SetText(strconv.Itoa(sv.current.AutoRefreshSecs))

	sv.confirmDelCheck = widget.NewCheck(sv.lang.T(lang.Settings_label_confirm_delete), nil)
	sv.confirmDelCheck.Checked = sv.current.ConfirmBeforeDelete

	langOptions := make([]string, 0, len(lang.AvailableLanguages))
	sv.langCodeToDisplay = make(map[string]string)
	for code, display := range lang.AvailableLanguages {
		langOptions = append(langOptions, display)
		sv.langCodeToDisplay[code] = display
	}
	sv.langSelect = widget.NewSelect(langOptions, nil)
	sv.langSelect.SetSelected(sv.langCodeToDisplay[sv.current.Language])

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_behavior)),
		widget.NewForm(
			widget.NewFormItem(sv.lang.T(lang.Settings_label_language), sv.langSelect),
			widget.NewFormItem("", sv.autoRefreshCheck),
			widget.NewFormItem(sv.lang.T(lang.Settings_label_refresh_interval), sv.refreshSecsEntry),
			widget.NewFormItem("", sv.confirmDelCheck),
		),
	)
}

func (sv *SettingsView) pageScanner() fyne.CanvasObject {
	sv.workersEntry = widget.NewEntry()
	sv.workersEntry.SetText(strconv.Itoa(sv.current.ScanWorkers))

	sv.scanTimeoutEntry = widget.NewEntry()
	sv.scanTimeoutEntry.SetText(strconv.Itoa(sv.current.ScanTimeoutSecs))

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_scanner)),
		widget.NewForm(
			widget.NewFormItem(sv.lang.T(lang.Settings_label_scan_workers), sv.workersEntry),
			widget.NewFormItem(sv.lang.T(lang.Settings_label_scan_timeout), sv.scanTimeoutEntry),
		),
	)
}

func (sv *SettingsView) pagePrivileges() fyne.CanvasObject {
	privOptions := []string{"none", "pkexec", "sudo"}
	pkexecNotInstalled := sv.lang.T(lang.Settings_privilege_pkexec_unavailable)
	if !settings.PkexecAvailable() {
		privOptions = []string{"none", pkexecNotInstalled, "sudo"}
	}
	sv.privSelect = widget.NewSelect(privOptions, nil)

	privSelected := sv.current.PrivilegeEscalation
	if !settings.PkexecAvailable() && privSelected == "pkexec" {
		privSelected = pkexecNotInstalled
	}
	sv.privSelect.SetSelected(privSelected)

	note := widget.NewRichTextFromMarkdown(sv.lang.T(lang.Settings_privilege_desc))
	note.Wrapping = fyne.TextWrapWord

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_card_privilege)),
		formHint(sv.lang.T(lang.Settings_card_privilege_desc)),
		widget.NewForm(
			widget.NewFormItem(sv.lang.T(lang.Settings_label_privilege_method), sv.privSelect)),
		note,
	)
}

// pageAbout surfaces the environment facts and the keyboard map, so the
// shortcuts are discoverable from the UI rather than only from the palette.
func (sv *SettingsView) pageAbout() fyne.CanvasObject {
	const labelWidth = 110

	privilege := sv.lang.T(lang.Main_chip_no_root)
	if os.Geteuid() == 0 {
		privilege = sv.lang.T(lang.Main_chip_root)
	}

	version := "dev"
	if meta := fyne.CurrentApp().Metadata(); meta.Version != "" {
		version = meta.Version
	}

	shortcuts := []struct{ key, desc string }{
		{"Ctrl+K", sv.lang.T(lang.Shortcuts_palette)},
		{"Ctrl+N", sv.lang.T(lang.Shortcuts_new)},
		{"Ctrl+I", sv.lang.T(lang.Shortcuts_import)},
		{"Ctrl+R", sv.lang.T(lang.Shortcuts_refresh)},
		{"Ctrl+F", sv.lang.T(lang.Shortcuts_filter)},
		{"Ctrl+B", sv.lang.T(lang.Shortcuts_backups)},
		{"Ctrl+,", sv.lang.T(lang.Shortcuts_settings)},
		{"Ctrl+E", sv.lang.T(lang.Shortcuts_edit)},
		{"Ctrl+P", sv.lang.T(lang.Shortcuts_peers)},
		{"Ctrl+G", sv.lang.T(lang.Shortcuts_config)},
		{"Ctrl+S", sv.lang.T(lang.Shortcuts_scan)},
		{"Space", sv.lang.T(lang.Shortcuts_toggle)},
		{"Esc", sv.lang.T(lang.Shortcuts_close)},
	}

	rows := make([]fyne.CanvasObject, 0, len(shortcuts))
	for _, s := range shortcuts {
		k := wgwidget.Value(s.key)
		keyBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(74, k.MinSize().Height)), k)
		rows = append(rows, container.NewBorder(nil, nil, keyBox, nil, widget.NewLabel(s.desc)))
	}

	return container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(sv.lang.T(lang.Settings_nav_about)),
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM),
			wgwidget.KeyValue(sv.lang.T(lang.Settings_about_version), version, labelWidth),
			wgwidget.KeyValue(sv.lang.T(lang.Settings_about_backend), sv.lang.T(lang.Main_chip_backend), labelWidth),
			wgwidget.KeyValue(sv.lang.T(lang.Settings_about_config), sv.current.WGConfigPath, labelWidth),
			wgwidget.KeyValue(sv.lang.T(lang.Settings_about_privilege), privilege, labelWidth),
		),
		wgwidget.SectionHeader(sv.lang.T(lang.Shortcuts_title)),
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXS), rows...),
	)
}

// save validates and applies the settings.
func (sv *SettingsView) save() {
	updated, err := sv.validate()
	if err != nil {
		helpers.ShowError(err, sv.win)
		return
	}
	updated.Save(fyne.CurrentApp().Preferences())
	if sv.onApply != nil {
		sv.onApply(updated)
	}
	sv.win.Close()
}

// reset restores every control to its default.
func (sv *SettingsView) reset() {
	helpers.ShowConfirm(sv.lang.T(lang.Settings_reset_confirm_title),
		sv.lang.T(lang.Settings_reset_confirm_msg), func(yes bool) {
			if !yes {
				return
			}

			sv.wgPathEntry.SetText(settings.DefaultWGConfigPath)
			sv.clientDirEntry.SetText(settings.DefaultClientConfigDir)
			sv.widthEntry.SetText(strconv.Itoa(settings.DefaultWindowWidth))
			sv.heightEntry.SetText(strconv.Itoa(settings.DefaultWindowHeight))
			sv.fullscreenCheck.SetChecked(settings.DefaultStartFullscreen)
			sv.autoRefreshCheck.SetChecked(settings.DefaultAutoRefreshEnabled)
			sv.refreshSecsEntry.SetText(strconv.Itoa(settings.DefaultAutoRefreshSecs))
			sv.confirmDelCheck.SetChecked(settings.DefaultConfirmBeforeDelete)
			sv.themeSelect.SetSelected(settings.DefaultThemeVariant)
			sv.langSelect.SetSelected(sv.langCodeToDisplay[settings.DefaultLanguage])
			sv.densitySelect.SetSelected(settings.DefaultDensity)
			sv.workersEntry.SetText(strconv.Itoa(settings.DefaultScanWorkers))
			sv.scanTimeoutEntry.SetText(strconv.Itoa(settings.DefaultScanTimeoutSecs))
			sv.privSelect.SetSelected(settings.DefaultPrivilegeEscalation)
			sv.fontSizeSelect.SetSelected(settings.DefaultFontSize)
			sv.customFontCheck.SetChecked(settings.DefaultUseCustomFont)

			// One loop over the token tables replaces the 38-line block the
			// previous version had to keep in sync by hand.
			for _, groups := range [][]colorGroup{darkColorGroups(), lightColorGroups()} {
				for _, g := range groups {
					for _, t := range g.tokens {
						if btn, ok := sv.colorButtons[t.key]; ok {
							if col, err := wgwidget.HexToColor(t.def); err == nil {
								btn.SetColor(col)
							}
						}
					}
				}
			}
		}, sv.win)
}

// validate turns the controls into a settings struct, rejecting bad input.
func (sv *SettingsView) validate() (*settings.AppSettings, error) {
	if sv.wgPathEntry.Text == "" {
		return nil, errors.New(sv.lang.T(lang.Error_wg_path_empty))
	}
	if sv.clientDirEntry.Text == "" {
		return nil, errors.New(sv.lang.T(lang.Error_client_dir_empty))
	}

	width, err := strconv.Atoi(sv.widthEntry.Text)
	if err != nil || width < 400 {
		return nil, errors.New(sv.lang.T(lang.Error_width_invalid))
	}
	height, err := strconv.Atoi(sv.heightEntry.Text)
	if err != nil || height < 300 {
		return nil, errors.New(sv.lang.T(lang.Error_height_invalid))
	}
	refreshSecs, err := strconv.Atoi(sv.refreshSecsEntry.Text)
	if err != nil || refreshSecs < 1 {
		return nil, errors.New(sv.lang.T(lang.Error_refresh_invalid))
	}
	workers, err := strconv.Atoi(sv.workersEntry.Text)
	if err != nil || workers < 1 {
		return nil, errors.New(sv.lang.T(lang.Error_workers_invalid))
	}
	scanTimeout, err := strconv.Atoi(sv.scanTimeoutEntry.Text)
	if err != nil || scanTimeout < 1 {
		return nil, errors.New(sv.lang.T(lang.Error_timeout_invalid))
	}

	privMethod := sv.privSelect.Selected
	if privMethod == sv.lang.T(lang.Settings_privilege_pkexec_unavailable) {
		privMethod = "pkexec"
	}
	if privMethod != "none" && privMethod != "pkexec" && privMethod != "sudo" {
		return nil, errors.New(sv.lang.T(lang.Error_privilege_invalid))
	}

	themeVariant := sv.themeSelect.Selected
	if themeVariant != "system" && themeVariant != "light" && themeVariant != "dark" {
		return nil, errors.New(sv.lang.T(lang.Error_theme_invalid))
	}

	fontSize := sv.fontSizeSelect.Selected
	if fontSize != "small" && fontSize != "normal" && fontSize != "large" {
		return nil, errors.New(sv.lang.T(lang.Error_fontsize_invalid))
	}

	density := sv.densitySelect.Selected
	if density != settings.DensityCompact && density != settings.DensityComfortable {
		density = settings.DefaultDensity
	}

	langCode := settings.DefaultLanguage
	for code, display := range lang.AvailableLanguages {
		if display == sv.langSelect.Selected {
			langCode = code
			break
		}
	}

	h := sv.hex

	return &settings.AppSettings{
		WGConfigPath:        sv.wgPathEntry.Text,
		ClientConfigDir:     sv.clientDirEntry.Text,
		WindowWidth:         width,
		WindowHeight:        height,
		StartFullscreen:     sv.fullscreenCheck.Checked,
		AutoRefreshEnabled:  sv.autoRefreshCheck.Checked,
		AutoRefreshSecs:     refreshSecs,
		ConfirmBeforeDelete: sv.confirmDelCheck.Checked,
		ThemeVariant:        themeVariant,
		ScanWorkers:         workers,
		ScanTimeoutSecs:     scanTimeout,
		PrivilegeEscalation: privMethod,
		FontSize:            fontSize,
		Language:            langCode,
		Density:             density,
		UseCustomFont:       sv.customFontCheck.Checked,
		AccentColor:         h(settings.KeyLightAccentColor, settings.DefaultLightAccentColor),

		LightAccentColor:           h(settings.KeyLightAccentColor, settings.DefaultLightAccentColor),
		LightBackgroundColor:       h(settings.KeyLightBackgroundColor, settings.DefaultLightBackgroundColor),
		LightSurfaceColor:          h(settings.KeyLightSurfaceColor, settings.DefaultLightSurfaceColor),
		LightSurfaceRaisedColor:    h(settings.KeyLightSurfaceRaisedColor, settings.DefaultLightSurfaceRaisedColor),
		LightTextPrimaryColor:      h(settings.KeyLightTextPrimaryColor, settings.DefaultLightTextPrimaryColor),
		LightTextSecondaryColor:    h(settings.KeyLightTextSecondaryColor, settings.DefaultLightTextSecondaryColor),
		LightTextDimColor:          h(settings.KeyLightTextDimColor, settings.DefaultLightTextDimColor),
		LightTextDisabledColor:     h(settings.KeyLightTextDisabledColor, settings.DefaultLightTextDisabledColor),
		LightBorderColor:           h(settings.KeyLightBorderColor, settings.DefaultLightBorderColor),
		LightBorderStrongColor:     h(settings.KeyLightBorderStrongColor, settings.DefaultLightBorderStrongColor),
		LightSuccessColor:          h(settings.KeyLightSuccessColor, settings.DefaultLightSuccessColor),
		LightErrorColor:            h(settings.KeyLightErrorColor, settings.DefaultLightErrorColor),
		LightWarningColor:          h(settings.KeyLightWarningColor, settings.DefaultLightWarningColor),
		LightInfoColor:             h(settings.KeyLightInfoColor, settings.DefaultLightInfoColor),
		LightInputBackgroundColor:  h(settings.KeyLightInputBackgroundColor, settings.DefaultLightInputBackgroundColor),
		LightInputBorderColor:      h(settings.KeyLightInputBorderColor, settings.DefaultLightInputBorderColor),
		LightHoverColor:            h(settings.KeyLightHoverColor, settings.DefaultLightHoverColor),
		LightPressedColor:          h(settings.KeyLightPressedColor, settings.DefaultLightPressedColor),
		LightSelectionColor:        h(settings.KeyLightSelectionColor, settings.DefaultLightSelectionColor),
		LightHeaderBackgroundColor: h(settings.KeyLightHeaderBackgroundColor, settings.DefaultLightHeaderBackgroundColor),
		LightMenuBackgroundColor:   h(settings.KeyLightMenuBackgroundColor, settings.DefaultLightMenuBackgroundColor),
		LightScrollbarColor:        h(settings.KeyLightScrollbarColor, settings.DefaultLightScrollbarColor),

		DarkAccentColor:           h(settings.KeyDarkAccentColor, settings.DefaultDarkAccentColor),
		DarkBackgroundColor:       h(settings.KeyDarkBackgroundColor, settings.DefaultDarkBackgroundColor),
		DarkSurfaceColor:          h(settings.KeyDarkSurfaceColor, settings.DefaultDarkSurfaceColor),
		DarkSurfaceRaisedColor:    h(settings.KeyDarkSurfaceRaisedColor, settings.DefaultDarkSurfaceRaisedColor),
		DarkTextPrimaryColor:      h(settings.KeyDarkTextPrimaryColor, settings.DefaultDarkTextPrimaryColor),
		DarkTextSecondaryColor:    h(settings.KeyDarkTextSecondaryColor, settings.DefaultDarkTextSecondaryColor),
		DarkTextDimColor:          h(settings.KeyDarkTextDimColor, settings.DefaultDarkTextDimColor),
		DarkTextDisabledColor:     h(settings.KeyDarkTextDisabledColor, settings.DefaultDarkTextDisabledColor),
		DarkBorderColor:           h(settings.KeyDarkBorderColor, settings.DefaultDarkBorderColor),
		DarkBorderStrongColor:     h(settings.KeyDarkBorderStrongColor, settings.DefaultDarkBorderStrongColor),
		DarkSuccessColor:          h(settings.KeyDarkSuccessColor, settings.DefaultDarkSuccessColor),
		DarkErrorColor:            h(settings.KeyDarkErrorColor, settings.DefaultDarkErrorColor),
		DarkWarningColor:          h(settings.KeyDarkWarningColor, settings.DefaultDarkWarningColor),
		DarkInfoColor:             h(settings.KeyDarkInfoColor, settings.DefaultDarkInfoColor),
		DarkInputBackgroundColor:  h(settings.KeyDarkInputBackgroundColor, settings.DefaultDarkInputBackgroundColor),
		DarkInputBorderColor:      h(settings.KeyDarkInputBorderColor, settings.DefaultDarkInputBorderColor),
		DarkHoverColor:            h(settings.KeyDarkHoverColor, settings.DefaultDarkHoverColor),
		DarkPressedColor:          h(settings.KeyDarkPressedColor, settings.DefaultDarkPressedColor),
		DarkSelectionColor:        h(settings.KeyDarkSelectionColor, settings.DefaultDarkSelectionColor),
		DarkHeaderBackgroundColor: h(settings.KeyDarkHeaderBackgroundColor, settings.DefaultDarkHeaderBackgroundColor),
		DarkMenuBackgroundColor:   h(settings.KeyDarkMenuBackgroundColor, settings.DefaultDarkMenuBackgroundColor),
		DarkScrollbarColor:        h(settings.KeyDarkScrollbarColor, settings.DefaultDarkScrollbarColor),
	}, nil
}
