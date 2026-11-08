package ui

import (
	"errors"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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
	hint.Wrapping = fyne.TextWrapWord

	copied := func(v string) {
		parent.Clipboard().SetContent(v)
		wgwidget.ShowToast(parent, wgwidget.ToastSuccess, l.T(lang.Common_copied))
	}

	// Naming the endpoint the code was built with is the quickest way to catch
	// a peer that inherited the tunnel default when it needed its own.
	endpointField := wgwidget.NewCopyField(wgtheme.IconGlobe, l.T(lang.Detail_label_endpoint), endpoint, copied)

	copyBtn := wgwidget.NewButton(l.T(lang.Qr_btn_copy_config), wgtheme.IconCopy, wgwidget.Secondary, func() {
		copied(text)
	})
	saveBtn := wgwidget.NewButton(l.T(lang.Qr_btn_save), wgtheme.IconDownload, wgwidget.Secondary, func() {
		helpers.SaveFile(parent, helpers.FileOpts{
			Title:       l.T(lang.Qr_btn_save),
			FilterLabel: "WireGuard",
			Extensions:  []string{".conf"},
			FileName:    peerFileName(peer.Name) + ".conf",
		}, func(path string) {
			if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
				helpers.ShowError(err, parent)
				return
			}
			wgwidget.ShowToast(parent, wgwidget.ToastSuccess, path)
		})
	})

	// The QR stays black on white whatever the theme: scanners need it.
	side := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceLG),
		hint,
		wgwidget.Card(endpointField, 0),
		wgwidget.Banner(wgwidget.BannerWarning, l.T(lang.Common_warning), l.T(lang.Qr_warning_secret), nil),
		container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceSM), copyBtn, saveBtn),
	)
	sideBox := container.New(layout.NewGridWrapLayout(fyne.NewSize(330, side.MinSize().Height)), side)

	body := container.New(layout.NewCustomPaddedHBoxLayout(wgtheme.SpaceXL),
		container.NewVBox(qr), sideBox)

	var s *wgwidget.Surface
	closeBtn := wgwidget.NewButton(l.T(lang.Common_close), nil, wgwidget.Primary, func() { s.Close() })
	s = wgwidget.ShowModal(parent, l.Tf(lang.Qr_title, peer.Name), 0, body, closeBtn)
}

// peerFileName makes a peer name safe to use as a file name.
func peerFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "peer"
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			return '_'
		}
		return r
	}, name)
}
