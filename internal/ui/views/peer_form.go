package views

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	"wgAdmin/internal/wireguard"
)

// PeerForm handles peer configuration
type PeerForm struct {
	nameLabel                *widget.Label
	publicKeyEntry           *widget.Entry
	allowedIPsEntry          *widget.Entry
	endpointEntry            *widget.Entry
	persistentKeepaliveEntry *widget.Entry
	presharedKeyEntry        *widget.Entry

	onSave   func(peer models.Peer)
	onCancel func()
}

// NewPeerForm creates a new peer form
func NewPeerForm(existing *models.Peer, onSave func(models.Peer), onCancel func()) *PeerForm {
	f := &PeerForm{
		nameLabel:                widget.NewLabel(existing.Name),
		publicKeyEntry:           widget.NewEntry(),
		allowedIPsEntry:          widget.NewEntry(),
		endpointEntry:            widget.NewEntry(),
		persistentKeepaliveEntry: widget.NewEntry(),
		presharedKeyEntry:        widget.NewEntry(),
		onSave:                   onSave,
		onCancel:                 onCancel,
	}

	f.publicKeyEntry.SetPlaceHolder("Base64 encoded public key")
	f.allowedIPsEntry.SetPlaceHolder("e.g., 10.0.0.0/24, 192.168.1.0/24")
	f.endpointEntry.SetPlaceHolder("e.g., vpn.example.com:51820 (optional)")
	f.persistentKeepaliveEntry.SetPlaceHolder("e.g., 25 (seconds, optional)")
	f.presharedKeyEntry.SetPlaceHolder("Base64 encoded key (optional)")

	if existing != nil {
		f.publicKeyEntry.SetText(existing.PublicKey)
		f.allowedIPsEntry.SetText(existing.AllowedIPs)
		f.endpointEntry.SetText(existing.Endpoint)
		if existing.PersistentKeepalive > 0 {
			f.persistentKeepaliveEntry.SetText(strconv.Itoa(existing.PersistentKeepalive))
		}
		f.presharedKeyEntry.SetText(existing.PresharedKey)
	}

	return f
}

// Show displays the peer form dialog
func (f *PeerForm) Show(parent fyne.Window) {
	form := container.NewVBox(
		widget.NewLabel("Public Key *"),
		f.publicKeyEntry,

		widget.NewLabel("Allowed IPs *"),
		f.allowedIPsEntry,

		widget.NewLabel("Endpoint"),
		f.endpointEntry,

		widget.NewLabel("Persistent Keepalive"),
		f.persistentKeepaliveEntry,

		widget.NewLabel("Preshared Key"),
		f.presharedKeyEntry,
	)

	d := dialog.NewCustomConfirm("Peer Configuration", "Save", "Cancel", form, func(confirmed bool) {
		if !confirmed {
			if f.onCancel != nil {
				f.onCancel()
			}
			return
		}

		peer, errs := f.validate()
		if len(errs) > 0 {
			dialog.ShowError(errs[0], parent)
			return
		}

		if f.onSave != nil {
			f.onSave(peer)
		}
	}, parent)

	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

func (f *PeerForm) validate() (models.Peer, []error) {
	peer := models.Peer{
		PublicKey:    f.publicKeyEntry.Text,
		AllowedIPs:   f.allowedIPsEntry.Text,
		Endpoint:     f.endpointEntry.Text,
		PresharedKey: f.presharedKeyEntry.Text,
	}

	if f.persistentKeepaliveEntry.Text != "" {
		keepalive, err := strconv.Atoi(f.persistentKeepaliveEntry.Text)
		if err != nil {
			return peer, []error{wireguard.ValidationError{Field: "PersistentKeepalive", Message: "must be a number"}}
		}
		peer.PersistentKeepalive = keepalive
	}

	// Create a temp config to use the validator
	tempConfig := &models.WireGuardConfig{
		PrivateKey: "dGVzdGtleWZvcnZhbGlkYXRpb24wMTIzNDU2Nzg5MA==", // dummy valid key
		Address:    "10.0.0.1/32",
		Peers:      []models.Peer{peer},
	}

	errs := wireguard.ValidateConfig(tempConfig)
	// Filter to only peer errors
	var peerErrs []error
	for _, e := range errs {
		if ve, ok := e.(wireguard.ValidationError); ok {
			if len(ve.Field) > 4 && ve.Field[:4] == "Peer" {
				peerErrs = append(peerErrs, e)
			}
		}
	}

	return peer, peerErrs
}
