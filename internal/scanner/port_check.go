package scanner

import (
	"fmt"
	"net"
	"time"
)

const (
	DialTimeout = 2 * time.Second
)

// IsPortOpen checks if a TCP port is open on the given IP
func IsPortOpen(ip string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), DialTimeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// CheckWebPorts returns a list of open web ports (80, 443)
func CheckWebPorts(ip string) []int {
	var open []int
	ports := []int{80, 443}

	for _, port := range ports {
		if IsPortOpen(ip, port) {
			open = append(open, port)
		}
	}
	return open
}
