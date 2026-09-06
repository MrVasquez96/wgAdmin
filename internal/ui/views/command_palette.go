package ui

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"
)

// command is one entry in the palette.
type command struct {
	title    string
	group    string
	shortcut string
	icon     fyne.Resource
	run      func()
}

// paletteRow is a reusable list row: icon, title, group and shortcut hint.
// It is a widget rather than an ad-hoc container so the update function can
// address its parts by name instead of by index.
type paletteRow struct {
	widget.BaseWidget

	icon      *widget.Icon
	title     *widget.Label
	group     *canvas.Text
	shortcut  *canvas.Text
	container *fyne.Container
}

func newPaletteRow() *paletteRow {
	r := &paletteRow{
		icon:     widget.NewIcon(wgtheme.IconChevronR),
		title:    widget.NewLabel(""),
		group:    wgwidget.Caption(""),
		shortcut: wgwidget.Caption(""),
	}
	r.title.Truncation = fyne.TextTruncateEllipsis

	right := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD),
		container.NewCenter(r.group),
		container.NewCenter(r.shortcut),
	)

	// The right padding keeps the shortcut column clear of the list scrollbar,
	// which otherwise sits on top of it.
	r.container = container.New(layout.NewCustomPaddedLayout(0, 0, 0, wgtheme.SpaceLG),
		container.NewBorder(nil, nil,
			container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(16, 16)), r.icon)),
			right,
			r.title,
		))

	r.ExtendBaseWidget(r)
	return r
}

func (r *paletteRow) Set(c command) {
	r.icon.SetResource(c.icon)
	r.title.SetText(c.title)
	r.group.Text = strings.ToUpper(c.group)
	r.shortcut.Text = c.shortcut
	r.group.Refresh()
	r.shortcut.Refresh()
}

func (r *paletteRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.container)
}

// commands builds the palette's contents: the global actions first, then one
// entry per tunnel so any tunnel can be reached by name without the mouse.
func (v *MainView) commands() []command {
	l := v.lang
	actions := l.T(lang.Palette_group_actions)

	cmds := []command{
		{l.T(lang.Main_add_tunnel), actions, "Ctrl+N", wgtheme.IconAdd, v.showAddTunnelForm},
		{l.T(lang.Main_import), actions, "Ctrl+I", wgtheme.IconImport, v.importConfig},
		{l.T(lang.Common_refresh), actions, "Ctrl+R", wgtheme.IconRefresh, v.Refresh},
		{l.T(lang.Main_backups), actions, "Ctrl+B", wgtheme.IconFolder, v.showBackupsDialog},
		{l.T(lang.Common_settings), actions, "Ctrl+,", wgtheme.IconSettings, v.showSettings},
		{v.fullscreenCommandTitle(), actions, "F11", wgtheme.IconFullscreen, v.toggleFullscreen},
		{l.T(lang.Shortcuts_title), actions, "", wgtheme.IconKeyboard, v.showShortcuts},
	}

	// Actions that need a selection are only offered when one exists, so the
	// palette never lists something that would silently do nothing.
	if name := v.selectedName(); name != "" {
		if iface, ok := v.findInterface(name); ok {
			toggleTitle, icon := l.T(lang.Card_btn_activate), wgtheme.IconPlay
			if iface.Active {
				toggleTitle, icon = l.T(lang.Card_btn_deactivate), wgtheme.IconStop
			}
			cmds = append(cmds,
				command{toggleTitle + " · " + name, actions, "Space", icon, v.toggleSelected},
				command{l.T(lang.Card_btn_scan) + " · " + name, actions, "Ctrl+S", wgtheme.IconScan, v.scanSelected},
				command{l.T(lang.Card_btn_edit) + " · " + name, actions, "Ctrl+E", wgtheme.IconEdit,
					func() { v.showEditTunnelForm(name) }},
				command{l.T(lang.Card_btn_peers) + " · " + name, actions, "Ctrl+P", wgtheme.IconPeers,
					func() { v.showEditPeersTunnelForm(name) }},
				command{l.T(lang.Detail_btn_config) + " · " + name, actions, "Ctrl+G", wgtheme.IconDoc,
					func() { v.showConfigView(name) }},
			)
		}
	}

	tunnels := l.T(lang.Palette_group_tunnels)
	for _, name := range v.tunnelNames() {
		n := name
		cmds = append(cmds, command{
			title: l.Tf(lang.Palette_select_tunnel, n),
			group: tunnels,
			icon:  wgtheme.IconGlobe,
			run: func() {
				v.selectTunnel(n)
				v.syncSidebarSelection()
			},
		})
	}

	return cmds
}

// fullscreenCommandTitle names the action by what it does next, so the palette
// offers a way out of fullscreen rather than a way further in.
func (v *MainView) fullscreenCommandTitle() string {
	if v.window.FullScreen() {
		return v.lang.T(lang.Main_exit_fullscreen)
	}
	return v.lang.T(lang.Main_enter_fullscreen)
}

// matchCommands filters and ranks commands against a query. Matching falls
// back to a subsequence test — typing "scwg" finds "Scan · wgCloud" — which is
// what makes a palette feel quick compared with plain substring search.
func matchCommands(cmds []command, query string) []command {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return cmds
	}

	type scored struct {
		cmd   command
		score int
	}
	var out []scored

	for _, c := range cmds {
		title := strings.ToLower(c.title)
		switch {
		case strings.HasPrefix(title, q):
			out = append(out, scored{c, 0}) // best: the title starts with it
		case strings.Contains(title, q):
			out = append(out, scored{c, 1})
		case isSubsequence(q, title):
			out = append(out, scored{c, 2})
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].score < out[j].score })

	res := make([]command, 0, len(out))
	for _, s := range out {
		res = append(res, s.cmd)
	}
	return res
}

// isSubsequence reports whether every byte of needle appears in haystack in
// order, allowing gaps. Both are expected to be lower-cased already.
func isSubsequence(needle, haystack string) bool {
	if needle == "" {
		return true
	}
	i := 0
	for j := 0; j < len(haystack); j++ {
		if needle[i] == haystack[j] {
			i++
			if i == len(needle) {
				return true
			}
		}
	}
	return false
}

// showCommandPalette opens the Ctrl+K palette.
func (v *MainView) showCommandPalette() {
	all := v.commands()
	shown := all
	selected := 0

	var d dialog.Dialog
	var list *widget.List

	entry := widget.NewEntry()
	entry.SetPlaceHolder(v.lang.T(lang.Palette_placeholder))

	run := func(i int) {
		if i < 0 || i >= len(shown) {
			return
		}
		cmd := shown[i]
		d.Hide()
		cmd.run()
	}

	list = widget.NewList(
		func() int { return len(shown) },
		func() fyne.CanvasObject { return newPaletteRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row, ok := obj.(*paletteRow)
			if !ok || id >= len(shown) {
				return
			}
			row.Set(shown[id])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		selected = id
		run(id)
	}

	// Enter runs the top match, which is the whole point of type-then-Enter.
	entry.OnSubmitted = func(string) { run(selected) }

	hint := wgwidget.Caption(v.lang.T(lang.Palette_hint))

	// An empty list would otherwise render as blank space, leaving the user
	// unsure whether the palette is broken or simply has no match.
	noResults := widget.NewLabel(v.lang.T(lang.Palette_no_results))
	noResults.Alignment = fyne.TextAlignCenter
	noResults.Importance = widget.LowImportance
	noResults.Hide()

	results := container.NewStack(list, container.NewCenter(noResults))

	syncEmptyState := func() {
		if len(shown) == 0 {
			list.Hide()
			noResults.Show()
		} else {
			noResults.Hide()
			list.Show()
		}
	}

	entry.OnChanged = func(q string) {
		shown = matchCommands(all, q)
		selected = 0
		list.UnselectAll()
		list.Refresh()
		syncEmptyState()
	}

	body := container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, 0), entry),
		container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceMD, 0, 0, 0), hint),
		nil, nil,
		results,
	)

	d = dialog.NewCustom(v.lang.T(lang.Palette_title), v.lang.T(lang.Common_close), body, v.window)
	d.Resize(fyne.NewSize(600, 440))
	d.Show()

	v.window.Canvas().Focus(entry)
}

// showShortcuts lists the keyboard map, so the shortcuts are discoverable
// rather than folklore.
func (v *MainView) showShortcuts() {
	l := v.lang

	rows := []struct{ key, desc string }{
		{"Ctrl+K", l.T(lang.Shortcuts_palette)},
		{"Ctrl+N", l.T(lang.Shortcuts_new)},
		{"Ctrl+I", l.T(lang.Shortcuts_import)},
		{"Ctrl+R", l.T(lang.Shortcuts_refresh)},
		{"Ctrl+F", l.T(lang.Shortcuts_filter)},
		{"Ctrl+B", l.T(lang.Shortcuts_backups)},
		{"Ctrl+,", l.T(lang.Shortcuts_settings)},
		{"Ctrl+E", l.T(lang.Shortcuts_edit)},
		{"Ctrl+P", l.T(lang.Shortcuts_peers)},
		{"Ctrl+G", l.T(lang.Shortcuts_config)},
		{"Ctrl+S", l.T(lang.Shortcuts_scan)},
		{"Space", l.T(lang.Shortcuts_toggle)},
		{"F11", l.T(lang.Shortcuts_fullscreen)},
		{"Esc", l.T(lang.Shortcuts_close)},
	}

	items := make([]fyne.CanvasObject, 0, len(rows))
	for _, r := range rows {
		key := wgwidget.Value(r.key)
		keyBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(78, key.MinSize().Height)), key)
		desc := widget.NewLabel(r.desc)
		items = append(items, container.NewBorder(nil, nil, keyBox, nil, desc))
	}

	body := container.NewVScroll(container.New(
		layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXS), items...))

	d := dialog.NewCustom(l.T(lang.Shortcuts_title), l.T(lang.Common_close), body, v.window)
	d.Resize(fyne.NewSize(460, 460))
	d.Show()
}
