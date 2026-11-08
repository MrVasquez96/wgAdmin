package ui

import (
	"os"

	"wgAdmin/internal/wgstat"

	"github.com/MrVasquez96/go-wg/wg"
	"github.com/MrVasquez96/go-wg/wg/config"
)

// Data sources for MainView. The app always uses these defaults; cmd/preview
// swaps them for fakes so the console can be developed and screenshotted with
// live-looking tunnels without root or a real WireGuard setup.
var (
	ListInterfaces = func(ctrl *wg.WG) ([]config.Interface, error) { return ctrl.ListInterfaces() }
	ReadSnapshot   = wgstat.Read
	IsRoot         = func() bool { return os.Geteuid() == 0 }
)
