package helpers

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// FileOpts describes a file picker request.
type FileOpts struct {
	Title       string
	FilterLabel string   // e.g. "WireGuard config"
	Extensions  []string // e.g. ".conf"; empty means any file
	FileName    string   // suggested name, save dialogs only
}

// pickerBusy stops a second picker opening while a native one is on screen.
// The native dialog is a separate process, so the main window stays clickable
// behind it and nothing else would prevent a double open.
var pickerBusy atomic.Bool

// OpenFile asks for an existing file and calls onPick with its path on the UI
// thread. Cancelling calls nothing.
//
// The desktop's own picker (kdialog or zenity) is used when one is installed:
// it has places, recent files, type-ahead and the user's bookmarks, none of
// which Fyne's built-in dialog offers. Fyne's dialog is the fallback.
func OpenFile(parent fyne.Window, opts FileOpts, onPick func(path string)) {
	pick(parent, opts, false, onPick)
}

// SaveFile asks for a destination path and calls onPick with it on the UI
// thread. Cancelling calls nothing.
func SaveFile(parent fyne.Window, opts FileOpts, onPick func(path string)) {
	pick(parent, opts, true, onPick)
}

func pick(parent fyne.Window, opts FileOpts, save bool, onPick func(string)) {
	cmd := nativePicker(opts, save)
	if cmd == nil {
		fynePicker(parent, opts, save, onPick)
		return
	}
	if !pickerBusy.CompareAndSwap(false, true) {
		return
	}

	go func() {
		defer pickerBusy.Store(false)

		out, err := cmd.Output()
		path := strings.TrimSpace(string(out))

		var exitErr *exec.ExitError
		switch {
		case err == nil && path != "":
			fyne.Do(func() { onPick(path) })
		case errors.As(err, &exitErr) && exitErr.ExitCode() == 1:
			// Both kdialog and zenity exit 1 on cancel.
		default:
			// The helper is installed but could not run (no display access,
			// crashed); fall back rather than leave the click doing nothing.
			fyne.Do(func() { fynePicker(parent, opts, save, onPick) })
		}
	}()
}

// nativePicker builds the command for the desktop's picker, or nil when none
// is available. KDE gets kdialog first so the dialog matches Dolphin.
func nativePicker(opts FileOpts, save bool) *exec.Cmd {
	order := []string{"zenity", "kdialog"}
	if strings.Contains(strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP")), "KDE") {
		order = []string{"kdialog", "zenity"}
	}

	start := userHome()
	if save && opts.FileName != "" {
		start = filepath.Join(start, opts.FileName)
	}

	patterns := make([]string, 0, len(opts.Extensions))
	for _, ext := range opts.Extensions {
		patterns = append(patterns, "*"+ext)
	}
	glob := strings.Join(patterns, " ")

	for _, tool := range order {
		bin, err := exec.LookPath(tool)
		if err != nil {
			continue
		}
		switch tool {
		case "kdialog":
			mode := "--getopenfilename"
			if save {
				mode = "--getsavefilename"
			}
			args := []string{"--title", opts.Title, mode, start}
			if glob != "" {
				args = append(args, glob+"|"+opts.FilterLabel+" ("+glob+")")
			}
			return exec.Command(bin, args...)
		case "zenity":
			args := []string{"--file-selection", "--title=" + opts.Title}
			if save {
				args = append(args, "--save", "--confirm-overwrite", "--filename="+start)
			} else {
				args = append(args, "--filename="+start+string(filepath.Separator))
			}
			if glob != "" {
				args = append(args, "--file-filter="+opts.FilterLabel+" | "+glob, "--file-filter=* | *")
			}
			return exec.Command(bin, args...)
		}
	}
	return nil
}

// fynePicker is the built-in fallback, started in the user's home directory and
// filtered to the requested extensions so it is at least short to scan.
func fynePicker(parent fyne.Window, opts FileOpts, save bool, onPick func(string)) {
	var d *dialog.FileDialog
	if save {
		d = dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil {
				ShowError(err, parent)
				return
			}
			if w == nil {
				return
			}
			path := w.URI().Path()
			_ = w.Close()
			onPick(path)
		}, parent)
		if opts.FileName != "" {
			d.SetFileName(opts.FileName)
		}
	} else {
		d = dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				ShowError(err, parent)
				return
			}
			if r == nil {
				return
			}
			path := r.URI().Path()
			_ = r.Close()
			onPick(path)
		}, parent)
	}

	if len(opts.Extensions) > 0 {
		d.SetFilter(storage.NewExtensionFileFilter(opts.Extensions))
	}
	if l, err := storage.ListerForURI(storage.NewFileURI(userHome())); err == nil {
		d.SetLocation(l)
	}
	d.Resize(fyne.NewSize(820, 600))
	d.Show()
}

// userHome is the home of the person at the keyboard. Once relaunched through
// pkexec or sudo the process HOME is /root, which is never where their configs
// were downloaded to.
func userHome() string {
	if uid := os.Getenv("PKEXEC_UID"); uid != "" {
		if u, err := user.LookupId(uid); err == nil {
			return u.HomeDir
		}
	}
	if name := os.Getenv("SUDO_USER"); name != "" {
		if u, err := user.Lookup(name); err == nil {
			return u.HomeDir
		}
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "/"
}
