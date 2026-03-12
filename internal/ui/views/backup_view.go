package ui

import (
	"fmt"
	"time"

	"wgAdmin/internal/ui/helpers"

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
	onRestore func()

	win           fyne.Window
	listContainer *fyne.Container
}

// NewBackupView creates a new backup/restore view
func NewBackupView(parent fyne.Window, ctrl *wg.WG, onRestore func()) *BackupView {
	return &BackupView{
		window:        parent,
		ctrl:          ctrl,
		onRestore:     onRestore,
		listContainer: container.NewVBox(),
	}
}

// Show opens the backup/restore window
func (bv *BackupView) Show() {
	bv.win = fyne.CurrentApp().NewWindow("Backups")
	bv.win.Resize(fyne.NewSize(700, 500))

	// Clean old backups button
	cleanBtn := widget.NewButtonWithIcon("Delete Backups Older Than 30 Days", theme.DeleteIcon(), func() {
		helpers.ShowConfirm("Clean Old Backups",
			"Delete all backup files older than 30 days?",
			func(yes bool) {
				if !yes {
					return
				}
				removed, err := bv.ctrl.CleanOldBackups(30 * 24 * time.Hour)
				if err != nil {
					helpers.ShowError(fmt.Errorf("cleanup failed: %w", err), bv.win)
					return
				}
				helpers.ShowInformation("Cleanup Complete",
					fmt.Sprintf("Removed %d old backup(s).", removed), bv.win)
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
		bv.listContainer.Add(widget.NewLabel(fmt.Sprintf("Error loading backups: %v", err)))
		return
	}

	if len(backups) == 0 {
		bv.listContainer.Add(widget.NewLabel("No backups found."))
		return
	}

	for _, b := range backups {
		backup := b // capture for closure

		nameLabel := widget.NewLabel(backup.Name)
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		timeLabel := widget.NewLabel(backup.Timestamp.Format("2006-01-02 15:04:05"))

		restoreBtn := widget.NewButtonWithIcon("Restore", theme.HistoryIcon(), func() {
			msg := fmt.Sprintf("Restore '%s' from backup?\n\nThis will overwrite %s.conf if it exists.",
				backup.Filename, backup.Name)
			helpers.ShowConfirm("Restore Backup", msg, func(yes bool) {
				if !yes {
					return
				}
				if err := bv.ctrl.RestoreBackup(backup.Filename); err != nil {
					helpers.ShowError(fmt.Errorf("restore failed: %w", err), bv.win)
					return
				}
				helpers.ShowInformation("Restored",
					fmt.Sprintf("Successfully restored %s", backup.Name), bv.win)
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
