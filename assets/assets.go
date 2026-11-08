// Package assets embeds the binary resources the UI ships with.
//
// Only the faces the theme actually uses are embedded. assets/fonts holds the
// full Inter release for reference, and anything not listed here costs nothing
// in the binary.
package assets

import _ "embed"

var (
	//go:embed fonts/Inter-Regular.ttf
	InterRegular []byte
	//go:embed fonts/Inter-Italic.ttf
	InterItalic []byte
	//go:embed fonts/Inter-Medium.ttf
	InterMedium []byte
	//go:embed fonts/Inter-SemiBold.ttf
	InterSemiBold []byte
	//go:embed fonts/Inter-Bold.ttf
	InterBold []byte
	//go:embed fonts/Inter-BoldItalic.ttf
	InterBoldItalic []byte
	//go:embed fonts/JetBrainsMono-Regular.ttf
	JetBrainsMono []byte
)
