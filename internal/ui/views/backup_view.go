package ui

import (
	"fmt"
	"sort"
	"time"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	countLabel    *canvas.Text
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
	bv.win.Resize(fyne.NewSize(720, 560))

	cleanBtn := widget.NewButtonWithIcon(bv.lang.T(lang.Backup_btn_clean_old), wgtheme.IconTrash, func() {
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

	bv.countLabel = wgwidget.Caption("")

	header := container.NewBorder(nil, nil,
		container.NewCenter(bv.countLabel), cleanBtn)

	content := container.NewBorder(
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD),
			header, wgwidget.Hairline()),
		nil, nil, nil,
		container.NewVScroll(bv.listContainer),
	)

	bv.win.SetContent(container.New(layout.NewCustomPaddedLayout(
		wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG), content))

	bv.refresh()
	bv.win.Show()
}

func (bv *BackupView) refresh() {
	bv.listContainer.RemoveAll()

	backups, err := bv.ctrl.ListBackups()
	if err != nil {
		bv.listContainer.Add(emptyNote(bv.lang.Tf(lang.Backup_error_loading, err)))
		bv.setCount(0)
		return
	}

	if len(backups) == 0 {
		bv.listContainer.Add(emptyNote(bv.lang.T(lang.Backup_no_backups)))
		bv.setCount(0)
		return
	}

	// Newest first, so the backup most likely to be wanted is at the top.
	sort.SliceStable(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	now := time.Now()
	var lastName string

	for _, b := range backups {
		backup := b // capture for the closure

		// Group by tunnel: a heading is emitted whenever the name changes, so
		// a long flat list becomes scannable per tunnel.
		if backup.Name != lastName {
			if lastName != "" {
				bv.listContainer.Add(container.New(
					layout.NewCustomPaddedLayout(wgtheme.SpaceMD, 0, 0, 0), wgwidget.Hairline()))
			}
			bv.listContainer.Add(wgwidget.SectionHeader(backup.Name))
			lastName = backup.Name
		}

		relative := wgwidget.Value(wgwidget.RelativeTime(backup.Timestamp, now))
		exact := wgwidget.DimValue(backup.Timestamp.Format("2006-01-02 15:04:05"))
		exact.TextSize = 11

		restoreBtn := widget.NewButtonWithIcon(bv.lang.T(lang.Common_restore), wgtheme.IconClock, func() {
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
				bv.refresh()
			}, bv.win)
		})

		info := container.New(layout.NewCustomPaddedVBoxLayout(1), relative, exact)
		row := container.New(layout.NewCustomPaddedLayout(
			wgtheme.SpaceSM, wgtheme.SpaceSM, wgtheme.SpaceMD, 0),
			container.NewBorder(nil, nil, info, restoreBtn))

		bv.listContainer.Add(row)
	}

	bv.setCount(len(backups))
}

// setCount updates the header counter.
func (bv *BackupView) setCount(n int) {
	if bv.countLabel == nil {
		return
	}
	bv.countLabel.Text = bv.lang.Tf(lang.Backup_count, n)
	bv.countLabel.Refresh()
}
