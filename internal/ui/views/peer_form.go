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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// PeerForm handles peer configuration
type PeerForm struct {
	nameEntry                *widget.Entry
	privateKeyEntry          *widget.Entry
	publicKeyEntry           *widget.Entry
	publicKeyLabel           *widget.Label
	allowedIPsEntry          *widget.Entry
	persistentKeepaliveEntry *widget.Entry
	presharedKeyEntry        *widget.Entry

	generatedPrivateKey string // In-memory only, never persisted to server config

	onSave   func(peer config.PeerConfig, privateKey string)
	onCancel func()
}

// NewPeerForm creates a new peer form
func NewPeerForm(existing *config.PeerConfig, onSave func(config.PeerConfig, string), onCancel func()) *PeerForm {
	f := &PeerForm{
		nameEntry:                widget.NewEntry(),
		privateKeyEntry:          widget.NewEntry(),
		publicKeyEntry:           widget.NewEntry(),
		publicKeyLabel:           widget.NewLabel(""),
		allowedIPsEntry:          widget.NewEntry(),
		persistentKeepaliveEntry: widget.NewEntry(),
		presharedKeyEntry:        widget.NewEntry(),
		onSave:                   onSave,
		onCancel:                 onCancel,
	}

	f.nameEntry.SetPlaceHolder("e.g., laptop, phone, office (required)")
	f.privateKeyEntry.SetPlaceHolder("Base64 encoded private key (optional)")
	f.publicKeyEntry.SetPlaceHolder("Base64 encoded public key")
	f.allowedIPsEntry.SetPlaceHolder("e.g., 10.0.0.2/32")
	f.persistentKeepaliveEntry.SetPlaceHolder("e.g., 25 (seconds, optional)")
	f.presharedKeyEntry.SetPlaceHolder("Base64 encoded key (optional)")

	// Update public key when private key changes
	f.privateKeyEntry.OnChanged = func(s string) {
		f.updatePublicKey()
	}

	if existing != nil {
		f.nameEntry.SetText(existing.Name)
		f.publicKeyEntry.SetText(existing.PublicKey.String())

		// Convert AllowedIPs to string
		ips := make([]string, len(existing.AllowedIPs))
		for i, ip := range existing.AllowedIPs {
			ips[i] = ip.String()
		}
		f.allowedIPsEntry.SetText(strings.Join(ips, ", "))

		if existing.PersistentKeepalive > 0 {
			f.persistentKeepaliveEntry.SetText(strconv.Itoa(int(existing.PersistentKeepalive.Seconds())))
		}
		if existing.PresharedKey != nil {
			f.presharedKeyEntry.SetText(existing.PresharedKey.String())
		}
	}

	return f
}

func (f *PeerForm) updatePublicKey() {
	if f.privateKeyEntry.Text == "" {
		f.publicKeyLabel.SetText("")
		f.publicKeyEntry.Enable()
		f.generatedPrivateKey = ""
		return
	}
	pubKey, err := wg.DerivePublicKey(f.privateKeyEntry.Text)
	if err != nil {
		f.publicKeyLabel.SetText("(invalid key)")
		f.publicKeyEntry.Enable()
		f.generatedPrivateKey = ""
		return
	}
	f.publicKeyLabel.SetText(pubKey)
	f.publicKeyEntry.SetText(pubKey)
	f.publicKeyEntry.Disable()
	f.generatedPrivateKey = f.privateKeyEntry.Text
}

// Show displays the peer form dialog
func (f *PeerForm) Show(parent fyne.Window) {
	generateKeyBtn := widget.NewButtonWithIcon("Generate Keys", theme.ViewRefreshIcon(), func() {
		priv, pub, err := wg.GenerateKeyPair()
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		f.privateKeyEntry.SetText(priv)
		f.publicKeyLabel.SetText(pub)
		f.publicKeyEntry.SetText(pub)
		f.publicKeyEntry.Disable()
		f.generatedPrivateKey = priv
	})

	copyPubKeyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		text := f.publicKeyEntry.Text
		if text != "" && text != "(invalid key)" {
			parent.Clipboard().SetContent(text)
		}
	})

	form := container.NewVBox(
		widget.NewLabel("Peer Name *"),
		f.nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Private Key (for client config generation)"),
		container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry),
		widget.NewLabel("Public Key (derived):"),
		container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyLabel),
		widget.NewSeparator(),
		widget.NewLabel("Public Key *"),
		f.publicKeyEntry,

		widget.NewLabel("Allowed IPs * (client's VPN address)"),
		f.allowedIPsEntry,

		widget.NewLabel("Persistent Keepalive (for client config)"),
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
			f.onSave(peer, f.generatedPrivateKey)
		}
	}, parent)

	d.Resize(fyne.NewSize(500, 600))
	d.Show()
}

func (f *PeerForm) validate() (config.PeerConfig, []error) {
	peer := config.PeerConfig{}

	// Parse Name
	name := strings.TrimSpace(f.nameEntry.Text)
	if name == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.Name", Message: "required"}}
	}
	peer.Name = name

	// Parse PublicKey
	if f.publicKeyEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: "required"}}
	}
	if !wg.ValidateKey(f.publicKeyEntry.Text) {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: "invalid format"}}
	}
	pubKey, err := wgtypes.ParseKey(f.publicKeyEntry.Text)
	if err != nil {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: "invalid key"}}
	}
	peer.PublicKey = pubKey

	// Parse AllowedIPs
	if f.allowedIPsEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: "required"}}
	}
	for _, cidr := range strings.Split(f.allowedIPsEntry.Text, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: fmt.Sprintf("invalid CIDR: %s", cidr)}}
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *ipNet)
	}

	// Parse PersistentKeepalive
	if f.persistentKeepaliveEntry.Text != "" {
		keepalive, err := strconv.Atoi(f.persistentKeepaliveEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: "must be a number"}}
		}
		if keepalive < 0 || keepalive > 65535 {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: "must be 0-65535"}}
		}
		peer.PersistentKeepalive = time.Duration(keepalive) * time.Second
	}

	// Parse PresharedKey
	if f.presharedKeyEntry.Text != "" {
		if !wg.ValidateKey(f.presharedKeyEntry.Text) {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: "invalid format"}}
		}
		psk, err := wgtypes.ParseKey(f.presharedKeyEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: "invalid key"}}
		}
		peer.PresharedKey = &psk
	}

	return peer, nil
}
