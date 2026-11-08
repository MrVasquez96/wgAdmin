package helpers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"
)

// These helpers keep the signatures every view already calls, but render the
// application's own modal (icon, wrapped message, styled buttons) instead of
// Fyne's stock dialog, so all ~50 call sites change look without being edited.

// Button labels for the helpers. The main view sets them from the active
// language; English is only the fallback.
var labels = struct{ OK, Cancel, Yes, No string }{"OK", "Cancel", "Yes", "No"}

// SetLabels installs translated button labels.
func SetLabels(ok, cancel, yes, no string) {
	labels.OK, labels.Cancel, labels.Yes, labels.No = ok, cancel, yes, no
}

// message lays out a status icon beside a word-wrapped message. The label is
// given a fixed width to wrap against; a wrapping label inside a centring
// container collapses to one glyph per line.
func message(icon fyne.Resource, text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	img := widget.NewIcon(icon)
	iconBox := container.NewVBox(container.New(layout.NewGridWrapLayout(fyne.NewSize(26, 26)), img))
	body := container.New(layout.NewGridWrapLayout(fyne.NewSize(380, wrappedHeight(l, 380))), l)
	return container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceLG), iconBox, body)
}

// wrappedHeight estimates the height text needs when wrapped at width w.
func wrappedHeight(l *widget.Label, w float32) float32 {
	l.Resize(fyne.NewSize(w, 0))
	return l.MinSize().Height
}

func show(title string, icon fyne.Resource, text string, parent fyne.Window,
	buttons func(close func()) []fyne.CanvasObject) {
	var s *wgwidget.Surface
	body := message(icon, text)
	s = wgwidget.ShowModal(parent, title, 460, body, buttons(func() { s.Close() })...)
}

// ShowConfirm asks a yes/no question. callback receives the answer; closing
// the dialog any other way counts as no.
func ShowConfirm(title, msg string, callback func(bool), parent fyne.Window) {
	answered := false
	var s *wgwidget.Surface
	body := message(wgtheme.IconAlert.For(wgtheme.Current().Warning), msg)
	reply := func(v bool) {
		answered = true
		s.Close()
		callback(v)
	}
	no := wgwidget.NewButton(labels.No, nil, wgwidget.Secondary, func() { reply(false) })
	yes := wgwidget.NewButton(labels.Yes, nil, wgwidget.Primary, func() { reply(true) })
	s = wgwidget.ShowModal(parent, title, 460, body, no, yes)
	s.OnClose = func() {
		if !answered {
			answered = true
			callback(false)
		}
	}
}

// ShowError reports a failure.
func ShowError(err error, parent fyne.Window) {
	show(errorTitle(), wgtheme.IconError.For(wgtheme.Current().Error), err.Error(), parent, func(close func()) []fyne.CanvasObject {
		return []fyne.CanvasObject{wgwidget.NewButton(labels.OK, nil, wgwidget.Primary, close)}
	})
}

// ShowInformation shows a notice.
func ShowInformation(title, msg string, parent fyne.Window) {
	show(title, wgtheme.IconInfo.For(wgtheme.Current().Info), msg, parent, func(close func()) []fyne.CanvasObject {
		return []fyne.CanvasObject{wgwidget.NewButton(labels.OK, nil, wgwidget.Primary, close)}
	})
}

// errorTitle is the error dialog heading.
var errorTitle = func() string { return "Error" }

// SetErrorTitle installs the translated error heading.
func SetErrorTitle(t string) { errorTitle = func() string { return t } }
