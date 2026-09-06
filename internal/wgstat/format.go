package wgstat

import (
	"fmt"
	"time"
)

// Dash is what the UI shows where a value does not exist, rather than a zero
// that would read as "measured, and it is nothing".
const Dash = "—"

// FormatBytes renders a byte count in binary units, with the precision falling
// off as the magnitude grows so columns stay narrow and easy to compare.
func FormatBytes(n int64) string {
	if n < 0 {
		return Dash
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit && exp < 4; v /= unit {
		div *= unit
		exp++
	}
	val := float64(n) / float64(div)
	suffix := [...]string{"KiB", "MiB", "GiB", "TiB", "PiB"}[exp]
	switch {
	case val >= 100:
		return fmt.Sprintf("%.0f %s", val, suffix)
	case val >= 10:
		return fmt.Sprintf("%.1f %s", val, suffix)
	default:
		return fmt.Sprintf("%.2f %s", val, suffix)
	}
}

// FormatRate renders a throughput in bytes per second.
func FormatRate(bytesPerSec float64) string {
	if bytesPerSec <= 0 {
		return Dash
	}
	return FormatBytes(int64(bytesPerSec)) + "/s"
}

// FormatAge renders how long ago t was, in the compact form operators expect
// from wg(8): "12s", "3m", "2h14m", "4d". Zero times render as a dash.
func FormatAge(t time.Time, now time.Time) string {
	if t.IsZero() {
		return Dash
	}
	d := now.Sub(t)
	if d < 0 {
		d = 0
	}

	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int(d.Minutes()) - h*60
		if m == 0 {
			return fmt.Sprintf("%dh", h)
		}
		return fmt.Sprintf("%dh%dm", h, m)
	default:
		days := int(d.Hours()) / 24
		h := int(d.Hours()) - days*24
		if h == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd%dh", days, h)
	}
}

// HandshakeAge is the peer's last-handshake age in compact form.
func (p Peer) HandshakeAge(now time.Time) string {
	return FormatAge(p.Handshake, now)
}

// ShortKey abbreviates a base64 WireGuard key for display in tight columns.
// The leading characters are enough to tell keys apart by eye.
func ShortKey(key string) string {
	const keep = 12
	if len(key) <= keep {
		return key
	}
	return key[:keep] + "…"
}
