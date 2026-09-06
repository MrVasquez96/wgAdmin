package wgstat

import (
	"sync"
	"time"
)

// HistoryLength is how many samples the tracker keeps per interface. At the
// default five-second refresh that is about four minutes of history, which is
// enough for a sparkline without holding meaningful memory.
const HistoryLength = 48

// Sample is one throughput reading derived from two consecutive snapshots.
type Sample struct {
	At time.Time
	RX float64 // bytes per second
	TX float64
}

// Tracker turns the cumulative byte counters in successive snapshots into
// throughput rates and keeps a short rolling history per interface.
//
// It is safe for concurrent use: the UI refresh goroutine writes while the
// render path reads.
type Tracker struct {
	mu      sync.Mutex
	last    map[string]counters
	history map[string][]Sample
}

type counters struct {
	rx, tx int64
	at     time.Time
}

// NewTracker creates an empty tracker.
func NewTracker() *Tracker {
	return &Tracker{
		last:    map[string]counters{},
		history: map[string][]Sample{},
	}
}

// Observe folds a snapshot into the history, computing per-interface rates.
// Interfaces that have gone away are forgotten so a tunnel that is brought back
// up starts with a clean history rather than one huge synthetic spike.
func (t *Tracker) Observe(snap Snapshot) {
	if !snap.Available {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for name := range t.last {
		if _, still := snap.Devices[name]; !still {
			delete(t.last, name)
			delete(t.history, name)
		}
	}

	for name, dev := range snap.Devices {
		prev, seen := t.last[name]
		cur := counters{rx: dev.RX, tx: dev.TX, at: snap.Taken}
		t.last[name] = cur

		if !seen {
			continue // need two readings before a rate exists
		}

		elapsed := cur.at.Sub(prev.at).Seconds()
		if elapsed <= 0 {
			continue
		}

		// Counters only ever climb while an interface stays up. A decrease
		// means it was restarted, so drop the delta instead of reporting a
		// negative or absurd rate.
		rx, tx := cur.rx-prev.rx, cur.tx-prev.tx
		if rx < 0 || tx < 0 {
			t.history[name] = nil
			continue
		}

		h := append(t.history[name], Sample{
			At: cur.at,
			RX: float64(rx) / elapsed,
			TX: float64(tx) / elapsed,
		})
		if len(h) > HistoryLength {
			h = h[len(h)-HistoryLength:]
		}
		t.history[name] = h
	}
}

// Rate returns the most recent throughput sample for an interface.
func (t *Tracker) Rate(name string) (Sample, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	h := t.history[name]
	if len(h) == 0 {
		return Sample{}, false
	}
	return h[len(h)-1], true
}

// History returns a copy of the rolling samples for an interface, oldest first.
func (t *Tracker) History(name string) []Sample {
	t.mu.Lock()
	defer t.mu.Unlock()

	h := t.history[name]
	if len(h) == 0 {
		return nil
	}
	out := make([]Sample, len(h))
	copy(out, h)
	return out
}

// Reset forgets everything known about an interface. Call it when a tunnel is
// toggled so the next reading does not appear as a burst.
func (t *Tracker) Reset(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.last, name)
	delete(t.history, name)
}
