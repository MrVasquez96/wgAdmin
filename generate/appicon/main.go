// Command appicon renders the application icon (icon.png, referenced by
// FyneApp.toml) from the brand mark in internal/ui/theme/icons/brand.svg: the
// duotone shield in the accent green on a dark rounded tile with a soft glow.
//
//	go run ./generate/appicon            # writes icon.png
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	accent     = "#00d472"
	duoOpacity = 0.32 // keep in sync with theme.DuoOpacity
)

func main() {
	out := flag.String("o", "icon.png", "output PNG")
	size := flag.Int("size", 512, "edge length in pixels")
	flag.Parse()

	data, err := os.ReadFile("internal/ui/theme/icons/brand.svg")
	if err != nil {
		fail(err)
	}
	var doc struct {
		Paths []struct {
			Layer string `xml:"data-layer,attr"`
			D     string `xml:"d,attr"`
		} `xml:"path"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		fail(err)
	}

	n := *size
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	drawTile(img)

	// The 24px mark at ~62% of the tile, centred. oksvg ignores group
	// transforms, so the scaling comes from SetTarget instead.
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">`)
	for _, p := range doc.Paths {
		if p.Layer == "duo" {
			fmt.Fprintf(&b, `<path fill="%s" fill-opacity="%g" d="%s"/>`, accent, duoOpacity, p.D)
		}
	}
	for _, p := range doc.Paths {
		if p.Layer != "duo" {
			fmt.Fprintf(&b, `<path fill="%s" d="%s"/>`, accent, p.D)
		}
	}
	b.WriteString(`</svg>`)

	icon, err := oksvg.ReadIconStream(strings.NewReader(b.String()))
	if err != nil {
		fail(err)
	}
	mark := float64(n) * 0.62
	off := (float64(n) - mark) / 2
	icon.SetTarget(off, off, mark, mark)
	scanner := rasterx.NewScannerGV(n, n, img, img.Bounds())
	icon.Draw(rasterx.NewDasher(n, n, scanner), 1)

	f, err := os.Create(*out)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fail(err)
	}
	fmt.Println("wrote", *out)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// drawTile paints the rounded dark tile: a vertical gradient with a soft
// accent glow behind the mark, anti-aliased at the corners.
func drawTile(img *image.NRGBA) {
	n := float64(img.Bounds().Dx())
	inset, radius := n*0.03, n*0.22
	top := [3]float64{0x1a, 0x21, 0x30}
	bot := [3]float64{0x0b, 0x0e, 0x14}
	glow := [3]float64{0x00, 0xd4, 0x72}

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			cov := roundedCoverage(fx, fy, inset, n-inset, radius)
			if cov <= 0 {
				continue
			}
			t := (fy - inset) / (n - 2*inset)
			var c [3]float64
			for i := range c {
				c[i] = top[i] + (bot[i]-top[i])*t
			}
			// Radial glow centred slightly above middle.
			dx, dy := (fx-n/2)/(n/2), (fy-n*0.46)/(n/2)
			g := 1 - math.Sqrt(dx*dx+dy*dy)
			if g > 0 {
				g = g * g * 0.34
				for i := range c {
					c[i] += (glow[i] - c[i]) * g
				}
			}
			img.SetNRGBA(x, y, color.NRGBA{uint8(c[0]), uint8(c[1]), uint8(c[2]), uint8(255 * cov)})
		}
	}
}

// roundedCoverage is how much of the pixel at (x, y) lies inside the rounded
// square spanning lo..hi with corner radius r, for smooth edges.
func roundedCoverage(x, y, lo, hi, r float64) float64 {
	cx := math.Max(lo+r, math.Min(x, hi-r))
	cy := math.Max(lo+r, math.Min(y, hi-r))
	d := math.Hypot(x-cx, y-cy) - r
	if x < lo || x > hi || y < lo || y > hi {
		return 0
	}
	return math.Max(0, math.Min(1, 0.5-d))
}
