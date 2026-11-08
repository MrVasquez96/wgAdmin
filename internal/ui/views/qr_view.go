package ui

import (
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgcfg"
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

// qrRequest is everything needed to rebuild a peer's client config on demand.
//
// The private key and the endpoint are both optional here: either can be
// recovered from what is already on disk, and resolving them inside the dialog
// keeps that recovery in one place instead of at every call site.
type qrRequest struct {
	tunnel     string
	serverCfg  *config.Config
	peer       config.PeerConfig
	privateKey string // in-session key, when the pair was generated just now
	clientDir  string // root of the generated client configs
	endpoint   string // overrides the config's endpoint, for unsaved form edits
}

// resolve fills in whatever the caller could not supply.
func (r qrRequest) resolve() (privateKey, endpoint string) {
	privateKey = r.privateKey
	if privateKey == "" {
		privateKey = storedPeerPrivateKey(r.clientDir, r.tunnel, r.peer.PublicKey.String())
	}

	endpoint = strings.TrimSpace(r.peer.Endpoint)
	if endpoint == "" {
		endpoint = strings.TrimSpace(r.endpoint)
	}
	if endpoint == "" {
		endpoint = wgcfg.EndpointFor(r.serverCfg, r.peer)
	}
	return privateKey, endpoint
}

// ShowPeerQR displays a scannable client config for a peer.
//
// The peer's private key is required: a client config is useless without it.
// WireGuard servers only ever store the public half, so the key comes either
// from this session or from the client config wgAdmin wrote when the peer was
// created. A peer added by public key alone has neither, and the dialog says so
// rather than showing a QR that would not work.
func ShowPeerQR(parent fyne.Window, req qrRequest, l *lang.Language) {
	peer := req.peer
	peerPrivateKey, endpoint := req.resolve()

	if endpoint == "" {
		helpers.ShowInformation(l.T(lang.Qr_btn_show), l.T(lang.Qr_no_endpoint), parent)
		return
	}
	if peerPrivateKey == "" {
		helpers.ShowInformation(l.T(lang.Qr_btn_show), l.T(lang.Qr_no_private_key), parent)
		return
	}

	text, err := buildClientConfig(req.serverCfg, peer, peerPrivateKey, endpoint)
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

	// Naming the endpoint the code was built with is the quickest way to catch
	// a peer that inherited the tunnel default when it needed its own.
	endpointLine := wgwidget.DimValue(l.Tf(lang.Qr_endpoint, endpoint))
	endpointLine.TextSize = 11

	// The hint is added bare, not centred: a centred child is shrunk to its
	// MinSize, and a word-wrapping label's minimum width is a single glyph,
	// which stacked the text one letter per line. Full width plus centre
	// alignment wraps it by words instead.
	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceMD),
		container.NewCenter(qr),
		hint,
		container.NewCenter(endpointLine),
		container.NewCenter(warning),
		container.NewCenter(copyBtn),
	)

	d = dialog.NewCustom(l.Tf(lang.Qr_title, peer.Name), l.T(lang.Common_close),
		container.New(layout.NewCustomPaddedLayout(
			wgtheme.SpaceMD, wgtheme.SpaceMD, wgtheme.SpaceXL, wgtheme.SpaceXL), body), parent)
	d.Resize(fyne.NewSize(400, 520))
	d.Show()
}
