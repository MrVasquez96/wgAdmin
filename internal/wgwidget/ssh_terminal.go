package wgwidget

import (
	"fmt"
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"

	"wgAdmin/internal/ui/lang"
)

// SSHTerminal opens a window with a scrollable output area and command input
// backed by an SSH session. Call Show() to connect and display the window.
type SSHTerminal struct {
	host     string
	username string
	password string
	lang     *lang.Language

	win         fyne.Window
	outputBuf   strings.Builder
	outputLabel *widget.Label
	inputEntry  *widget.Entry
	statusLabel *widget.Label

	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
}

// NewSSHTerminal creates a new SSH terminal component for the given host.
func NewSSHTerminal(host, username, password string, language *lang.Language) *SSHTerminal {
	return &SSHTerminal{
		host:     host,
		username: username,
		password: password,
		lang:     language,
	}
}

// Show opens the terminal window and initiates the SSH connection.
func (t *SSHTerminal) Show() {
	t.win = fyne.CurrentApp().NewWindow(t.lang.Tf(lang.Ssh_window_title, t.host))
	t.win.Resize(fyne.NewSize(800, 500))
	t.win.SetOnClosed(t.cleanup)

	t.statusLabel = widget.NewLabel(t.lang.Tf(lang.Ssh_status_connecting, t.host))

	t.outputLabel = widget.NewLabel("")
	t.outputLabel.TextStyle = fyne.TextStyle{Monospace: true}
	t.outputLabel.Wrapping = fyne.TextWrapOff
	outputScroll := container.NewVScroll(t.outputLabel)

	t.inputEntry = widget.NewEntry()
	t.inputEntry.SetPlaceHolder(t.lang.T(lang.Ssh_input_placeholder))
	t.inputEntry.Disable()
	t.inputEntry.OnSubmitted = t.sendCommand

	t.win.SetContent(container.NewPadded(container.NewBorder(
		t.statusLabel,
		container.NewPadded(t.inputEntry),
		nil, nil,
		outputScroll,
	)))
	t.win.Show()

	go t.connect()
}

func (t *SSHTerminal) connect() {
	cfg := &ssh.ClientConfig{
		User: t.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(t.password),
		},
		// InsecureIgnoreHostKey is acceptable for an internal LAN/VPN tool.
		// A production implementation would use knownhosts.New() here.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", t.host), cfg)
	if err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_connect, err) + "\n")
			t.statusLabel.SetText(t.lang.T(lang.Ssh_status_disconnected))
		})
		return
	}
	t.client = client

	session, err := client.NewSession()
	if err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_session, err) + "\n")
			t.statusLabel.SetText(t.lang.T(lang.Ssh_status_disconnected))
		})
		return
	}
	t.session = session

	stdoutPipe, _ := session.StdoutPipe()
	stderrPipe, _ := session.StderrPipe()
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_shell, err) + "\n")
			t.statusLabel.SetText(t.lang.T(lang.Ssh_status_disconnected))
		})
		return
	}
	t.stdin = stdinPipe

	// Request a pseudo-terminal so interactive commands work correctly.
	// Failure is non-fatal; the terminal still functions without PTY support.
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	_ = session.RequestPty("vt100", 40, 120, modes)

	if err := session.Shell(); err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_shell, err) + "\n")
			t.statusLabel.SetText(t.lang.T(lang.Ssh_status_disconnected))
		})
		return
	}

	fyne.Do(func() {
		t.statusLabel.SetText(t.lang.Tf(lang.Ssh_status_connected, t.host))
		t.inputEntry.Enable()
	})

	go t.pumpReader(stdoutPipe)
	go t.pumpReader(stderrPipe)

	_ = session.Wait()

	fyne.Do(func() {
		t.appendOutput("\n[" + t.lang.T(lang.Ssh_status_disconnected) + "]\n")
		t.statusLabel.SetText(t.lang.T(lang.Ssh_status_disconnected))
		t.inputEntry.Disable()
	})
}

func (t *SSHTerminal) pumpReader(r io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			text := stripANSI(string(buf[:n]))
			fyne.Do(func() {
				t.appendOutput(text)
			})
		}
		if err != nil {
			break
		}
	}
}

// appendOutput appends text to the output label. Must be called from the UI goroutine.
func (t *SSHTerminal) appendOutput(text string) {
	t.outputBuf.WriteString(text)
	t.outputLabel.SetText(t.outputBuf.String())
	t.outputLabel.Refresh()
}

func (t *SSHTerminal) sendCommand(cmd string) {
	if t.stdin == nil {
		return
	}
	t.inputEntry.SetText("")
	_, _ = fmt.Fprintln(t.stdin, cmd)
}

func (t *SSHTerminal) cleanup() {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.session != nil {
		_ = t.session.Close()
	}
	if t.client != nil {
		_ = t.client.Close()
	}
}

// stripANSI removes ANSI escape sequences (ESC [ ... letter) from s.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < 'A' || s[i] > 'Z') && (s[i] < 'a' || s[i] > 'z') {
				i++
			}
			if i < len(s) {
				i++ // skip terminating letter
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}
