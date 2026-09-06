package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"
)

// peerListRow is one row of the peers editor: name, allowed IPs, and the
// per-peer actions. Written as a widget so the list's update function can set
// its parts by name instead of indexing into a container.
type peerListRow struct {
	widget.BaseWidget

	title    *canvas.Text
	subtitle *canvas.Text
	qrBtn    *widget.Button
	editBtn  *widget.Button
	delBtn   *widget.Button

	onQR     func()
	onEdit   func()
	onDelete func()

	container *fyne.Container
}

func newPeerListRow() *peerListRow {
	r := &peerListRow{}

	r.title = wgwidget.Value("")
	r.title.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	r.subtitle = wgwidget.DimValue("")
	r.subtitle.TextSize = 11

	r.qrBtn = widget.NewButtonWithIcon("", wgtheme.IconQR, func() {
		if r.onQR != nil {
			r.onQR()
		}
	})
	r.editBtn = widget.NewButtonWithIcon("", wgtheme.IconEdit, func() {
		if r.onEdit != nil {
			r.onEdit()
		}
	})
	r.delBtn = widget.NewButtonWithIcon("", wgtheme.IconTrash, func() {
		if r.onDelete != nil {
			r.onDelete()
		}
	})
	r.delBtn.Importance = widget.DangerImportance

	text := container.New(layout.NewCustomPaddedVBoxLayout(1), r.title, r.subtitle)
	actions := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceXS),
		r.qrBtn, r.editBtn, r.delBtn)

	r.container = container.New(layout.NewCustomPaddedLayout(
		wgtheme.SpaceSM, wgtheme.SpaceSM, wgtheme.SpaceMD, wgtheme.SpaceMD),
		container.NewBorder(nil, nil, nil, actions, text))

	r.ExtendBaseWidget(r)
	return r
}

// Set updates the row. canProvision controls whether the QR action is offered:
// a client config cannot be built without the peer's private key.
func (r *peerListRow) Set(title, subtitle string, canProvision bool) {
	r.title.Text = title
	r.subtitle.Text = subtitle
	r.title.Refresh()
	r.subtitle.Refresh()

	if canProvision {
		r.qrBtn.Enable()
	} else {
		r.qrBtn.Disable()
	}
}

func (r *peerListRow) MinSize() fyne.Size {
	return fyne.NewSize(320, 46)
}

func (r *peerListRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.container)
}
