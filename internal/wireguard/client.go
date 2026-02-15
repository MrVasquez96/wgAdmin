package wireguard

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"wgAdmin/internal/models"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// LogLevel represents the verbosity of logging.
type LogLevel int

const (
	LogLevelSilent LogLevel = iota
	LogLevelError
	LogLevelVerbose
)

// Client represents a WireGuard VPN client.
type Client struct {
	interfaceName string
	config        *models.Config
	logLevel      LogLevel

	tunDevice    tun.Device
	wgDevice     *device.Device
	wgClient     *wgctrl.Client
	netConfig    *NetworkConfig
	uapiListener net.Listener

	mu      sync.Mutex
	running bool
	done    chan struct{}
}

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

// WithLogLevel sets the logging verbosity.
func WithLogLevel(level LogLevel) ClientOption {
	return func(c *Client) {
		c.logLevel = level
	}
}

// WithInterfaceName sets a custom interface name.
func WithInterfaceName(name string) ClientOption {
	return func(c *Client) {
		c.interfaceName = name
	}
}

// NewClient creates a new WireGuard client from a configuration.
func NewClient(config *models.Config, opts ...ClientOption) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	c := &Client{
		interfaceName: "wg0",
		config:        config,
		logLevel:      LogLevelError,
		done:          make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// NewClientFromFile creates a new WireGuard client from a configuration file.
func NewClientFromFile(configPath string, opts ...ClientOption) (*Client, error) {
	config, err := ParseConfig(configPath)
	if err != nil {
		return nil, err
	}
	return NewClient(config, opts...)
}

// Start initializes and starts the WireGuard tunnel.
func (c *Client) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("client is already running")
	}
	c.running = true
	c.mu.Unlock()

	// Execute PreUp script if defined
	if c.config.Interface.PreUp != "" {
		if err := c.executeScript("PreUp", c.config.Interface.PreUp); err != nil {
			c.cleanup()
			return err
		}
	}

	// Step 1: Create TUN device
	mtu := c.config.Interface.MTU
	if mtu == 0 {
		mtu = device.DefaultMTU
	}

	tunDev, err := tun.CreateTUN(c.interfaceName, mtu)
	if err != nil {
		c.cleanup()
		return fmt.Errorf("failed to create TUN device: %w", err)
	}
	c.tunDevice = tunDev

	// Get actual interface name (may differ on some platforms)
	actualName, err := tunDev.Name()
	if err != nil {
		c.cleanup()
		return fmt.Errorf("failed to get TUN device name: %w", err)
	}
	c.interfaceName = actualName

	// Step 2: Create WireGuard device
	var logger *device.Logger
	switch c.logLevel {
	case LogLevelSilent:
		logger = device.NewLogger(device.LogLevelSilent, fmt.Sprintf("(%s) ", c.interfaceName))
	case LogLevelError:
		logger = device.NewLogger(device.LogLevelError, fmt.Sprintf("(%s) ", c.interfaceName))
	case LogLevelVerbose:
		logger = device.NewLogger(device.LogLevelVerbose, fmt.Sprintf("(%s) ", c.interfaceName))
	}

	c.wgDevice = device.NewDevice(tunDev, conn.NewDefaultBind(), logger)
	fileUAPI, err := ipc.UAPIOpen(c.interfaceName)
	if err != nil {
		c.cleanup()
		return fmt.Errorf("failed to open UAPI socket: %w", err)
	}

	uapiListener, err := ipc.UAPIListen(c.interfaceName, fileUAPI)
	if err != nil {
		c.cleanup()
		return fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	// Store this in your Client struct so you can close it in c.cleanup()
	c.uapiListener = uapiListener

	// Handle UAPI requests in the background so wgctrl can communicate with the device
	go func() {
		for {
			conn, err := c.uapiListener.Accept()
			if err != nil {
				// Listener was closed, safely exit the goroutine
				return
			}
			go c.wgDevice.IpcHandle(conn)
		}
	}()
	// Step 3: Configure WireGuard via wgctrl
	wgClient, err := wgctrl.New()
	if err != nil {
		c.cleanup()
		return fmt.Errorf("failed to create wgctrl client: %w", err)
	}
	c.wgClient = wgClient

	// Build peer configurations
	var peers []wgtypes.PeerConfig
	for _, p := range c.config.Peers {
		peerCfg := wgtypes.PeerConfig{
			PublicKey:         p.PublicKey,
			Endpoint:          p.Endpoint,
			ReplaceAllowedIPs: true,
			AllowedIPs:        p.AllowedIPs,
		}

		if p.PresharedKey != nil {
			peerCfg.PresharedKey = p.PresharedKey
		}

		if p.PersistentKeepalive > 0 {
			keepalive := p.PersistentKeepalive
			peerCfg.PersistentKeepaliveInterval = &keepalive
		}

		peers = append(peers, peerCfg)
	}

	// Build interface configuration
	wgConfig := wgtypes.Config{
		PrivateKey:   &c.config.Interface.PrivateKey,
		ReplacePeers: true,
		Peers:        peers,
	}

	if c.config.Interface.ListenPort != nil {
		wgConfig.ListenPort = c.config.Interface.ListenPort
	}

	if c.config.Interface.FwMark != nil {
		wgConfig.FirewallMark = c.config.Interface.FwMark
	}

	// Apply WireGuard configuration
	if err := c.wgClient.ConfigureDevice(c.interfaceName, wgConfig); err != nil {
		c.cleanup()
		return fmt.Errorf("failed to configure WireGuard device: %w", err)
	}

	// Step 4: Configure OS-level networking (interface up, IPs, routes)
	c.netConfig = NewNetworkConfig(c.interfaceName, c.config)
	if err := c.netConfig.Apply(); err != nil {
		c.cleanup()
		return fmt.Errorf("failed to apply network configuration: %w", err)
	}

	// Execute PostUp script if defined
	if c.config.Interface.PostUp != "" {
		if err := c.executeScript("PostUp", c.config.Interface.PostUp); err != nil {
			c.cleanup()
			return err
		}
	}

	// Start device
	c.wgDevice.Up()

	return nil
}

// Stop gracefully shuts down the WireGuard tunnel.
func (c *Client) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	close(c.done)

	// Execute PreDown script if defined
	if c.config.Interface.PreDown != "" {
		if err := c.executeScript("PreDown", c.config.Interface.PreDown); err != nil {
			// Log but continue with cleanup
			fmt.Fprintf(os.Stderr, "PreDown script failed: %v\n", err)
		}
	}

	if err := c.cleanup(); err != nil {
		return err
	}

	// Execute PostDown script if defined
	if c.config.Interface.PostDown != "" {
		if err := c.executeScript("PostDown", c.config.Interface.PostDown); err != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "PostDown script failed: %v\n", err)
		}
	}

	return nil
}

func (c *Client) cleanup() error {
	var errs []error

	// Remove network configuration
	if c.netConfig != nil {
		if err := c.netConfig.Remove(); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove network config: %w", err))
		}
		c.netConfig = nil
	}

	// Close WireGuard device
	if c.wgDevice != nil {
		c.wgDevice.Close()
		c.wgDevice = nil
	}

	// Close wgctrl client
	if c.wgClient != nil {
		c.wgClient.Close()
		c.wgClient = nil
	}

	if c.uapiListener != nil {
		c.uapiListener.Close()
	}
	// Close TUN device
	if c.tunDevice != nil {
		if err := c.tunDevice.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close TUN device: %w", err))
		}
		c.tunDevice = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}
	return nil
}

// Wait blocks until the client is stopped.
func (c *Client) Wait() {
	<-c.done
}

// IsRunning returns whether the client is currently running.
func (c *Client) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}

// InterfaceName returns the name of the WireGuard interface.
func (c *Client) InterfaceName() string {
	return c.interfaceName
}

// Status returns the current status of the WireGuard device.
func (c *Client) Status() (*wgtypes.Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running || c.wgClient == nil {
		return nil, fmt.Errorf("client is not running")
	}

	return c.wgClient.Device(c.interfaceName)
}

// Stats returns transfer statistics for the tunnel.
func (c *Client) Stats() ([]PeerStats, error) {
	dev, err := c.Status()
	if err != nil {
		return nil, err
	}

	var stats []PeerStats
	for _, peer := range dev.Peers {
		s := PeerStats{
			PublicKey:        peer.PublicKey,
			Endpoint:         peer.Endpoint,
			LastHandshake:    peer.LastHandshakeTime,
			BytesReceived:    peer.ReceiveBytes,
			BytesTransmitted: peer.TransmitBytes,
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// PeerStats contains statistics for a single peer.
type PeerStats struct {
	PublicKey        wgtypes.Key
	Endpoint         *net.UDPAddr
	LastHandshake    time.Time
	BytesReceived    int64
	BytesTransmitted int64
}

func (c *Client) executeScript(name, script string) error {
	// Replace %i with interface name
	script = replaceInterfacePlaceholder(script, c.interfaceName)

	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WIREGUARD_INTERFACE=%s", c.interfaceName),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s script failed: %w", name, err)
	}
	return nil
}

func replaceInterfacePlaceholder(script, ifaceName string) string {
	// wg-quick style %i replacement
	result := ""
	for i := 0; i < len(script); i++ {
		if i+1 < len(script) && script[i] == '%' && script[i+1] == 'i' {
			result += ifaceName
			i++ // skip 'i'
		} else {
			result += string(script[i])
		}
	}
	return result
}
