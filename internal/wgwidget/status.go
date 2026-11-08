package wgwidget

import (
	"image/color"
	"strconv"
	"time"

	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgstat"
)

// StateTone maps a peer state onto its palette colour.
func StateTone(s wgstat.PeerState) color.NRGBA {
	p := wgtheme.Current()
	switch s {
	case wgstat.StateOnline:
		return p.Success
	case wgstat.StateIdle:
		return p.Warning
	case wgstat.StateStale:
		return p.Error
	default:
		return p.TextDim
	}
}

// RelativeTime renders a timestamp as "just now" / "12m ago" / "3d ago" for
// human-facing lists such as backups, where an exact clock time is noise.
func RelativeTime(t time.Time, now time.Time) string {
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return plural(int(d.Minutes()), "minute") + " ago"
	case d < 24*time.Hour:
		return plural(int(d.Hours()), "hour") + " ago"
	case d < 30*24*time.Hour:
		return plural(int(d.Hours())/24, "day") + " ago"
	default:
		return t.Format("2006-01-02")
	}
}

func plural(n int, unit string) string {
	s := ""
	if n != 1 {
		s = "s"
	}
	return strconv.Itoa(n) + " " + unit + s
}
