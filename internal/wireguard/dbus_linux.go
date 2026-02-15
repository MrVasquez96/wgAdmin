//go:build linux

package wireguard

import (
	"github.com/godbus/dbus/v5"
)

// dbusCon establishes a connection to the system D-Bus.
func dbusCon() (*dbus.Conn, error) {
	return dbus.SystemBus()
}
