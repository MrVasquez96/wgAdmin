package ui

import (
	"fmt"
	"time"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/MrVasquez96/go-wg/wg"
)

// BackupView displays available backups and allows restoring them
type BackupView struct {
	window    fyne.Window
	ctrl      *wg.WG
	lang      *lang.Language
	onRestore func()

	win           fyne.Window
	listContainer *fyne.Container
}

// NewBackupView creates a new backup/restore view
func NewBackupView(parent fyne.Window, ctrl *wg.WG, onRestore func(), language *lang.Language) *BackupView {
	return &BackupView{
		window:        parent,
		ctrl:          ctrl,
		lang:          language,
		onRestore:     onRestore,
		listContainer: container.NewVBox(),
	}
}

// Show opens the backup/restore window
func (bv *BackupView) Show() {
	bv.win = fyne.CurrentApp().NewWindow(bv.lang.T(lang.Backup_title))
	bv.win.Resize(fyne.NewSize(700, 500))

	// Clean old backups button
	cleanBtn := widget.NewButtonWithIcon(bv.lang.T(lang.Backup_btn_clean_old), theme.DeleteIcon(), func() {
		helpers.ShowConfirm(bv.lang.T(lang.Backup_clean_confirm_title),
			bv.lang.T(lang.Backup_clean_confirm_msg),
			func(yes bool) {
				if !yes {
					return
				}
				removed, err := bv.ctrl.CleanOldBackups(30 * 24 * time.Hour)
				if err != nil {
					helpers.ShowError(fmt.Errorf(bv.lang.T(lang.Backup_error_cleanup), err), bv.win)
					return
				}
				helpers.ShowInformation(bv.lang.T(lang.Backup_clean_success_title),
					bv.lang.Tf(lang.Backup_clean_success_msg, removed), bv.win)
				bv.refresh()
			}, bv.win)
	})
	cleanBtn.Importance = widget.DangerImportance

	header := container.NewHBox(cleanBtn)

	scroll := container.NewVScroll(bv.listContainer)
	scroll.SetMinSize(fyne.NewSize(660, 400))

	content := container.NewBorder(container.NewPadded(header), nil, nil, nil, scroll)
	bv.win.SetContent(container.NewPadded(content))

	bv.refresh()
	bv.win.Show()
}

func (bv *BackupView) refresh() {
	bv.listContainer.RemoveAll()

	backups, err := bv.ctrl.ListBackups()
	if err != nil {
		bv.listContainer.Add(widget.NewLabel(bv.lang.Tf(lang.Backup_error_loading, err)))
		return
	}

	if len(backups) == 0 {
		bv.listContainer.Add(widget.NewLabel(bv.lang.T(lang.Backup_no_backups)))
		return
	}

	for _, b := range backups {
		backup := b // capture for closure

		nameLabel := widget.NewLabel(backup.Name)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		timeLabel := widget.NewLabel(backup.Timestamp.Format("2006-01-02 15:04:05"))

		restoreBtn := widget.NewButtonWithIcon(bv.lang.T(lang.Common_restore), theme.HistoryIcon(), func() {
			msg := bv.lang.Tf(lang.Backup_restore_confirm_msg, backup.Filename, backup.Name)
			helpers.ShowConfirm(bv.lang.T(lang.Backup_restore_confirm_title), msg, func(yes bool) {
				if !yes {
					return
				}
				if err := bv.ctrl.RestoreBackup(backup.Filename); err != nil {
					helpers.ShowError(fmt.Errorf(bv.lang.T(lang.Backup_error_restore), err), bv.win)
					return
				}
				helpers.ShowInformation(bv.lang.T(lang.Backup_restore_success_title),
					bv.lang.Tf(lang.Backup_restore_success_msg, backup.Name), bv.win)
				if bv.onRestore != nil {
					bv.onRestore()
				}
			}, bv.win)
		})
		restoreBtn.Importance = widget.HighImportance

		infoBox := container.NewVBox(nameLabel, timeLabel)
		row := container.NewBorder(nil, nil, infoBox, restoreBtn)
		bv.listContainer.Add(container.NewPadded(row))
		bv.listContainer.Add(widget.NewSeparator())
	}
}
