package views

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	"wgAdmin/internal/wireguard"
)

// TunnelForm handles tunnel creation and editing
type TunnelForm struct {
	window fyne.Window
	isEdit bool
	name   string

	// Interface fields
	nameEntry       *widget.Entry
	privateKeyEntry *widget.Entry
	publicKeyLabel  *widget.Label
	addressEntry    *widget.Entry
	dnsEntry        *widget.Entry
	listenPortEntry *widget.Entry
	mtuEntry        *widget.Entry

	// Peers
	peers     []models.Peer
	peersList *widget.List

	// Callbacks
	onSave   func(name string, config *models.WireGuardConfig) error
	onCancel func()
}

// NewTunnelForm creates a new tunnel form
func NewTunnelForm(parent fyne.Window, existingName string, existingConfig *models.WireGuardConfig, onSave func(string, *models.WireGuardConfig) error, onCancel func()) *TunnelForm {
	tunnelName := existingName
	isEdit := existingName != ""
	if isEdit {
		tunnelName = existingConfig.Name
	}

	f := &TunnelForm{
		window:          parent,
		isEdit:          existingName != "",
		name:            tunnelName,
		nameEntry:       widget.NewEntry(),
		privateKeyEntry: widget.NewEntry(),
		publicKeyLabel:  widget.NewLabel(""),
		addressEntry:    widget.NewEntry(),
		dnsEntry:        widget.NewEntry(),
		listenPortEntry: widget.NewEntry(),
		mtuEntry:        widget.NewEntry(),
		peers:           []models.Peer{},
		onSave:          onSave,
		onCancel:        onCancel,
	}

	f.nameEntry.SetPlaceHolder("e.g., wg0")
	f.privateKeyEntry.SetPlaceHolder("Base64 encoded private key")
	f.addressEntry.SetPlaceHolder("e.g., 10.0.0.2/32")
	f.dnsEntry.SetPlaceHolder("e.g., 1.1.1.1 (optional)")
	f.listenPortEntry.SetPlaceHolder("e.g., 51820 (optional)")
	f.mtuEntry.SetPlaceHolder("e.g., 1420 (optional)")

	if existingConfig != nil {
		f.nameEntry.SetText(existingName)
		if isEdit {
			f.nameEntry.Disable()
		} else {
			f.nameEntry.Enable()

		}
		f.privateKeyEntry.SetText(existingConfig.PrivateKey)
		f.addressEntry.SetText(existingConfig.Address)
		f.dnsEntry.SetText(existingConfig.DNS)
		if existingConfig.ListenPort > 0 {
			f.listenPortEntry.SetText(strconv.Itoa(existingConfig.ListenPort))
		}
		if existingConfig.MTU > 0 {
			f.mtuEntry.SetText(strconv.Itoa(existingConfig.MTU))
		}
		f.peers = existingConfig.Peers
		f.updatePublicKey()
	}

	// Update public key when private key changes
	f.privateKeyEntry.OnChanged = func(s string) {
		f.updatePublicKey()
	}

	return f
}

func (f *TunnelForm) updatePublicKey() {
	if f.privateKeyEntry.Text == "" {
		f.publicKeyLabel.SetText("")
		return
	}
	pubKey, err := wireguard.DerivePublicKey(f.privateKeyEntry.Text)
	if err != nil {
		f.publicKeyLabel.SetText("(invalid key)")
	} else {
		f.publicKeyLabel.SetText(pubKey)
	}
}

// Show displays the tunnel form
func (f *TunnelForm) Show() {
	title := "Add Tunnel"
	if f.isEdit {
		title = "Edit Tunnel: " + f.name
	}

	win := fyne.CurrentApp().NewWindow(title)
	win.Resize(fyne.NewSize(600, 700))

	// Interface section
	generateKeyBtn := widget.NewButtonWithIcon("Generate Keys", theme.ViewRefreshIcon(), func() {
		priv, pub, err := wireguard.GenerateKeyPair()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		f.privateKeyEntry.SetText(priv)
		f.publicKeyLabel.SetText(pub)
	})

	interfaceForm := container.NewVBox(
		widget.NewLabel("Tunnel Name *"),
		f.nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Private Key *"),
		container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry),
		widget.NewLabel("Public Key (derived):"),
		f.publicKeyLabel,
		widget.NewSeparator(),
		widget.NewLabel("Address (CIDR) *"),
		f.addressEntry,
		widget.NewLabel("DNS"),
		f.dnsEntry,
		widget.NewLabel("Listen Port"),
		f.listenPortEntry,
		widget.NewLabel("MTU"),
		f.mtuEntry,
	)

	// Peers section
	f.peersList = widget.NewList(
		func() int { return len(f.peers) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Peer"),
				layout.NewSpacer(),
				widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), nil),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := obj.(*fyne.Container)
			label := c.Objects[0].(*widget.Label)
			editBtn := c.Objects[2].(*widget.Button)
			deleteBtn := c.Objects[3].(*widget.Button)

			peer := f.peers[id]
			endpoint := peer.Endpoint
			if endpoint == "" {
				endpoint = "(no endpoint)"
			}
			label.SetText(fmt.Sprintf("%s... - %s", peer.PublicKey[:12], peer.Name))

			editBtn.OnTapped = func() {
				peerCopy := f.peers[id]
				peerForm := NewPeerForm(&peerCopy, func(p models.Peer) {
					f.peers[id] = p
					f.peersList.Refresh()
				}, nil)
				peerForm.Show(win)
			}

			deleteBtn.OnTapped = func() {
				dialog.ShowConfirm("Delete Peer", "Remove this peer?", func(yes bool) {
					if yes {
						f.peers = append(f.peers[:id], f.peers[id+1:]...)

						f.peersList.Refresh()
					}
				}, win)
			}
		},
	)
	//f.peersList.Resize(fyne.NewSize(560, float32(f.peersList.Length()*100)))
	// f.peersList.Resize(fyne.NewSize(560, 150))
	addPeerBtn := widget.NewButtonWithIcon("Add Peer", theme.ContentAddIcon(), func() {
		peerForm := NewPeerForm(nil, func(p models.Peer) {
			f.peers = append(f.peers, p)
			f.peersList.Refresh()
		}, nil)
		peerForm.Show(win)
	})
	sizedList := container.NewGridWrap(
		fyne.NewSize(560, float32(f.peersList.Length()*50)),
		f.peersList,
	)

	peersSection := container.NewBorder(
		widget.NewLabel("Peers"),
		addPeerBtn,
		nil, nil,
		sizedList, // Use the sized wrapper here
	)
	// Action buttons
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		config, errs := f.validate()
		if len(errs) > 0 {
			dialog.ShowError(errs[0], win)
			return
		}

		name := f.nameEntry.Text
		if f.isEdit {
			name = f.name
		}

		if err := f.onSave(name, config); err != nil {
			dialog.ShowError(err, win)
			return
		}

		win.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		if f.onCancel != nil {
			f.onCancel()
		}
		win.Close()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)

	// Main layout
	content := container.NewBorder(
		nil,
		buttons,
		nil, nil,
		container.NewVScroll(container.NewVBox(
			widget.NewCard("Interface Configuration", "", interfaceForm),
			widget.NewCard("Peers", "", peersSection),
		)),
	)

	win.SetContent(container.NewPadded(content))
	win.Show()
}

func (f *TunnelForm) validate() (*models.WireGuardConfig, []error) {
	config := &models.WireGuardConfig{
		PrivateKey: f.privateKeyEntry.Text,
		Address:    f.addressEntry.Text,
		DNS:        f.dnsEntry.Text,
		Peers:      f.peers,
	}

	if f.listenPortEntry.Text != "" {
		port, err := strconv.Atoi(f.listenPortEntry.Text)
		if err != nil {
			return nil, []error{wireguard.ValidationError{Field: "ListenPort", Message: "must be a number"}}
		}
		config.ListenPort = port
	}

	if f.mtuEntry.Text != "" {
		mtu, err := strconv.Atoi(f.mtuEntry.Text)
		if err != nil {
			return nil, []error{wireguard.ValidationError{Field: "MTU", Message: "must be a number"}}
		}
		config.MTU = mtu
	}

	// Validate name for new tunnels
	if !f.isEdit {
		if !wireguard.ValidateName(f.nameEntry.Text) {
			return nil, []error{wireguard.ValidationError{Field: "Name", Message: "must be 1-15 alphanumeric characters"}}
		}
		if wireguard.ConfigExists(f.nameEntry.Text) {
			return nil, []error{wireguard.ValidationError{Field: "Name", Message: "tunnel already exists"}}
		}
	}

	errs := wireguard.ValidateConfig(config)
	return config, errs
}
