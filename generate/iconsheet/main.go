// Command iconsheet renders every icon in internal/ui/theme/icons into one PNG
// contact sheet, for reviewing the set by eye.
//
// Each cell shows the icon at 96, 48, 24 and 16px on the dark background, at
// 24px in the accent colour, and at 24 and 16px on the light background. The
// rendering uses the same layer markup as theme.Icon.For and no fill-rule, so
// a hole wound the wrong way shows up here exactly as it would in the app.
//
//	go run ./generate/iconsheet                    # all icons -> iconsheet.png
//	go run ./generate/iconsheet -names copy,qr -o /tmp/x.png
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const duoOpacity = 0.32 // keep in sync with theme.DuoOpacity

var (
	darkBG   = color.NRGBA{0x0b, 0x0e, 0x14, 0xff}
	darkFG   = "#e6edf3"
	accent   = "#00d472"
	lightBG  = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	lightFG  = "#1f2328"
	labelCol = color.NRGBA{0x8b, 0x94, 0x9e, 0xff}
)

type doc struct {
	ViewBox string `xml:"viewBox,attr"`
	Paths   []struct {
		Layer string `xml:"data-layer,attr"`
		D     string `xml:"d,attr"`
	} `xml:"path"`
}

func markup(d doc, fg string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="%s">`, d.ViewBox)
	for _, p := range d.Paths {
		if p.Layer == "duo" {
			fmt.Fprintf(&b, `<path fill="%s" fill-opacity="%g" d="%s"/>`, fg, duoOpacity, p.D)
		}
	}
	for _, p := range d.Paths {
		if p.Layer != "duo" {
			fmt.Fprintf(&b, `<path fill="%s" d="%s"/>`, fg, p.D)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}

func render(dst *image.RGBA, svg string, x, y, size int) error {
	icon, err := oksvg.ReadIconStream(strings.NewReader(svg))
	if err != nil {
		return err
	}
	icon.SetTarget(float64(x), float64(y), float64(size), float64(size))
	scanner := rasterx.NewScannerGV(dst.Bounds().Dx(), dst.Bounds().Dy(), dst, dst.Bounds())
	icon.Draw(rasterx.NewDasher(dst.Bounds().Dx(), dst.Bounds().Dy(), scanner), 1)
	return nil
}

func label(dst *image.RGBA, s string, x, y int) {
	(&font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(labelCol),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}).DrawString(s)
}

func main() {
	dir := flag.String("dir", "internal/ui/theme/icons", "icon directory")
	out := flag.String("o", "iconsheet.png", "output PNG")
	only := flag.String("names", "", "comma-separated icon names (default: all)")
	cols := flag.Int("cols", 4, "cells per row")
	flag.Parse()

	files, _ := filepath.Glob(filepath.Join(*dir, "*.svg"))
	want := map[string]bool{}
	for _, n := range strings.Split(*only, ",") {
		if n = strings.TrimSpace(n); n != "" {
			want[n] = true
		}
	}

	type entry struct {
		name string
		d    doc
	}
	var icons []entry
	for _, fpath := range files {
		name := strings.TrimSuffix(filepath.Base(fpath), ".svg")
		if len(want) > 0 && !want[name] {
			continue
		}
		data, err := os.ReadFile(fpath)
		if err != nil {
			fail(err)
		}
		var d doc
		if err := xml.Unmarshal(data, &d); err != nil {
			fail(fmt.Errorf("%s: %w", name, err))
		}
		icons = append(icons, entry{name, d})
	}
	sort.Slice(icons, func(i, j int) bool { return icons[i].name < icons[j].name })
	if len(icons) == 0 {
		fail(fmt.Errorf("no icons found in %s", *dir))
	}

	const (
		cellW, cellH = 330, 124
		pad          = 10
	)
	rows := (len(icons) + *cols - 1) / *cols
	img := image.NewRGBA(image.Rect(0, 0, cellW**cols, cellH*rows))
	draw.Draw(img, img.Bounds(), image.NewUniform(darkBG), image.Point{}, draw.Src)

	for i, ic := range icons {
		cx, cy := (i%*cols)*cellW, (i / *cols)*cellH

		// Light panel on the right of each cell.
		lightX := cx + 96 + 48 + 24 + 16 + 24 + pad*6
		draw.Draw(img, image.Rect(lightX-pad/2, cy+pad-4, cx+cellW-4, cy+pad+100),
			image.NewUniform(lightBG), image.Point{}, draw.Src)

		x := cx + pad
		for _, s := range []int{96, 48, 24, 16} {
			if err := render(img, markup(ic.d, darkFG), x, cy+pad, s); err != nil {
				fail(fmt.Errorf("%s: %w", ic.name, err))
			}
			x += s + pad
		}
		_ = render(img, markup(ic.d, accent), x, cy+pad, 24)

		x = lightX + pad/2
		for _, s := range []int{24, 16} {
			_ = render(img, markup(ic.d, lightFG), x, cy+pad, s)
			x += s + pad
		}

		label(img, ic.name, cx+pad, cy+cellH-8)
	}

	f, err := os.Create(*out)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fail(err)
	}
	fmt.Printf("rendered %d icons to %s\n", len(icons), *out)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
