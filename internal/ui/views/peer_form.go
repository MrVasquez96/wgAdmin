package ui

import (
	"net"
	"strconv"
	"strings"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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

	f.nameEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_name))
	f.privateKeyEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_private_key))
	f.publicKeyEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_public_key))
	f.allowedIPsEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_allowed_ips))
	f.persistentKeepaliveEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_keepalive))
	f.presharedKeyEntry.SetPlaceHolder(language.T(lang.Peer_placeholder_preshared))

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
		f.publicKeyLabel.SetText(f.lang.T(lang.Peer_public_key_invalid))
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
	generateKeyBtn := widget.NewButtonWithIcon(f.lang.T(lang.Peer_btn_generate_keys), wgtheme.IconRefresh, func() {
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

	copyPubKeyBtn := widget.NewButtonWithIcon("", wgtheme.IconCopy, func() {
		text := f.publicKeyEntry.Text
		if text != "" && text != "(invalid key)" {
			parent.Clipboard().SetContent(text)
		}
	})
	form := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		wgwidget.SectionHeader(f.lang.T(lang.Tunnel_section_identity)),
		widget.NewForm(
			widget.NewFormItem(f.lang.T(lang.Peer_label_name), f.nameEntry),
		),

		wgwidget.SectionHeader(f.lang.T(lang.Peer_section_keys)),
		formHint(f.lang.T(lang.Peer_keys_hint)),
		widget.NewForm(
			widget.NewFormItem(f.lang.T(lang.Peer_label_private_key),
				container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry)),
			widget.NewFormItem(f.lang.T(lang.Peer_label_public_key),
				container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyEntry)),
			widget.NewFormItem(f.lang.T(lang.Peer_label_preshared_key), f.presharedKeyEntry),
		),

		wgwidget.SectionHeader(f.lang.T(lang.Tunnel_section_network)),
		widget.NewForm(
			widget.NewFormItem(f.lang.T(lang.Peer_label_allowed_ips_server), f.allowedIPsEntry),
			widget.NewFormItem(f.lang.T(lang.Peer_label_keepalive), f.persistentKeepaliveEntry),
		),
	)

	scrolled := container.NewVScroll(container.New(
		layout.NewCustomPaddedLayout(0, wgtheme.SpaceMD, 0, wgtheme.SpaceMD), form))

	d := dialog.NewCustomConfirm(f.lang.T(lang.Peer_form_title), f.lang.T(lang.Common_save), f.lang.T(lang.Common_cancel), scrolled, func(confirmed bool) {
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
		return peer, []error{wg.ValidationError{Field: "Peer.Name", Message: f.lang.T(lang.Error_required)}}
	}
	peer.Name = name

	if f.publicKeyEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T(lang.Error_required)}}
	}
	if !wg.ValidateKey(f.publicKeyEntry.Text) {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T(lang.Error_invalid_format)}}
	}
	pubKey, err := wgtypes.ParseKey(f.publicKeyEntry.Text)
	if err != nil {
		return peer, []error{wg.ValidationError{Field: "Peer.PublicKey", Message: f.lang.T(lang.Error_invalid_key)}}
	}
	peer.PublicKey = pubKey

	if f.allowedIPsEntry.Text == "" {
		return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: f.lang.T(lang.Error_required)}}
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
			return peer, []error{wg.ValidationError{Field: "Peer.AllowedIPs", Message: f.lang.Tf(lang.Error_invalid_cidr_value, cidr)}}
		}
		peer.AllowedIPs = append(peer.AllowedIPs, *ipNet)
	}

	if f.persistentKeepaliveEntry.Text != "" {
		keepalive, err := strconv.Atoi(f.persistentKeepaliveEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: f.lang.T(lang.Error_invalid_number)}}
		}
		if keepalive < 0 || keepalive > 65535 {
			return peer, []error{wg.ValidationError{Field: "Peer.PersistentKeepalive", Message: f.lang.T(lang.Error_invalid_range_0)}}
		}
		peer.PersistentKeepalive = keepalive
	}

	if f.presharedKeyEntry.Text != "" {
		if !wg.ValidateKey(f.presharedKeyEntry.Text) {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: f.lang.T(lang.Error_invalid_format)}}
		}
		psk, err := wgtypes.ParseKey(f.presharedKeyEntry.Text)
		if err != nil {
			return peer, []error{wg.ValidationError{Field: "Peer.PresharedKey", Message: f.lang.T(lang.Error_invalid_key)}}
		}
		peer.PresharedKey = &psk
	}

	return peer, nil
}
