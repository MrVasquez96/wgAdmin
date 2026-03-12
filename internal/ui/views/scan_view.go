package ui

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"wgAdmin/internal/ui/helpers"

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
}

// NewScanView creates a new scan view
func NewScanView(ifaceName, ip string, workers, timeoutSecs int) *ScanView {
	return &ScanView{
		ifaceName:   ifaceName,
		cidr:        scanner.DeriveCIDR24(ip),
		workers:     workers,
		timeoutSecs: timeoutSecs,
	}
}

func (v *ScanView) Show() {
	log.Printf("[DEBUG] Show() called with CIDR: %s, Interface: %s", v.cidr, v.ifaceName)

	if v.cidr == "" {
		log.Println("[WARN] Scan abort: CIDR is empty")
		return
	}

	win := fyne.CurrentApp().NewWindow("Scan: " + v.ifaceName)
	win.Resize(fyne.NewSize(850, 600))

	progress := widget.NewProgressBar()
	progressNote := widget.NewLabel("0%")
	resultsContainer := container.NewVBox()
	resultsScroll := container.NewVScroll(resultsContainer)

	cidrLabel := widget.NewLabel(v.cidr)
	cidrLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Scanning Network:"),
			cidrLabel,
		),
		container.NewPadded(container.NewBorder(nil, nil, widget.NewLabel("Progress:"), progressNote, progress)),
		widget.NewSeparator(),
	)

	win.SetContent(container.NewPadded(container.NewBorder(header, nil, nil, nil, resultsScroll)))
	win.Show()

	timeout := time.Duration(v.timeoutSecs) * time.Second
	log.Println("[DEBUG] Initializing scanner...", v.cidr)
	s, err := scanner.NewScanner(v.cidr, v.workers, timeout)
	if err != nil {
		log.Printf("[ERROR] Scanner initialization failed: %v", err)
		helpers.ShowError(err, win)
		return
	}

	// Start non-blocking scan
	s.Run()

	// Result listener: update UI as results arrive
	go func() {
		log.Println("[DEBUG] Listener goroutine started: Waiting for results...")
		for res := range s.Results {
			resCopy := res
			scanned, total := s.Progress()
			pct := float64(scanned) / float64(total)

			fyne.Do(func() {
				bar := pct
				progress.SetValue(bar)
				progressNote.SetText(fmt.Sprintf("%d%%", int(bar*100)))
				resultsContainer.Add(makeScanRow(resCopy))
				resultsContainer.Refresh()
			})
		}

		log.Println("[DEBUG] Results channel closed. Scan complete.")
		fyne.Do(func() {
			progress.SetValue(1)
			progressNote.SetText("100%")
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
				progressNote.SetText(fmt.Sprintf("%d%%", int(pct*100)))
			})
			if scanned >= total {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func makeScanRow(r scanner.ScanResult) fyne.CanvasObject {
	host := ""
	if len(r.Hostnames) > 0 {
		host = r.Hostnames[0]
	}

	portStr := "No open ports"
	if len(r.PortsOpen) > 0 {
		var ports []string
		for _, p := range r.PortsOpen {
			ports = append(ports, strconv.Itoa(p))
		}
		portStr = "Ports: " + strings.Join(ports, ", ")
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
func NewScanViewWithProgress(ifaceName, ip string, workers, timeoutSecs int) *ScanViewWithProgress {
	var count uint32
	return &ScanViewWithProgress{
		ScanView:     NewScanView(ifaceName, ip, workers, timeoutSecs),
		scannedCount: &count,
	}
}

// Progress returns the current scan progress
func (v *ScanViewWithProgress) Progress() (scanned, total uint32) {
	return atomic.LoadUint32(v.scannedCount), v.totalIPs
}
