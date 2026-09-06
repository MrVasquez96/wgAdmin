package wgwidget

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
)

// SSHTerminal opens a window with a scrollable output area and command input
// backed by an SSH session. Call Show() to connect and display the window.
type SSHTerminal struct {
	host     string
	username string
	password string
	hostname string
	port     int
	lang     *lang.Language

	win         fyne.Window
	outputBuf   strings.Builder
	output      *widget.TextGrid
	outputScrol *container.Scroll
	inputEntry  *commandEntry
	statusLabel *canvas.Text
	statusDot   *canvas.Circle

	// history holds previously sent commands for up/down recall.
	history    []string
	historyPos int

	lastCommand string
	lastLine    string

	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
}

func getHostname() string {
	contentBytes, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(contentBytes), "\n", "")
}

// NewSSHTerminal creates a new SSH terminal component for the given host.
func NewSSHTerminal(host, username, password string, port int, language *lang.Language) *SSHTerminal {

	return &SSHTerminal{
		host:     host,
		username: username,
		password: password,
		hostname: getHostname(),
		port:     port,
		lang:     language,
	}
}

// Show opens the terminal window and initiates the SSH connection.
func (t *SSHTerminal) Show() {
	p := wgtheme.Current()

	t.win = fyne.CurrentApp().NewWindow(t.lang.Tf(lang.Ssh_window_title, t.host))
	t.win.Resize(fyne.NewSize(880, 560))
	t.win.SetOnClosed(t.cleanup)

	t.statusDot = canvas.NewCircle(p.Warning)
	t.statusLabel = canvas.NewText(t.lang.Tf(lang.Ssh_status_connecting, t.host), p.TextSecondary)
	t.statusLabel.TextSize = 12

	// TextGrid rather than a Label: the previous implementation re-set the
	// entire accumulated buffer on every read from the socket, so a verbose
	// command turned into repeated full-buffer re-layouts.
	t.output = widget.NewTextGrid()

	if t.port < 1 || t.port > 65535 {
		t.appendOutput(fmt.Sprintf("port %d is out of range (1-65535); using 22\n", t.port))
		t.port = 22
	}

	t.outputScrol = container.NewVScroll(t.output)

	t.inputEntry = newCommandEntry(t.recallHistory)
	t.inputEntry.TextStyle = fyne.TextStyle{Monospace: true}
	t.inputEntry.SetPlaceHolder(t.lang.T(lang.Ssh_input_placeholder))
	t.inputEntry.Disable()
	t.inputEntry.OnSubmitted = t.sendCommand

	prompt := Value("$")
	prompt.Color = p.Accent

	header := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM),
		container.NewCenter(container.New(
			layout.NewGridWrapLayout(fyne.NewSize(8, 8)), t.statusDot)),
		container.NewCenter(t.statusLabel),
	)

	inputRow := container.NewBorder(nil, nil,
		container.NewCenter(prompt), nil, t.inputEntry)

	terminalBg := canvas.NewRectangle(p.Background)

	t.win.SetContent(container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceMD),
		container.NewBorder(
			container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM), header, Hairline()),
			container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceMD, 0, 0, 0), inputRow),
			nil, nil,
			container.NewStack(terminalBg, t.outputScrol),
		)))

	t.win.Show()
	t.win.Canvas().Focus(t.inputEntry)

	go t.connect()
}

// setStatus updates the connection indicator.
func (t *SSHTerminal) setStatus(text string, tone color.Color) {
	t.statusLabel.Text = text
	t.statusLabel.Refresh()
	t.statusDot.FillColor = tone
	t.statusDot.Refresh()
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

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", t.host, t.port), cfg)
	if err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_connect, err) + "\n")
			t.setStatus(t.lang.T(lang.Ssh_status_disconnected), wgtheme.Current().Error)
		})
		return
	}
	t.client = client

	session, err := client.NewSession()
	if err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_session, err) + "\n")
			t.setStatus(t.lang.T(lang.Ssh_status_disconnected), wgtheme.Current().Error)
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
			t.setStatus(t.lang.T(lang.Ssh_status_disconnected), wgtheme.Current().Error)
		})
		return
	}
	t.stdin = stdinPipe

	// Request a pseudo-terminal so interactive commands work correctly.
	// Failure is non-fatal; the terminal still functions without PTY support.
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	_ = session.RequestPty("vt100", 40, 120, modes)

	if err := session.Shell(); err != nil {
		fyne.Do(func() {
			t.appendOutput(t.lang.Tf(lang.Ssh_error_shell, err) + "\n")
			t.setStatus(t.lang.T(lang.Ssh_status_disconnected), wgtheme.Current().Error)
		})
		return
	}

	fyne.Do(func() {
		t.setStatus(t.lang.Tf(lang.Ssh_status_connected, t.host), wgtheme.Current().Success)
		t.inputEntry.Enable()
	})

	go t.pumpReader(stdoutPipe)
	go t.pumpReader(stderrPipe)

	_ = session.Wait()

	fyne.Do(func() {
		t.appendOutput("\n[" + t.lang.T(lang.Ssh_status_disconnected) + "]\n")
		t.setStatus(t.lang.T(lang.Ssh_status_disconnected), wgtheme.Current().Error)
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

// maxTerminalBytes caps the scrollback. An unbounded buffer is a slow leak in
// a long-lived session, and no one scrolls back through megabytes anyway.
const maxTerminalBytes = 256 * 1024

// appendOutput appends text to the terminal. Must be called on the UI goroutine.
func (t *SSHTerminal) appendOutput(text string) {
	t.outputBuf.WriteString(text)

	buf := t.outputBuf.String()
	if len(buf) > maxTerminalBytes {
		// Trim on a line boundary so the top of the view is not a half line.
		trimmed := buf[len(buf)-maxTerminalBytes:]
		if nl := strings.IndexByte(trimmed, '\n'); nl >= 0 {
			trimmed = trimmed[nl+1:]
		}
		t.outputBuf.Reset()
		t.outputBuf.WriteString(trimmed)
		buf = trimmed
	}

	t.output.SetText(buf)
	if t.outputScrol != nil {
		t.outputScrol.ScrollToBottom()
	}
}

func (t *SSHTerminal) sendCommand(cmd string) {
	if t.stdin == nil {
		return
	}
	t.lastCommand = cmd
	if strings.TrimSpace(cmd) != "" {
		t.history = append(t.history, cmd)
	}
	t.historyPos = len(t.history)
	t.inputEntry.SetText("")
	_, _ = fmt.Fprintln(t.stdin, cmd)
}

// commandEntry is a single-line entry that hands the up/down arrows to a
// callback instead of letting Entry move the caret, giving the terminal
// shell-style history recall.
type commandEntry struct {
	widget.Entry
	onRecall func(delta int)
}

func newCommandEntry(onRecall func(delta int)) *commandEntry {
	e := &commandEntry{onRecall: onRecall}
	e.ExtendBaseWidget(e)
	return e
}

func (e *commandEntry) TypedKey(ev *fyne.KeyEvent) {
	switch ev.Name {
	case fyne.KeyUp:
		e.onRecall(-1)
	case fyne.KeyDown:
		e.onRecall(1)
	default:
		e.Entry.TypedKey(ev)
	}
}

// recallHistory steps through previously sent commands.
func (t *SSHTerminal) recallHistory(delta int) {
	if len(t.history) == 0 {
		return
	}
	pos := t.historyPos + delta
	if pos < 0 {
		pos = 0
	}
	if pos >= len(t.history) {
		// Past the newest entry is an empty prompt, as in a real shell.
		t.historyPos = len(t.history)
		t.inputEntry.SetText("")
		return
	}
	t.historyPos = pos
	t.inputEntry.SetText(t.history[pos])
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
