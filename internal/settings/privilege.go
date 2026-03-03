package settings

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// PkexecAvailable checks if pkexec is installed on the system.
func PkexecAvailable() bool {
	_, err := exec.LookPath("pkexec")
	return err == nil
}

// RelaunchWithPkexec spawns a new root process through pkexec.
// pkexec strips most environment variables, so we use env(1) to explicitly
// pass the display-related vars the GUI needs to connect to X11/Wayland.
func RelaunchWithPkexec() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	pkexecPath, err := exec.LookPath("pkexec")
	if err != nil {
		return fmt.Errorf("pkexec not found: %w", err)
	}
	envPath, err := exec.LookPath("env")
	if err != nil {
		return fmt.Errorf("env not found: %w", err)
	}

	// pkexec strips env vars. We use: pkexec env DISPLAY=... XAUTHORITY=... /path/to/exe
	args := []string{pkexecPath, envPath}
	for _, key := range displayEnvKeys {
		if val := os.Getenv(key); val != "" {
			args = append(args, fmt.Sprintf("%s=%s", key, val))
		}
	}
	args = append(args, exe)
	args = append(args, os.Args[1:]...)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// ShowSudoRelaunchDialog shows a password dialog and relaunches the app with sudo.
// On success the current app is quit gracefully via app.Quit().
func ShowSudoRelaunchDialog(parent fyne.Window) {
	entry := widget.NewPasswordEntry()
	entry.SetPlaceHolder("Enter sudo password")

	formItem := widget.NewFormItem("Password", entry)
	d := dialog.NewForm("Root Access Required", "Authenticate", "Skip",
		[]*widget.FormItem{formItem},
		func(confirmed bool) {
			if !confirmed || entry.Text == "" {
				return
			}
			if err := relaunchWithSudo(entry.Text); err != nil {
				dialog.ShowError(fmt.Errorf("sudo failed: %w", err), parent)
				return
			}
			// New root process launched — quit this one gracefully
			fyne.CurrentApp().Quit()
		}, parent)
	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}

// displayEnvKeys are the environment variables needed for GUI apps to work
// when relaunched as root.
var displayEnvKeys = []string{
	"DISPLAY",
	"XAUTHORITY",
	"WAYLAND_DISPLAY",
	"XDG_RUNTIME_DIR",
	"DBUS_SESSION_BUS_ADDRESS",
}

func relaunchWithSudo(password string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// sudo --preserve-env passes the display vars so the GUI can connect.
	preserveList := strings.Join(displayEnvKeys, ",")
	args := []string{"-S", "--preserve-env=" + preserveList, exe}
	args = append(args, os.Args[1:]...)

	cmd := exec.Command("sudo", args...)
	cmd.Stdin = strings.NewReader(password + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Start()
}
