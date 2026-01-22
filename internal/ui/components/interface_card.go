package components

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	customTheme "wgAdmin/internal/ui/theme"
)

// InterfaceCardCallbacks holds callback functions for interface card actions
type InterfaceCardCallbacks struct {
	OnToggle func(name string, activate bool)
	OnScan   func(name, ip string)
	OnEdit   func(name string)
	OnDelete func(name string)
}

// InterfaceCard represents a card widget for a WireGuard interface
type InterfaceCard struct {
	widget.BaseWidget

	iface     models.Interface
	callbacks InterfaceCardCallbacks
	container *fyne.Container
}

// NewInterfaceCard creates a new interface card
func NewInterfaceCard(iface models.Interface, callbacks InterfaceCardCallbacks) *InterfaceCard {
	card := &InterfaceCard{
		iface:     iface,
		callbacks: callbacks,
	}
	card.ExtendBaseWidget(card)
	card.container = card.buildCard()
	return card
}

func (c *InterfaceCard) buildCard() *fyne.Container {
	variant := fyne.CurrentApp().Settings().ThemeVariant()

	// Title
	title := canvas.NewText(c.iface.Name, theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 18

	// Status badge
	var statusColor color.Color
	var statusText string
	if c.iface.Active {
		statusColor = customTheme.AppColors.Active(variant)
		statusText = "Active"
	} else {
		statusColor = customTheme.AppColors.Inactive(variant)
		statusText = "Inactive"
	}

	statusDot := canvas.NewCircle(statusColor)
	//statusDot.SetMinSize(fyne.NewSize(10, 10))
	statusLabel := widget.NewLabel(statusText)
	statusBadge := container.NewHBox(statusDot, statusLabel)

	// IP address
	ipLabel := widget.NewLabel(fmt.Sprintf("IP: %s", c.iface.IP))
	ipLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Action buttons
	var toggleBtn *widget.Button
	if c.iface.Active {
		toggleBtn = widget.NewButtonWithIcon("Deactivate", theme.MediaStopIcon(), func() {
			if c.callbacks.OnToggle != nil {
				c.callbacks.OnToggle(c.iface.Name, false)
			}
		})
		toggleBtn.Importance = widget.DangerImportance
	} else {
		toggleBtn = widget.NewButtonWithIcon("Activate", theme.MediaPlayIcon(), func() {
			if c.callbacks.OnToggle != nil {
				c.callbacks.OnToggle(c.iface.Name, true)
			}
		})
		toggleBtn.Importance = widget.SuccessImportance
	}

	scanBtn := widget.NewButtonWithIcon("Scan", theme.SearchIcon(), func() {
		if c.callbacks.OnScan != nil {
			c.callbacks.OnScan(c.iface.Name, c.iface.IP)
		}
	})
	if c.iface.IP == "unknown" {
		scanBtn.Disable()
	}

	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if c.callbacks.OnEdit != nil {
			c.callbacks.OnEdit(c.iface.Name)
		}
	})

	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if c.callbacks.OnDelete != nil {
			c.callbacks.OnDelete(c.iface.Name)
		}
	})
	deleteBtn.Importance = widget.DangerImportance

	// Layout
	statusContent := container.NewHBox(
		statusBadge,
		ipLabel,
	)
	leftContent := container.NewVBox(
		title,
		statusContent,
	)

	rightContent := container.NewHBox(
		layout.NewSpacer(),
		toggleBtn,
		scanBtn,
		editBtn,
		deleteBtn,
	)

	cardContent := container.NewBorder(
		nil, nil,
		leftContent,
		rightContent,
	)

	// Card background
	var bgColor color.Color
	if c.iface.Active {
		bgColor = customTheme.AppColors.CardActiveBackground(variant)
	} else {
		bgColor = customTheme.AppColors.CardBackground(variant)
	}

	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 12
	bg.StrokeColor = customTheme.AppColors.Border(variant)
	bg.StrokeWidth = 1

	padded := container.NewPadded(cardContent)
	return container.NewStack(bg, padded)
}

// CreateRenderer implements fyne.Widget
func (c *InterfaceCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.container)
}

// MinSize returns the minimum size of the card
func (c *InterfaceCard) MinSize() fyne.Size {
	return fyne.NewSize(800, 100)
}
