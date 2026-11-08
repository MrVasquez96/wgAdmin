// Command fields draws the status and data-field icons: status discs, the
// glyphs that label tunnel/peer fields, and traffic/tunnel state.
//
//	go run ./generate/icons/fields
package main

import (
	"math"

	. "wgAdmin/generate/icondsl"
)

// ellipse samples an axis-aligned ellipse as a closed point list.
func ellipse(cx, cy, rx, ry float64, n int) []Pt {
	pts := make([]Pt, 0, n)
	for i := 0; i < n; i++ {
		t := 2 * math.Pi * float64(i) / float64(n)
		pts = append(pts, P(cx+rx*math.Cos(t), cy+ry*math.Sin(t)))
	}
	return pts
}

// reversed returns pts in reverse order.
func reversed(pts []Pt) []Pt {
	out := make([]Pt, len(pts))
	for i, p := range pts {
		out[len(pts)-1-i] = p
	}
	return out
}

// mirrorX reflects pts about the vertical line x = 12.
func mirrorX(pts []Pt) []Pt {
	out := make([]Pt, len(pts))
	for i, p := range pts {
		out[i] = P(24-p.X, p.Y)
	}
	return out
}

// Shared outlines.
var (
	alertTri = []Pt{P(12, 4), P(20.75, 19.25), P(3.25, 19.25)}

	shieldOutline = func() []Pt {
		right := Join([]Pt{P(12, 3.25)}, Cubic(P(19, 5.75), P(19, 14), P(16.5, 18), P(12, 20.75), 16))
		left := reversed(mirrorX(right))
		return Join(right, left[1:len(left)-1])
	}()

	eyeOutline = Join(
		Cubic(P(2.75, 12), P(6.5, 6.25), P(17.5, 6.25), P(21.25, 12), 16),
		Cubic(P(21.25, 12), P(17.5, 17.75), P(6.5, 17.75), P(2.75, 12), 16)[1:16],
	)

	docOutline = Join(
		[]Pt{P(14, 3)},
		ArcPts(7.5, 5.5, 2.5, 270, 180, 6),
		ArcPts(7.5, 18.5, 2.5, 180, 90, 6),
		ArcPts(16.5, 18.5, 2.5, 90, 0, 6),
		[]Pt{P(19, 8)},
	)

	plugBody = Join([]Pt{P(6.5, 8), P(17.5, 8)}, ArcPts(12, 11.5, 5.5, 0, 180, 16))

	archOutline = Join([]Pt{P(3.75, 20)}, ArcPts(12, 12, 8.25, 180, 360, 24), []Pt{P(20.25, 20)})
)

func tunnel(name string, inner ...string) *Icon {
	return New(name).
		Duo(Poly(archOutline...)).
		Fg(PL(archOutline...)).
		Fg(inner...)
}

func main() {
	Main(
		New("success").
			Duo(Circle(12, 12, 9)).
			Fg(R(12, 12, 9), PL(P(8, 12.25), P(11, 15.25), P(16.25, 9.25))),

		New("error").
			Duo(Circle(12, 12, 9)).
			Fg(R(12, 12, 9), L(9, 9, 15, 15), L(15, 9, 9, 15)),

		New("alert").
			Duo(Poly(alertTri...)).
			Fg(PolyStroke(SW, alertTri...), L(12, 9.5, 12, 13.25), Circle(12, 16.25, 1.15)),

		New("clock").
			Duo(Circle(12, 12, 9)).
			Fg(R(12, 12, 9), PL(P(12, 7.25), P(12, 12), P(15.25, 14))),

		New("activity").
			Fg(PL(P(2.75, 12), P(6.75, 12), P(9.25, 5.5), P(14.75, 18.5), P(17.25, 12), P(21.25, 12))),

		// Three rising bars; outline + soft fill like other object icons.
		New("signal").
			Duo(RRect(4, 15, 4, 5, 1.25), RRect(10, 10, 4, 10, 1.25), RRect(16, 4.5, 4, 15.5, 1.25)).
			Fg(RR(4, 15, 4, 5, 1.25), RR(10, 10, 4, 10, 1.25), RR(16, 4.5, 4, 15.5, 1.25)),

		New("lock").
			Duo(RRect(5, 10.5, 14, 10, 2.5)).
			Fg(RR(5, 10.5, 14, 10, 2.5),
				L(8, 10.5, 8, 7.5), A(12, 7.5, 4, 180, 360), L(16, 7.5, 16, 10.5),
				L(12, 14.25, 12, 16.75)),

		New("shield").
			Duo(Poly(shieldOutline...)).
			Fg(PolyStroke(SW, shieldOutline...)),

		New("globe").
			Duo(Circle(12, 12, 9)).
			Fg(R(12, 12, 9), L(3, 12, 21, 12), PolyStroke(SW, ellipse(12, 12, 3.75, 9, 48)...)),

		New("plug").
			Duo(Poly(plugBody...)).
			Fg(PolyStroke(SW, plugBody...),
				L(9.25, 3.5, 9.25, 8), L(14.75, 3.5, 14.75, 8),
				L(12, 17, 12, 20.75)),

		// Two stacked server units, distinct from globe.
		New("dns").
			Duo(RRect(3.75, 3.75, 16.5, 6.5, 2), RRect(3.75, 13.75, 16.5, 6.5, 2)).
			Fg(RR(3.75, 3.75, 16.5, 6.5, 2), RR(3.75, 13.75, 16.5, 6.5, 2),
				Circle(7.75, 7, 1.15), Circle(7.75, 17, 1.15),
				L(12, 7, 16.25, 7), L(12, 17, 16.25, 17)),

		New("ruler").
			Duo(RRect(3, 7.5, 18, 9, 2)).
			Fg(RR(3, 7.5, 18, 9, 2),
				L(6.75, 7.5, 6.75, 10.5), L(10.25, 7.5, 10.25, 12),
				L(13.75, 7.5, 13.75, 10.5), L(17.25, 7.5, 17.25, 12)),

		New("table").
			Duo(RRect(3.5, 4.5, 17, 15, 2)).
			Fg(RR(3.5, 4.5, 17, 15, 2),
				L(3.5, 9.5, 20.5, 9.5), L(3.5, 14.5, 20.5, 14.5), L(9.5, 9.5, 9.5, 19.5)),

		// Start and end points joined by a switchback path.
		New("route").
			Duo(Circle(6, 18.5, 2.25), Circle(18, 5.5, 2.25)).
			Fg(R(6, 18.5, 2.25), R(18, 5.5, 2.25),
				L(8.25, 18.5, 16.75, 18.5), A(16.75, 15.25, 3.25, 270, 450),
				L(16.75, 12, 7.25, 12), A(7.25, 8.75, 3.25, 90, 270),
				L(7.25, 5.5, 15.75, 5.5)),

		New("doc").
			Duo(Poly(docOutline...)).
			Fg(PolyStroke(SW, docOutline...),
				PL(P(14, 3.5), P(14, 8), P(18.5, 8)),
				L(8.5, 12.5, 15.5, 12.5), L(8.5, 16, 13, 16)),

		New("eye").
			Duo(Poly(eyeOutline...)).
			Fg(PolyStroke(SW, eyeOutline...), R(12, 12, 2.75)),

		New("eye-off").
			Duo(Poly(eyeOutline...)).
			Fg(PolyStroke(SW, eyeOutline...), R(12, 12, 2.75), L(4.5, 4.5, 19.5, 19.5)),

		New("arrow-down").
			Fg(L(12, 4, 12, 15.5), PL(P(7, 10.75), P(12, 15.75), P(17, 10.75)), L(5.5, 20, 18.5, 20)),

		New("arrow-up").
			Fg(L(12, 16, 12, 4.5), PL(P(7, 9.25), P(12, 4.25), P(17, 9.25)), L(5.5, 20, 18.5, 20)),

		tunnel("tunnel-up",
			L(7.25, 15, 16.5, 15), PL(P(13.25, 11.75), P(16.5, 15), P(13.25, 18.25))),

		tunnel("tunnel-down",
			L(9, 12, 15, 18), L(15, 12, 9, 18)),

		New("cpu").
			Duo(RRect(6, 6, 12, 12, 2)).
			Fg(RR(6, 6, 12, 12, 2), RR(9.5, 9.5, 5, 5, 1),
				L(10, 3, 10, 6), L(14, 3, 14, 6),
				L(10, 18, 10, 21), L(14, 18, 14, 21),
				L(3, 10, 6, 10), L(3, 14, 6, 14),
				L(18, 10, 21, 10), L(18, 14, 21, 14)),
	)
}
