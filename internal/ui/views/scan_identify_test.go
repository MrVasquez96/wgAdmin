package ui

import "testing"

func TestDescribeSSHBanner(t *testing.T) {
	tests := []struct {
		banner string
		want   string
	}{
		// Banners seen on the wgFRS networks.
		{"SSH-2.0-OpenSSH_9.2p1 Debian-2+deb12u10", "Debian 12"},
		{"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.16", "Ubuntu"},

		{"SSH-2.0-OpenSSH_9.6p1 Ubuntu-3ubuntu13.5", "Ubuntu"},
		{"SSH-2.0-OpenSSH_7.9p1 Raspbian-10+deb10u2", "Raspberry Pi OS"},
		{"SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u3", "Debian 11"},
		{"SSH-2.0-dropbear_2020.81", "dropbear"},
		{"SSH-2.0-OpenSSH_9.6", ""},
		{"SSH-2.0-OpenSSH_for_Windows_9.5", ""},
		{"not an ssh banner", ""},
		{"", ""},
	}

	for _, tc := range tests {
		if got := describeSSHBanner(tc.banner); got != tc.want {
			t.Errorf("describeSSHBanner(%q) = %q, want %q", tc.banner, got, tc.want)
		}
	}
}

func TestIdentifyHostSkipsHostsWithoutSSH(t *testing.T) {
	if got := identifyHost("10.255.255.1", []int{80, 443}, 0); got != "" {
		t.Errorf("identifyHost without port 22 = %q, want empty", got)
	}
}
