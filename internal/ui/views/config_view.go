package ui

import (
	"errors"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

	sheet    *wgwidget.Surface
	viewer   *wgwidget.CodeView
	editor   *widget.Entry
	body     *fyne.Container
	note     *widget.Label
	editBtn  *wgwidget.Button
	saveBtn  *wgwidget.Button
	revertBt *wgwidget.Button
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

	c.viewer = wgwidget.NewCodeView(c.original)

	c.editor = widget.NewMultiLineEntry()
	c.editor.TextStyle = fyne.TextStyle{Monospace: true}
	c.editor.SetText(c.original)
	c.editor.Wrapping = fyne.TextWrapOff

	c.note = widget.NewLabel(c.lang.T(lang.Config_read_only_note))
	c.note.Importance = widget.LowImportance
	c.note.Wrapping = fyne.TextWrapWord

	c.body = container.NewStack(container.NewVScroll(c.viewer.Object()))

	copyBtn := wgwidget.NewButton(c.lang.T(lang.Common_copy), wgtheme.IconCopy, wgwidget.Secondary, func() {
		c.parent.Clipboard().SetContent(c.currentText())
		wgwidget.ShowToast(c.parent, wgwidget.ToastSuccess, c.lang.T(lang.Common_copied))
	})
	c.editBtn = wgwidget.NewButton(c.lang.T(lang.Config_btn_edit), wgtheme.IconEdit, wgwidget.Secondary, c.startEditing)
	c.revertBt = wgwidget.NewButton(c.lang.T(lang.Config_btn_revert), wgtheme.IconRefresh, wgwidget.Ghost, c.revert)
	c.revertBt.Hide()
	c.saveBtn = wgwidget.NewButton(c.lang.T(lang.Config_btn_save), wgtheme.IconCheck, wgwidget.Primary, c.save)
	c.saveBtn.Hide()

	// The code sits on the input surface so it reads as a document, framed
	// apart from the sheet's chrome.
	p := wgtheme.Current()
	codeBg := canvas.NewRectangle(p.InputBackground)
	codeBg.CornerRadius = wgtheme.RadiusLG
	codeBg.StrokeColor = p.Border
	codeBg.StrokeWidth = 1
	code := container.NewStack(codeBg, container.New(layout.NewCustomPaddedLayout(4, 4, 4, 4), c.body))

	header := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM),
		wgwidget.Card(wgwidget.NewCopyField(wgtheme.IconDoc, c.lang.T(lang.Detail_label_config), path, func(v string) {
			c.parent.Clipboard().SetContent(v)
			wgwidget.ShowToast(c.parent, wgwidget.ToastSuccess, c.lang.T(lang.Common_copied))
		}), 0),
		c.note,
	)

	content := container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, 0), header),
		nil, nil, nil, code)

	c.sheet = wgwidget.ShowSheetFill(c.parent, c.lang.Tf(lang.Config_title, c.name), wgwidget.SheetWide, content,
		copyBtn, c.editBtn, c.revertBt, c.saveBtn)
	c.sheet.Intercept = c.requestClose
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
		helpers.ShowError(errors.New(c.lang.Tf(lang.Config_invalid, err.Error())), c.parent)
		return
	}
	if err := parsed.Validate(); err != nil {
		helpers.ShowError(errors.New(c.lang.Tf(lang.Config_invalid, err.Error())), c.parent)
		return
	}

	// WriteConfig takes a backup, so a bad edit is recoverable from Backups.
	if err := c.ctrl.WriteConfig(c.name, *parsed); err != nil {
		helpers.ShowError(err, c.parent)
		return
	}

	c.original = text
	c.stopEditing(text)

	wgwidget.ShowToast(c.parent, wgwidget.ToastSuccess, c.lang.Tf(lang.Config_saved, c.name))

	if c.onSave != nil {
		c.onSave()
	}
}

// requestClose guards against losing unsaved edits.
func (c *ConfigView) requestClose(close func()) {
	if c.editing && c.editor.Text != c.original {
		helpers.ShowConfirm(c.lang.T(lang.Config_confirm_title),
			c.lang.T(lang.Config_confirm_message), func(discard bool) {
				if discard {
					close()
				}
			}, c.parent)
		return
	}
	close()
}
