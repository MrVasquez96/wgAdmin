// Command core draws the reference icons that set the style for the set.
//
//	go run ./generate/icons/core
package main

import . "wgAdmin/generate/icondsl"

func main() {
	Main(
		// Two sheets: the front one carries the soft fill, the back one is only
		// the visible top-left corner, so the pair never reads as one blob.
		New("copy").
			Duo(RRect(8, 8, 12, 12, 2.5)).
			Fg(RR(8, 8, 12, 12, 2.5),
				L(4, 13.5, 4, 6.5), A(6.5, 6.5, 2.5, 180, 270), L(6.5, 4, 13.5, 4)),

		New("info").
			Duo(Circle(12, 12, 9)).
			Fg(R(12, 12, 9), L(12, 11, 12, 16.5), Circle(12, 7.75, 1.2)),

		New("play").
			Duo(Poly(P(8, 5.5), P(18.5, 12), P(8, 18.5))).
			Fg(PolyStroke(SW, P(8, 5.5), P(18.5, 12), P(8, 18.5))),

		New("check").
			Fg(PL(P(5, 12.5), P(10, 17.25), P(19, 7))),
	)
}
