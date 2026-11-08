// Package icondsl is a small geometry library for drawing wgAdmin's duotone
// icons as filled SVG paths.
//
// Why not just write SVG by hand: Fyne recolours icons by rewriting fills and
// ignores strokes, and its SVG round-trip drops fill-rule, so an outline has to
// be a filled shape whose inner contour winds the opposite way. Getting that
// right by hand for sixty icons is error-prone; these helpers do it by
// construction.
//
// Conventions (y grows downward, as in SVG):
//   - Every filled shape winds clockwise. Under the default nonzero fill rule,
//     overlapping clockwise shapes simply union, so strokes can cross freely.
//   - Holes wind counter-clockwise and must sit inside a filled shape of the
//     same layer; they cancel only that shape.
//   - Strokes are centred on the geometry you give, SW wide, with round caps
//     and round joins.
//
// Each icon has two layers: Fg (outline/detail, full strength) and Duo (soft
// fill drawn underneath at theme.DuoOpacity). All subpaths of a layer go into
// a single <path>, so overlapping duo shapes never double their opacity.
package icondsl

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// SW is the standard outline weight on the 24px grid.
const SW = 1.75

// Pt is a point.
type Pt struct{ X, Y float64 }

// P is shorthand for Pt{x, y}.
func P(x, y float64) Pt { return Pt{x, y} }

func f(v float64) string {
	s := strconv.FormatFloat(math.Round(v*100)/100, 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}

func pt(p Pt) string { return f(p.X) + " " + f(p.Y) }

// ---------------------------------------------------------------- filled shapes

// Circle is a filled disc.
func Circle(cx, cy, r float64) string { return circle(cx, cy, r, true) }

// CircleHole is a disc-shaped hole.
func CircleHole(cx, cy, r float64) string { return circle(cx, cy, r, false) }

func circle(cx, cy, r float64, cw bool) string {
	s := "1"
	if !cw {
		s = "0"
	}
	return fmt.Sprintf("M%s %sA%s %s 0 1 %s %s %sA%s %s 0 1 %s %s %sZ",
		f(cx-r), f(cy), f(r), f(r), s, f(cx+r), f(cy), f(r), f(r), s, f(cx-r), f(cy))
}

// RRect is a filled rounded rectangle.
func RRect(x, y, w, h, r float64) string { return rrect(x, y, w, h, r, true) }

// RRectHole is a rounded-rectangle hole.
func RRectHole(x, y, w, h, r float64) string { return rrect(x, y, w, h, r, false) }

func rrect(x, y, w, h, r float64, cw bool) string {
	r = math.Max(0, math.Min(r, math.Min(w, h)/2))
	x2, y2 := x+w, y+h
	if r == 0 {
		if cw {
			return fmt.Sprintf("M%s %sH%sV%sH%sZ", f(x), f(y), f(x2), f(y2), f(x))
		}
		return fmt.Sprintf("M%s %sV%sH%sV%sZ", f(x), f(y), f(y2), f(x2), f(y))
	}
	a := func(px, py float64, sweep string) string {
		return fmt.Sprintf("A%s %s 0 0 %s %s %s", f(r), f(r), sweep, f(px), f(py))
	}
	if cw {
		return "M" + f(x+r) + " " + f(y) +
			"H" + f(x2-r) + a(x2, y+r, "1") +
			"V" + f(y2-r) + a(x2-r, y2, "1") +
			"H" + f(x+r) + a(x, y2-r, "1") +
			"V" + f(y+r) + a(x+r, y, "1") + "Z"
	}
	return "M" + f(x+r) + " " + f(y) +
		a(x, y+r, "0") + "V" + f(y2-r) +
		a(x+r, y2, "0") + "H" + f(x2-r) +
		a(x2, y2-r, "0") + "V" + f(y+r) +
		a(x2-r, y, "0") + "Z"
}

// Poly is a filled polygon; its winding is corrected automatically.
func Poly(pts ...Pt) string { return poly(pts, true) }

// PolyHole is a polygonal hole; its winding is corrected automatically.
func PolyHole(pts ...Pt) string { return poly(pts, false) }

func poly(pts []Pt, cw bool) string {
	if len(pts) < 3 {
		return ""
	}
	if (signedArea(pts) > 0) != cw {
		rev := make([]Pt, len(pts))
		for i, p := range pts {
			rev[len(pts)-1-i] = p
		}
		pts = rev
	}
	var b strings.Builder
	b.WriteString("M" + pt(pts[0]))
	for _, p := range pts[1:] {
		b.WriteString("L" + pt(p))
	}
	b.WriteString("Z")
	return b.String()
}

// signedArea is positive for clockwise polygons in y-down coordinates.
func signedArea(pts []Pt) float64 {
	a := 0.0
	for i := range pts {
		j := (i + 1) % len(pts)
		a += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return a
}

// RoundPoly is a filled polygon with rounded corners of radius r, built as the
// union of an inset polygon and circles/capsules along its edges.
func RoundPoly(r float64, pts ...Pt) string {
	return Poly(pts...) + PolyStroke(2*r, pts...)
}

// --------------------------------------------------------------------- strokes

// Line is a straight stroke with round caps.
func Line(a, b Pt, w float64) string {
	dx, dy := b.X-a.X, b.Y-a.Y
	l := math.Hypot(dx, dy)
	r := w / 2
	if l < 1e-6 {
		return Circle(a.X, a.Y, r)
	}
	nx, ny := dy/l*r, -dx/l*r
	A, B := Pt{a.X + nx, a.Y + ny}, Pt{b.X + nx, b.Y + ny}
	C, D := Pt{b.X - nx, b.Y - ny}, Pt{a.X - nx, a.Y - ny}
	return fmt.Sprintf("M%sL%sA%s %s 0 0 1 %sL%sA%s %s 0 0 1 %sZ",
		pt(A), pt(B), f(r), f(r), pt(C), pt(D), f(r), f(r), pt(A))
}

// L is Line at the standard weight.
func L(x1, y1, x2, y2 float64) string { return Line(P(x1, y1), P(x2, y2), SW) }

// Polyline is an open stroke through pts with round joins.
func Polyline(w float64, pts ...Pt) string {
	var b strings.Builder
	for i := 0; i+1 < len(pts); i++ {
		b.WriteString(Line(pts[i], pts[i+1], w))
	}
	return b.String()
}

// PL is Polyline at the standard weight.
func PL(pts ...Pt) string { return Polyline(SW, pts...) }

// PolyStroke is a closed polygon outline with round joins.
func PolyStroke(w float64, pts ...Pt) string {
	if len(pts) == 0 {
		return ""
	}
	return Polyline(w, append(append([]Pt{}, pts...), pts[0])...)
}

// Ring is a circle outline.
func Ring(cx, cy, r, w float64) string {
	return Circle(cx, cy, r+w/2) + CircleHole(cx, cy, math.Max(r-w/2, 0))
}

// R is Ring at the standard weight.
func R(cx, cy, r float64) string { return Ring(cx, cy, r, SW) }

// RRectRing is a rounded-rectangle outline; x, y, w, h, r describe the centre
// line of the stroke.
func RRectRing(x, y, w, h, r, sw float64) string {
	o := sw / 2
	return RRect(x-o, y-o, w+sw, h+sw, r+o) + RRectHole(x+o, y+o, w-sw, h-sw, math.Max(r-o, 0))
}

// RR is RRectRing at the standard weight.
func RR(x, y, w, h, r float64) string { return RRectRing(x, y, w, h, r, SW) }

// Arc is a circular stroke from a0 to a1 degrees (0 = +x, increasing
// clockwise on screen) with round caps.
func Arc(cx, cy, r, a0, a1, w float64) string {
	for a1 < a0 {
		a1 += 360
	}
	if a1-a0 >= 359.99 {
		return Ring(cx, cy, r, w)
	}
	ro, ri, rc := r+w/2, r-w/2, w/2
	at := func(rad, deg float64) Pt {
		t := deg * math.Pi / 180
		return Pt{cx + rad*math.Cos(t), cy + rad*math.Sin(t)}
	}
	large := "0"
	if a1-a0 > 180 {
		large = "1"
	}
	o0, o1, i1, i0 := at(ro, a0), at(ro, a1), at(ri, a1), at(ri, a0)
	return fmt.Sprintf("M%sA%s %s 0 %s 1 %sA%s %s 0 0 1 %sA%s %s 0 %s 0 %sA%s %s 0 0 1 %sZ",
		pt(o0), f(ro), f(ro), large, pt(o1),
		f(rc), f(rc), pt(i1),
		f(ri), f(ri), large, pt(i0),
		f(rc), f(rc), pt(o0))
}

// A is Arc at the standard weight.
func A(cx, cy, r, a0, a1 float64) string { return Arc(cx, cy, r, a0, a1, SW) }

// Quad samples a quadratic Bézier into points, for strokes or fills.
func Quad(p0, c, p1 Pt, n int) []Pt {
	pts := make([]Pt, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		u := 1 - t
		pts = append(pts, Pt{
			u*u*p0.X + 2*u*t*c.X + t*t*p1.X,
			u*u*p0.Y + 2*u*t*c.Y + t*t*p1.Y,
		})
	}
	return pts
}

// Cubic samples a cubic Bézier into points, for strokes or fills.
func Cubic(p0, c1, c2, p1 Pt, n int) []Pt {
	pts := make([]Pt, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		u := 1 - t
		pts = append(pts, Pt{
			u*u*u*p0.X + 3*u*u*t*c1.X + 3*u*t*t*c2.X + t*t*t*p1.X,
			u*u*u*p0.Y + 3*u*u*t*c1.Y + 3*u*t*t*c2.Y + t*t*t*p1.Y,
		})
	}
	return pts
}

// ArcPts samples a circular arc into points, for building custom outlines.
func ArcPts(cx, cy, r, a0, a1 float64, n int) []Pt {
	pts := make([]Pt, 0, n+1)
	for i := 0; i <= n; i++ {
		t := (a0 + (a1-a0)*float64(i)/float64(n)) * math.Pi / 180
		pts = append(pts, Pt{cx + r*math.Cos(t), cy + r*math.Sin(t)})
	}
	return pts
}

// Join concatenates point lists, for composing outlines from segments.
func Join(parts ...[]Pt) []Pt {
	var out []Pt
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// ------------------------------------------------------------------------ icons

// Icon is one glyph under construction.
type Icon struct {
	Name    string
	Size    float64 // viewBox edge; 24 for glyphs, 64 for illustrations
	fg, duo []string
}

// New starts a 24px icon.
func New(name string) *Icon { return &Icon{Name: name, Size: 24} }

// NewIllustration starts a 64px empty-state illustration.
func NewIllustration(name string) *Icon { return &Icon{Name: name, Size: 64} }

// Fg adds outline/detail subpaths.
func (i *Icon) Fg(parts ...string) *Icon { i.fg = append(i.fg, parts...); return i }

// Duo adds soft-fill subpaths.
func (i *Icon) Duo(parts ...string) *Icon { i.duo = append(i.duo, parts...); return i }

// SVG renders the icon file.
func (i *Icon) SVG() string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %s %s">`, f(i.Size), f(i.Size))
	if len(i.duo) > 0 {
		fmt.Fprintf(&b, "\n"+`<path data-layer="duo" d="%s"/>`, strings.Join(i.duo, ""))
	}
	fmt.Fprintf(&b, "\n"+`<path data-layer="fg" d="%s"/>`+"\n</svg>\n", strings.Join(i.fg, ""))
	return b.String()
}

// Main writes icons to -out (default internal/ui/theme/icons, run from the
// repository root). Each batch is its own main package calling this.
func Main(icons ...*Icon) {
	out := flag.String("out", "internal/ui/theme/icons", "output directory")
	flag.Parse()
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	seen := map[string]bool{}
	names := make([]string, 0, len(icons))
	for _, ic := range icons {
		if seen[ic.Name] {
			fmt.Fprintln(os.Stderr, "duplicate icon:", ic.Name)
			os.Exit(1)
		}
		if len(ic.fg) == 0 {
			fmt.Fprintln(os.Stderr, "icon has no fg layer:", ic.Name)
			os.Exit(1)
		}
		seen[ic.Name] = true
		names = append(names, ic.Name)
		if err := os.WriteFile(filepath.Join(*out, ic.Name+".svg"), []byte(ic.SVG()), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	sort.Strings(names)
	fmt.Printf("wrote %d icons: %s\n", len(names), strings.Join(names, ", "))
}
