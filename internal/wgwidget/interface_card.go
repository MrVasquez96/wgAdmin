package wgwidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/MrVasquez96/go-wg/wg/config"

	"wgAdmin/internal/ui/lang"
	customTheme "wgAdmin/internal/ui/theme"
)

// InterfaceCardCallbacks holds callback functions for interface card actions
type InterfaceCardCallbacks struct {
	OnToggle     func(name string, activate bool)
	OnScan       func(name, ip string)
	OnEdit       func(name string)
	OnPeers      func(name string)
	OnDelete     func(name string)
	OnCopyPubKey func(pubKey string)
}

// InterfaceCard represents a card widget for a WireGuard interface
type InterfaceCard struct {
	widget.BaseWidget

	iface     config.Interface
	callbacks InterfaceCardCallbacks
	lang      *lang.Language
	container *fyne.Container
}

// NewInterfaceCard creates a new interface card
func NewInterfaceCard(iface config.Interface, callbacks InterfaceCardCallbacks, language *lang.Language) *InterfaceCard {
	card := &InterfaceCard{
		iface:     iface,
		callbacks: callbacks,
		lang:      language,
	}
	card.ExtendBaseWidget(card)
	card.container = card.buildCard()
	return card
}

func (c *InterfaceCard) buildCard() *fyne.Container {
	variant := customTheme.CurrentVariant()

	// Title
	var titleColor color.Color
	if c.iface.Active {
		titleColor = theme.PrimaryColor()
	} else {
		titleColor = theme.ForegroundColor()
	}
	title := canvas.NewText(c.iface.Name, titleColor)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 20

	// Status badge
	var statusColor color.Color
	var statusText string
	if c.iface.Active {
		statusColor = customTheme.AppColors.Active(variant)
		statusText = c.lang.T(lang.Card_active)
	} else {
		statusColor = customTheme.AppColors.Inactive(variant)
		statusText = c.lang.T(lang.Card_inactive)
	}

	statusDot := canvas.NewCircle(statusColor)
	statusLabel := widget.NewLabel(statusText)
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	// IP address
	ipLabel := widget.NewLabel(c.lang.Tf(lang.Card_ip_label, c.iface.IP))
	ipLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Public key display with copy button
	var pubKeyRow *fyne.Container
	if c.iface.PublicKey != "" {
		displayKey := c.iface.PublicKey
		if len(displayKey) > 20 {
			displayKey = displayKey[:20] + "..."
		}
		pubKeyLabel := widget.NewLabel(c.lang.Tf(lang.Card_public_key_label, displayKey))
		pubKeyLabel.TextStyle = fyne.TextStyle{Monospace: true}

		copyKeyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
			if c.callbacks.OnCopyPubKey != nil {
				c.callbacks.OnCopyPubKey(c.iface.PublicKey)
			}
		})

		pubKeyRow = container.NewHBox(pubKeyLabel, copyKeyBtn)
	}

	// Action buttons
	var toggleBtn *widget.Button
	if c.iface.Active {
		toggleBtn = widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_deactivate), theme.MediaStopIcon(), func() {
			if c.callbacks.OnToggle != nil {
				c.callbacks.OnToggle(c.iface.Name, false)
			}
		})
		toggleBtn.Importance = widget.DangerImportance
	} else {
		toggleBtn = widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_activate), theme.MediaPlayIcon(), func() {
			if c.callbacks.OnToggle != nil {
				c.callbacks.OnToggle(c.iface.Name, true)
			}
		})
		toggleBtn.Importance = widget.SuccessImportance
	}

	scanBtn := widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_scan), theme.SearchIcon(), func() {
		if c.callbacks.OnScan != nil {
			c.callbacks.OnScan(c.iface.Name, c.iface.IP)
		}
	})
	if c.iface.IP == "unknown" {
		scanBtn.Disable()
	}

	editBtn := widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_edit), theme.DocumentCreateIcon(), func() {
		if c.callbacks.OnEdit != nil {
			c.callbacks.OnEdit(c.iface.Name)
		}
	})
	editPeersBtn := widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_peers), theme.DocumentCreateIcon(), func() {
		if c.callbacks.OnPeers != nil {
			c.callbacks.OnPeers(c.iface.Name)
		}
	})

	deleteBtn := widget.NewButtonWithIcon(c.lang.T(lang.Card_btn_delete), theme.DeleteIcon(), func() {
		if c.callbacks.OnDelete != nil {
			c.callbacks.OnDelete(c.iface.Name)
		}
	})
	deleteBtn.Importance = widget.DangerImportance

	// Layout: left info column, right button column (3 rows)
	statusBadge := container.NewHBox(statusDot, statusLabel)
	infoCol := container.NewVBox(
		container.NewHBox(statusBadge, title),
		ipLabel,
	)
	if pubKeyRow != nil {
		infoCol.Add(pubKeyRow)
	}

	scanPeersRow := container.NewHBox(scanBtn, editPeersBtn)
	editDeleteRow := container.NewHBox(editBtn, deleteBtn)
	btnCol := container.NewVBox(toggleBtn, scanPeersRow, editDeleteRow)

	cardContent := container.NewBorder(nil, nil, infoCol, btnCol)

	// Card background
	var bgColor color.Color
	if c.iface.Active {
		bgColor = customTheme.AppColors.CardActiveBackground(variant)
	} else {
		bgColor = customTheme.AppColors.CardBackground(variant)
	}

	// Shadow layer for depth
	shadow := canvas.NewRectangle(customTheme.AppColors.CardElevation(variant))
	shadow.CornerRadius = 16

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 16
	bg.StrokeColor = customTheme.AppColors.Border(variant)
	bg.StrokeWidth = 1.5

	padded := container.NewPadded(cardContent)

	// Active tunnel: neon accent bar on the left edge
	if c.iface.Active {
		accentBar := canvas.NewRectangle(customTheme.AppColors.Active(variant))
		accentBar.CornerRadius = 2
		accentBarContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(3, 0)), accentBar)
		inner := container.NewBorder(nil, nil, accentBarContainer, nil, padded)
		return container.NewStack(shadow, bg, inner)
	}
	return container.NewStack(shadow, bg, padded)
}

// CreateRenderer implements fyne.Widget
func (c *InterfaceCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

// MinSize returns the minimum size of the card
func (c *InterfaceCard) MinSize() fyne.Size {
	contentMin := c.container.MinSize()
	return fyne.NewSize(800, contentMin.Height)
}
