package ui

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// Hosts behind the tunnel have no reverse DNS and do not answer mDNS, so the
// SSH banner is the one identifying string they hand out for free. It names
// the distribution a device runs, which in practice separates the Raspberry Pi
// units from the Ubuntu machines.

// Debian stamps its release into the package version ("Debian-2+deb12u10");
// Ubuntu's banner only carries a package revision, so it stays unversioned.
var debianRelease = regexp.MustCompile(`deb(?:ian)?[-_]?(\d+)`)

// sshBanner reads the identification string an SSH server sends on connect.
func sshBanner(ip string, port int, timeout time.Duration) string {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, fmt.Sprint(port)), timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(timeout))
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil && line == "" {
		return ""
	}
	return strings.TrimSpace(line)
}

// describeSSHBanner turns an SSH identification string into a short label,
// e.g. "SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u10" -> "Debian 12".
func describeSSHBanner(banner string) string {
	banner = strings.TrimSpace(banner)
	if !strings.HasPrefix(banner, "SSH-") {
		return ""
	}

	// SSH-<protocol>-<software> [comments]
	parts := strings.SplitN(banner, "-", 3)
	if len(parts) < 3 {
		return ""
	}
	rest := parts[2]

	switch {
	case strings.Contains(rest, "Raspbian"):
		return "Raspberry Pi OS"
	case strings.Contains(rest, "Ubuntu"):
		return "Ubuntu"
	case debianRelease.MatchString(rest):
		return "Debian " + debianRelease.FindStringSubmatch(rest)[1]
	case strings.Contains(rest, "Debian"):
		return "Debian"
	}

	// Unknown flavour: fall back to the software name (dropbear on embedded
	// gear, OpenSSH_for_Windows, ...), which still tells the devices apart.
	software := strings.Fields(rest)[0]
	if i := strings.IndexAny(software, "_-"); i > 0 {
		software = software[:i]
	}
	if software == "" || software == "OpenSSH" {
		return ""
	}
	return software
}

// identifyHost returns a short description of what a host appears to be,
// derived from its SSH banner. Empty when the host gives nothing away.
func identifyHost(ip string, openPorts []int, timeout time.Duration) string {
	for _, p := range openPorts {
		if p != 22 {
			continue
		}
		return describeSSHBanner(sshBanner(ip, p, timeout))
	}
	return ""
}
