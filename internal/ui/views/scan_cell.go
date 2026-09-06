package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
)

// scanCell is a results-table cell that can render either a monospace value or
// the per-row action buttons. Table cells are recycled across columns, so one
// cell type has to cover both and hide whichever half is unused.
type scanCell struct {
	widget.BaseWidget

	label   *widget.Label
	copyBtn *widget.Button
	sshBtn  *widget.Button
	actions *fyne.Container

	container *fyne.Container
}

func newScanCell(l *lang.Language) *scanCell {
	c := &scanCell{}

	c.label = widget.NewLabel("")
	c.label.TextStyle = fyne.TextStyle{Monospace: true}
	c.label.Truncation = fyne.TextTruncateEllipsis

	c.copyBtn = widget.NewButtonWithIcon("", wgtheme.IconCopy, nil)
	c.copyBtn.Importance = widget.LowImportance

	c.sshBtn = widget.NewButtonWithIcon(l.T(lang.Ssh_connect_btn), wgtheme.IconTerminal, nil)
	c.sshBtn.Importance = widget.MediumImportance

	c.actions = container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceXS), c.copyBtn, c.sshBtn)
	c.actions.Hide()

	c.container = container.NewStack(c.label, c.actions)

	c.ExtendBaseWidget(c)
	return c
}

// SetText renders the cell as a value.
func (c *scanCell) SetText(text string) {
	c.actions.Hide()
	c.label.Show()
	c.label.SetText(text)
}

// SetActions renders the cell as the row's action buttons.
func (c *scanCell) SetActions(onCopy, onSSH func()) {
	c.label.Hide()
	c.actions.Show()
	c.copyBtn.OnTapped = onCopy
	c.sshBtn.OnTapped = onSSH
}

func (c *scanCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}
