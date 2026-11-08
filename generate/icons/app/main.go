// Command app draws the app, window, list, brand and empty-state icons.
//
//	go run ./generate/icons/app
package main

import (
	"math"

	. "wgAdmin/generate/icondsl"
)

// IW is the outline weight of the 64px empty-state illustrations.
const IW = 2.5

// ------------------------------------------------------------------ helpers

// resample returns pts with extra points so no segment is longer than step.
func resample(pts []Pt, step float64) []Pt {
	if len(pts) == 0 {
		return nil
	}
	out := []Pt{pts[0]}
	for i := 0; i+1 < len(pts); i++ {
		a, b := pts[i], pts[i+1]
		n := int(math.Ceil(math.Hypot(b.X-a.X, b.Y-a.Y) / step))
		for k := 1; k <= n; k++ {
			t := float64(k) / float64(n)
			out = append(out, P(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t))
		}
	}
	return out
}

// dashed strokes an open path as dashes of length on, separated by off.
func dashed(w, on, off float64, pts ...Pt) string {
	var s string
	pts = resample(pts, 0.1)
	var cur []Pt
	dist, drawing := 0.0, true
	for i, p := range pts {
		if i > 0 {
			q := pts[i-1]
			dist += math.Hypot(p.X-q.X, p.Y-q.Y)
		}
		lim := off
		if drawing {
			lim = on
		}
		if drawing {
			cur = append(cur, p)
		}
		if dist >= lim {
			if drawing && len(cur) > 1 {
				s += Polyline(w, cur...)
			}
			cur = nil
			drawing = !drawing
			dist = 0
			if drawing {
				cur = append(cur, p)
			}
		}
	}
	if drawing && len(cur) > 1 {
		s += Polyline(w, cur...)
	}
	return s
}

// clipOut strokes the path, leaving out every part closer than r to (cx, cy),
// so an overlapping badge or lens gets a clean gap.
func clipOut(w, cx, cy, r float64, pts ...Pt) string {
	var s string
	var cur []Pt
	for _, p := range resample(pts, 0.1) {
		if math.Hypot(p.X-cx, p.Y-cy) < r {
			if len(cur) > 1 {
				s += Polyline(w, cur...)
			}
			cur = nil
			continue
		}
		cur = append(cur, p)
	}
	if len(cur) > 1 {
		s += Polyline(w, cur...)
	}
	return s
}

// closed returns pts with the first point appended, for closed paths.
func closed(pts []Pt) []Pt { return append(append([]Pt{}, pts...), pts[0]) }

// rrectPts is the centre line of a rounded rectangle, clockwise from top-left.
func rrectPts(x, y, w, h, r float64) []Pt {
	return Join(
		ArcPts(x+w-r, y+r, r, 270, 360, 8),
		ArcPts(x+w-r, y+h-r, r, 0, 90, 8),
		ArcPts(x+r, y+h-r, r, 90, 180, 8),
		ArcPts(x+r, y+r, r, 180, 270, 8),
	)
}

// shield is the brand shield outline centred on cx, scaled by k from the 24px
// design (top at y=2.75, bottom at y=21.25).
func shield(cx, k, dy float64) []Pt {
	m := func(x, y float64) Pt { return P(cx+(x-12)*k, dy+y*k) }
	right := Join(
		[]Pt{m(12, 2.75), m(19.25, 5.25)},
		Cubic(m(19.25, 5.25), m(19.25, 7), m(19.25, 9.5), m(19.25, 11), 6),
		Cubic(m(19.25, 11), m(19.25, 16), m(15.75, 19.5), m(12, 21.25), 12),
	)
	pts := append([]Pt{}, right...)
	for i := len(right) - 2; i >= 1; i-- {
		p := right[i]
		pts = append(pts, P(2*cx-p.X, p.Y))
	}
	return pts
}

// link is the brand motif: two nodes joined by an S-shaped path, scaled from
// the 24px design like shield.
func link(cx, k, dy float64) (a, b Pt, path []Pt) {
	m := func(x, y float64) Pt { return P(cx+(x-12)*k, dy+y*k) }
	a, b = m(8.25, 14), m(15.75, 10)
	path = Cubic(a, m(12, 14), m(12, 10), b, 16)
	return
}

// crescent is the outline of circle 1 with circle 2 bitten out of it.
func crescent(x1, y1, r1, x2, y2, r2 float64) []Pt {
	d := math.Hypot(x2-x1, y2-y1)
	phi := math.Atan2(y2-y1, x2-x1) * 180 / math.Pi
	a := (r1*r1 - r2*r2 + d*d) / (2 * d)
	h := math.Sqrt(r1*r1 - a*a)
	alpha := math.Acos(a/r1) * 180 / math.Pi
	beta := math.Atan2(h, d-a) * 180 / math.Pi
	return Join(
		ArcPts(x1, y1, r1, phi+alpha, phi+360-alpha, 40),
		ArcPts(x2, y2, r2, phi+180+beta, phi+180-beta, 24),
	)
}

// person is a head and shoulders; body is the bottom-closed shoulder arch.
func person(cx, top, hr, br float64) (head Pt, body []Pt) {
	head = P(cx, top+hr)
	base := top + 2*hr + 4 + br
	body = Join(ArcPts(cx, base, br, 180, 360, 24))
	body = append(body, P(cx+br, base), P(cx-br, base))
	return
}

// --------------------------------------------------------------------- icons

func main() {
	sh := shield(12, 1, 0)
	la, lb, lpath := link(12, 1, 0)

	// Illustration geometry.
	ish := shield(28, 2.35, 3)
	ia, ib, ipath := link(28, 2.35, 3)
	const bx, by, br = 46.5, 46.5, 8.5

	pHead, pBody := person(23, 17, 7, 13)
	gHead, gBody := person(45, 23, 5.5, 9.5)

	card := rrectPts(8, 9, 34, 42, 5)
	const lx, ly, lr = 42, 40, 10.5

	Main(
		// ---- chevrons: glyphs, fg only.
		New("chevron-right").Fg(PL(P(9.5, 6.5), P(15, 12), P(9.5, 17.5))),
		New("chevron-left").Fg(PL(P(14.5, 6.5), P(9, 12), P(14.5, 17.5))),
		New("chevron-down").Fg(PL(P(6.5, 9.5), P(12, 15), P(17.5, 9.5))),
		New("chevron-up").Fg(PL(P(6.5, 14.5), P(12, 9), P(17.5, 14.5))),

		// ---- window.
		New("fullscreen").
			Duo(RRect(8, 8, 8, 8, 1.5)).
			Fg(
				PL(P(4, 9), P(4, 4), P(9, 4)), PL(P(15, 4), P(20, 4), P(20, 9)),
				PL(P(20, 15), P(20, 20), P(15, 20)), PL(P(9, 20), P(4, 20), P(4, 15)),
			),
		New("fullscreen-exit").
			Fg(
				PL(P(3.5, 8.5), P(8.5, 8.5), P(8.5, 3.5)), PL(P(15.5, 3.5), P(15.5, 8.5), P(20.5, 8.5)),
				PL(P(20.5, 15.5), P(15.5, 15.5), P(15.5, 20.5)), PL(P(8.5, 20.5), P(8.5, 15.5), P(3.5, 15.5)),
			),

		New("keyboard").
			Duo(RRect(3, 6, 18, 12, 2.5)).
			Fg(RR(3, 6, 18, 12, 2.5),
				Circle(7.5, 10, 1.15), Circle(12, 10, 1.15), Circle(16.5, 10, 1.15),
				L(8.5, 14, 15.5, 14)),

		// ⌘: four loops around a centre square.
		New("command").
			Duo(RRect(9, 9, 6, 6, 0)).
			Fg(
				L(9, 6.5, 9, 17.5), L(15, 6.5, 15, 17.5),
				L(6.5, 9, 17.5, 9), L(6.5, 15, 17.5, 15),
				A(6.5, 6.5, 2.5, 90, 360), A(17.5, 6.5, 2.5, 180, 450),
				A(17.5, 17.5, 2.5, 270, 540), A(6.5, 17.5, 2.5, 0, 270),
			),

		// ---- theme.
		New("sun").
			Duo(Circle(12, 12, 4)).
			Fg(R(12, 12, 4), sunRays()),
		New("moon").
			Duo(Poly(crescent(11.5, 12.5, 8, 16.5, 7.5, 6)...)).
			Fg(PolyStroke(SW, crescent(11.5, 12.5, 8, 16.5, 7.5, 6)...)),
		New("monitor").
			Duo(RRect(3, 4, 18, 12, 2)).
			Fg(RR(3, 4, 18, 12, 2), L(12, 16, 12, 19.75), L(8.5, 19.75, 15.5, 19.75)),

		// ---- lists: glyphs, fg only.
		New("sort-asc").Fg(L(12, 19, 12, 5.5), PL(P(6.5, 11), P(12, 5.5), P(17.5, 11))),
		New("sort-desc").Fg(L(12, 5, 12, 18.5), PL(P(6.5, 13), P(12, 18.5), P(17.5, 13))),
		New("grip").Fg(
			Circle(9, 6, 1.5), Circle(15, 6, 1.5),
			Circle(9, 12, 1.5), Circle(15, 12, 1.5),
			Circle(9, 18, 1.5), Circle(15, 18, 1.5),
		),

		// ---- brand: a shield carrying two peers joined by a tunnel path.
		New("brand").
			Duo(Poly(sh...)).
			Fg(PolyStroke(SW, sh...),
				Polyline(SW, lpath...),
				Circle(la.X, la.Y, 1.6), Circle(lb.X, lb.Y, 1.6)),

		// ---- empty states (64px).
		NewIllustration("empty-tunnels").
			Duo(Poly(ish...), Circle(bx, by, br)).
			Fg(clipOut(IW, bx, by, br+IW+2, closed(ish)...),
				dashed(IW, 3, 3.5, ipath...),
				Circle(ia.X, ia.Y, 3), Circle(ib.X, ib.Y, 3),
				Ring(bx, by, br, IW),
				Line(P(bx, by-4), P(bx, by+4), IW), Line(P(bx-4, by), P(bx+4, by), IW)),

		NewIllustration("empty-peers").
			Duo(Circle(pHead.X, pHead.Y, 7), Poly(pBody...)).
			Fg(Ring(pHead.X, pHead.Y, 7, IW), PolyStroke(IW, pBody...),
				dashed(IW, 3, 3.2, closed(ArcPts(gHead.X, gHead.Y, 5.5, -90, 270, 48))...),
				dashed(IW, 3, 3.2, closed(gBody)...)),

		NewIllustration("empty-search").
			Duo(Poly(card...), Circle(lx, ly, lr)).
			Fg(clipOut(IW, lx, ly, lr+IW+2, closed(card)...),
				clipOut(IW, lx, ly, lr+IW+2, P(15, 19), P(35, 19)),
				clipOut(IW, lx, ly, lr+IW+2, P(15, 27), P(31, 27)),
				clipOut(IW, lx, ly, lr+IW+2, P(15, 35), P(27, 35)),
				Ring(lx, ly, lr, IW),
				Line(P(lx+lr*0.707+2, ly+lr*0.707+2), P(55, 53), IW+0.5)),
	)
}

func sunRays() string {
	var s string
	for i := 0; i < 8; i++ {
		t := float64(i) * math.Pi / 4
		c, si := math.Cos(t), math.Sin(t)
		s += L(12+7*c, 12+7*si, 12+9*c, 12+9*si)
	}
	return s
}
