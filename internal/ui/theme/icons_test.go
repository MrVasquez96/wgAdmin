package theme

import (
	"encoding/xml"
	"image/color"
	"strings"
	"testing"
)

// TestIconFilesAreValid checks every generated SVG: well-formed, on the
// expected grid, with a non-empty fg layer and nothing Fyne would mangle.
// Fyne swallows malformed SVG and renders nothing at all, so this is the only
// place a broken file would be noticed.
func TestIconFilesAreValid(t *testing.T) {
	files, err := iconFS.ReadDir("icons")
	if err != nil || len(files) == 0 {
		t.Fatalf("no icon files: %v", err)
	}
	for _, f := range files {
		name := strings.TrimSuffix(f.Name(), ".svg")
		data, err := iconFS.ReadFile("icons/" + f.Name())
		if err != nil {
			t.Fatal(err)
		}
		var doc iconDoc
		if err := xml.Unmarshal(data, &doc); err != nil {
			t.Errorf("icon %q is not well-formed XML: %v", name, err)
			continue
		}
		want := "0 0 24 24"
		if strings.HasPrefix(name, "empty-") {
			want = "0 0 64 64"
		}
		if doc.ViewBox != want {
			t.Errorf("icon %q viewBox is %q, want %q", name, doc.ViewBox, want)
		}
		fg := false
		for _, p := range doc.Paths {
			if p.Layer != "fg" && p.Layer != "duo" {
				t.Errorf("icon %q has a path with layer %q, want fg or duo", name, p.Layer)
			}
			if strings.TrimSpace(p.D) == "" {
				t.Errorf("icon %q has an empty %s path", name, p.Layer)
			}
			fg = fg || p.Layer == "fg"
		}
		if !fg {
			t.Errorf("icon %q has no fg layer", name)
		}
		// Fyne recolours fills only and drops fill-rule on its round trip, so
		// either attribute would make the icon render differently in stock
		// widgets than in the icon sheet.
		for _, bad := range []string{"stroke=", "fill-rule=", "fill="} {
			if strings.Contains(string(data), bad) {
				t.Errorf("icon %q contains %q; colours and winding belong to the generator", name, bad)
			}
		}
	}
}

// TestThemedViewOmitsDuoLayer guards the reason Icon exists: Fyne's recolouring
// would flatten the soft layer into a solid shape, so the themed view handed to
// stock widgets must carry the outline only.
func TestThemedViewOmitsDuoLayer(t *testing.T) {
	for name, ic := range Icons() {
		if strings.Contains(string(ic.Content()), "fill-opacity") {
			t.Errorf("icon %q themed content includes the duo layer", name)
		}
	}
}

// TestIconColourCache checks that coloured renders are cached per colour and
// that their resource names differ, since Fyne caches rasters by name.
func TestIconColourCache(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}

	a, b := IconCopy.For(red), IconCopy.For(red)
	if a != b {
		t.Error("same colour produced two resources")
	}
	c := IconCopy.For(blue)
	if a.Name() == c.Name() || a.Name() == IconCopy.Name() {
		t.Errorf("resource names collide: %q %q %q", a.Name(), c.Name(), IconCopy.Name())
	}
	if !strings.Contains(string(a.Content()), "#ff0000") {
		t.Error("coloured render does not use the requested colour")
	}
}

// TestIconOverridesResolve checks the Fyne-name -> app-icon mapping is intact,
// since these are what existing theme.XxxIcon() call sites resolve through.
func TestIconOverridesResolve(t *testing.T) {
	if len(iconOverrides) == 0 {
		t.Fatal("no icon overrides registered")
	}
	for name, res := range iconOverrides {
		if res == nil {
			t.Errorf("override for %q is nil", name)
			continue
		}
		got, ok := overrideIcon(name)
		if !ok || got != res {
			t.Errorf("overrideIcon(%q) did not return the registered resource", name)
		}
	}
}

// TestEveryIconFileIsRegistered catches an icon that was generated but never
// declared in icons.go, which would silently never be used.
func TestEveryIconFileIsRegistered(t *testing.T) {
	files, _ := iconFS.ReadDir("icons")
	for _, f := range files {
		name := strings.TrimSuffix(f.Name(), ".svg")
		if _, ok := Icons()[name]; !ok {
			t.Errorf("icons/%s has no declaration in icons.go", f.Name())
		}
	}
}
