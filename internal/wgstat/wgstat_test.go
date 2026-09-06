package wgstat

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KiB"},
		{1536, "1.50 KiB"},
		{10 * 1024, "10.0 KiB"},
		{100 * 1024, "100 KiB"},
		{1024 * 1024, "1.00 MiB"},
		{1288490188, "1.20 GiB"},
		{-1, Dash},
	}
	for _, tc := range tests {
		if got := FormatBytes(tc.in); got != tc.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatAge(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		ago  time.Duration
		want string
	}{
		{0, "0s"},
		{12 * time.Second, "12s"},
		{90 * time.Second, "1m"},
		{45 * time.Minute, "45m"},
		{2*time.Hour + 14*time.Minute, "2h14m"},
		{3 * time.Hour, "3h"},
		{50 * time.Hour, "2d2h"},
		{72 * time.Hour, "3d"},
	}
	for _, tc := range tests {
		if got := FormatAge(now.Add(-tc.ago), now); got != tc.want {
			t.Errorf("FormatAge(-%v) = %q, want %q", tc.ago, got, tc.want)
		}
	}

	// A never-handshaken peer must read as a dash, not as "0s", which would
	// wrongly suggest it just connected.
	if got := FormatAge(time.Time{}, now); got != Dash {
		t.Errorf("FormatAge(zero) = %q, want %q", got, Dash)
	}

	// Clock skew between the snapshot and now must not produce a negative age.
	if got := FormatAge(now.Add(time.Minute), now); got != "0s" {
		t.Errorf("FormatAge(future) = %q, want %q", got, "0s")
	}
}

func TestPeerState(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		hs   time.Time
		want PeerState
	}{
		{"never", time.Time{}, StateNever},
		{"just now", now.Add(-5 * time.Second), StateOnline},
		{"within window", now.Add(-2 * time.Minute), StateOnline},
		{"just past window", now.Add(-4 * time.Minute), StateIdle},
		{"idle edge", now.Add(-14 * time.Minute), StateIdle},
		{"stale", now.Add(-time.Hour), StateStale},
	}
	for _, tc := range tests {
		p := Peer{Handshake: tc.hs}
		if got := p.State(now); got != tc.want {
			t.Errorf("%s: State() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTrackerComputesRates(t *testing.T) {
	tr := NewTracker()
	base := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

	snap := func(at time.Time, rx, tx int64) Snapshot {
		return Snapshot{
			Available: true,
			Taken:     at,
			Devices:   map[string]Device{"wg0": {Name: "wg0", RX: rx, TX: tx}},
		}
	}

	// The first observation establishes a baseline and yields no rate: without
	// two readings there is nothing to divide.
	tr.Observe(snap(base, 1000, 500))
	if _, ok := tr.Rate("wg0"); ok {
		t.Fatal("expected no rate after a single observation")
	}

	// 10 KiB received over 10s = 1 KiB/s.
	tr.Observe(snap(base.Add(10*time.Second), 1000+10240, 500+5120))
	got, ok := tr.Rate("wg0")
	if !ok {
		t.Fatal("expected a rate after two observations")
	}
	if got.RX != 1024 || got.TX != 512 {
		t.Errorf("rate = RX %.0f TX %.0f, want RX 1024 TX 512", got.RX, got.TX)
	}
}

func TestTrackerHandlesCounterReset(t *testing.T) {
	tr := NewTracker()
	base := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	snap := func(at time.Time, rx int64) Snapshot {
		return Snapshot{
			Available: true,
			Taken:     at,
			Devices:   map[string]Device{"wg0": {Name: "wg0", RX: rx}},
		}
	}

	tr.Observe(snap(base, 1_000_000))
	tr.Observe(snap(base.Add(5*time.Second), 1_010_000))
	if len(tr.History("wg0")) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(tr.History("wg0")))
	}

	// Bringing the interface down and up resets the kernel counters. That must
	// clear the history rather than record a huge negative or bogus rate.
	tr.Observe(snap(base.Add(10*time.Second), 200))
	if h := tr.History("wg0"); len(h) != 0 {
		t.Errorf("expected history cleared after counter reset, got %d samples", len(h))
	}
}

func TestTrackerForgetsRemovedInterfaces(t *testing.T) {
	tr := NewTracker()
	base := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

	tr.Observe(Snapshot{Available: true, Taken: base,
		Devices: map[string]Device{"wg0": {Name: "wg0", RX: 100}}})
	tr.Observe(Snapshot{Available: true, Taken: base.Add(5 * time.Second),
		Devices: map[string]Device{"wg0": {Name: "wg0", RX: 200}}})

	// wg0 is gone — the tunnel was deactivated.
	tr.Observe(Snapshot{Available: true, Taken: base.Add(10 * time.Second),
		Devices: map[string]Device{}})
	if h := tr.History("wg0"); h != nil {
		t.Errorf("expected history dropped for removed interface, got %d samples", len(h))
	}
}

func TestTrackerIgnoresUnavailableSnapshots(t *testing.T) {
	tr := NewTracker()
	base := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)

	tr.Observe(Snapshot{Available: true, Taken: base,
		Devices: map[string]Device{"wg0": {Name: "wg0", RX: 100}}})

	// A failed read (e.g. lost privileges) must not be mistaken for the
	// counters dropping to zero.
	tr.Observe(Snapshot{Available: false, Taken: base.Add(5 * time.Second)})
	tr.Observe(Snapshot{Available: true, Taken: base.Add(10 * time.Second),
		Devices: map[string]Device{"wg0": {Name: "wg0", RX: 10340}}})

	if _, ok := tr.Rate("wg0"); !ok {
		t.Error("expected the rate to survive an unavailable snapshot")
	}
}

func TestOnlinePeers(t *testing.T) {
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	d := Device{Peers: []Peer{
		{Handshake: now.Add(-10 * time.Second)},
		{Handshake: now.Add(-2 * time.Minute)},
		{Handshake: now.Add(-30 * time.Minute)},
		{}, // never
	}}
	if got := d.OnlinePeers(now); got != 2 {
		t.Errorf("OnlinePeers() = %d, want 2", got)
	}
}

func TestShortKey(t *testing.T) {
	long := "VAm4R8lFDNtEtddTlF0Bxyz1234567890abcdefghijk="
	if got := ShortKey(long); got != "VAm4R8lFDNtE…" {
		t.Errorf("ShortKey() = %q", got)
	}
	if got := ShortKey("short"); got != "short" {
		t.Errorf("ShortKey(short) = %q, want unchanged", got)
	}
}
