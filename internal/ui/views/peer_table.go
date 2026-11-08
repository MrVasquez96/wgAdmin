package ui

import (
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
}

// Column indexes into peerColumns.
const (
	colState = iota
	colQR
	colPeer
	colAllowedIPs
	colEndpoint
	colHandshake
	colRX
	colTX
)

// peerColumns defines the table layout. Widths are fixed so values line up
// down each column, and the total is kept under the width of a detail pane at
// the default window size so the table does not need horizontal scrolling to
// show the traffic counters — the columns people came to read.
var peerColumns = []struct {
	titleKey string
	width    float32
}{
	{lang.Detail_col_state, 24},
	// The QR action sits next to the state dot, inside the sticky columns, so
	// it stays in reach when the table is scrolled sideways.
	{lang.Detail_col_qr, 40},
	{lang.Detail_col_peer, 112},
	{lang.Detail_col_allowed_ips, 140},
	{lang.Detail_col_endpoint, 132},
	// Wide enough for the "HANDSHAKE" heading itself, which otherwise ran
	// straight into the RX column.
	{lang.Detail_col_handshake, 96},
	{lang.Detail_col_rx, 76},
	{lang.Detail_col_tx, 76},
}

// Row metrics used to size the peers section. Fyne clamps a table row to the
// natural height of its cell content, so these mirror what the table actually
// lays out at the current type scale rather than trying to force it smaller —
// guessing low clipped the final row.
const (
	peerRowHeight    = 40
	peerHeaderHeight = 32
)

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

// peerCell is the reusable cell body: a mono label plus a state dot slot that
// only the first column uses.
type peerCell struct {
	label *widget.Label
	dot   *fyne.Container
	obj   *fyne.Container
}

func newPeerCell() *peerCell {
	label := widget.NewLabel("")
	label.TextStyle = fyne.TextStyle{Monospace: true}
	label.Truncation = fyne.TextTruncateEllipsis

	dot := container.NewCenter()

	return &peerCell{
		label: label,
		dot:   dot,
		obj:   container.NewStack(dot, label),
	}
}

// qrCellIcon is the QR action glyph, in the accent colour so it reads as a
// control rather than as data.
var qrCellIcon = theme.NewPrimaryThemedResource(wgtheme.IconQR)

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

// newPeerTable builds the live peer table. Clicking the QR cell calls onQR with
// the row's public key; clicking a data cell calls onCopy with the column title
// and the cell's full value. Deliberately, a plain click on data never opens a
// dialog — those cells look like text, so they behave like text.
func newPeerTable(rows *[]peerRow, l *lang.Language,
	onQR func(publicKey string), onCopy func(column, value string)) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (int, int) { return len(*rows), len(peerColumns) },
		func() fyne.CanvasObject { return newPeerCell().obj },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			box, ok := obj.(*fyne.Container)
			if !ok || id.Row >= len(*rows) {
				return
			}
			dotBox, _ := box.Objects[0].(*fyne.Container)
			label, _ := box.Objects[1].(*widget.Label)
			if label == nil || dotBox == nil {
				return
			}

			r := (*rows)[id.Row]
			dotBox.Objects = nil
			label.TextStyle = fyne.TextStyle{Monospace: true}

			switch id.Col {
			case colState:
				// State is a dot, not text — it needs no width and reads fast.
				label.SetText("")
				if r.live {
					dotBox.Objects = []fyne.CanvasObject{wgwidget.PeerStateDot(r.state)}
				}
			case colQR:
				label.SetText("")
				dotBox.Objects = []fyne.CanvasObject{widget.NewIcon(qrCellIcon)}
			case colPeer:
				label.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
				label.SetText(r.name)
			default:
				label.SetText(r.cellValue(id.Col))
			}

			dotBox.Refresh()
			label.Refresh()
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
		// Row headers are suppressed; only the column strip carries titles.
		if id.Col < 0 || id.Col >= len(peerColumns) {
			h.SetText("")
			return
		}
		h.SetText(strings.ToUpper(l.T(peerColumns[id.Col].titleKey)))
	}

	table.ShowHeaderRow = true
	table.ShowHeaderColumn = false
	table.StickyColumnCount = colQR + 1

	for i, c := range peerColumns {
		table.SetColumnWidth(i, c.width)
	}

	table.OnSelected = func(id widget.TableCellID) {
		defer table.UnselectAll()
		if id.Row < 0 || id.Row >= len(*rows) {
			return
		}
		r := (*rows)[id.Row]
		switch id.Col {
		case colState:
		case colQR:
			onQR(r.publicKey)
		default:
			if v := r.cellValue(id.Col); v != "" && v != wgstat.Dash {
				onCopy(l.T(peerColumns[id.Col].titleKey), v)
			}
		}
	}

	return table
}

// peerTableHeight sizes the table to its content so it sits naturally inside
// the scrolling detail pane instead of fighting it for space.
func peerTableHeight(rowCount int) float32 {
	if rowCount == 0 {
		return peerHeaderHeight
	}
	// The +6 covers the table's own border and horizontal scrollbar so the
	// final row is fully visible.
	h := float32(rowCount)*peerRowHeight + peerHeaderHeight + 6
	if max := float32(360); h > max {
		return max
	}
	return h
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
