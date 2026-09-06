package ui

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgwidget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"
)

// buildClientConfig renders the .conf a peer's device needs in order to join,
// without writing anything to disk. It reuses the same generator the file-based
// export uses, so the QR and the exported file can never drift apart.
func buildClientConfig(serverCfg *config.Config, peer config.PeerConfig, peerPrivateKey, endpoint string) (string, error) {
	allowedIPs := make([]string, 0, len(peer.AllowedIPs))
	for _, n := range peer.AllowedIPs {
		allowedIPs = append(allowedIPs, n.String())
	}

	opts := wg.PeerOpts{
		Name:                peer.Name,
		PublicKey:           peer.PublicKey.String(),
		AllowedIPs:          allowedIPs,
		Endpoint:            endpoint,
		PersistentKeepalive: peer.PersistentKeepalive,
	}

	// save=false keeps this purely in memory.
	ctrl := wg.New("")
	clientCfg, err := ctrl.NewClientConfig(*serverCfg, opts, peerPrivateKey, false)
	if err != nil {
		return "", err
	}

	return config.GenerateConfigString(&clientCfg), nil
}

// ShowPeerQR displays a scannable client config for a peer.
//
// The peer's private key is required: a client config is useless without it,
// and wgAdmin only holds it for peers whose keys were generated here. Peers
// added by public key alone cannot be provisioned this way, and the dialog
// says so rather than showing a QR that would not work.
func ShowPeerQR(parent fyne.Window, serverCfg *config.Config, peer config.PeerConfig,
	peerPrivateKey, endpoint string, l *lang.Language) {

	if endpoint == "" {
		helpers.ShowInformation(l.T(lang.Qr_btn_show), l.T(lang.Qr_no_endpoint), parent)
		return
	}
	if peerPrivateKey == "" {
		helpers.ShowInformation(l.T(lang.Qr_btn_show), l.T(lang.Qr_no_private_key), parent)
		return
	}

	text, err := buildClientConfig(serverCfg, peer, peerPrivateKey, endpoint)
	if err != nil {
		helpers.ShowError(errors.New(l.Tf(lang.Qr_error_generate, err.Error())), parent)
		return
	}

	qr, err := wgwidget.NewQRCode(text)
	if err != nil {
		helpers.ShowError(errors.New(l.Tf(lang.Qr_error_generate, err.Error())), parent)
		return
	}

	hint := widget.NewLabel(l.T(lang.Qr_hint))
	hint.Alignment = fyne.TextAlignCenter
	hint.Wrapping = fyne.TextWrapWord

	warning := wgwidget.Caption(l.T(lang.Qr_warning_secret))
	warning.Color = wgtheme.Current().Warning
	warning.Alignment = fyne.TextAlignCenter

	var d dialog.Dialog
	copyBtn := widget.NewButtonWithIcon(l.T(lang.Qr_btn_copy_config), wgtheme.IconCopy, func() {
		parent.Clipboard().SetContent(text)
	})

	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD),
		container.NewCenter(qr),
		container.NewCenter(hint),
		container.NewCenter(warning),
		container.NewCenter(copyBtn),
	)

	d = dialog.NewCustom(l.Tf(lang.Qr_title, peer.Name), l.T(lang.Common_close),
		container.New(layout.NewCustomPaddedLayout(
			wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceXL, wgtheme.SpaceXL), body), parent)
	d.Resize(fyne.NewSize(400, 520))
	d.Show()
}
