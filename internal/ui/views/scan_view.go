package ui

import (
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	scanner "github.com/MrVasquez96/go-ipscan"
)

// ScanView displays network scan results
type ScanView struct {
	ifaceName    string
	cidr         string
	serverIP     string
	workers      int
	timeoutSecs  int
	lang         *lang.Language
	sortedResults []scanner.ScanResult
	resultsMu    sync.Mutex
}

// NewScanView creates a new scan view
func NewScanView(ifaceName, ip string, workers, timeoutSecs int, language *lang.Language) *ScanView {
	return &ScanView{
		ifaceName:   ifaceName,
		cidr:        scanner.DeriveCIDR24(ip),
		serverIP:    ip,
		workers:     workers,
		timeoutSecs: timeoutSecs,
		lang:        language,
	}
}

func (v *ScanView) Show() {
	if v.cidr == "" {
		return
	}

	win := fyne.CurrentApp().NewWindow(v.lang.Tf(lang.Scan_title, v.ifaceName))
	win.Resize(fyne.NewSize(900, 600))

	progress := widget.NewProgressBar()
	progressNote := widget.NewLabel("0%")
	resultsContainer := container.NewVBox()
	resultsScroll := container.NewVScroll(resultsContainer)

	cidrLabel := widget.NewLabel(v.cidr)
	cidrLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel(v.lang.T(lang.Scan_label_scanning)),
			cidrLabel,
		),
		container.NewPadded(container.NewBorder(nil, nil, widget.NewLabel(v.lang.T(lang.Scan_label_progress)), progressNote, progress)),
		widget.NewSeparator(),
	)

	win.SetContent(container.NewPadded(container.NewBorder(header, nil, nil, nil, resultsScroll)))
	win.Show()

	timeout := time.Duration(v.timeoutSecs) * time.Second
	s, err := scanner.NewScanner(v.cidr, v.workers, timeout)
	if err != nil {
		helpers.ShowError(err, win)
		return
	}

	// Start non-blocking scan
	s.Run()

	// Result listener: insert-sort each result and rebuild VBox
	go func() {
		for res := range s.Results {
			resCopy := res
			scanned, total := s.Progress()
			pct := float64(scanned) / float64(total)

			v.resultsMu.Lock()
			v.sortedResults = insertSorted(v.sortedResults, resCopy, v.serverIP)
			snapshot := make([]scanner.ScanResult, len(v.sortedResults))
			copy(snapshot, v.sortedResults)
			v.resultsMu.Unlock()

			fyne.Do(func() {
				progress.SetValue(pct)
				progressNote.SetText(v.lang.Tf(lang.Scan_progress_pct, int(pct*100)))
				resultsContainer.RemoveAll()
				for _, r := range snapshot {
					resultsContainer.Add(v.makeScanRow(r, win))
				}
				resultsContainer.Refresh()
			})
		}

		fyne.Do(func() {
			progress.SetValue(1)
			progressNote.SetText(v.lang.T(lang.Scan_progress_100))
		})
	}()

	// Progress poller: update bar even when no new results arrive
	go func() {
		for {
			scanned, total := s.Progress()
			if total == 0 {
				break
			}
			pct := float64(scanned) / float64(total)
			fyne.Do(func() {
				progress.SetValue(pct)
				progressNote.SetText(v.lang.Tf(lang.Scan_progress_pct, int(pct*100)))
			})
			if scanned >= total {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (v *ScanView) makeScanRow(r scanner.ScanResult, win fyne.Window) fyne.CanvasObject {
	host := ""
	if len(r.Hostnames) > 0 {
		host = r.Hostnames[0]
	}

	portStr := v.lang.T(lang.Scan_no_open_ports)
	hasSSH := false
	if len(r.PortsOpen) > 0 {
		var ports []string
		for _, p := range r.PortsOpen {
			ports = append(ports, strconv.Itoa(p))
			if p == 22 {
				hasSSH = true
			}
		}
		portStr = v.lang.Tf(lang.Scan_ports_label, strings.Join(ports, ", "))
	}

	ipLabel := widget.NewLabel(r.IP)
	ipLabel.TextStyle = fyne.TextStyle{Monospace: true}

	portLabel := widget.NewLabel(portStr)
	portLabel.TextStyle = fyne.TextStyle{Monospace: true}

	actionCol := container.NewHBox()
	if hasSSH {
		ip := r.IP
		sshBtn := widget.NewButton(v.lang.T(lang.Ssh_connect_btn), func() {
			v.showSSHDialog(ip, win)
		})
		sshBtn.Importance = widget.LowImportance
		actionCol.Add(sshBtn)
	}

	return container.NewPadded(container.NewGridWithColumns(4,
		ipLabel,
		widget.NewLabel(host),
		portLabel,
		actionCol,
	))
}

func (v *ScanView) showSSHDialog(ip string, parent fyne.Window) {
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder(v.lang.T(lang.Ssh_dialog_placeholder_user))

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(v.lang.T(lang.Ssh_dialog_placeholder_pass))

	d := dialog.NewForm(
		v.lang.T(lang.Ssh_dialog_title),
		v.lang.T(lang.Common_save),
		v.lang.T(lang.Common_cancel),
		[]*widget.FormItem{
			{Text: v.lang.T(lang.Ssh_dialog_label_user), Widget: userEntry},
			{Text: v.lang.T(lang.Ssh_dialog_label_pass), Widget: passEntry},
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			wgwidget.NewSSHTerminal(ip, userEntry.Text, passEntry.Text, v.lang).Show()
		},
		parent,
	)
	d.Resize(fyne.NewSize(400, 220))
	d.Show()
}

// insertSorted inserts res into results maintaining numeric IP order,
// with serverIP always at index 0.
func insertSorted(results []scanner.ScanResult, res scanner.ScanResult, serverIP string) []scanner.ScanResult {
	results = append(results, res)
	sort.Slice(results, func(i, j int) bool {
		if results[i].IP == serverIP {
			return true
		}
		if results[j].IP == serverIP {
			return false
		}
		return ipToUint32(results[i].IP) < ipToUint32(results[j].IP)
	})
	return results
}

// ipToUint32 converts an IPv4 string to uint32 for numeric comparison.
func ipToUint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// ScanViewWithProgress is a version with external progress tracking
type ScanViewWithProgress struct {
	*ScanView
	scannedCount *uint32
	totalIPs     uint32
}

// NewScanViewWithProgress creates a scan view with progress tracking
func NewScanViewWithProgress(ifaceName, ip string, workers, timeoutSecs int, language *lang.Language) *ScanViewWithProgress {
	var count uint32
	return &ScanViewWithProgress{
		ScanView:     NewScanView(ifaceName, ip, workers, timeoutSecs, language),
		scannedCount: &count,
	}
}

// Progress returns the current scan progress
func (v *ScanViewWithProgress) Progress() (scanned, total uint32) {
	return atomic.LoadUint32(v.scannedCount), v.totalIPs
}
