package ui

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"
	wgtheme "wgAdmin/internal/ui/theme"
	"wgAdmin/internal/wgcfg"
	"wgAdmin/internal/wgstat"
	"wgAdmin/internal/wgwidget"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// TunnelForm handles tunnel creation and editing
type TunnelForm struct {
	window          fyne.Window
	ctrl            *wg.WG
	lang            *lang.Language
	isEdit          bool
	isPeer          bool
	name            string
	clientConfigDir string

	// Interface fields
	nameEntry           *widget.Entry
	privateKeyEntry     *widget.Entry
	publicKeyLabel      *widget.Label
	addressEntry        *widget.Entry
	dnsEntry            *widget.Entry
	listenPortEntry     *widget.Entry
	publicEndpointEntry *widget.Entry
	mtuEntry            *widget.Entry
	postUpEntry         *widget.Entry
	postDownEntry       *widget.Entry

	// Peers
	peers           []config.PeerConfig
	peersList       *widget.List
	peerPrivateKeys map[string]string

	// storedKeys holds the private keys recovered from client configs written
	// in earlier sessions, read once so the peer list does not re-scan the
	// directory on every row it draws.
	storedKeys map[string]string

	// Callbacks
	onSave   func(name string, config *config.Config) error
	onCancel func()
}

// NewTunnelForm creates a new tunnel form
func NewTunnelForm(parent fyne.Window, ctrl *wg.WG, existingName string, existingConfig *config.Config, onSave func(string, *config.Config) error, onCancel func(), clientConfigDir string, language *lang.Language) *TunnelForm {
	// The interface name is the file's name. The config's own Name field is a
	// display string go-wg builds as "[wg0] Public key: …", so using it here put
	// that whole line in the window title.
	isEdit := existingName != ""
	f := &TunnelForm{
		window:              parent,
		ctrl:                ctrl,
		lang:                language,
		isEdit:              isEdit,
		name:                existingName,
		clientConfigDir:     clientConfigDir,
		nameEntry:           widget.NewEntry(),
		privateKeyEntry:     widget.NewEntry(),
		publicKeyLabel:      widget.NewLabel(""),
		addressEntry:        widget.NewEntry(),
		dnsEntry:            widget.NewEntry(),
		listenPortEntry:     widget.NewEntry(),
		publicEndpointEntry: widget.NewEntry(),
		mtuEntry:            widget.NewEntry(),
		postUpEntry:         widget.NewMultiLineEntry(),
		postDownEntry:       widget.NewMultiLineEntry(),
		peers:               []config.PeerConfig{},
		peerPrivateKeys:     make(map[string]string),
		onSave:              onSave,
		onCancel:            onCancel,
	}

	f.nameEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_name))
	f.privateKeyEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_private_key))
	f.addressEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_address))
	f.dnsEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_dns))
	f.listenPortEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_listen_port))
	f.publicEndpointEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_endpoint))
	f.mtuEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_mtu))
	f.postUpEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_post_up))
	f.postDownEntry.SetPlaceHolder(language.T(lang.Tunnel_placeholder_post_down))
	f.postUpEntry.Wrapping = fyne.TextWrapWord
	f.postDownEntry.Wrapping = fyne.TextWrapWord

	if existingConfig != nil {
		f.nameEntry.SetText(existingName)
		if isEdit {
			f.nameEntry.Disable()
		} else {
			f.nameEntry.Enable()
		}
		f.privateKeyEntry.SetText(existingConfig.Interface.PrivateKey.String())

		addrs := make([]string, len(existingConfig.Interface.Address))
		for i, addr := range existingConfig.Interface.Address {
			addrs[i] = addr.String()
		}
		f.addressEntry.SetText(strings.Join(addrs, ", "))

		dnsAddrs := make([]string, len(existingConfig.Interface.DNS))
		for i, dns := range existingConfig.Interface.DNS {
			dnsAddrs[i] = dns.String()
		}
		f.dnsEntry.SetText(strings.Join(dnsAddrs, ", "))

		if existingConfig.Interface.ListenPort != nil {
			f.listenPortEntry.SetText(strconv.Itoa(*existingConfig.Interface.ListenPort))
		}
		f.publicEndpointEntry.SetText(existingConfig.PublicEndpoint)
		if existingConfig.Interface.MTU > 0 && existingConfig.Interface.MTU != 1420 {
			f.mtuEntry.SetText(strconv.Itoa(existingConfig.Interface.MTU))
		}
		f.postUpEntry.SetText(existingConfig.Interface.PostUp)
		f.postDownEntry.SetText(existingConfig.Interface.PostDown)
		f.peers = existingConfig.Peers
		f.updatePublicKey()
	}

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
	pubKey, err := wg.DerivePublicKey(f.privateKeyEntry.Text)
	if err != nil {
		f.publicKeyLabel.SetText(f.lang.T(lang.Tunnel_public_key_invalid))
	} else {
		f.publicKeyLabel.SetText(pubKey)
	}
}

func (f *TunnelForm) getTunnelName() string {
	if !f.isEdit {
		return f.nameEntry.Text
	}
	if i := strings.Index(f.name, "]"); i > 0 && strings.HasPrefix(f.name, "[") {
		return f.name[1:i]
	}
	return f.nameEntry.Text
}

// Show opens the interface form as a sheet over the main window.
func (f *TunnelForm) Show() {
	title := f.lang.T(lang.Tunnel_form_title_new)
	if f.isEdit {
		title = f.lang.Tf(lang.Tunnel_form_title_edit, f.name)
	}
	win := f.window

	generateKeyBtn := wgwidget.NewButton(f.lang.T(lang.Tunnel_btn_generate_keys), wgtheme.IconRefresh, wgwidget.Secondary, func() {
		priv, pub, err := wg.GenerateKeyPair()
		if err != nil {
			helpers.ShowError(err, win)
			return
		}
		f.privateKeyEntry.SetText(priv)
		f.publicKeyLabel.SetText(pub)
	})

	copyPubKeyBtn := wgwidget.NewIconButton(wgtheme.IconCopy, f.lang.T(lang.Common_copy), func() {
		if f.publicKeyLabel.Text != "" && f.publicKeyLabel.Text != f.lang.T(lang.Tunnel_public_key_invalid) {
			win.Clipboard().SetContent(f.publicKeyLabel.Text)
			wgwidget.ShowToast(win, wgwidget.ToastSuccess, f.lang.T(lang.Status_copied_pubkey))
		}
	})
	f.publicKeyLabel.TextStyle = fyne.TextStyle{Monospace: true}
	f.publicKeyLabel.Truncation = fyne.TextTruncateEllipsis

	f.postUpEntry.MultiLine = true
	f.postUpEntry.SetMinRowsVisible(3)
	f.postDownEntry.MultiLine = true
	f.postDownEntry.SetMinRowsVisible(3)

	// Grouped into the things a tunnel actually consists of, rather than one
	// undifferentiated column of fields.
	body := container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceXL+6),
		wgwidget.FormSection(wgtheme.IconKey, f.lang.T(lang.Tunnel_section_identity), "",
			wgwidget.Field(f.lang.T(lang.Tunnel_label_name), f.nameEntry, ""),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_private_key),
				container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry), ""),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_public_key),
				container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyLabel), ""),
		),
		wgwidget.FormSection(wgtheme.IconRoute, f.lang.T(lang.Tunnel_section_network), "",
			wgwidget.Field(f.lang.T(lang.Tunnel_label_address), f.addressEntry, ""),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_public_endpoint), f.publicEndpointEntry, ""),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_dns), f.dnsEntry, ""),
			wgwidget.FieldRow(
				wgwidget.Field(f.lang.T(lang.Tunnel_label_listen_port), f.listenPortEntry, ""),
				wgwidget.Field(f.lang.T(lang.Tunnel_label_mtu), f.mtuEntry, ""),
			),
		),
		wgwidget.FormSection(wgtheme.IconTerminal, f.lang.T(lang.Tunnel_section_hooks), f.lang.T(lang.Tunnel_hooks_hint),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_post_up), f.postUpEntry, ""),
			wgwidget.Field(f.lang.T(lang.Tunnel_label_post_down), f.postDownEntry, ""),
		),
	)

	var sheet *wgwidget.Surface
	cancelBtn := wgwidget.NewButton(f.lang.T(lang.Common_cancel), nil, wgwidget.Secondary, func() { sheet.Close() })
	saveBtn := wgwidget.NewButton(f.lang.T(lang.Common_save), wgtheme.IconCheck, wgwidget.Primary, func() {
		if f.save(win) {
			sheet.Close()
		}
	})

	sheet = wgwidget.ShowSheet(win, title, wgwidget.SheetWide, body, cancelBtn, saveBtn)
	f.closeHooks(sheet)
	if !f.isEdit {
		win.Canvas().Focus(f.nameEntry)
	}
}

// save validates and writes the tunnel, reporting problems over the sheet.
// It returns whether the sheet can close.
func (f *TunnelForm) save(win fyne.Window) bool {
	cfg, errs := f.validate()
	if len(errs) > 0 {
		helpers.ShowError(errs[0], win)
		return false
	}
	name := f.getTunnelName()
	if err := f.onSave(name, cfg); err != nil {
		helpers.ShowError(err, win)
		return false
	}
	f.generateClientConfigs(name, cfg, win)
	return true
}

// closeHooks runs onCancel when the sheet is dismissed without saving.
func (f *TunnelForm) closeHooks(sheet *wgwidget.Surface) {
	saved := false
	origSave := f.onSave
	f.onSave = func(name string, cfg *config.Config) error {
		err := origSave(name, cfg)
		if err == nil {
			saved = true
		}
		return err
	}
	sheet.OnClose = func() {
		f.onSave = origSave
		if !saved && f.onCancel != nil {
			f.onCancel()
		}
	}
}

// formHint renders the explanatory line under a section header.
func formHint(text string) fyne.CanvasObject {
	l := widget.NewLabel(text)
	l.Wrapping = fyne.TextWrapWord
	l.Importance = widget.LowImportance
	return l
}

// peer list columns
const (
	plColName = iota
	plColAllowed
	plColEndpoint
	plColClient
)

// ShowPeers opens the peer editor as a sheet: the tunnel's peers in a table
// with QR, edit and delete actions per row, and Add peer above it.
func (f *TunnelForm) ShowPeers() {
	title := f.lang.T(lang.Tunnel_peers_window_title)
	if f.isEdit {
		title = f.lang.Tf(lang.Tunnel_edit_peer_title, f.name)
	}
	win := f.window
	l := f.lang

	// Peers created in earlier sessions can still be provisioned: their private
	// keys live in the client configs wgAdmin wrote at the time.
	f.storedKeys = storedPeerKeys(f.clientConfigDir, f.getTunnelName())

	var table *wgwidget.DataTable
	cols := []wgwidget.Column{
		{Title: l.T(lang.Detail_col_peer), MinWidth: 120, Weight: 1},
		{Title: l.T(lang.Detail_col_allowed_ips), MinWidth: 130, Weight: 1.3, Mono: true},
		{Title: l.T(lang.Detail_col_endpoint), MinWidth: 120, Weight: 1.3, Mono: true},
		{Title: l.T(lang.Tunnel_col_client_config), MinWidth: 100},
	}
	table = wgwidget.NewDataTable(cols, func() int { return len(f.peers) }, func(r, c int) wgwidget.Cell {
		peer := f.peers[r]
		pubKeyStr := peer.PublicKey.String()
		switch c {
		case plColName:
			// The parser writes "Unknown" for unnamed peers; the key prefix
			// identifies them far better than that placeholder does.
			name := strings.TrimSpace(peer.Name)
			if name == "" || name == "Unknown" {
				name = wgstat.ShortKey(pubKeyStr)
			}
			return wgwidget.Cell{Text: name, Strong: true}
		case plColAllowed:
			ips := make([]string, 0, len(peer.AllowedIPs))
			for _, n := range peer.AllowedIPs {
				ips = append(ips, n.String())
			}
			return wgwidget.Cell{Text: strings.Join(ips, ", ")}
		case plColEndpoint:
			if peer.Endpoint == "" {
				return wgwidget.Cell{Text: wgstat.Dash, Dim: true}
			}
			return wgwidget.Cell{Text: peer.Endpoint}
		}
		// Provisioning needs the peer's private key: one generated this session
		// or the one in the client config written when the peer was created.
		_, generated := f.peerPrivateKeys[pubKeyStr]
		_, stored := f.storedKeys[pubKeyStr]
		if generated || stored {
			return wgwidget.Cell{Text: l.T(lang.Tunnel_client_ready)}
		}
		return wgwidget.Cell{Text: wgstat.Dash, Dim: true}
	})

	table.Actions = []wgwidget.RowAction{
		{Icon: wgtheme.IconQR, Tooltip: l.T(lang.Detail_tip_qr), OnTap: func(id int) {
			cfg, errs := f.validate()
			if len(errs) > 0 {
				helpers.ShowError(errs[0], win)
				return
			}
			req := f.qrRequestFor(id)
			req.serverCfg = cfg
			ShowPeerQR(win, req, l)
		}},
		{Icon: wgtheme.IconEdit, Tooltip: l.T(lang.Common_edit), OnTap: func(id int) {
			peerCopy := f.peers[id]
			oldPubKey := peerCopy.PublicKey.String()
			NewPeerForm(&peerCopy, f.publicEndpointEntry.Text, func(p config.PeerConfig, privateKey string) {
				delete(f.peerPrivateKeys, oldPubKey)
				f.peers[id] = p
				if privateKey != "" {
					f.peerPrivateKeys[p.PublicKey.String()] = privateKey
				}
				table.Refresh()
			}, nil, l).Show(win)
		}},
		{Icon: wgtheme.IconTrash, Tooltip: l.T(lang.Common_delete), OnTap: func(id int) {
			helpers.ShowConfirm(l.T(lang.Tunnel_delete_peer_title), l.T(lang.Tunnel_delete_peer_msg), func(yes bool) {
				if yes {
					delete(f.peerPrivateKeys, f.peers[id].PublicKey.String())
					f.peers = append(f.peers[:id], f.peers[id+1:]...)
					table.Refresh()
				}
			}, win)
		}},
	}

	addPeerBtn := wgwidget.NewButton(l.T(lang.Tunnel_btn_add_peer), wgtheme.IconAdd, wgwidget.Primary, func() {
		NewPeerForm(nil, f.publicEndpointEntry.Text, func(p config.PeerConfig, privateKey string) {
			f.peers = append(f.peers, p)
			if privateKey != "" {
				f.peerPrivateKeys[p.PublicKey.String()] = privateKey
			}
			table.Refresh()
		}, nil, l).Show(win)
	})
	addPeerBtn.Scale = wgwidget.ScaleSM

	// The table fills the sheet rather than being sized to its own contents, so
	// a tunnel with many peers scrolls and one with a single peer is not lost
	// in empty space.
	content := container.NewBorder(
		container.New(layout.NewCustomPaddedVBoxLayout(wgtheme.SpaceSM),
			wgwidget.SectionTitle(wgtheme.IconPeers, l.T(lang.Detail_section_peers), addPeerBtn),
			formHint(l.T(lang.Tunnel_peers_hint)),
		),
		nil, nil, nil,
		container.New(layout.NewCustomPaddedLayout(wgtheme.SpaceSM, 0, 0, 0), table),
	)

	var sheet *wgwidget.Surface
	cancelBtn := wgwidget.NewButton(l.T(lang.Common_cancel), nil, wgwidget.Secondary, func() { sheet.Close() })
	saveBtn := wgwidget.NewButton(l.T(lang.Common_save), wgtheme.IconCheck, wgwidget.Primary, func() {
		if f.save(win) {
			sheet.Close()
		}
	})

	sheet = wgwidget.ShowSheetFill(win, title, wgwidget.SheetFull, content, cancelBtn, saveBtn)
	f.closeHooks(sheet)
}

// qrRequestFor describes the client config for the peer at index id. The
// endpoint carries the form's current text rather than the saved config's, so
// an endpoint typed but not yet saved still produces a usable QR code.
func (f *TunnelForm) qrRequestFor(id int) qrRequest {
	if id < 0 || id >= len(f.peers) {
		return qrRequest{}
	}
	peer := f.peers[id]
	pubKey := peer.PublicKey.String()

	// A key generated in this session wins over a stored one: it is the key the
	// peer is about to be saved with.
	privateKey := f.peerPrivateKeys[pubKey]
	if privateKey == "" {
		privateKey = f.storedKeys[pubKey]
	}

	return qrRequest{
		tunnel:     f.getTunnelName(),
		peer:       peer,
		privateKey: privateKey,
		clientDir:  f.clientConfigDir,
		endpoint:   strings.TrimSpace(f.publicEndpointEntry.Text),
	}
}

func (f *TunnelForm) generateClientConfigs(tunnelName string, serverCfg *config.Config, win fyne.Window) {
	if len(f.peerPrivateKeys) == 0 {
		return
	}

	clientDir := clientDirFor(f.clientConfigDir, tunnelName)
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		helpers.ShowError(fmt.Errorf(f.lang.T(lang.Tunnel_error_create_client_dir), err), win)
		return
	}

	clientCtrl := wg.New(clientDir)
	var generated []string
	var missingEndpoint []string

	for _, peer := range serverCfg.Peers {
		pubKeyStr := peer.PublicKey.String()
		privKey, ok := f.peerPrivateKeys[pubKeyStr]
		if !ok {
			continue
		}

		// Each peer dials its own endpoint when it has one, so a config with a
		// second gateway or a non-standard port no longer forces every client
		// onto the same address.
		endpoint := wgcfg.EndpointFor(serverCfg, peer)
		if endpoint == "" {
			missingEndpoint = append(missingEndpoint, peer.Name)
			continue
		}
		if !wg.ValidateEndpoint(endpoint) {
			helpers.ShowError(fmt.Errorf(f.lang.T(lang.Tunnel_error_invalid_endpoint),
				errors.New(endpoint)), win)
			continue
		}

		allowedIPs := make([]string, len(peer.AllowedIPs))
		for i, ip := range peer.AllowedIPs {
			allowedIPs[i] = ip.String()
		}

		peerOpts := wg.PeerOpts{
			Name:                peer.Name,
			PublicKey:           pubKeyStr,
			AllowedIPs:          allowedIPs,
			Endpoint:            endpoint,
			PersistentKeepalive: peer.PersistentKeepalive,
		}

		clientPath := filepath.Join(clientDir, peer.Name+".conf")
		if _, err := os.Stat(clientPath); err == nil {
			// A stale copy would otherwise block NewClientConfig, which refuses
			// to overwrite an existing file.
			os.Rename(clientPath, clientPath+".bkp")
		}
		if _, err := clientCtrl.NewClientConfig(*serverCfg, peerOpts, privKey, true); err != nil {
			helpers.ShowError(fmt.Errorf(f.lang.Tf(lang.Tunnel_error_client_config, peer.Name), err), win)
			continue
		}
		generated = append(generated, clientPath)
	}

	// One report covering both outcomes: stacking two dialogs on top of each
	// other after a save is worse than a single message that says what was
	// written and what was not.
	var report []string
	if len(generated) > 0 {
		absDir, _ := filepath.Abs(clientDir)
		report = append(report, f.lang.Tf(lang.Tunnel_client_configs_success, len(generated), absDir))
	}
	if len(missingEndpoint) > 0 {
		report = append(report, f.lang.Tf(lang.Tunnel_client_configs_no_endpoint,
			strings.Join(missingEndpoint, ", ")))
	}
	if len(report) > 0 {
		helpers.ShowInformation(f.lang.T(lang.Tunnel_client_configs_title),
			strings.Join(report, "\n\n"), win)
	}
}

func (f *TunnelForm) validate() (*config.Config, []error) {
	// Wee need to get the endpoint from main if applied there.

	cfg := &config.Config{
		Interface: config.InterfaceConfig{
			MTU:   1420,
			Table: "auto",
		},
		Peers: f.peers,
	}

	if f.privateKeyEntry.Text == "" {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: f.lang.T(lang.Error_required)}}
	}
	if !wg.ValidateKey(f.privateKeyEntry.Text) {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: f.lang.T(lang.Error_invalid_format)}}
	}
	privKey, err := wgtypes.ParseKey(f.privateKeyEntry.Text)
	if err != nil {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: f.lang.T(lang.Error_invalid_key)}}
	}
	cfg.Interface.PrivateKey = privKey

	if f.addressEntry.Text == "" {
		return nil, []error{wg.ValidationError{Field: "Address", Message: f.lang.T(lang.Error_required)}}
	}
	for _, addr := range strings.Split(f.addressEntry.Text, ",") {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		if !strings.Contains(addr, "/") {
			if strings.Contains(addr, ":") {
				addr += "/128"
			} else {
				addr += "/32"
			}
		}

		_, ipNet, err := net.ParseCIDR(addr)
		if err != nil {
			return nil, []error{wg.ValidationError{Field: "Address", Message: f.lang.T(lang.Error_invalid_cidr)}}
		}
		cfg.Interface.Address = append(cfg.Interface.Address, *ipNet)
	}

	if f.dnsEntry.Text != "" {
		for _, dns := range strings.Split(f.dnsEntry.Text, ",") {
			dns = strings.TrimSpace(dns)
			if dns == "" {
				continue
			}
			ip := net.ParseIP(dns)
			if ip == nil {
				return nil, []error{wg.ValidationError{Field: "DNS", Message: f.lang.Tf(lang.Error_invalid_dns, dns)}}
			}
			cfg.Interface.DNS = append(cfg.Interface.DNS, ip)
		}
	}

	if f.listenPortEntry.Text != "" {
		port, err := strconv.Atoi(f.listenPortEntry.Text)
		if err != nil {
			return nil, []error{wg.ValidationError{Field: "ListenPort", Message: f.lang.T(lang.Error_invalid_number)}}
		}
		if port < 0 || port > 65535 {
			return nil, []error{wg.ValidationError{Field: "ListenPort", Message: f.lang.T(lang.Error_invalid_range_0)}}
		}
		cfg.Interface.ListenPort = &port
	}

	// The tunnel endpoint is the default clients dial; it is not stamped onto
	// every [Peer] section. Doing that wrote the server's own address into its
	// peer list — which is meaningless for a hub — and made the default
	// indistinguishable from a genuine per-peer override once saved. Peers that
	// need their own host or port set it in the peer form; the rest inherit
	// this value when their client config is generated.
	if endpoint := strings.TrimSpace(f.publicEndpointEntry.Text); endpoint != "" {
		if !wg.ValidateEndpoint(endpoint) {
			return nil, []error{wg.ValidationError{Field: "PublicEndpoint", Message: f.lang.T(lang.Error_invalid_endpoint)}}
		}
		cfg.PublicEndpoint = endpoint
	}

	if f.mtuEntry.Text != "" {
		mtu, err := strconv.Atoi(f.mtuEntry.Text)
		if err != nil {
			return nil, []error{wg.ValidationError{Field: "MTU", Message: f.lang.T(lang.Error_invalid_number)}}
		}
		cfg.Interface.MTU = mtu
	}

	cfg.Interface.PostUp = strings.TrimSpace(f.postUpEntry.Text)
	cfg.Interface.PostDown = strings.TrimSpace(f.postDownEntry.Text)

	if !f.isEdit {
		if !wg.ValidateName(f.nameEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "Name", Message: f.lang.T(lang.Error_tunnel_name_format)}}
		}
		if f.ctrl.ConfigExists(f.nameEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "Name", Message: f.lang.T(lang.Error_tunnel_exists)}}
		}
	}

	errs := wg.ValidateConfig(cfg)
	return cfg, errs
}
