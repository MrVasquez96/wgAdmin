package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// registerShortcuts wires the keyboard map onto the main window.
//
// Shortcuts are registered on the canvas so they work wherever focus sits,
// with one exception: the bare keys (Space, Esc, F11, arrows) are handled
// through the typed-key hook and deliberately ignored while a text field has
// focus, so typing a space into the filter box does not toggle a tunnel.
//
// Fyne only builds a shortcut when a modifier is held, so an unmodified key
// cannot be registered above and reaches this hook only when nothing is
// focused. That is why leaving fullscreen has a toolbar button as well: a
// keyboard-only escape hatch would be one stray click away from not working.
func (v *MainView) registerShortcuts() {
	canvas := v.window.Canvas()

	add := func(key fyne.KeyName, mod fyne.KeyModifier, fn func()) {
		canvas.AddShortcut(&desktop.CustomShortcut{KeyName: key, Modifier: mod},
			func(fyne.Shortcut) { fn() })
	}

	ctrl := fyne.KeyModifierControl

	add(fyne.KeyK, ctrl, v.showCommandPalette)
	add(fyne.KeyN, ctrl, v.showAddTunnelForm)
	add(fyne.KeyI, ctrl, v.importConfig)
	add(fyne.KeyR, ctrl, v.Refresh)
	add(fyne.KeyF, ctrl, v.focusFilter)
	add(fyne.KeyB, ctrl, v.showBackupsDialog)
	add(fyne.KeyComma, ctrl, v.showSettings)
	add(fyne.KeyS, ctrl, v.scanSelected)

	add(fyne.KeyE, ctrl, func() {
		if name := v.selectedName(); name != "" {
			v.showEditTunnelForm(name)
		}
	})
	add(fyne.KeyP, ctrl, func() {
		if name := v.selectedName(); name != "" {
			v.showEditPeersTunnelForm(name)
		}
	})
	add(fyne.KeyG, ctrl, func() {
		if name := v.selectedName(); name != "" {
			v.showConfigView(name)
		}
	})

	canvas.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		// Never steal plain keys from a focused text field.
		if isTextFocused(canvas) {
			if ev.Name == fyne.KeyEscape {
				v.clearFilter()
				canvas.Unfocus()
			}
			return
		}

		switch ev.Name {
		case fyne.KeySpace:
			v.toggleSelected()
		case fyne.KeyF11:
			v.toggleFullscreen()
		case fyne.KeyEscape:
			// Leaving fullscreen is the more urgent of the two, and a filter
			// that survives the first Esc is cleared by the second.
			if v.exitFullscreen() {
				return
			}
			v.clearFilter()
		case fyne.KeyDown:
			v.moveSelection(1)
		case fyne.KeyUp:
			v.moveSelection(-1)
		}
	})
}

// isTextFocused reports whether focus is on a widget that consumes ordinary
// typing. Canvas.Focused() is already a fyne.Focusable, so the check has to be
// against the concrete text widgets — buttons and lists are focusable too, and
// those should not block Space or the arrow keys.
func isTextFocused(c fyne.Canvas) bool {
	switch c.Focused().(type) {
	case *widget.Entry, *widget.SelectEntry:
		return true
	default:
		return false
	}
}
