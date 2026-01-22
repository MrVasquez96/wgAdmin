package models

// ScanResult represents the result of scanning a single IP
type ScanResult struct {
	IP        string
	Hostnames []string
	PortsOpen []int
}
