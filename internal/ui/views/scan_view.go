package ui

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	scanner "github.com/MrVasquez96/go-ipscan"
)

// ScanView displays network scan results
type ScanView struct {
	ifaceName   string
	cidr        string
	workers     int
	timeoutSecs int
	lang        *lang.Language
}

// NewScanView creates a new scan view
func NewScanView(ifaceName, ip string, workers, timeoutSecs int, language *lang.Language) *ScanView {
	return &ScanView{
		ifaceName:   ifaceName,
		cidr:        scanner.DeriveCIDR24(ip),
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
	win.Resize(fyne.NewSize(850, 600))

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

	// Result listener: update UI as results arrive
	go func() {
		for res := range s.Results {
			resCopy := res
			scanned, total := s.Progress()
			pct := float64(scanned) / float64(total)

			fyne.Do(func() {
				bar := pct
				progress.SetValue(bar)
				progressNote.SetText(v.lang.Tf(lang.Scan_progress_pct, int(bar*100)))
				resultsContainer.Add(makeScanRow(resCopy, v.lang))
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

func makeScanRow(r scanner.ScanResult, language *lang.Language) fyne.CanvasObject {
	host := ""
	if len(r.Hostnames) > 0 {
		host = r.Hostnames[0]
	}

	portStr := language.T(lang.Scan_no_open_ports)
	if len(r.PortsOpen) > 0 {
		var ports []string
		for _, p := range r.PortsOpen {
			ports = append(ports, strconv.Itoa(p))
		}
		portStr = language.Tf(lang.Scan_ports_label, strings.Join(ports, ", "))
	}

	ipLabel := widget.NewLabel(r.IP)
	ipLabel.TextStyle = fyne.TextStyle{Monospace: true}

	portLabel := widget.NewLabel(portStr)
	portLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewPadded(container.NewGridWithColumns(3,
		ipLabel,
		widget.NewLabel(host),
		portLabel,
	))
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
