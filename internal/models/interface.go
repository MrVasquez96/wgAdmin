package models

// Interface represents a WireGuard interface
type Interface struct {
	Name   string
	IP     string
	Active bool
}
