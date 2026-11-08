package wgwidget

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// Secondary windows (SSH terminals, a popped-out scan) are hidden and reused
// rather than closed.
//
// Fyne 2.7 gives every window its own, unshared GL context, and when a window
// closes it frees that window's textures without first making its context
// current (glfw window.Close → cache.RangeTexturesFor(…, Painter().Free), and
// DeleteTextTexturesFor). Texture names are numbered per context, so the
// deletes land on whichever window's context is current — usually the main
// window — and its long-lived text and icons turn into black boxes. A window
// that is only hidden frees nothing, so reusing it sidesteps the bug entirely.
var windowPool struct {
	sync.Mutex
	idle []fyne.Window
}

// SecondaryWindow returns a window for a short-lived secondary view: a
// previously released one when available, otherwise a new one. onClose runs
// when the user closes it; the window is then released for reuse instead of
// destroyed.
func SecondaryWindow(title string, size fyne.Size, onClose func()) fyne.Window {
	windowPool.Lock()
	var w fyne.Window
	if n := len(windowPool.idle); n > 0 {
		w = windowPool.idle[n-1]
		windowPool.idle = windowPool.idle[:n-1]
	}
	windowPool.Unlock()

	if w == nil {
		w = fyne.CurrentApp().NewWindow(title)
	}
	w.SetTitle(title)
	w.Resize(size)
	w.SetCloseIntercept(func() {
		if onClose != nil {
			onClose()
		}
		releaseWindow(w)
	})
	return w
}

// releaseWindow hides w and keeps it for the next SecondaryWindow call. Its
// content is replaced so the old view (and any session it holds) can be
// garbage collected.
func releaseWindow(w fyne.Window) {
	w.Hide()
	w.SetContent(container.NewStack())
	windowPool.Lock()
	windowPool.idle = append(windowPool.idle, w)
	windowPool.Unlock()
}
