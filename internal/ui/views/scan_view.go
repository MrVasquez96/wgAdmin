package ui

import (
	"encoding/csv"
	"errors"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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

	// UI handles kept so results can be repainted in place instead of the
	// whole list being rebuilt on every hit.
	win        fyne.Window
	table      *widget.Table
	countLabel *canvas.Text
}

// scanColumns defines the results table.
var scanColumns = []struct {
	titleKey string
	width    float32
}{
	{lang.Scan_col_address, 150},
	{lang.Scan_col_identity, 260},
	{lang.Scan_col_ports, 260},
	{lang.Scan_col_actions, 190},
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
	win.Resize(fyne.NewSize(960, 640))
	win.SetOnClosed(func() { v.stopped.Store(true) })
	v.win = win

	progress := widget.NewProgressBar()
	progress.TextFormatter = func() string { return "" } // the count carries the detail
	progressNote := wgwidget.Value("0%")
	statusLabel := wgwidget.DimValue("")
	countLabel := wgwidget.Caption("")

	v.table = v.buildResultsTable(win)

	allOption := v.lang.Tf(lang.Scan_all_networks, len(v.plan.Direct)+len(v.plan.Wide))
	options := []string{allOption}
	options = append(options, v.plan.Direct...)
	for _, w := range v.plan.Wide {
		options = append(options, v.lang.Tf(lang.Scan_wide_option, w))
	}

	networkSelect := widget.NewSelectEntry(options)
	networkSelect.SetText(allOption)

	rescanBtn := widget.NewButtonWithIcon(v.lang.T(lang.Scan_rescan), wgtheme.IconRefresh, nil)
	rescanBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButtonWithIcon(v.lang.T(lang.Scan_btn_stop), wgtheme.IconStop, func() {
		v.stopped.Store(true)
	})
	stopBtn.Importance = widget.DangerImportance

	exportBtn := widget.NewButtonWithIcon(v.lang.T(lang.Scan_btn_export), wgtheme.IconDownload, func() {
		v.exportCSV(win)
	})

	controls := container.NewBorder(nil, nil,
		wgwidget.Caption(v.lang.T(lang.Scan_label_network)),
		container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM), stopBtn, rescanBtn, exportBtn),
		networkSelect,
	)

	progressRow := container.NewBorder(nil, nil,
		container.NewCenter(countLabel),
		container.NewCenter(progressNote),
		container.NewCenter(progress),
	)

	headerRows := []fyne.CanvasObject{
		controls,
		progressRow,
		statusLabel,
	}
	if len(v.plan.Skipped) > 0 {
		// Networks too large to scan are worth naming: silently ignoring
		// 0.0.0.0/0 looks like the scanner simply missed them.
		note := widget.NewLabel(v.lang.Tf(lang.Scan_skipped_networks, strings.Join(v.plan.Skipped, ", ")))
		note.Wrapping = fyne.TextWrapWord
		note.Importance = widget.LowImportance
		headerRows = append(headerRows, note)
	}
	headerRows = append(headerRows, wgwidget.Hairline())

	header := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD), headerRows...)

	win.SetContent(container.New(
		layout.NewCustomPaddedLayout(wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG, wgtheme.SpaceLG),
		container.NewBorder(header, nil, nil, nil, v.table)))
	win.Show()

	v.countLabel = countLabel
	v.updateCount()

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
		stopBtn.Enable()
		v.table.Refresh()
		v.updateCount()
		progress.SetValue(0)
		progressNote.Text = v.lang.Tf(lang.Scan_progress_pct, 0)
		progressNote.Refresh()

		v.run(direct, wide, win, progress, progressNote, statusLabel, nil, func() {
			networkSelect.Enable()
			rescanBtn.Enable()
			stopBtn.Disable()
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
func (v *ScanView) run(direct, wide []string, win fyne.Window, progress *widget.ProgressBar, progressNote *canvas.Text, statusLabel *canvas.Text, resultsContainer *fyne.Container, done func()) {
	v.stopped.Store(false)
	timeout := time.Duration(v.timeoutSecs) * time.Second

	setText := func(t *canvas.Text, s string) {
		t.Text = s
		t.Refresh()
	}

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
			setText(progressNote, v.lang.Tf(lang.Scan_progress_pct, int(pct*100)))
		})
	}
	setStatus := func(text string) {
		fyne.Do(func() { setText(statusLabel, text) })
	}

	finish := func() {
		fyne.Do(func() {
			progress.SetValue(1)
			setText(progressNote, v.lang.T(lang.Scan_progress_100))
			setText(statusLabel, "")
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
				setText(progressNote, v.lang.T(lang.Scan_progress_100))
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

// addResult inserts a result in IP order, repaints the table and asks the host
// what it is in the background.
func (v *ScanView) addResult(res scanner.ScanResult, win fyne.Window, _ *fyne.Container) {
	v.resultsMu.Lock()
	v.sortedResults = insertSorted(v.sortedResults, res, v.serverIP)
	v.resultsMu.Unlock()

	v.repaint()

	go func() {
		id := identifyHost(res.IP, res.PortsOpen, time.Duration(v.timeoutSecs)*time.Second)
		if id == "" || v.stopped.Load() {
			return
		}
		v.resultsMu.Lock()
		v.identities[res.IP] = id
		v.resultsMu.Unlock()

		v.repaint()
	}()
}

// repaint refreshes the results table. The table pulls rows from the shared
// slice on demand, so this only has to invalidate it — the previous version
// destroyed and rebuilt every row widget on each new host found, which grew
// quadratically over a large scan.
func (v *ScanView) repaint() {
	fyne.Do(func() {
		if v.table != nil {
			v.table.Refresh()
		}
		v.updateCount()
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

// hostIdentity renders the name column: reverse DNS rarely resolves across a
// tunnel, so the fingerprinted identity stands in, and both are shown when the
// name is known.
func (v *ScanView) hostIdentity(r scanner.ScanResult, identity string) string {
	host := ""
	if len(r.Hostnames) > 0 {
		host = strings.TrimSuffix(r.Hostnames[0], ".")
	}
	switch {
	case host == "" && identity != "":
		return identity
	case host != "" && identity != "":
		return host + " (" + identity + ")"
	case host != "":
		return host
	default:
		return wgstat.Dash
	}
}

// portsText renders the open ports, naming the well-known ones.
func (v *ScanView) portsText(r scanner.ScanResult) string {
	if len(r.PortsOpen) == 0 {
		return v.lang.T(lang.Scan_no_open_ports)
	}
	names := map[int]string{22: "ssh", 80: "http", 443: "https", 53: "dns", 8080: "http-alt", 3389: "rdp", 445: "smb"}

	parts := make([]string, 0, len(r.PortsOpen))
	for _, p := range r.PortsOpen {
		if n, ok := names[p]; ok {
			parts = append(parts, strconv.Itoa(p)+"/"+n)
		} else {
			parts = append(parts, strconv.Itoa(p))
		}
	}
	return strings.Join(parts, " ")
}

// snapshot returns a copy of the current results and identities.
func (v *ScanView) snapshot() ([]scanner.ScanResult, map[string]string) {
	v.resultsMu.Lock()
	defer v.resultsMu.Unlock()

	results := make([]scanner.ScanResult, len(v.sortedResults))
	copy(results, v.sortedResults)
	ids := make(map[string]string, len(v.identities))
	for ip, id := range v.identities {
		ids[ip] = id
	}
	return results, ids
}

// buildResultsTable creates the results table. Each row carries its own copy
// and connect actions so a host can be acted on without leaving the scan.
func (v *ScanView) buildResultsTable(win fyne.Window) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (int, int) {
			v.resultsMu.Lock()
			defer v.resultsMu.Unlock()
			return len(v.sortedResults), len(scanColumns)
		},
		func() fyne.CanvasObject { return newScanCell(v.lang) },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			cell, ok := obj.(*scanCell)
			if !ok {
				return
			}

			v.resultsMu.Lock()
			if id.Row >= len(v.sortedResults) {
				v.resultsMu.Unlock()
				cell.SetText("")
				return
			}
			r := v.sortedResults[id.Row]
			identity := v.identities[r.IP]
			v.resultsMu.Unlock()

			switch id.Col {
			case 0:
				text := r.IP
				if r.IP == v.serverIP {
					text += "  " + v.lang.T(lang.Scan_this_host)
				}
				cell.SetText(text)
			case 1:
				cell.SetText(v.hostIdentity(r, identity))
			case 2:
				cell.SetText(v.portsText(r))
			case 3:
				ip := r.IP
				cell.SetActions(
					func() {
						win.Clipboard().SetContent(ip)
					},
					func() { v.showSSHDialog(ip, win) },
				)
			}
		},
	)

	table.CreateHeader = func() fyne.CanvasObject {
		h := widget.NewLabel("")
		h.TextStyle = fyne.TextStyle{Bold: true}
		return h
	}
	table.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		h, ok := obj.(*widget.Label)
		if !ok {
			return
		}
		if id.Col < 0 || id.Col >= len(scanColumns) {
			h.SetText("")
			return
		}
		h.SetText(strings.ToUpper(v.lang.T(scanColumns[id.Col].titleKey)))
	}

	table.ShowHeaderRow = true
	table.ShowHeaderColumn = false
	table.StickyColumnCount = 1

	for i, c := range scanColumns {
		table.SetColumnWidth(i, c.width)
	}
	// Rows carry buttons, so they need a little more height than the peer
	// table's plain text rows.
	table.SetRowHeight(-1, 28)

	return table
}

// updateCount refreshes the "N hosts" counter above the table.
func (v *ScanView) updateCount() {
	if v.countLabel == nil {
		return
	}
	v.resultsMu.Lock()
	n := len(v.sortedResults)
	v.resultsMu.Unlock()

	v.countLabel.Text = v.lang.Tf(lang.Scan_hosts_found, n)
	v.countLabel.Refresh()
}

// exportCSV writes the current results to a file the user picks.
func (v *ScanView) exportCSV(win fyne.Window) {
	results, ids := v.snapshot()
	if len(results) == 0 {
		helpers.ShowInformation(v.lang.T(lang.Scan_btn_export), v.lang.T(lang.Scan_no_results), win)
		return
	}

	helpers.SaveFile(win, helpers.FileOpts{
		Title:       v.lang.T(lang.Scan_btn_export),
		FilterLabel: "CSV",
		Extensions:  []string{".csv"},
		FileName:    v.ifaceName + "-scan.csv",
	}, func(path string) {
		f, err := os.Create(path)
		if err != nil {
			helpers.ShowError(errors.New(v.lang.Tf(lang.Scan_export_error, err.Error())), win)
			return
		}
		defer f.Close()

		cw := csv.NewWriter(f)
		_ = cw.Write([]string{"address", "identity", "ports"})
		for _, r := range results {
			ports := make([]string, 0, len(r.PortsOpen))
			for _, p := range r.PortsOpen {
				ports = append(ports, strconv.Itoa(p))
			}
			if err := cw.Write([]string{
				r.IP,
				v.hostIdentity(r, ids[r.IP]),
				strings.Join(ports, " "),
			}); err != nil {
				helpers.ShowError(errors.New(v.lang.Tf(lang.Scan_export_error, err.Error())), win)
				return
			}
		}
		cw.Flush()
		if err := cw.Error(); err != nil {
			helpers.ShowError(errors.New(v.lang.Tf(lang.Scan_export_error, err.Error())), win)
			return
		}

		helpers.ShowInformation(v.lang.T(lang.Scan_btn_export),
			v.lang.Tf(lang.Scan_export_success, path), win)
	})
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
