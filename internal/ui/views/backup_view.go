package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	"github.com/MrVasquez96/go-wg/wg"
)

// BackupView lists the config backups wgAdmin writes before every change and
// restores any of them, as a sheet over the main window.
type BackupView struct {
	window    fyne.Window
	ctrl      *wg.WG
	lang      *lang.Language
	onRestore func()

	backups []wg.BackupInfo
	table   *wgwidget.DataTable
	body    *fyne.Container
	count   *fyne.Container
}

// NewBackupView creates a new backup/restore view
func NewBackupView(parent fyne.Window, ctrl *wg.WG, onRestore func(), language *lang.Language) *BackupView {
	return &BackupView{window: parent, ctrl: ctrl, lang: language, onRestore: onRestore}
}

const (
	backupColTunnel = iota
	backupColTaken
	backupColDate
)

// Show opens the backups sheet.
func (bv *BackupView) Show() {
	l := bv.lang

	cols := []wgwidget.Column{
		{Title: l.T(lang.Backup_col_tunnel), MinWidth: 140, Weight: 1, Sortable: true},
		{Title: l.T(lang.Backup_col_taken), MinWidth: 110, Sortable: true},
		{Title: "", MinWidth: 160, Weight: 1, Mono: true},
	}
	now := time.Now()
	bv.table = wgwidget.NewDataTable(cols, func() int { return len(bv.backups) }, func(r, c int) wgwidget.Cell {
		b := bv.backups[r]
		switch c {
		case backupColTunnel:
			return wgwidget.Cell{Text: b.Name, Strong: true}
		case backupColTaken:
			return wgwidget.Cell{Text: wgwidget.RelativeTime(b.Timestamp, now)}
		}
		return wgwidget.Cell{Text: b.Timestamp.Format("2006-01-02 15:04:05"), Dim: true}
	})
	bv.table.SortCol, bv.table.SortAsc = backupColTaken, true
	bv.table.OnSort = func(col int, asc bool) { bv.sort(col, asc) }
	bv.table.Actions = []wgwidget.RowAction{
		{Icon: wgtheme.IconRefresh, Tooltip: l.T(lang.Common_restore), OnTap: func(r int) {
			bv.restore(bv.backups[r])
		}},
	}

	clean := wgwidget.NewButton(l.T(lang.Backup_btn_clean_old), wgtheme.IconTrash, wgwidget.Danger, bv.clean)
	clean.Scale = wgwidget.ScaleSM

	bv.count = container.NewStack()
	bv.body = container.NewStack()
	header := container.NewBorder(nil, nil, container.NewCenter(bv.count), clean)

	content := container.NewBorder(
		container.New(layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, 0), header),
		nil, nil, nil, bv.body)

	wgwidget.ShowSheetFill(bv.window, l.T(lang.Backup_title), wgwidget.SheetWide, content)
	bv.refresh()
}

// sort orders the backups. "Taken" ascending means newest first, which is the
// default: the backup most likely to be wanted is the most recent one.
func (bv *BackupView) sort(col int, asc bool) {
	sort.SliceStable(bv.backups, func(i, j int) bool {
		a, b := bv.backups[i], bv.backups[j]
		var less bool
		switch col {
		case backupColTunnel:
			if a.Name == b.Name {
				return a.Timestamp.After(b.Timestamp)
			}
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		default:
			less = a.Timestamp.After(b.Timestamp)
		}
		if asc {
			return less
		}
		return !less
	})
}

func (bv *BackupView) refresh() {
	backups, err := bv.ctrl.ListBackups()
	bv.backups = backups

	switch {
	case err != nil:
		bv.body.Objects = []fyne.CanvasObject{emptyNote(bv.lang.Tf(lang.Backup_error_loading, err))}
	case len(backups) == 0:
		bv.body.Objects = []fyne.CanvasObject{emptyNote(bv.lang.T(lang.Backup_no_backups))}
	default:
		bv.sort(bv.table.SortCol, bv.table.SortAsc)
		bv.body.Objects = []fyne.CanvasObject{bv.table}
		bv.table.Refresh()
	}
	bv.body.Refresh()

	bv.count.Objects = []fyne.CanvasObject{
		wgwidget.Badge(bv.lang.Tf(lang.Backup_count, len(bv.backups)), false),
	}
	bv.count.Refresh()
}

func (bv *BackupView) restore(backup wg.BackupInfo) {
	msg := bv.lang.Tf(lang.Backup_restore_confirm_msg, backup.Filename, backup.Name)
	helpers.ShowConfirm(bv.lang.T(lang.Backup_restore_confirm_title), msg, func(yes bool) {
		if !yes {
			return
		}
		if err := bv.ctrl.RestoreBackup(backup.Filename); err != nil {
			helpers.ShowError(fmt.Errorf(bv.lang.T(lang.Backup_error_restore), err), bv.window)
			return
		}
		wgwidget.ShowToast(bv.window, wgwidget.ToastSuccess,
			bv.lang.Tf(lang.Backup_restore_success_msg, backup.Name))
		if bv.onRestore != nil {
			bv.onRestore()
		}
		bv.refresh()
	}, bv.window)
}

func (bv *BackupView) clean() {
	helpers.ShowConfirm(bv.lang.T(lang.Backup_clean_confirm_title),
		bv.lang.T(lang.Backup_clean_confirm_msg), func(yes bool) {
			if !yes {
				return
			}
			removed, err := bv.ctrl.CleanOldBackups(30 * 24 * time.Hour)
			if err != nil {
				helpers.ShowError(fmt.Errorf(bv.lang.T(lang.Backup_error_cleanup), err), bv.window)
				return
			}
			wgwidget.ShowToast(bv.window, wgwidget.ToastSuccess,
				bv.lang.Tf(lang.Backup_clean_success_msg, removed))
			bv.refresh()
		}, bv.window)
}
