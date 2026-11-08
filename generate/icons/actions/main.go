// Command actions draws the toolbar/action icons.
//
//	go run ./generate/icons/actions
package main

import (
	"math"

	. "wgAdmin/generate/icondsl"
)

// rounded softens the corners of a polyline with quadratic fillets of radius r
// (clamped to half of each adjoining edge). Open polylines keep their ends.
func rounded(r float64, closed bool, pts ...Pt) []Pt {
	n := len(pts)
	var out []Pt
	for i := 0; i < n; i++ {
		c := pts[i]
		if !closed && (i == 0 || i == n-1) {
			out = append(out, c)
			continue
		}
		p, q := pts[(i-1+n)%n], pts[(i+1)%n]
		toward := func(t Pt) Pt {
			dx, dy := t.X-c.X, t.Y-c.Y
			l := math.Hypot(dx, dy)
			d := math.Min(r, l/2)
			return P(c.X+dx/l*d, c.Y+dy/l*d)
		}
		out = append(out, Quad(toward(p), c, toward(q), 6)...)
	}
	return out
}

// rot rotates points by deg (clockwise on screen) about (cx, cy).
func rot(deg, cx, cy float64, pts ...Pt) []Pt {
	t := deg * math.Pi / 180
	s, c := math.Sin(t), math.Cos(t)
	out := make([]Pt, len(pts))
	for i, p := range pts {
		x, y := p.X-cx, p.Y-cy
		out[i] = P(cx+x*c-y*s, cy+x*s+y*c)
	}
	return out
}

// tray is the open-top box shared by download and export.
func tray() []Pt {
	return rounded(2.5, false, P(4, 14.5), P(4, 20), P(20, 20), P(20, 14.5))
}

func gear() []Pt {
	const teeth = 8
	const ro, ri = 9, 6.6
	var pts []Pt
	for i := 0; i < teeth; i++ {
		c := float64(i)*360/teeth - 90
		at := func(r, a float64) Pt {
			t := (c + a) * math.Pi / 180
			return P(12+r*math.Cos(t), 12+r*math.Sin(t))
		}
		pts = append(pts, at(ri, -22.5+5), at(ri, -12), at(ro, -9), at(ro, 9), at(ri, 12))
	}
	return rounded(0.9, true, pts...)
}

func main() {
	// pencil drawn horizontally (tip left) then turned to the usual diagonal.
	pencil := func(pts ...Pt) []Pt {
		return rot(-45, 0, 0, pts...)
	}
	shift := func(dx, dy float64, pts []Pt) []Pt {
		out := make([]Pt, len(pts))
		for i, p := range pts {
			out[i] = P(p.X+dx, p.Y+dy)
		}
		return out
	}
	pBody := shift(11.6, 12.4, pencil(P(-9, 0), P(-5.25, -2.75), P(8.25, -2.75), P(8.25, 2.75), P(-5.25, 2.75)))
	pBand := shift(11.6, 12.4, pencil(P(4.5, -2.75), P(4.5, 2.75)))

	peerBody := rounded(3.5, false, P(5, 20.5), P(5, 14.5), P(19, 14.5), P(19, 20.5))

	doc := []Pt{P(6.5, 3), P(13.5, 3), P(19, 8.5), P(19, 21), P(6.5, 21)}

	Main(
		New("stop").
			Duo(RRect(5, 5, 14, 14, 3)).
			Fg(RR(5, 5, 14, 14, 3)),

		New("power").
			Fg(A(12, 12.75, 8, 300, 240), L(12, 3.25, 12, 11)),

		// Radar: two range rings and a beam with a soft trailing sweep.
		New("scan").
			Duo(Poly(Join([]Pt{P(12, 12)}, ArcPts(12, 12, 9, 240, 315, 12))...)).
			Fg(R(12, 12, 9), R(12, 12, 4.5), L(12, 12, 12+6.36, 12-6.36),
				Circle(12, 12, 1.3)),

		New("peer").
			Duo(Circle(12, 7.5, 3.75), Poly(peerBody...)).
			Fg(R(12, 7.5, 3.75), PL(peerBody...)),

		New("peers").
			Duo(Circle(9, 7.5, 3.5), Poly(rounded(3.5, false, P(3, 20.5), P(3, 14.5), P(15, 14.5), P(15, 20.5))...)).
			Fg(R(9, 7.5, 3.5), PL(rounded(3.5, false, P(3, 20.5), P(3, 14.5), P(15, 14.5), P(15, 20.5))...),
				A(15, 7.5, 3.5, -70, 70),
				PL(rounded(3.5, false, P(18, 14.5), P(21, 14.5), P(21, 20.5))...)),

		New("edit").
			Duo(Poly(pBody...)).
			Fg(PolyStroke(SW, rounded(1, true, pBody...)...), PL(pBand...)),

		New("trash").
			Duo(Poly(rounded(2.5, false, P(6, 6.5), P(6, 20.5), P(18, 20.5), P(18, 6.5))...)).
			Fg(L(4, 6.5, 20, 6.5),
				PL(rounded(1.5, false, P(9, 6.5), P(9, 3.5), P(15, 3.5), P(15, 6.5))...),
				PL(rounded(2.5, false, P(6, 6.5), P(6, 20.5), P(18, 20.5), P(18, 6.5))...),
				L(10, 10.5, 10, 16.5), L(14, 10.5, 14, 16.5)),

		New("refresh").
			Fg(A(12, 12, 7.75, 200, 330), A(12, 12, 7.75, 20, 150),
				arrowHead(12, 12, 7.75, 330), arrowHead(12, 12, 7.75, 150)),

		New("folder").
			Duo(Poly(folder()...)).
			Fg(PolyStroke(SW, folder()...)),

		New("archive").
			Duo(Poly(rounded(2.5, false, P(4.5, 8.5), P(4.5, 20), P(19.5, 20), P(19.5, 8.5))...)).
			Fg(RR(3, 4, 18, 4.5, 1.5),
				PL(rounded(2.5, false, P(4.5, 8.5), P(4.5, 20), P(19.5, 20), P(19.5, 8.5))...),
				L(10, 12.5, 14, 12.5)),

		New("settings").
			Duo(Poly(gear()...), CircleHole(12, 12, 3)).
			Fg(PolyStroke(SW, gear()...), R(12, 12, 3)),

		New("terminal").
			Duo(RRect(3, 4.5, 18, 15, 2.5)).
			Fg(RR(3, 4.5, 18, 15, 2.5), PL(P(7, 9.5), P(10, 12), P(7, 14.5)), L(12.5, 14.5, 17, 14.5)),

		New("qr").
			Duo(RRect(4, 4, 6, 6, 1.5), RRect(14, 4, 6, 6, 1.5), RRect(4, 14, 6, 6, 1.5)).
			Fg(RR(4, 4, 6, 6, 1.5), RR(14, 4, 6, 6, 1.5), RR(4, 14, 6, 6, 1.5),
				RRect(13.1, 13.1, 2.8, 2.8, 0.7), RRect(18.1, 18.1, 2.8, 2.8, 0.7),
				RRect(18.1, 13.1, 2.8, 2.8, 0.7), RRect(13.1, 18.1, 2.8, 2.8, 0.7)),

		New("filter").
			Duo(Poly(funnel()...)).
			Fg(PolyStroke(SW, funnel()...)),

		New("search").
			Duo(Circle(10.5, 10.5, 6.5)).
			Fg(R(10.5, 10.5, 6.5), L(15.4, 15.4, 20, 20)),

		New("add").
			Fg(L(12, 5, 12, 19), L(5, 12, 19, 12)),

		New("close").
			Fg(L(6.5, 6.5, 17.5, 17.5), L(17.5, 6.5, 6.5, 17.5)),

		New("download").
			Fg(PL(tray()...), L(12, 4, 12, 14.5), PL(P(7.75, 10.25), P(12, 14.5), P(16.25, 10.25))),

		New("export").
			Fg(PL(tray()...), L(12, 4, 12, 14.5), PL(P(7.75, 8.25), P(12, 4), P(16.25, 8.25))),

		New("import").
			Duo(Poly(rounded(2.5, true, doc...)...)).
			Fg(PL(rounded(2.5, false, P(6.5, 8.5), P(6.5, 3), P(13.5, 3), P(19, 8.5), P(19, 21), P(6.5, 21), P(6.5, 17.5))...),
				PL(rounded(1.5, false, P(13.5, 3), P(13.5, 8.5), P(19, 8.5))...),
				L(3, 13, 12.5, 13), PL(P(9.25, 9.75), P(12.5, 13), P(9.25, 16.25))),

		New("external").
			Duo(RRect(4, 6, 14, 14, 2.5)).
			Fg(PL(rounded(2.5, false, P(10.5, 6), P(4, 6), P(4, 20), P(18, 20), P(18, 13.5))...),
				L(11, 13, 20, 4), PL(P(14, 4), P(20, 4), P(20, 10))),

		New("more-vertical").
			Fg(Circle(12, 5.5, 1.6), Circle(12, 12, 1.6), Circle(12, 18.5, 1.6)),

		New("more-horizontal").
			Fg(Circle(5.5, 12, 1.6), Circle(12, 12, 1.6), Circle(18.5, 12, 1.6)),

		New("link").
			Fg(PL(rot(-45, 12, 12, Join([]Pt{P(10, 16.5)}, ArcPts(7.5, 12, 4.5, 90, 270, 16), []Pt{P(10, 7.5)})...)...),
				PL(rot(-45, 12, 12, Join([]Pt{P(14, 7.5)}, ArcPts(16.5, 12, 4.5, -90, 90, 16), []Pt{P(14, 16.5)})...)...),
				PL(rot(-45, 12, 12, P(9, 12), P(15, 12))...)),

		New("key").
			Duo(Circle(7.5, 16.5, 4)).
			Fg(R(7.5, 16.5, 4), L(10.6, 13.4, 19.5, 4.5), L(17.5, 6.5, 20.25, 9.25), L(14.5, 9.5, 16.75, 11.75)),
	)
}

// arrowHead is a chevron at angle deg on a circle, pointing along the
// clockwise tangent.
func arrowHead(cx, cy, r, deg float64) string {
	t := deg * math.Pi / 180
	tip := P(cx+r*math.Cos(t), cy+r*math.Sin(t))
	tx, ty := -math.Sin(t), math.Cos(t) // clockwise tangent
	nx, ny := math.Cos(t), math.Sin(t)  // outward normal
	const k = 3.25
	return PL(P(tip.X+(-tx+nx)*k*0.7071, tip.Y+(-ty+ny)*k*0.7071), tip,
		P(tip.X+(-tx-nx)*k*0.7071, tip.Y+(-ty-ny)*k*0.7071))
}

func folder() []Pt {
	return rounded(2, true, P(3, 4.5), P(9, 4.5), P(11, 7), P(21, 7), P(21, 19.5), P(3, 19.5))
}

func funnel() []Pt {
	return rounded(1.25, true, P(3.5, 4.5), P(20.5, 4.5), P(14, 12), P(14, 18.5), P(10, 20.5), P(10, 12))
}
