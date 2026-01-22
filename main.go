package main

import (
	"fmt"
	"os"

	"wgAdmin/internal/ui"
)

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Warning: This application typically requires root privileges for WireGuard operations.")
		fmt.Fprintln(os.Stderr, "Some features may not work correctly without elevated permissions.")
		fmt.Fprintln(os.Stderr, "")
	}

	app := ui.NewApp()
	app.Run()
}
