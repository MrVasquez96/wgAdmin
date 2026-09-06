package theme

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// The application icon set.
//
// IMPORTANT: every icon here must be built from *filled* paths. Fyne recolours
// SVG resources by rewriting the `fill` attribute only and never touches
// `stroke` (see fyne/internal/svg.Colorize -> replaceFillColor). Stroke-drawn
// icon sets such as Lucide or Feather would therefore render in their baked-in
// colour and ignore both the theme and button importance. All paths below are
// solid-fill, 24x24, in the Material Symbols idiom.

// iconSources keeps the pre-colourisation markup for every icon, keyed by
// name. ThemedResource.Content() rewrites fills and needs a running app, so
// tests assert against these sources instead.
var iconSources = map[string]string{}

// svgIcon wraps a 24x24 path (or path group) in a themeable SVG resource.
func svgIcon(name, body string) fyne.Resource {
	const tmpl = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="24" height="24">%s</svg>`
	markup := fmt.Sprintf(tmpl, body)
	iconSources[name] = markup
	return theme.NewThemedResource(fyne.NewStaticResource(name+".svg", []byte(markup)))
}

// path is shorthand for a single filled path element.
func path(d string) string {
	return `<path fill="#ffffff" d="` + d + `"/>`
}

var (
	// Brand / identity
	IconShield = svgIcon("wg-shield", path("M12 1 3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4z"))
	IconGlobe  = svgIcon("wg-globe", path("M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm6.92 6h-2.95a15.7 15.7 0 0 0-1.38-3.56A8.03 8.03 0 0 1 18.92 8zM12 4.04c.83 1.2 1.48 2.53 1.91 3.96h-3.82c.43-1.43 1.08-2.76 1.91-3.96zM4.26 14A8.1 8.1 0 0 1 4 12c0-.69.1-1.36.26-2h3.38a16.5 16.5 0 0 0 0 4H4.26zm.82 2h2.95c.32 1.25.78 2.45 1.38 3.56A7.99 7.99 0 0 1 5.08 16zm2.95-8H5.08a7.99 7.99 0 0 1 4.33-3.56A15.7 15.7 0 0 0 8.03 8zM12 19.96c-.83-1.2-1.48-2.53-1.91-3.96h3.82c-.43 1.43-1.08 2.76-1.91 3.96zM14.34 14H9.66a14.7 14.7 0 0 1 0-4h4.68a14.7 14.7 0 0 1 0 4zm.25 5.56c.6-1.11 1.06-2.31 1.38-3.56h2.95a8.03 8.03 0 0 1-4.33 3.56zM16.36 14a16.5 16.5 0 0 0 0-4h3.38c.16.64.26 1.31.26 2s-.1 1.36-.26 2h-3.38z"))

	// Tunnel state
	IconPlay  = svgIcon("wg-play", path("M8 5v14l11-7z"))
	IconStop  = svgIcon("wg-stop", path("M6 6h12v12H6z"))
	IconPower = svgIcon("wg-power", path("M13 3h-2v10h2V3zm4.83 2.17-1.42 1.42A6.92 6.92 0 0 1 19 12c0 3.87-3.13 7-7 7s-7-3.13-7-7c0-2.22 1.03-4.19 2.64-5.46L6.22 5.13A8.94 8.94 0 0 0 3 12a9 9 0 0 0 18 0c0-2.74-1.23-5.19-3.17-6.83z"))

	// Actions
	IconScan     = svgIcon("wg-scan", path("M12 4a8 8 0 1 0 5.66 13.66l3.34 3.34 1.41-1.41-3.34-3.34A8 8 0 0 0 12 4zm0 2a6 6 0 1 1 0 12 6 6 0 0 1 0-12zm0 2.5A3.5 3.5 0 1 0 15.5 12 3.5 3.5 0 0 0 12 8.5zm0 1.75A1.75 1.75 0 1 1 10.25 12 1.75 1.75 0 0 1 12 10.25z"))
	IconKey      = svgIcon("wg-key", path("M12.65 10A6 6 0 1 0 7 18a6 6 0 0 0 5.65-4H17v4h4v-4h2v-4H12.65zM7 14a2 2 0 1 1 0-4 2 2 0 0 1 0 4z"))
	IconPeers    = svgIcon("wg-peers", path("M16 11a3 3 0 1 0-3-3 3 3 0 0 0 3 3zm-8 0a3 3 0 1 0-3-3 3 3 0 0 0 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05A4.2 4.2 0 0 1 17 16.5V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"))
	IconEdit     = svgIcon("wg-edit", path("M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04a1 1 0 0 0 0-1.41l-2.34-2.34a1 1 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"))
	IconTrash    = svgIcon("wg-trash", path("M6 19a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"))
	IconCopy     = svgIcon("wg-copy", path("M16 1H4a2 2 0 0 0-2 2v14h2V3h12V1zm3 4H8a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2zm0 16H8V7h11v14z"))
	IconRefresh  = svgIcon("wg-refresh", path("M17.65 6.35A7.96 7.96 0 0 0 12 4a8 8 0 1 0 7.73 10h-2.08A6 6 0 1 1 12 6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"))
	IconFolder   = svgIcon("wg-folder", path("M10 4H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-8l-2-2z"))
	IconSettings = svgIcon("wg-settings", path("M19.14 12.94a7.07 7.07 0 0 0 0-1.88l2.03-1.58a.5.5 0 0 0 .12-.62l-1.92-3.32a.5.5 0 0 0-.6-.22l-2.39.96a7.03 7.03 0 0 0-1.62-.94l-.36-2.54a.5.5 0 0 0-.5-.42h-3.84a.5.5 0 0 0-.5.42l-.36 2.54c-.58.24-1.12.55-1.62.94l-2.39-.96a.5.5 0 0 0-.6.22L2.67 8.86a.5.5 0 0 0 .12.62l2.03 1.58a7.07 7.07 0 0 0 0 1.88l-2.03 1.58a.5.5 0 0 0-.12.62l1.92 3.32c.12.22.39.3.6.22l2.39-.96c.5.39 1.04.7 1.62.94l.36 2.54c.04.24.25.42.5.42h3.84c.25 0 .46-.18.5-.42l.36-2.54c.58-.24 1.12-.55 1.62-.94l2.39.96c.22.08.48 0 .6-.22l1.92-3.32a.5.5 0 0 0-.12-.62l-2.03-1.58zM12 15.6A3.6 3.6 0 1 1 15.6 12 3.6 3.6 0 0 1 12 15.6z"))
	IconTerminal = svgIcon("wg-terminal", path("M20 4H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2zm0 14H4V8h16v10zM6.5 9.5 5.44 10.56 7.38 12.5l-1.94 1.94L6.5 15.5l3-3-3-3zM12 14h5v1.5h-5V14z"))
	IconQR       = svgIcon("wg-qr", path("M3 3h8v8H3V3zm2 2v4h4V5H5zm8-2h8v8h-8V3zm2 2v4h4V5h-4zM3 13h8v8H3v-8zm2 2v4h4v-4H5zm8-2h3v2h-3v-2zm5 0h3v2h-3v-2zm-5 4h2v2h-2v-2zm3 0h2v2h-2v-2zm3 0h2v4h-2v-4zm-6 3h2v2h-2v-2zm3 0h2v2h-2v-2z"))
	IconFilter   = svgIcon("wg-filter", path("M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"))
	IconAdd      = svgIcon("wg-add", path("M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"))
	IconClose    = svgIcon("wg-close", path("M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12 19 6.41z"))
	IconCheck    = svgIcon("wg-check", path("M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41L9 16.17z"))
	IconDownload = svgIcon("wg-download", path("M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"))
	IconImport   = svgIcon("wg-import", path("M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm-1 15v-3h-2v3H8l4 4 4-4h-3zm0-8V3.5L18.5 9H13z"))
	IconClock    = svgIcon("wg-clock", path("M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm4.28 13.54-5.28-3.1V7h1.5v4.6l4.53 2.66-.75 1.28z"))
	IconAlert    = svgIcon("wg-alert", path("M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"))
	IconError    = svgIcon("wg-error", path("M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"))
	IconInfo     = svgIcon("wg-info", path("M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"))
	IconActivity = svgIcon("wg-activity", path("M3.5 12.75h3.2l1.9-5.7a1 1 0 0 1 1.9.04l2.6 9.1 1.6-4.1a1 1 0 0 1 .93-.64h4.87v1.9h-4.22l-2.4 6.14a1 1 0 0 1-1.89-.09l-2.55-8.9-1.2 3.6a1 1 0 0 1-.95.68H3.5v-2.03z"))
	IconChevronR = svgIcon("wg-chevron-right", path("M10 6 8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6-6-6z"))
	IconChevronD = svgIcon("wg-chevron-down", path("M7.41 8.59 12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"))
	IconRoute    = svgIcon("wg-route", path("M18 4a3 3 0 0 0-2.83 4H12a4 4 0 0 0 0 8h-.17A3 3 0 1 0 12 18h.17A6 6 0 0 1 12 6h3.17A3 3 0 1 0 18 4zm0 2a1 1 0 1 1-1 1 1 1 0 0 1 1-1zM6 18a1 1 0 1 1 1 1 1 1 0 0 1-1-1z"))
	IconDoc      = svgIcon("wg-doc", path("M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"))
	IconSearch   = svgIcon("wg-search", path("M15.5 14h-.79l-.28-.27A6.47 6.47 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0A4.5 4.5 0 1 1 14 9.5 4.5 4.5 0 0 1 9.5 14z"))
	// Material Symbols "fullscreen" / "fullscreen_exit": four corner brackets
	// pointing outward, then inward.
	IconFullscreen     = svgIcon("wg-fullscreen", path("M7 14H5v5h5v-2H7v-3zm-2-4h2V7h3V5H5v5zm12 7h-3v2h5v-5h-2v3zM14 5v2h3v3h2V5h-5z"))
	IconFullscreenExit = svgIcon("wg-fullscreen-exit", path("M5 16h3v3h2v-5H5v2zm3-8H5v2h5V5H8v3zm6 11h2v-3h3v-2h-5v5zm2-11V5h-2v5h5V8h-3z"))

	IconKeyboard = svgIcon("wg-keyboard", path("M20 5H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V7a2 2 0 0 0-2-2zM5 8h2v2H5V8zm0 3h2v2H5v-2zm3-3h2v2H8V8zm0 3h2v2H8v-2zm-2 3h9v2H6v-2zm5-3h2v2h-2v-2zm0-3h2v2h-2V8zm3 3h2v2h-2v-2zm0-3h2v2h-2V8zm3 6h2v2h-2v-2zm0-3h2v2h-2v-2zm0-3h2v2h-2V8z"))
)

// overrideIcon maps Fyne's built-in icon names onto the application set, so
// every existing theme.XxxIcon() call site picks up the new artwork for free.
func overrideIcon(name fyne.ThemeIconName) (fyne.Resource, bool) {
	res, ok := iconOverrides[name]
	return res, ok
}

var iconOverrides = map[fyne.ThemeIconName]fyne.Resource{
	theme.IconNameSearch:         IconSearch,
	theme.IconNameDelete:         IconTrash,
	theme.IconNameCancel:         IconClose,
	theme.IconNameConfirm:        IconCheck,
	theme.IconNameContentAdd:     IconAdd,
	theme.IconNameContentCopy:    IconCopy,
	theme.IconNameContentClear:   IconClose,
	theme.IconNameDocument:       IconDoc,
	theme.IconNameDocumentCreate: IconEdit,
	theme.IconNameDocumentSave:   IconImport,
	theme.IconNameFolder:         IconFolder,
	theme.IconNameFolderOpen:     IconFolder,
	theme.IconNameSettings:       IconSettings,
	theme.IconNameViewRefresh:    IconRefresh,
	theme.IconNameMediaPlay:      IconPlay,
	theme.IconNameMediaStop:      IconStop,
	theme.IconNameHistory:        IconClock,
	theme.IconNameInfo:           IconInfo,
	theme.IconNameError:          IconError,
	theme.IconNameWarning:        IconAlert,
	theme.IconNameDownload:       IconDownload,
	theme.IconNameNavigateNext:   IconChevronR,
	theme.IconNameArrowDropDown:  IconChevronD,
	theme.IconNameMenuExpand:     IconChevronR,
	theme.IconNameComputer:       IconGlobe,
	theme.IconNameAccount:        IconPeers,
}
