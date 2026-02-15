package views

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/models"
	"wgAdmin/internal/wireguard"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// PeerForm handles peer configuration
type PeerForm struct {
	nameLabel                *widget.Label
	publicKeyEntry           *widget.Entry
	allowedIPsEntry          *widget.Entry
	endpointEntry            *widget.Entry
	persistentKeepaliveEntry *widget.Entry
	presharedKeyEntry        *widget.Entry

	onSave   func(peer models.PeerConfig)
	onCancel func()
}

// NewPeerForm creates a new peer form
func NewPeerForm(existing *models.PeerConfig, onSave func(models.PeerConfig), onCancel func()) *PeerForm {
	peerName := ""
	if existing != nil {
		peerName = existing.Name
	}
	f := &PeerForm{
		nameLabel:                widget.NewLabel(peerName),
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
		f.publicKeyEntry.SetText(existing.PublicKey.String())

		// Convert AllowedIPs to string
		ips := make([]string, len(existing.AllowedIPs))
		for i, ip := range existing.AllowedIPs {
			ips[i] = ip.String()
		}
		f.allowedIPsEntry.SetText(strings.Join(ips, ", "))

		if existing.Endpoint != nil {
			f.endpointEntry.SetText(existing.Endpoint.String())
		}
		if existing.PersistentKeepalive > 0 {
			f.persistentKeepaliveEntry.SetText(strconv.Itoa(int(existing.PersistentKeepalive.Seconds())))
		}
		if existing.PresharedKey != nil {
			f.presharedKeyEntry.SetText(existing.PresharedKey.String())
		}
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

func (f *PeerForm) validate() (models.PeerConfig, []error) {
	peer := models.PeerConfig{}

	// Parse PublicKey
	if f.publicKeyEntry.Text == "" {
		return peer, []error{wireguard.ValidationError{Field: "Peer[0].PublicKey", Message: "required"}}
	}
	if !wireguard.ValidateKey(f.publicKeyEntry.Text) {
		return peer, []error{wireguard.ValidationError{Field: "Peer[0].PublicKey", Message: "invalid format"}}
	}
	pubKey, err := wgtypes.ParseKey(f.publicKeyEntry.Text)
	if err != nil {
		return peer, []error{wireguard.ValidationError{Field: "Peer[0].PublicKey", Message: "invalid key"}}
	}
	peer.PublicKey = pubKey

	// Parse AllowedIPs
	if f.allowedIPsEntry.Text == "" {
		return peer, []error{wireguard.ValidationError{Field: "Peer[0].AllowedIPs", Message: "required"}}
	}
	for _, cidr := range strings.Split(f.allowedIPsEntry.Text, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].AllowedIPs", Message: fmt.Sprintf("invalid CIDR: %s", cidr)}}
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *ipNet)
	}

	// Parse Endpoint
	if f.endpointEntry.Text != "" {
		if !wireguard.ValidateEndpoint(f.endpointEntry.Text) {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].Endpoint", Message: "invalid format (host:port)"}}
		}
		endpoint, err := net.ResolveUDPAddr("udp", f.endpointEntry.Text)
		if err != nil {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].Endpoint", Message: "cannot resolve endpoint"}}
		}
		peer.Endpoint = endpoint
	}

	// Parse PersistentKeepalive
	if f.persistentKeepaliveEntry.Text != "" {
		keepalive, err := strconv.Atoi(f.persistentKeepaliveEntry.Text)
		if err != nil {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].PersistentKeepalive", Message: "must be a number"}}
		}
		if keepalive < 0 || keepalive > 65535 {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].PersistentKeepalive", Message: "must be 0-65535"}}
		}
		peer.PersistentKeepalive = time.Duration(keepalive) * time.Second
	}

	// Parse PresharedKey
	if f.presharedKeyEntry.Text != "" {
		if !wireguard.ValidateKey(f.presharedKeyEntry.Text) {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].PresharedKey", Message: "invalid format"}}
		}
		psk, err := wgtypes.ParseKey(f.presharedKeyEntry.Text)
		if err != nil {
			return peer, []error{wireguard.ValidationError{Field: "Peer[0].PresharedKey", Message: "invalid key"}}
		}
		peer.PresharedKey = &psk
	}

	return peer, nil
}
