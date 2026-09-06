package ui

import (
	"errors"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgcfg"
	"wgAdmin/internal/wgwidget"

	"github.com/MrVasquez96/go-wg/wg"
)

// ConfigView shows a tunnel's .conf file as it is on disk, and optionally lets
// it be edited directly — the escape hatch an advanced user wants when the
// form does not expose some option.
type ConfigView struct {
	parent fyne.Window
	ctrl   *wg.WG
	name   string
	lang   *lang.Language
	onSave func()

	original string
	editing  bool

	win      fyne.Window
	viewer   *wgwidget.CodeView
	editor   *widget.Entry
	body     *fyne.Container
	note     *widget.Label
	editBtn  *widget.Button
	saveBtn  *widget.Button
	revertBt *widget.Button
}

// NewConfigView creates a raw config viewer for a tunnel.
func NewConfigView(parent fyne.Window, ctrl *wg.WG, name string, language *lang.Language, onSave func()) *ConfigView {
	return &ConfigView{parent: parent, ctrl: ctrl, name: name, lang: language, onSave: onSave}
}

// Show opens the config window.
func (c *ConfigView) Show() {
	path := c.ctrl.GetConfigPath(c.name)

	raw, err := os.ReadFile(path)
	if err != nil {
		helpers.ShowError(errors.New(c.lang.Tf(lang.Config_error_read, err.Error())), c.parent)
		return
	}
	c.original = string(raw)

	c.win = fyne.CurrentApp().NewWindow(c.lang.Tf(lang.Config_title, c.name))
	c.win.Resize(fyne.NewSize(760, 620))

	c.viewer = wgwidget.NewCodeView(c.original)

	c.editor = widget.NewMultiLineEntry()
	c.editor.TextStyle = fyne.TextStyle{Monospace: true}
	c.editor.SetText(c.original)
	c.editor.Wrapping = fyne.TextWrapOff

	c.note = widget.NewLabel(c.lang.T(lang.Config_read_only_note))
	c.note.Importance = widget.LowImportance
	c.note.Wrapping = fyne.TextWrapWord

	c.body = container.NewStack(container.NewVScroll(c.viewer.Object()))

	copyBtn := widget.NewButtonWithIcon(c.lang.T(lang.Common_copy), wgtheme.IconCopy, func() {
		c.win.Clipboard().SetContent(c.currentText())
	})

	c.editBtn = widget.NewButtonWithIcon(c.lang.T(lang.Config_btn_edit), wgtheme.IconEdit, c.startEditing)

	c.saveBtn = widget.NewButtonWithIcon(c.lang.T(lang.Config_btn_save), wgtheme.IconImport, c.save)
	c.saveBtn.Importance = widget.HighImportance
	c.saveBtn.Hide()

	c.revertBt = widget.NewButtonWithIcon(c.lang.T(lang.Config_btn_revert), wgtheme.IconRefresh, c.revert)
	c.revertBt.Hide()

	closeBtn := widget.NewButton(c.lang.T(lang.Common_close), c.requestClose)

	header := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM),
		wgwidget.KeyValue(c.lang.T(lang.Detail_label_config), path, 70),
		c.note,
	)

	actions := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		copyBtn, c.editBtn, c.revertBt,
		layout.NewSpacer(),
		closeBtn, c.saveBtn,
	)

	content := container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, 0), header),
		container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceMD, 0, 0, 0), actions),
		nil, nil,
		c.body,
	)

	c.win.SetContent(container.New(layout.NewCustomPaddedLayout(
		wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG), content))
	c.win.SetCloseIntercept(c.requestClose)
	c.win.Show()
}

// currentText returns whichever pane is live.
func (c *ConfigView) currentText() string {
	if c.editing {
		return c.editor.Text
	}
	return c.viewer.Text()
}

func (c *ConfigView) startEditing() {
	c.editing = true
	c.editor.SetText(c.viewer.Text())

	c.body.Objects = []fyne.CanvasObject{container.NewVScroll(c.editor)}
	c.body.Refresh()

	c.note.SetText(c.lang.T(lang.Config_edit_note))
	c.editBtn.Hide()
	c.saveBtn.Show()
	c.revertBt.Show()
}

func (c *ConfigView) stopEditing(text string) {
	c.editing = false
	c.viewer.SetText(text)

	c.body.Objects = []fyne.CanvasObject{container.NewVScroll(c.viewer.Object())}
	c.body.Refresh()

	c.note.SetText(c.lang.T(lang.Config_read_only_note))
	c.editBtn.Show()
	c.saveBtn.Hide()
	c.revertBt.Hide()
}

func (c *ConfigView) revert() {
	c.editor.SetText(c.original)
	c.stopEditing(c.original)
}

// save validates the edited text before writing. Parsing goes through the same
// code path the rest of the app uses, so a file that saves here is a file the
// app can load — a syntax error never reaches disk, and an endpoint edited by
// hand survives the trip back out through the writer.
func (c *ConfigView) save() {
	text := c.editor.Text

	parsed, err := wgcfg.LoadString(text)
	if err != nil {
		helpers.ShowError(errors.New(c.lang.Tf(lang.Config_invalid, err.Error())), c.win)
		return
	}
	if err := parsed.Validate(); err != nil {
		helpers.ShowError(errors.New(c.lang.Tf(lang.Config_invalid, err.Error())), c.win)
		return
	}

	// WriteConfig takes a backup, so a bad edit is recoverable from Backups.
	if err := c.ctrl.WriteConfig(c.name, *parsed); err != nil {
		helpers.ShowError(err, c.win)
		return
	}

	c.original = text
	c.stopEditing(text)

	helpers.ShowInformation(c.lang.T(lang.Common_success),
		c.lang.Tf(lang.Config_saved, c.name), c.win)

	if c.onSave != nil {
		c.onSave()
	}
}

// requestClose guards against losing unsaved edits.
func (c *ConfigView) requestClose() {
	if c.editing && c.editor.Text != c.original {
		helpers.ShowConfirm(c.lang.T(lang.Config_confirm_title),
			c.lang.T(lang.Config_confirm_message), func(discard bool) {
				if discard {
					c.win.Close()
				}
			}, c.win)
		return
	}
	c.win.Close()
}
