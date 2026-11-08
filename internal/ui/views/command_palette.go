package ui

import (
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
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
	icon     *wgtheme.Icon
	run      func()
}

// paletteRow is a reusable list row: icon, title, group and shortcut keycap,
// with its own highlight for the keyboard-selected entry.
type paletteRow struct {
	widget.BaseWidget

	bg       *canvas.Rectangle
	icon     *canvas.Image
	title    *canvas.Text
	group    *canvas.Text
	shortcut *fyne.Container
	box      *fyne.Container
}

func newPaletteRow() *paletteRow {
	p := wgtheme.Current()
	r := &paletteRow{
		bg:       canvas.NewRectangle(color.Transparent),
		icon:     &canvas.Image{FillMode: canvas.ImageFillContain},
		title:    canvas.NewText("", p.TextPrimary),
		group:    canvas.NewText("", p.TextDim),
		shortcut: container.NewStack(),
	}
	r.bg.CornerRadius = wgtheme.RadiusMD
	r.icon.SetMinSize(fyne.NewSize(18, 18))
	r.title.TextSize = wgtheme.TextSize(wgtheme.TextMD)
	r.title.FontSource = wgtheme.Font(wgtheme.WeightMedium)
	r.group.TextSize = wgtheme.TextSize(wgtheme.TextXS)
	r.group.FontSource = wgtheme.Font(wgtheme.WeightSemiBold)

	right := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceMD),
		container.NewCenter(r.group), container.NewCenter(r.shortcut))
	// The right padding keeps the shortcut clear of the list scrollbar.
	r.box = container.NewStack(r.bg, container.New(layout.NewCustomPaddedLayout(7, 7, 10, 18),
		container.NewBorder(nil, nil,
			container.New(layout.NewCustomPaddedLayout(0, 0, 0, 10), container.NewCenter(r.icon)),
			right, container.NewVBox(layout.NewSpacer(), r.title, layout.NewSpacer()))))

	r.ExtendBaseWidget(r)
	return r
}

func (r *paletteRow) Set(c command, selected bool) {
	p := wgtheme.Current()
	tone := p.TextSecondary
	r.bg.FillColor = color.Transparent
	if selected {
		tone = p.Accent
		r.bg.FillColor = p.AccentTint()
	}
	if c.icon != nil {
		r.icon.Resource = c.icon.For(tone)
	}
	r.title.Text = c.title
	r.group.Text = c.group
	r.shortcut.Objects = nil
	if c.shortcut != "" {
		r.shortcut.Objects = []fyne.CanvasObject{wgwidget.Kbd(c.shortcut)}
	}
	r.bg.Refresh()
	r.icon.Refresh()
	r.title.Refresh()
	r.group.Refresh()
	r.shortcut.Refresh()
}

func (r *paletteRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

// paletteEntry is the search field; it hands Up/Down to the result list so
// the palette works without leaving the keyboard.
type paletteEntry struct {
	widget.Entry
	onMove func(delta int)
}

func newPaletteEntry(onMove func(int)) *paletteEntry {
	e := &paletteEntry{onMove: onMove}
	e.ExtendBaseWidget(e)
	return e
}

func (e *paletteEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyDown:
		e.onMove(1)
	case fyne.KeyUp:
		e.onMove(-1)
	default:
		e.Entry.TypedKey(ev)
	}
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
		{l.T(lang.Main_backups), actions, "Ctrl+B", wgtheme.IconArchive, v.showBackupsDialog},
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
			cmds = append(cmds, command{toggleTitle + " · " + name, actions, "Space", icon, v.toggleSelected})
			// Scanning needs the tunnel up, so it is offered only then.
			if canScan(*iface) {
				cmds = append(cmds, command{l.T(lang.Card_btn_scan) + " · " + name, actions, "Ctrl+S", wgtheme.IconScan, v.scanSelected})
			}
			cmds = append(cmds,
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
		icon := wgtheme.IconTunnelDown
		if iface, ok := v.findInterface(n); ok && iface.Active {
			icon = wgtheme.IconTunnelUp
		}
		cmds = append(cmds, command{
			title: l.Tf(lang.Palette_select_tunnel, n),
			group: tunnels,
			icon:  icon,
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

	var s *wgwidget.Surface
	var list *widget.List

	run := func(i int) {
		if i < 0 || i >= len(shown) {
			return
		}
		cmd := shown[i]
		s.Close()
		cmd.run()
	}

	entry := newPaletteEntry(func(delta int) {
		if len(shown) == 0 {
			return
		}
		selected = (selected + delta + len(shown)) % len(shown)
		list.Refresh()
		list.ScrollTo(selected)
	})
	entry.SetPlaceHolder(v.lang.T(lang.Palette_placeholder))
	entry.ActionItem = widget.NewIcon(wgtheme.IconSearch)

	list = widget.NewList(
		func() int { return len(shown) },
		func() fyne.CanvasObject { return newPaletteRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			row, ok := obj.(*paletteRow)
			if !ok || id >= len(shown) {
				return
			}
			row.Set(shown[id], id == selected)
		},
	)
	list.HideSeparators = true
	list.OnSelected = func(id widget.ListItemID) {
		list.UnselectAll()
		run(id)
	}

	// Enter runs the highlighted match, which is the whole point of
	// type-then-Enter.
	entry.OnSubmitted = func(string) { run(selected) }

	// An empty list would otherwise render as blank space, leaving the user
	// unsure whether the palette is broken or simply has no match.
	noResults := widget.NewLabel(v.lang.T(lang.Palette_no_results))
	noResults.Alignment = fyne.TextAlignCenter
	noResults.Importance = widget.LowImportance
	noResults.Hide()

	entry.OnChanged = func(q string) {
		shown = matchCommands(all, q)
		selected = 0
		list.Refresh()
		if len(shown) == 0 {
			list.Hide()
			noResults.Show()
		} else {
			noResults.Hide()
			list.Show()
		}
	}

	hint := widget.NewLabel(v.lang.T(lang.Palette_hint))
	hint.Importance = widget.LowImportance

	body := container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, 0), entry),
		hint, nil, nil,
		container.NewStack(list, container.NewCenter(noResults)),
	)

	s = wgwidget.ShowModalSized(v.window, v.lang.T(lang.Palette_title), fyne.NewSize(600, 400), body)
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
		keyBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(84, 26)),
			container.NewHBox(wgwidget.Kbd(r.key)))
		items = append(items, container.NewBorder(nil, nil, keyBox, nil, widget.NewLabel(r.desc)))
	}

	var s *wgwidget.Surface
	closeBtn := wgwidget.NewButton(l.T(lang.Common_close), nil, wgwidget.Primary, func() { s.Close() })
	s = wgwidget.ShowModal(v.window, l.T(lang.Shortcuts_title), 440,
		container.New(layout.NewCustomPaddedVBoxLayout(0), items...), closeBtn)
}
