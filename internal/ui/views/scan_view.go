package ui

import (
	"errors"
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
	ifaceName     string
	plan          ScanPlan
	serverIP      string
	workers       int
	timeoutSecs   int
	lang          *lang.Language
	sortedResults []scanner.ScanResult
	identities    map[string]string // IP -> what the host looks like
	resultsMu     sync.Mutex
	stopped       atomic.Bool
}

// NewScanView creates a new scan view for the networks in plan (see PlanScan).
// When the plan is empty, the /24 around ip is scanned.
func NewScanView(ifaceName, ip string, plan ScanPlan, workers, timeoutSecs int, language *lang.Language) *ScanView {
	if plan.IsEmpty() {
		if fallback := scanner.DeriveCIDR24(ip); fallback != "" {
			plan.Direct = []string{fallback}
		}
	}
	return &ScanView{
		ifaceName:   ifaceName,
		plan:        plan,
		serverIP:    ip,
		workers:     workers,
		timeoutSecs: timeoutSecs,
		lang:        language,
		identities:  make(map[string]string),
	}
}

func (v *ScanView) Show() {
	if v.plan.IsEmpty() {
		return
	}

	win := fyne.CurrentApp().NewWindow(v.lang.Tf(lang.Scan_title, v.ifaceName))
	win.Resize(fyne.NewSize(900, 600))
	win.SetOnClosed(func() { v.stopped.Store(true) })

	progress := widget.NewProgressBar()
	progressNote := widget.NewLabel("0%")
	statusLabel := widget.NewLabel("")
	resultsContainer := container.NewVBox()
	resultsScroll := container.NewVScroll(resultsContainer)

	allOption := v.lang.Tf(lang.Scan_all_networks, len(v.plan.Direct)+len(v.plan.Wide))
	options := []string{allOption}
	options = append(options, v.plan.Direct...)
	for _, w := range v.plan.Wide {
		options = append(options, v.lang.Tf(lang.Scan_wide_option, w))
	}

	networkSelect := widget.NewSelectEntry(options)
	networkSelect.SetText(allOption)
	rescanBtn := widget.NewButton(v.lang.T(lang.Scan_rescan), nil)

	headerRows := []fyne.CanvasObject{
		container.NewBorder(nil, nil,
			widget.NewLabel(v.lang.T(lang.Scan_label_scanning)),
			rescanBtn,
			networkSelect,
		),
		container.NewPadded(container.NewBorder(nil, nil, widget.NewLabel(v.lang.T(lang.Scan_label_progress)), progressNote, progress)),
		statusLabel,
	}
	if len(v.plan.Skipped) > 0 {
		note := widget.NewLabel(v.lang.Tf(lang.Scan_skipped_networks, strings.Join(v.plan.Skipped, ", ")))
		note.Wrapping = fyne.TextWrapWord
		headerRows = append(headerRows, note)
	}
	headerRows = append(headerRows, widget.NewSeparator())
	header := container.NewVBox(headerRows...)

	win.SetContent(container.NewPadded(container.NewBorder(header, nil, nil, nil, resultsScroll)))
	win.Show()

	start := func() {
		direct, wide, err := v.resolveSelection(networkSelect.Text, allOption)
		if err != nil {
			helpers.ShowError(err, win)
			return
		}

		v.resultsMu.Lock()
		v.sortedResults = nil
		v.identities = make(map[string]string)
		v.resultsMu.Unlock()

		networkSelect.Disable()
		rescanBtn.Disable()
		resultsContainer.RemoveAll()
		resultsContainer.Refresh()
		progress.SetValue(0)
		progressNote.SetText(v.lang.Tf(lang.Scan_progress_pct, 0))

		v.run(direct, wide, win, progress, progressNote, statusLabel, resultsContainer, func() {
			networkSelect.Enable()
			rescanBtn.Enable()
		})
	}

	// SelectEntry fires OnChanged on every keystroke, so only the dropdown
	// entries auto-start; typed CIDRs are launched with the Rescan button.
	networkSelect.OnChanged = func(text string) {
		for _, opt := range options {
			if text == opt {
				start()
				return
			}
		}
	}
	rescanBtn.OnTapped = start

	start()
}

// resolveSelection turns the selector's text into the networks to scan. It
// accepts the "all networks" option, one of the listed networks, or any CIDR
// typed by the user.
func (v *ScanView) resolveSelection(text, allOption string) (direct, wide []string, err error) {
	text = strings.TrimSpace(text)

	if text == "" || text == allOption {
		return v.plan.Direct, v.plan.Wide, nil
	}
	for _, d := range v.plan.Direct {
		if text == d {
			return []string{d}, nil, nil
		}
	}
	for _, w := range v.plan.Wide {
		if text == w || text == v.lang.Tf(lang.Scan_wide_option, w) {
			return nil, []string{w}, nil
		}
	}

	// Anything else is treated as a hand-typed CIDR.
	_, ipnet, parseErr := net.ParseCIDR(text)
	if parseErr != nil || ipnet.IP.To4() == nil {
		return nil, nil, errors.New(v.lang.Tf(lang.Scan_invalid_network, text))
	}
	ones, _ := ipnet.Mask.Size()
	switch size := uint64(1) << uint(32-ones); {
	case size <= maxScanHosts:
		return []string{ipnet.String()}, nil, nil
	case size/256 <= maxSweepSubnets:
		return nil, []string{ipnet.String()}, nil
	default:
		return nil, nil, errors.New(v.lang.Tf(lang.Scan_network_too_large, ipnet.String()))
	}
}

// run discovers the active subnets of the wide routes, then scans those
// together with the direct networks. done runs on the UI thread when finished.
func (v *ScanView) run(direct, wide []string, win fyne.Window, progress *widget.ProgressBar, progressNote *widget.Label, statusLabel *widget.Label, resultsContainer *fyne.Container, done func()) {
	v.stopped.Store(false)
	timeout := time.Duration(v.timeoutSecs) * time.Second

	setProgress := func(scanned, total uint32) {
		if total == 0 {
			return
		}
		pct := float64(scanned) / float64(total)
		if pct > 1 {
			pct = 1
		}
		fyne.Do(func() {
			progress.SetValue(pct)
			progressNote.SetText(v.lang.Tf(lang.Scan_progress_pct, int(pct*100)))
		})
	}
	setStatus := func(text string) {
		fyne.Do(func() { statusLabel.SetText(text) })
	}

	finish := func() {
		fyne.Do(func() {
			progress.SetValue(1)
			progressNote.SetText(v.lang.T(lang.Scan_progress_100))
			statusLabel.SetText("")
			done()
		})
	}

	go func() {
		// Phase 1: narrow wide routes down to the /24s that answer.
		targets := append([]string{}, direct...)
		for _, cidr := range wide {
			if v.stopped.Load() {
				finish()
				return
			}
			setStatus(v.lang.Tf(lang.Scan_phase_discovering, cidr))
			found := discoverSubnets(cidr, v.workers, timeout, &v.stopped, func(doneCount, total int) {
				setProgress(uint32(doneCount), uint32(total))
			})
			targets = append(targets, found...)
		}
		targets = dedupeCIDRs(targets)

		if len(targets) == 0 {
			setStatus(v.lang.T(lang.Scan_no_active_subnets))
			fyne.Do(func() {
				progress.SetValue(1)
				progressNote.SetText(v.lang.T(lang.Scan_progress_100))
				done()
			})
			return
		}

		// Phase 2: scan every target address by address.
		grandTotal := totalHosts(targets)
		var completed uint32
		var current atomic.Pointer[scanner.Scanner]

		report := func() {
			scanned := atomic.LoadUint32(&completed)
			if s := current.Load(); s != nil {
				partial, _ := s.Progress()
				scanned += partial
			}
			setProgress(scanned, grandTotal)
		}

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		stopPoll := make(chan struct{})
		defer close(stopPoll)
		go func() {
			for {
				select {
				case <-stopPoll:
					return
				case <-ticker.C:
					report()
				}
			}
		}()

		for _, cidr := range targets {
			if v.stopped.Load() {
				break
			}
			setStatus(v.lang.Tf(lang.Scan_phase_scanning, cidr))

			s, err := scanner.NewScanner(cidr, v.workers, timeout)
			if err != nil {
				helpers.ShowError(err, win)
				continue
			}
			current.Store(s)
			s.Run()

			for res := range s.Results {
				v.addResult(res, win, resultsContainer)
				report()
			}

			current.Store(nil)
			atomic.AddUint32(&completed, s.TotalIPs)
			report()
		}

		finish()
	}()
}

// addResult inserts a result in IP order, repaints the list and asks the host
// what it is in the background.
func (v *ScanView) addResult(res scanner.ScanResult, win fyne.Window, resultsContainer *fyne.Container) {
	v.resultsMu.Lock()
	v.sortedResults = insertSorted(v.sortedResults, res, v.serverIP)
	v.resultsMu.Unlock()

	v.repaint(win, resultsContainer)

	go func() {
		id := identifyHost(res.IP, res.PortsOpen, time.Duration(v.timeoutSecs)*time.Second)
		if id == "" || v.stopped.Load() {
			return
		}
		v.resultsMu.Lock()
		v.identities[res.IP] = id
		v.resultsMu.Unlock()

		v.repaint(win, resultsContainer)
	}()
}

// repaint rebuilds the result rows from the current snapshot.
func (v *ScanView) repaint(win fyne.Window, resultsContainer *fyne.Container) {
	v.resultsMu.Lock()
	snapshot := make([]scanner.ScanResult, len(v.sortedResults))
	copy(snapshot, v.sortedResults)
	identities := make(map[string]string, len(v.identities))
	for ip, id := range v.identities {
		identities[ip] = id
	}
	v.resultsMu.Unlock()

	fyne.Do(func() {
		resultsContainer.RemoveAll()
		for _, r := range snapshot {
			resultsContainer.Add(v.makeScanRow(r, identities[r.IP], win))
		}
		resultsContainer.Refresh()
	})
}

// dedupeCIDRs removes repeats while keeping numeric network order.
func dedupeCIDRs(cidrs []string) []string {
	seen := make(map[string]bool, len(cidrs))
	out := make([]string, 0, len(cidrs))
	for _, c := range cidrs {
		if seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		return ipToUint32(strings.Split(out[i], "/")[0]) < ipToUint32(strings.Split(out[j], "/")[0])
	})
	return out
}

func (v *ScanView) makeScanRow(r scanner.ScanResult, identity string, win fyne.Window) fyne.CanvasObject {
	host := ""
	if len(r.Hostnames) > 0 {
		host = strings.TrimSuffix(r.Hostnames[0], ".")
	}

	// Reverse DNS rarely resolves across a tunnel; show what the host looks
	// like instead, and both when the name is known.
	nameLabel := widget.NewLabel(host)
	switch {
	case host == "" && identity != "":
		nameLabel = widget.NewLabelWithStyle(identity, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	case host != "" && identity != "":
		nameLabel.SetText(host + " (" + identity + ")")
	}

	portStr := v.lang.T(lang.Scan_no_open_ports)
	if len(r.PortsOpen) > 0 {
		var portsStrs []string
		for _, p := range r.PortsOpen {
			if p == 22 {
				portsStrs = append(portsStrs, strconv.Itoa(p)+" (ssh)")
			} else {
				portsStrs = append(portsStrs, strconv.Itoa(p))

			}
		}
		portStr = v.lang.Tf(lang.Scan_ports_label, strings.Join(portsStrs, ", "))
	}

	ipLabel := widget.NewLabel(r.IP)
	ipLabel.TextStyle = fyne.TextStyle{Monospace: true}

	portLabel := widget.NewLabel(portStr)
	portLabel.TextStyle = fyne.TextStyle{Monospace: true}

	actionCol := container.NewHBox()
	ip := r.IP
	sshBtn := widget.NewButton(v.lang.T(lang.Ssh_connect_btn), func() {
		v.showSSHDialog(ip, win)
	})
	sshBtn.Importance = widget.LowImportance
	actionCol.Add(sshBtn)

	return container.NewPadded(container.NewGridWithColumns(4,
		ipLabel,
		nameLabel,
		portLabel,
		actionCol,
	))
}

func (v *ScanView) showSSHDialog(ip string, parent fyne.Window) {
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder(v.lang.T(lang.Ssh_dialog_placeholder_user))

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(v.lang.T(lang.Ssh_dialog_placeholder_pass))

	sshPortEntry := widget.NewEntry()
	sshPortEntry.SetPlaceHolder(v.lang.T(lang.Ssh_dialog_placeholder_ssh))

	sshPort := ""
	sshPortEntry.OnChanged = func(s string) {
		portInt, _ := strconv.Atoi(s)
		if portInt != 0 {
			sshPort = s
		} else {
			sshPortEntry.SetText(sshPort)

		}
	}
	if sshPort == "" {
		sshPort = "22"
	}
	d := dialog.NewForm(
		v.lang.T(lang.Ssh_dialog_title),
		v.lang.T(lang.Common_connect),
		v.lang.T(lang.Common_cancel),
		[]*widget.FormItem{
			{Text: v.lang.T(lang.Ssh_dialog_label_user), Widget: userEntry},
			{Text: v.lang.T(lang.Ssh_dialog_label_pass), Widget: passEntry},
			{Text: v.lang.T(lang.Ssh_dialog_label_ssh), Widget: sshPortEntry},
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			portInt, _ := strconv.Atoi(sshPort)
			wgwidget.NewSSHTerminal(ip, userEntry.Text, passEntry.Text, portInt, v.lang).Show()

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
func NewScanViewWithProgress(ifaceName, ip string, plan ScanPlan, workers, timeoutSecs int, language *lang.Language) *ScanViewWithProgress {
	var count uint32
	return &ScanViewWithProgress{
		ScanView:     NewScanView(ifaceName, ip, plan, workers, timeoutSecs, language),
		scannedCount: &count,
	}
}

// Progress returns the current scan progress
func (v *ScanViewWithProgress) Progress() (scanned, total uint32) {
	return atomic.LoadUint32(v.scannedCount), v.totalIPs
}
