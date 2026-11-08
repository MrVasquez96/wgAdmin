package ui

import (
	"image/color"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"github.com/MrVasquez96/go-wg/wg/config"
)

// peerRow is one row of the peer table: what the config says joined with what
// the kernel reports. The join is on the public key, which is the only stable
// identifier a peer has — names live in the config, counters live in the
// kernel, and neither side knows about the other.
type peerRow struct {
	name       string
	publicKey  string
	allowedIPs string
	endpoint   string
	handshake  string
	rx, tx     string
	state      wgstat.PeerState
	live       bool

	// Raw values, so sorting by traffic or recency is numeric rather than by
	// the formatted text ("9 KiB" must sort below "10 MiB").
	rxBytes, txBytes int64
	handshakeAt      time.Time
}

// Column indexes into peerColumns.
const (
	colState = iota
	colPeer
	colAllowedIPs
	colEndpoint
	colHandshake
	colRX
	colTX
)

// peerColumns defines the table layout. Machine values are monospace; the
// address and endpoint columns take the spare width because they are the
// ones that get truncated.
func peerColumns(l *lang.Language) []wgwidget.Column {
	return []wgwidget.Column{
		{Dot: true, MinWidth: 24},
		{Title: l.T(lang.Detail_col_peer), MinWidth: 110, Weight: 1, Sortable: true},
		{Title: l.T(lang.Detail_col_allowed_ips), MinWidth: 120, Weight: 1.4, Mono: true, Sortable: true},
		{Title: l.T(lang.Detail_col_endpoint), MinWidth: 120, Weight: 1.6, Mono: true, Sortable: true},
		{Title: l.T(lang.Detail_col_handshake), MinWidth: 92, Sortable: true},
		{Title: l.T(lang.Detail_col_rx), MinWidth: 88, Mono: true, Sortable: true},
		{Title: l.T(lang.Detail_col_tx), MinWidth: 88, Mono: true, Sortable: true},
	}
}

// buildPeerRows joins the configured peers with the live device state.
func buildPeerRows(peers []config.PeerConfig, dev wgstat.Device, live bool, l *lang.Language) []peerRow {
	now := time.Now()
	rows := make([]peerRow, 0, len(peers))

	for _, p := range peers {
		key := p.PublicKey.String()

		// go-wg's parser fills in the literal "Unknown" for peers that carry no
		// name comment. A column of "Unknown" identifies nothing, so fall back
		// to the key prefix, which is what actually distinguishes peers.
		name := strings.TrimSpace(p.Name)
		if name == "" || name == "Unknown" {
			name = wgstat.ShortKey(key)
		}

		ips := make([]string, 0, len(p.AllowedIPs))
		for _, n := range p.AllowedIPs {
			ips = append(ips, n.String())
		}

		row := peerRow{
			name:       name,
			publicKey:  key,
			allowedIPs: strings.Join(ips, ", "),
			endpoint:   p.Endpoint,
			handshake:  wgstat.Dash,
			rx:         wgstat.Dash,
			tx:         wgstat.Dash,
			state:      wgstat.StateNever,
			live:       live,
		}

		// Only overwrite with kernel data when we actually have it. Without
		// this a down tunnel would show hard zeroes, which reads as "connected
		// but idle" rather than "not running".
		if live {
			if lp, ok := dev.Peer(key); ok {
				row.state = lp.State(now)
				row.handshake = lp.HandshakeAge(now)
				row.rx = wgstat.FormatBytes(lp.RX)
				row.tx = wgstat.FormatBytes(lp.TX)
				row.rxBytes, row.txBytes, row.handshakeAt = lp.RX, lp.TX, lp.Handshake
				if lp.Endpoint != "" {
					row.endpoint = lp.Endpoint
				}
			}
		}

		if row.endpoint == "" {
			row.endpoint = wgstat.Dash
		}
		if row.allowedIPs == "" {
			row.allowedIPs = wgstat.Dash
		}

		rows = append(rows, row)
	}

	return rows
}

// cellValue is the full text behind a data column, which the cell itself may
// show truncated. The peer column yields the public key: the displayed name is
// either a comment or a key prefix, and neither is what gets pasted elsewhere.
func (r peerRow) cellValue(col int) string {
	switch col {
	case colPeer:
		return r.publicKey
	case colAllowedIPs:
		return r.allowedIPs
	case colEndpoint:
		return r.endpoint
	case colHandshake:
		return r.handshake
	case colRX:
		return r.rx
	case colTX:
		return r.tx
	}
	return ""
}

// sortPeerRows orders rows by a column. Peers without data always sort last,
// whichever the direction, so a column of dashes never buries the live ones.
func sortPeerRows(rows []peerRow, col int, asc bool) {
	less := func(a, b peerRow) bool {
		switch col {
		case colAllowedIPs:
			return a.allowedIPs < b.allowedIPs
		case colEndpoint:
			return a.endpoint < b.endpoint
		case colHandshake:
			// Most recent first reads as "ascending age".
			return a.handshakeAt.After(b.handshakeAt)
		case colRX:
			return a.rxBytes < b.rxBytes
		case colTX:
			return a.txBytes < b.txBytes
		}
		return strings.ToLower(a.name) < strings.ToLower(b.name)
	}
	empty := func(r peerRow) bool {
		switch col {
		case colHandshake:
			return r.handshakeAt.IsZero()
		case colRX, colTX:
			return r.handshake == wgstat.Dash
		case colEndpoint:
			return r.endpoint == wgstat.Dash
		}
		return false
	}
	sort.SliceStable(rows, func(i, j int) bool {
		ei, ej := empty(rows[i]), empty(rows[j])
		if ei != ej {
			return ej
		}
		if asc {
			return less(rows[i], rows[j])
		}
		return less(rows[j], rows[i])
	})
}

// peerTable wires a DataTable to a slice of peer rows. Sort state lives here
// so it survives the detail pane being rebuilt on every refresh tick.
type peerTable struct {
	sortCol int
	sortAsc bool
}

func (pt *peerTable) build(rows *[]peerRow, l *lang.Language,
	onQR func(publicKey string), onCopy func(column, value string)) *wgwidget.DataTable {

	cols := peerColumns(l)

	if pt.sortCol > 0 {
		sortPeerRows(*rows, pt.sortCol, pt.sortAsc)
	}

	t := wgwidget.NewDataTable(cols, func() int { return len(*rows) }, func(r, c int) wgwidget.Cell {
		row := (*rows)[r]
		switch c {
		case colState:
			if !row.live {
				return wgwidget.Cell{Dot: color.NRGBA{}}
			}
			return wgwidget.Cell{Dot: wgwidget.StateTone(row.state), Live: row.state == wgstat.StateOnline}
		case colPeer:
			return wgwidget.Cell{Text: row.name, Strong: true}
		case colHandshake:
			return wgwidget.Cell{Text: row.handshake, Dim: true}
		default:
			v := row.cellValue(c)
			return wgwidget.Cell{Text: v, Dim: v == wgstat.Dash}
		}
	})
	t.SortCol, t.SortAsc = -1, true
	if pt.sortCol > 0 {
		t.SortCol, t.SortAsc = pt.sortCol, pt.sortAsc
	}
	t.OnSort = func(col int, asc bool) {
		pt.sortCol, pt.sortAsc = col, asc
		sortPeerRows(*rows, col, asc)
	}
	t.OnCellTap = func(r, c int) {
		if v := (*rows)[r].cellValue(c); v != "" && v != wgstat.Dash {
			onCopy(cols[c].Title, v)
		}
	}
	t.Actions = []wgwidget.RowAction{
		{Icon: wgtheme.IconQR, Tooltip: l.T(lang.Detail_tip_qr), OnTap: func(r int) {
			onQR((*rows)[r].publicKey)
		}},
		{Icon: wgtheme.IconKey, Tooltip: l.T(lang.Detail_tip_copy_key), OnTap: func(r int) {
			onCopy(l.T(lang.Detail_label_public_key), (*rows)[r].publicKey)
		}},
	}
	return t
}

// fixedHeight constrains an object's height while letting it take the full
// available width.
//
// A GridWrap layout cannot do this: it sizes cells to the size it is given, so
// a zero width collapses the child to nothing. Stack instead sizes every child
// to the container, and the container's minimum height comes from the
// invisible spacer.
func fixedHeight(obj fyne.CanvasObject, h float32) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, h))
	return container.NewStack(spacer, obj)
}

// emptyNote renders the placeholder shown when a section has no content.
func emptyNote(text string) fyne.CanvasObject {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Importance = widget.LowImportance
	return container.New(layout.NewCustomPaddedLayout(
		wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceSM, wgtheme.SpaceSM), label)
}
