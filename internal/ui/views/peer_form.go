package ui

import (
	"net"
	"strconv"
	"strings"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"

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

	generatedPrivateKey string

	lang     *lang.Language
	onSave   func(peer config.PeerConfig, privateKey string)
	onCancel func()
}

// NewPeerForm creates a new peer form
func NewPeerForm(existing *config.PeerConfig, onSave func(config.PeerConfig, string), onCancel func(), language *lang.Language) *PeerForm {
	f := &PeerForm{
		nameEntry:                widget.NewEntry(),
		privateKeyEntry:          widget.NewEntry(),
		publicKeyEntry:           widget.NewEntry(),
		publicKeyLabel:           widget.NewLabel(""),
		allowedIPsEntry:          widget.NewEntry(),
		persistentKeepaliveEntry: widget.NewEntry(),
		presharedKeyEntry:        widget.NewEntry(),
		lang:                     language,
		onSave:                   onSave,
		onCancel:                 onCancel,
	}

	f.nameEntry.SetPlaceHolder(language.T("peer.placeholder_name"))
	f.privateKeyEntry.SetPlaceHolder(language.T("peer.placeholder_private_key"))
	f.publicKeyEntry.SetPlaceHolder(language.T("peer.placeholder_public_key"))
	f.allowedIPsEntry.SetPlaceHolder(language.T("peer.placeholder_allowed_ips"))
	f.persistentKeepaliveEntry.SetPlaceHolder(language.T("peer.placeholder_keepalive"))
	f.presharedKeyEntry.SetPlaceHolder(language.T("peer.placeholder_preshared_key"))

	f.privateKeyEntry.OnChanged = func(s string) {
		f.updatePublicKey()
	}

	if existing != nil {
		f.nameEntry.SetText(existing.Name)
		f.publicKeyEntry.SetText(existing.PublicKey.String())

		ips := make([]string, len(existing.AllowedIPs))
		for i, ip := range existing.AllowedIPs {
			ips[i] = ip.String()
		}
		f.allowedIPsEntry.SetText(strings.Join(ips, ", "))

		if existing.PersistentKeepalive > 0 {
			f.persistentKeepaliveEntry.SetText(strconv.Itoa(existing.PersistentKeepalive))
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
		f.publicKeyLabel.SetText(f.lang.T("peer.invalid_key"))
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
	generateKeyBtn := widget.NewButtonWithIcon(f.lang.T("peer.btn_generate_keys"), theme.ViewRefreshIcon(), func() {
		priv, pub, err := wg.GenerateKeyPair()
		if err != nil {
			helpers.ShowError(err, parent)
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
		widget.NewLabel(f.lang.T("peer.label_name")),
		f.nameEntry,
		widget.NewSeparator(),
		widget.NewLabel(f.lang.T("peer.label_private_key")),
		container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry),
		widget.NewLabel(f.lang.T("peer.label_public_key")),
		container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyEntry),

		widget.NewSeparator(),
		widget.NewLabel(f.lang.T("peer.label_allowed_ips")),
		f.allowedIPsEntry,

		widget.NewLabel(f.lang.T("peer.label_keepalive")),
		f.persistentKeepaliveEntry,

		widget.NewLabel(f.lang.T("peer.label_preshared_key")),
		f.presharedKeyEntry,
	)

	d := dialog.NewCustomConfirm(f.lang.T("peer.title"), f.lang.T("common.save"), f.lang.T("common.cancel"), form, func(confirmed bool) {
		if !confirmed {
			if f.onCancel != nil {
				f.onCancel()
			}
			return
		}

		peer, errs := f.validate()
		if len(errs) > 0 {
			helpers.ShowError(errs[0], parent)
			return
		}

		if f.onSave != nil {
			f.onSave(peer, f.generatedPrivateKey)
		}
	}, parent)

	d.Resize(fyne.NewSize(600, 700))
	d.Show()
}

func (f *PeerForm) validate() (config.PeerConfig, []error) {
	peer := config.PeerConfig{}

	name := strings.TrimSpace(f.nameEntry.Text)
	if name == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.Name", Message: f.lang.T("error.required")}}
	}
	peer.Name = name

	if f.publicKeyEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T("error.required")}}
	}
	if !wg.ValidateKey(f.publicKeyEntry.Text) {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T("error.invalid_format")}}
	}
	pubKey, err := wgtypes.ParseKey(f.publicKeyEntry.Text)
	if err != nil {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T("error.invalid_key")}}
	}
	peer.PublicKey = pubKey

	if f.allowedIPsEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: f.lang.T("error.required")}}
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
			return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: f.lang.Tf("error.invalid_cidr", cidr)}}
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *ipNet)
	}

	if f.persistentKeepaliveEntry.Text != "" {
		keepalive, err := strconv.Atoi(f.persistentKeepaliveEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: f.lang.T("error.must_be_number")}}
		}
		if keepalive < 0 || keepalive > 65535 {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: f.lang.T("error.keepalive_range")}}
		}
		peer.PersistentKeepalive = keepalive
	}

	if f.presharedKeyEntry.Text != "" {
		if !wg.ValidateKey(f.presharedKeyEntry.Text) {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: f.lang.T("error.invalid_format")}}
		}
		psk, err := wgtypes.ParseKey(f.presharedKeyEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: f.lang.T("error.invalid_key")}}
		}
		peer.PresharedKey = &psk
	}

	return peer, nil
}
