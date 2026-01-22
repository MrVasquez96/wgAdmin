package scanner

import (
	"net"
	"sync"
	"sync/atomic"

	"wgAdmin/internal/models"
)

const (
	WorkerCount = 100
)

// Scanner performs network scans on a CIDR range
type Scanner struct {
	cidr         string
	Results      chan models.ScanResult
	wg           sync.WaitGroup
	scannedCount uint32
	TotalIPs     uint32
}

// NewScanner creates a new scanner for the given CIDR
func NewScanner(cidr string) (*Scanner, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var totalIPs uint32
	for i := ip.Mask(ipnet.Mask); ipnet.Contains(i); inc(i) {
		totalIPs++
	}

	return &Scanner{
		cidr:     cidr,
		TotalIPs: totalIPs,
		Results:  make(chan models.ScanResult, WorkerCount),
	}, nil
}

// Run starts the scan
func (s *Scanner) Run() {
	jobs := make(chan string)

	for i := 0; i < WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker(jobs)
	}

	ip, ipnet, _ := net.ParseCIDR(s.cidr)
	for i := ip.Mask(ipnet.Mask); ipnet.Contains(i); inc(i) {
		jobs <- i.String()
	}
	close(jobs)

	s.wg.Wait()
	close(s.Results)
}

// Progress returns the current scan progress
func (s *Scanner) Progress() (scanned, total uint32) {
	return atomic.LoadUint32(&s.scannedCount), s.TotalIPs
}

func (s *Scanner) worker(jobs <-chan string) {
	defer s.wg.Done()
	for ip := range jobs {
		s.processIP(ip)
		atomic.AddUint32(&s.scannedCount, 1)
	}
}

func (s *Scanner) processIP(ip string) {
	ports := CheckWebPorts(ip)
	if len(ports) == 0 {
		return
	}

	hostnames, _ := net.LookupAddr(ip)

	s.Results <- models.ScanResult{
		IP:        ip,
		Hostnames: hostnames,
		PortsOpen: ports,
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// DeriveCIDR24 converts an IP to a /24 CIDR
func DeriveCIDR24(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() == nil {
		return ""
	}
	ipv4 := parsed.To4()
	return net.IPv4(ipv4[0], ipv4[1], ipv4[2], 0).String() + "/24"
}
