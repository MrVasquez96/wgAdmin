package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// WGAdminTheme is a modern custom theme for the WireGuard Admin app
type WGAdminTheme struct{}

var _ fyne.Theme = (*WGAdminTheme)(nil)

// NewWGAdminTheme creates a new WGAdminTheme
func NewWGAdminTheme() *WGAdminTheme {
	return &WGAdminTheme{}
}

// Color returns the color for the specified name
func (t *WGAdminTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 248, G: 249, B: 250, A: 255} // #f8f9fa
		}
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // #121212

	case theme.ColorNameButton:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 26, G: 115, B: 232, A: 255} // #1a73e8
		}
		return color.NRGBA{R: 138, G: 180, B: 248, A: 255} // #8ab4f8

	case theme.ColorNameDisabledButton:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 189, G: 189, B: 189, A: 255}
		}
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255}

	case theme.ColorNameDisabled:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 158, G: 158, B: 158, A: 255}
		}
		return color.NRGBA{R: 117, G: 117, B: 117, A: 255}

	case theme.ColorNameError:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 234, G: 67, B: 53, A: 255} // #ea4335
		}
		return color.NRGBA{R: 242, G: 139, B: 130, A: 255} // #f28b82

	case theme.ColorNameFocus:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 26, G: 115, B: 232, A: 255} // #1a73e8
		}
		return color.NRGBA{R: 138, G: 180, B: 248, A: 255} // #8ab4f8

	case theme.ColorNameForeground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 32, G: 33, B: 36, A: 255} // #202124
		}
		return color.NRGBA{R: 232, G: 234, B: 237, A: 255} // #e8eaed

	case theme.ColorNameHover:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 232, G: 240, B: 254, A: 255}
		}
		return color.NRGBA{R: 48, G: 48, B: 48, A: 255}

	case theme.ColorNameInputBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}

	case theme.ColorNameInputBorder:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
		}
		return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043

	case theme.ColorNameMenuBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	case theme.ColorNameOverlayBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255}

	case theme.ColorNamePlaceHolder:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 95, G: 99, B: 104, A: 255} // #5f6368
		}
		return color.NRGBA{R: 154, G: 160, B: 166, A: 255} // #9aa0a6

	case theme.ColorNamePressed:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 210, G: 227, B: 252, A: 255}
		}
		return color.NRGBA{R: 60, G: 60, B: 60, A: 255}

	case theme.ColorNamePrimary:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 26, G: 115, B: 232, A: 255} // #1a73e8
		}
		return color.NRGBA{R: 138, G: 180, B: 248, A: 255} // #8ab4f8

	case theme.ColorNameScrollBar:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 189, G: 189, B: 189, A: 255}
		}
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255}

	case theme.ColorNameSelection:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 26, G: 115, B: 232, A: 60}
		}
		return color.NRGBA{R: 138, G: 180, B: 248, A: 60}

	case theme.ColorNameSeparator:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
		}
		return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043

	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 40}

	case theme.ColorNameSuccess:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 52, G: 168, B: 83, A: 255} // #34a853
		}
		return color.NRGBA{R: 129, G: 201, B: 149, A: 255} // #81c995

	case theme.ColorNameWarning:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 251, G: 188, B: 5, A: 255} // #fbbc05
		}
		return color.NRGBA{R: 253, G: 214, B: 99, A: 255}

	case theme.ColorNameHeaderBackground:
		if variant == theme.VariantLight {
			return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	}

	return theme.DefaultTheme().Color(name, variant)
}

// Font returns the font for the specified style
func (t *WGAdminTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon for the specified name
func (t *WGAdminTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified name
func (t *WGAdminTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 18
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 4
	}
	return theme.DefaultTheme().Size(name)
}

// Colors returns pre-defined colors for use in the app
type Colors struct{}

// Active returns the active/success color
func (c Colors) Active(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 52, G: 168, B: 83, A: 255} // #34a853
	}
	return color.NRGBA{R: 129, G: 201, B: 149, A: 255} // #81c995
}

// Inactive returns the inactive/error color
func (c Colors) Inactive(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 234, G: 67, B: 53, A: 255} // #ea4335
	}
	return color.NRGBA{R: 242, G: 139, B: 130, A: 255} // #f28b82
}

// CardBackground returns the card background color
func (c Colors) CardBackground(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	return color.NRGBA{R: 30, G: 30, B: 30, A: 255} // #1e1e1e
}

// CardActiveBackground returns the background for active cards
func (c Colors) CardActiveBackground(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 232, G: 245, B: 233, A: 255} // light green tint
	}
	return color.NRGBA{R: 30, G: 50, B: 35, A: 255} // dark green tint
}

// Border returns the border color
func (c Colors) Border(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.NRGBA{R: 218, G: 220, B: 224, A: 255} // #dadce0
	}
	return color.NRGBA{R: 60, G: 64, B: 67, A: 255} // #3c4043
}

// AppColors provides easy access to theme colors
var AppColors = Colors{}
