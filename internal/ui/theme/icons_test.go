package theme

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestIconsAreWellFormedSVG guards against a typo in the path data breaking an
// icon: Fyne swallows malformed SVG and renders nothing at all.
func TestIconsAreWellFormedSVG(t *testing.T) {
	if len(iconSources) == 0 {
		t.Fatal("no icons registered")
	}
	for name, markup := range iconSources {
		var doc struct {
			XMLName xml.Name
			ViewBox string `xml:"viewBox,attr"`
			Paths   []struct {
				D    string `xml:"d,attr"`
				Fill string `xml:"fill,attr"`
			} `xml:"path"`
		}
		if err := xml.Unmarshal([]byte(markup), &doc); err != nil {
			t.Errorf("icon %q is not well-formed XML: %v", name, err)
			continue
		}
		if doc.XMLName.Local != "svg" {
			t.Errorf("icon %q root element is %q, want svg", name, doc.XMLName.Local)
		}
		if doc.ViewBox != "0 0 24 24" {
			t.Errorf("icon %q viewBox is %q, want 0 0 24 24", name, doc.ViewBox)
		}
		if len(doc.Paths) == 0 {
			t.Errorf("icon %q has no path elements", name)
		}
		for i, p := range doc.Paths {
			if strings.TrimSpace(p.D) == "" {
				t.Errorf("icon %q path %d has empty path data", name, i)
			}
			if p.Fill == "" {
				t.Errorf("icon %q path %d has no fill; it will not follow the theme", name, i)
			}
		}
	}
}

// TestIconsUseFillNotStroke enforces the constraint that makes theming work:
// Fyne's SVG colouriser rewrites `fill` and ignores `stroke`, so a stroke-drawn
// icon would keep its baked-in colour and ignore both the theme and button
// importance.
func TestIconsUseFillNotStroke(t *testing.T) {
	for name, markup := range iconSources {
		if strings.Contains(markup, "stroke=") {
			t.Errorf("icon %q uses stroke, which Fyne cannot recolour; convert it to filled paths", name)
		}
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
