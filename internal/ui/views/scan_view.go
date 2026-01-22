package views

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	"wgAdmin/internal/scanner"
)

// ScanView displays network scan results
type ScanView struct {
	ifaceName string
	cidr      string
}

// NewScanView creates a new scan view
func NewScanView(ifaceName, ip string) *ScanView {
	return &ScanView{
		ifaceName: ifaceName,
		cidr:      scanner.DeriveCIDR24(ip),
	}
}

// Show displays the scan window
func (v *ScanView) Show() {
	if v.cidr == "" {
		return
	}

	win := fyne.CurrentApp().NewWindow("Scan: " + v.ifaceName)
	win.Resize(fyne.NewSize(700, 520))

	progress := widget.NewProgressBar()
	progressNote := widget.NewLabel("0%")
	resultsContainer := container.NewVBox()
	resultsScroll := container.NewVScroll(resultsContainer)

	// Header
	header := container.NewHBox(
		widget.NewLabel("CIDR:"),
		widget.NewLabel(v.cidr),
		widget.NewSeparator(),
		progress,
		progressNote,
	)

	win.SetContent(container.NewBorder(
		header,
		nil, nil, nil,
		resultsScroll,
	))
	win.Show()

	// Start scan
	s, err := scanner.NewScanner(v.cidr)
	if err != nil {
		dialog.ShowError(err, win)
		return
	}

	go func() {
		for res := range s.Results {
			resCopy := res
			scanned, total := s.Progress()
			pct := 0.0
			if total > 0 {
				pct = float64(scanned) / float64(total)
			}
			fyne.Do(func() {
				progress.SetValue(pct)
				progressNote.SetText(fmt.Sprintf("%d%%", int(pct*100)))
				resultsContainer.Add(makeScanRow(resCopy))
				resultsContainer.Refresh()
			})
		}
		fyne.Do(func() {
			progress.SetValue(1)
			progressNote.SetText("100%")
		})
	}()

	go s.Run()
}

func makeScanRow(r models.ScanResult) fyne.CanvasObject {
	host := ""
	if len(r.Hostnames) > 0 {
		host = r.Hostnames[0]
	}

	var ports []string
	for _, p := range r.PortsOpen {
		ports = append(ports, strconv.Itoa(p))
	}
	portStr := "Ports: " + strings.Join(ports, ", ")

	return container.NewGridWithColumns(3,
		widget.NewLabel(r.IP),
		widget.NewLabel(host),
		widget.NewLabel(portStr),
	)
}

// ScanViewWithProgress is a version with external progress tracking
type ScanViewWithProgress struct {
	*ScanView
	scannedCount *uint32
	totalIPs     uint32
}

// NewScanViewWithProgress creates a scan view with progress tracking
func NewScanViewWithProgress(ifaceName, ip string) *ScanViewWithProgress {
	var count uint32
	return &ScanViewWithProgress{
		ScanView:     NewScanView(ifaceName, ip),
		scannedCount: &count,
	}
}

// Progress returns the current scan progress
func (v *ScanViewWithProgress) Progress() (scanned, total uint32) {
	return atomic.LoadUint32(v.scannedCount), v.totalIPs
}
