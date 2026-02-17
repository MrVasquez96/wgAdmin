package views

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// TunnelForm handles tunnel creation and editing
type TunnelForm struct {
	window fyne.Window
	isEdit bool
	name   string

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
	peerPrivateKeys map[string]string // peer public key -> peer private key (for client config generation)

	// Callbacks
	onSave   func(name string, config *config.Config) error
	onCancel func()
}

// NewTunnelForm creates a new tunnel form
func NewTunnelForm(parent fyne.Window, existingName string, existingConfig *config.Config, onSave func(string, *config.Config) error, onCancel func()) *TunnelForm {
	tunnelName := existingName
	isEdit := existingName != ""
	if isEdit {
		tunnelName = existingConfig.Name
	}

	f := &TunnelForm{
		window:              parent,
		isEdit:              existingName != "",
		name:                tunnelName,
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

	f.nameEntry.SetPlaceHolder("e.g., wg0")
	f.privateKeyEntry.SetPlaceHolder("Base64 encoded private key")
	f.addressEntry.SetPlaceHolder("e.g., 10.0.0.1/24")
	f.dnsEntry.SetPlaceHolder("e.g., 1.1.1.1 (optional)")
	f.listenPortEntry.SetPlaceHolder("e.g., 51820 (optional)")
	f.publicEndpointEntry.SetPlaceHolder("e.g., vpn.example.com:51820 (for client configs)")
	f.mtuEntry.SetPlaceHolder("e.g., 1420 (optional)")
	f.postUpEntry.SetPlaceHolder("e.g., iptables -A FORWARD -i %i -j ACCEPT (optional)")
	f.postDownEntry.SetPlaceHolder("e.g., iptables -D FORWARD -i %i -j ACCEPT (optional)")
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

		// Convert addresses to string
		addrs := make([]string, len(existingConfig.Interface.Address))
		for i, addr := range existingConfig.Interface.Address {
			addrs[i] = addr.String()
		}
		f.addressEntry.SetText(strings.Join(addrs, ", "))

		// Convert DNS to string
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

	// Update public key when private key changes
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
		f.publicKeyLabel.SetText("(invalid key)")
	} else {
		f.publicKeyLabel.SetText(pubKey)
	}
}

// getTunnelName returns the clean tunnel name (e.g., "wg0") from the form state.
func (f *TunnelForm) getTunnelName() string {
	if !f.isEdit {
		return f.nameEntry.Text
	}
	// In edit mode, f.name contains the config comment name like "[wg0] Public key: ..."
	// Extract the raw name from it
	if i := strings.Index(f.name, "]"); i > 0 && strings.HasPrefix(f.name, "[") {
		return f.name[1:i]
	}
	// Fallback: use the disabled name entry text
	return f.nameEntry.Text
}

// Show displays the tunnel form
func (f *TunnelForm) Show() {
	title := "Add Tunnel"
	if f.isEdit {
		title = "Edit Tunnel: " + f.name
	}

	win := fyne.CurrentApp().NewWindow(title)
	win.Resize(fyne.NewSize(600, 700))

	// Interface section
	generateKeyBtn := widget.NewButtonWithIcon("Generate Keys", theme.ViewRefreshIcon(), func() {
		priv, pub, err := wg.GenerateKeyPair()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		f.privateKeyEntry.SetText(priv)
		f.publicKeyLabel.SetText(pub)
	})

	copyPubKeyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		if f.publicKeyLabel.Text != "" && f.publicKeyLabel.Text != "(invalid key)" {
			win.Clipboard().SetContent(f.publicKeyLabel.Text)
		}
	})

	interfaceForm := container.NewVBox(
		widget.NewLabel("Tunnel Name *"),
		f.nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Private Key *"),
		container.NewBorder(nil, nil, nil, generateKeyBtn, f.privateKeyEntry),
		widget.NewLabel("Public Key (derived):"),
		container.NewBorder(nil, nil, nil, copyPubKeyBtn, f.publicKeyLabel),
		widget.NewSeparator(),
		widget.NewLabel("Address (CIDR) *"),
		f.addressEntry,
		widget.NewLabel("DNS"),
		f.dnsEntry,
		widget.NewLabel("Listen Port"),
		f.listenPortEntry,
		widget.NewLabel("Public Endpoint (for client configs)"),
		f.publicEndpointEntry,
		widget.NewLabel("MTU"),
		f.mtuEntry,
		widget.NewSeparator(),
		widget.NewLabel("PostUp"),
		container.NewGridWrap(fyne.NewSize(560, 60), f.postUpEntry),
		widget.NewLabel("PostDown"),
		container.NewGridWrap(fyne.NewSize(560, 60), f.postDownEntry),
	)

	// Peers section
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
				peerName = "(unnamed)"
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
				}, nil)
				peerForm.Show(win)
			}

			deleteBtn.OnTapped = func() {
				dialog.ShowConfirm("Delete Peer", "Remove this peer?", func(yes bool) {
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
	addPeerBtn := widget.NewButtonWithIcon("Add Peer", theme.ContentAddIcon(), func() {
		peerForm := NewPeerForm(nil, func(p config.PeerConfig, privateKey string) {
			f.peers = append(f.peers, p)
			if privateKey != "" {
				f.peerPrivateKeys[p.PublicKey.String()] = privateKey
			}
			f.peersList.Refresh()
		}, nil)
		peerForm.Show(win)
	})
	sizedList := container.NewGridWrap(
		fyne.NewSize(560, float32(f.peersList.Length()*50)),
		f.peersList,
	)

	peersSection := container.NewBorder(
		widget.NewLabel("Peers"),
		addPeerBtn,
		nil, nil,
		sizedList,
	)
	// Action buttons
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		cfg, errs := f.validate()
		if len(errs) > 0 {
			dialog.ShowError(errs[0], win)
			return
		}

		name := f.getTunnelName()
		if err := f.onSave(name, cfg); err != nil {
			fmt.Println("error saving", err)
			dialog.ShowError(err, win)
			return
		}
		// Generate client configs for peers with private keys
		f.generateClientConfigs(name, cfg, win)

		win.Close()
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		if f.onCancel != nil {
			f.onCancel()
		}
		win.Close()
	})

	buttons := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)

	// Main layout
	content := container.NewBorder(
		nil,
		buttons,
		nil, nil,
		container.NewVScroll(container.NewVBox(
			widget.NewCard("Interface Configuration", "", interfaceForm),
			widget.NewCard("Peers", "", peersSection),
		)),
	)

	win.SetContent(container.NewPadded(content))
	win.Show()
}

// generateClientConfigs creates client .conf files for each peer that has a generated private key.
// Configs are saved to ./clients/{tunnelName}/{peerName}.conf relative to the project root.
func (f *TunnelForm) generateClientConfigs(tunnelName string, serverCfg *config.Config, win fyne.Window) {
	if len(f.peerPrivateKeys) == 0 {
		return
	}

	if f.publicEndpointEntry.Text == "" {
		// Can't generate client configs without a server endpoint
		dialog.ShowInformation("Client Configs",
			"No public endpoint set - skipping client config generation.\nSet the Public Endpoint field to generate client configs.", win)
		fmt.Println("Empty endpoint")
		return
	}

	endpoint := f.publicEndpointEntry.Text
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		fmt.Println("invalid endpoint")
		dialog.ShowError(fmt.Errorf("invalid public endpoint for client configs: %w", err), win)
		return
	}
	port, _ := strconv.Atoi(portStr)

	// Save client configs to ./clients/{tunnelName}/ relative to project root
	clientDir := filepath.Join("clients", tunnelName)
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("failed to create client config directory: %w", err), win)
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
			EndpointIP:          host,
			EndpointPort:        port,
			PersistentKeepalive: int(peer.PersistentKeepalive.Seconds()),
		}

		// Rename existing client config to allow overwrite
		clientPath := filepath.Join(clientDir, peer.Name+".conf")
		os.Rename(clientPath, clientPath+".bkp")
		_, err := clientCtrl.NewClientConfig(*serverCfg, peerOpts, privKey, true)
		if err != nil {
			dialog.ShowError(fmt.Errorf("client config for '%s': %w", peer.Name, err), win)
			continue
		}
		generated = append(generated, clientPath)
	}

	if len(generated) > 0 {
		absDir, _ := filepath.Abs(clientDir)
		msg := fmt.Sprintf("Generated %d client config(s) in:\n%s", len(generated), absDir)
		dialog.ShowInformation("Client Configs", msg, win)
	}
}

func (f *TunnelForm) validate() (*config.Config, []error) {
	cfg := &config.Config{
		Interface: config.InterfaceConfig{
			MTU:   1420,
			Table: "auto",
		},
		Peers: f.peers,
	}

	// Parse PrivateKey
	if f.privateKeyEntry.Text == "" {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: "required"}}
	}
	if !wg.ValidateKey(f.privateKeyEntry.Text) {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: "invalid format"}}
	}
	privKey, err := wgtypes.ParseKey(f.privateKeyEntry.Text)
	if err != nil {
		return nil, []error{wg.ValidationError{Field: "PrivateKey", Message: "invalid key"}}
	}
	cfg.Interface.PrivateKey = privKey

	// Parse Address
	if f.addressEntry.Text == "" {
		return nil, []error{wg.ValidationError{Field: "Address", Message: "required"}}
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
			return nil, []error{wg.ValidationError{Field: "Address", Message: "invalid CIDR format"}}
		}
		cfg.Interface.Address = append(cfg.Interface.Address, *ipNet)
	}

	// Parse DNS
	if f.dnsEntry.Text != "" {
		for _, dns := range strings.Split(f.dnsEntry.Text, ",") {
			dns = strings.TrimSpace(dns)
			if dns == "" {
				continue
			}
			ip := net.ParseIP(dns)
			if ip == nil {
				return nil, []error{wg.ValidationError{Field: "DNS", Message: fmt.Sprintf("invalid DNS address: %s", dns)}}
			}
			cfg.Interface.DNS = append(cfg.Interface.DNS, ip)
		}
	}

	// Parse ListenPort
	if f.listenPortEntry.Text != "" {
		port, err := strconv.Atoi(f.listenPortEntry.Text)
		if err != nil {
			return nil, []error{wg.ValidationError{Field: "ListenPort", Message: "must be a number"}}
		}
		if port < 0 || port > 65535 {
			return nil, []error{wg.ValidationError{Field: "ListenPort", Message: "must be 0-65535"}}
		}
		cfg.Interface.ListenPort = &port
	}

	// Parse and store Public Endpoint (optional)
	if f.publicEndpointEntry.Text != "" {
		if !wg.ValidateEndpoint(f.publicEndpointEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "PublicEndpoint", Message: "invalid format (host:port)"}}
		}
		cfg.PublicEndpoint = f.publicEndpointEntry.Text
	}

	// Parse MTU
	if f.mtuEntry.Text != "" {
		mtu, err := strconv.Atoi(f.mtuEntry.Text)
		if err != nil {
			return nil, []error{wg.ValidationError{Field: "MTU", Message: "must be a number"}}
		}
		cfg.Interface.MTU = mtu
	}

	// PostUp / PostDown
	cfg.Interface.PostUp = strings.TrimSpace(f.postUpEntry.Text)
	cfg.Interface.PostDown = strings.TrimSpace(f.postDownEntry.Text)

	// Validate name for new tunnels
	if !f.isEdit {
		if !wg.ValidateName(f.nameEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "Name", Message: "must be 1-15 alphanumeric characters"}}
		}
		if wgController.ConfigExists(f.nameEntry.Text) {
			return nil, []error{wg.ValidationError{Field: "Name", Message: "tunnel already exists"}}
		}
	}

	errs := wg.ValidateConfig(cfg)
	return cfg, errs
}
