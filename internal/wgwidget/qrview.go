package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	qrcode "github.com/skip2/go-qrcode"
)

// QRSize is the rendered edge length of a QR code, large enough for a phone
// camera to lock on from a comfortable distance.
const QRSize = 260

// NewQRCode renders content as a scannable QR image.
//
// The code is always drawn black-on-white regardless of theme: scanners rely on
// that contrast and polarity, and an inverted or tinted code fails to scan on
// many phones. The white quiet zone around it is part of the specification, not
// decoration.
func NewQRCode(content string) (fyne.CanvasObject, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = false

	img := canvas.NewImageFromImage(qr.Image(QRSize))
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScalePixels // keep the modules crisp
	img.SetMinSize(fyne.NewSize(QRSize, QRSize))

	// A white plate guarantees the quiet zone even on a dark background.
	plate := canvas.NewRectangle(color.White)
	plate.CornerRadius = 4

	return container.NewStack(plate,
		container.New(layout.NewCustomPaddedLayout(8, 8, 8, 8), img),
	), nil
}
