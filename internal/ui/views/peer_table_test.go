package ui

import (
	"testing"
	"time"

	"wgAdmin/internal/wgstat"
)

func TestSortPeerRowsKeepsEmptyLast(t *testing.T) {
	now := time.Now()
	rows := []peerRow{
		{name: "idle", handshake: wgstat.Dash},
		{name: "big", handshake: "1s", rxBytes: 10 << 20, handshakeAt: now.Add(-time.Hour)},
		{name: "small", handshake: "1s", rxBytes: 9 << 10, handshakeAt: now},
	}

	names := func() []string {
		out := make([]string, len(rows))
		for i, r := range rows {
			out[i] = r.name
		}
		return out
	}

	sortPeerRows(rows, colRX, true)
	if got := names(); got[0] != "small" || got[1] != "big" || got[2] != "idle" {
		t.Errorf("RX ascending = %v, want numeric order with the dash row last", got)
	}
	sortPeerRows(rows, colRX, false)
	if got := names(); got[0] != "big" || got[2] != "idle" {
		t.Errorf("RX descending = %v, the dash row must stay last", got)
	}
	sortPeerRows(rows, colHandshake, true)
	if got := names(); got[0] != "small" || got[2] != "idle" {
		t.Errorf("handshake ascending = %v, want most recent first and never-seen last", got)
	}
}

func TestSplitUnit(t *testing.T) {
	for in, want := range map[string][2]string{
		"1.4 GiB":  {"1.4", "GiB"},
		"96 KiB/s": {"96", "KiB/s"},
		"—":        {"—", ""},
	} {
		v, u := splitUnit(in)
		if v != want[0] || u != want[1] {
			t.Errorf("splitUnit(%q) = %q, %q", in, v, u)
		}
	}
}
