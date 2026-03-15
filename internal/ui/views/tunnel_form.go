package ui

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"wgAdmin/internal/ui/helpers"
	"wgAdmin/internal/ui/lang"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

	// Callbacks
	onSave   func(name string, config *config.Config) error
	onCancel func()
}

// NewTunnelForm creates a new tunnel form
func NewTunnelForm(parent fyne.Window, ctrl *wg.WG, existingName string, existingConfig *config.Config, onSave func(string, *config.Config) error, onCancel func(), clientConfigDir string, language *lang.Language) *TunnelForm {
	tunnelName := existingName
	isEdit := existingName != ""
	if isEdit {
		tunnelName = existingConfig.Name
	}
	f := &TunnelForm{
		window:              parent,
		ctrl:                ctrl,
		lang:                language,
		isEdit:              existingName != "",
		name:                tunnelName,
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

// Show displays the tunnel form
func (f *TunnelForm) Show() {
	title := f.lang.T(lang.Tunnel_form_title_new)
	if f.isEdit {
		title = f.lang.Tf(lang.Tunnel_form_title_edit, f.name)
	}

	win := fyne.CurrentApp().NewWindow(title)
	win.Resize(fyne.NewSize(700, 750))

	generateKeyBtn := widget.NewButtonWithIcon(f.lang.T(lang.Tunnel_btn_generate_keys), theme.ViewRefreshIcon(), func() {
		priv, pub, err := wg.GenerateKeyPair()
		if err != nil {
			helpers.ShowError(err, win)
			return
		}
		f.privateKeyEntry.SetText(priv)
		f.publicKeyLabel.SetText(pub)
	})

	copyPubKeyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		if f.publicKeyLabel.Text != "" && f.publicKeyLabel.Text != f.lang.T(lang.Tunnel_public_key_invalid) {
			win.Clipboard().SetContent(f.publicKeyLabel.Text)
		}
	})

	privKeyRow := container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry)
	pubKeyRow := container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyLabel)

	f.postUpEntry.MultiLine = true
	f.postUpEntry.SetMinRowsVisible(3)
	f.postDownEntry.MultiLine = true
	f.postDownEntry.SetMinRowsVisible(3)

	interfaceForm := widget.NewForm(
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_name), f.nameEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_private_key), privKeyRow),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_public_key), pubKeyRow),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_address), f.addressEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_dns), f.dnsEntry),
		//widget.NewFormItem(f.lang.T(lang.Tunnel_label_listen_port), f.listenPortEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_public_endpoint), f.publicEndpointEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_mtu), f.mtuEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_post_up), f.postUpEntry),
		widget.NewFormItem(f.lang.T(lang.Tunnel_label_post_down), f.postDownEntry),
	)

	saveBtn := widget.NewButtonWithIcon(f.lang.T(lang.Common_save), theme.DocumentSaveIcon(), func() {
		cfg, errs := f.validate()
		if len(errs) > 0 {
			helpers.ShowError(errs[0], win)
			return
		}

		name := f.getTunnelName()
		if err := f.onSave(name, cfg); err != nil {
			helpers.ShowError(err, win)
			return
		}
		f.generateClientConfigs(name, cfg, win)
		win.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton(f.lang.T(lang.Common_cancel), func() {
		if f.onCancel != nil {
			f.onCancel()
		}
		win.Close()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)

	content := container.NewBorder(
		nil,
		container.NewPadded(buttons),
		nil, nil,
		container.NewVScroll(container.NewVBox(
			container.NewPadded(widget.NewCard(f.lang.T(lang.Tunnel_tab_interface), "", interfaceForm)),
		)),
	)

	win.SetContent(container.NewPadded(content))
	win.Show()
}

// ShowPeers displays the peers editing form
func (f *TunnelForm) ShowPeers() {
	title := f.lang.T(lang.Tunnel_peers_window_title)
	if f.isEdit {
		title = f.lang.Tf(lang.Tunnel_edit_peer_title, f.name)
	}

	win := fyne.CurrentApp().NewWindow(title)
	win.Resize(fyne.NewSize(700, 750))

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
			pubKeyStr := peer.PublicKey.String()
			displayKey := pubKeyStr
			if len(displayKey) > 12 {
				displayKey = displayKey[:12]
			}
			peerName := peer.Name
			if peerName == "" {
				peerName = f.lang.T(lang.Tunnel_peer_list_label)
			}
			label.SetText(fmt.Sprintf("%s... - %s", displayKey, peerName))

			editBtn.OnTapped = func() {
				peerCopy := f.peers[id]
				oldPubKey := peerCopy.PublicKey.String()
				peerForm := NewPeerForm(&peerCopy, func(p config.PeerConfig, privateKey string) {
					delete(f.peerPrivateKeys, oldPubKey)
					f.peers[id] = p
					if privateKey != "" {
						f.peerPrivateKeys[p.PublicKey.String()] = privateKey
					}
					f.peersList.Refresh()
				}, nil, f.lang)
				peerForm.Show(win)
			}

			deleteBtn.OnTapped = func() {
				helpers.ShowConfirm(f.lang.T(lang.Tunnel_delete_peer_title), f.lang.T(lang.Tunnel_delete_peer_msg), func(yes bool) {
					if yes {
						pubKey := f.peers[id].PublicKey.String()
						delete(f.peerPrivateKeys, pubKey)
						f.peers = append(f.peers[:id], f.peers[id+1:]...)
						f.peersList.Refresh()
					}
				}, win)
			}
		},
	)

	addPeerBtn := widget.NewButtonWithIcon(f.lang.T(lang.Tunnel_btn_add_peer), theme.ContentAddIcon(), func() {
		peerForm := NewPeerForm(nil, func(p config.PeerConfig, privateKey string) {
			f.peers = append(f.peers, p)
			if privateKey != "" {
				f.peerPrivateKeys[p.PublicKey.String()] = privateKey
			}
			f.peersList.Refresh()
		}, nil, f.lang)
		peerForm.Show(win)
	})

	sizedList := container.NewGridWrap(
		fyne.NewSize(560, float32(f.peersList.Length()*50)),
		f.peersList,
	)

	peersSection := container.NewBorder(
		widget.NewLabel(f.lang.T(lang.Tunnel_peer_list_label)),
		addPeerBtn,
		nil, nil,
		sizedList,
	)

	saveBtn := widget.NewButtonWithIcon(f.lang.T(lang.Common_save), theme.DocumentSaveIcon(), func() {
		cfg, errs := f.validate()
		if len(errs) > 0 {
			helpers.ShowError(errs[0], win)
			return
		}

		name := f.getTunnelName()
		if err := f.onSave(name, cfg); err != nil {
			helpers.ShowError(err, win)
			return
		}
		f.generateClientConfigs(name, cfg, win)
		win.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton(f.lang.T(lang.Common_cancel), func() {
		if f.onCancel != nil {
			f.onCancel()
		}
		win.Close()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)

	content := container.NewBorder(
		nil,
		container.NewPadded(buttons),
		nil, nil,
		container.NewVScroll(container.NewVBox(
			container.NewPadded(widget.NewCard(f.lang.T(lang.Tunnel_tab_peers), "", peersSection)),
		)),
	)

	win.SetContent(container.NewPadded(content))
	win.Show()
}

func (f *TunnelForm) generateClientConfigs(tunnelName string, serverCfg *config.Config, win fyne.Window) {
	if len(f.peerPrivateKeys) == 0 {
		return
	}

	if f.publicEndpointEntry.Text == "" {
		helpers.ShowInformation(f.lang.T(lang.Tunnel_client_configs_title),
			f.lang.T(lang.Tunnel_client_configs_no_endpoint), win)
		return
	}

	endpoint := f.publicEndpointEntry.Text
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		helpers.ShowError(fmt.Errorf(f.lang.T(lang.Tunnel_error_invalid_endpoint), err), win)
		return
	}
	port, _ := strconv.Atoi(portStr)

	clientDir := filepath.Join(f.clientConfigDir, tunnelName)
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		helpers.ShowError(fmt.Errorf(f.lang.T(lang.Tunnel_error_create_client_dir), err), win)
		return
	}

	clientCtrl := wg.New(clientDir)
	var generated []string

	for _, peer := range serverCfg.Peers {
		pubKeyStr := peer.PublicKey.String()
		privKey, ok := f.peerPrivateKeys[pubKeyStr]
		if !ok {
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
			Endpoint:            fmt.Sprintf("%s:%d", host, port),
			PersistentKeepalive: peer.PersistentKeepalive,
		}

		clientPath := filepath.Join(clientDir, peer.Name+".conf")
		_, err := os.Stat(clientPath)
		if os.IsExist(err) {
			os.Rename(clientPath, clientPath+".bkp")
		}
		_, err = clientCtrl.NewClientConfig(*serverCfg, peerOpts, privKey, true)
		if err != nil {
			helpers.ShowError(fmt.Errorf(f.lang.Tf(lang.Tunnel_error_client_config, peer.Name), err), win)
			continue
		}
		generated = append(generated, clientPath)
	}

	if len(generated) > 0 {
		absDir, _ := filepath.Abs(clientDir)
		msg := f.lang.Tf(lang.Tunnel_client_configs_success, len(generated), absDir)
		helpers.ShowInformation(f.lang.T(lang.Tunnel_client_configs_title), msg, win)
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

	if f.publicEndpointEntry.Text != "" {
		if !wg.ValidateEndpoint(f.publicEndpointEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "PublicEndpoint", Message: f.lang.T(lang.Error_invalid_endpoint)}}
		}
		cfg.PublicEndpoint = f.publicEndpointEntry.Text
		for i := range cfg.Peers {
			if cfg.Peers[i].Endpoint == "" {
				cfg.Peers[i].Endpoint = cfg.PublicEndpoint
			}
		}
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
