package ui

import (
	"net"
	"reflect"
	"testing"

	"github.com/MrVasquez96/go-wg/wg/config"
)

func mustCIDR(t *testing.T, s string) net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatalf("bad cidr %s: %v", s, err)
	}
	return *n
}

func TestPlanScan(t *testing.T) {
	tests := []struct {
		name       string
		addresses  []string
		allowedIPs []string
		fallbackIP string
		want       ScanPlan
	}{
		{
			// The real wgPersonell config: a /32 address and two /17 routes.
			name:       "hub and spoke: wide peer routes get a discovery pass",
			addresses:  []string{"10.0.144.1/32"},
			allowedIPs: []string{"10.0.128.0/17", "10.1.0.0/17"},
			want: ScanPlan{
				Direct: []string{"10.0.144.0/24"},
				Wide:   []string{"10.0.128.0/17", "10.1.0.0/17"},
			},
		},
		{
			name:       "site routes small enough to scan directly",
			addresses:  []string{"10.0.144.1/32"},
			allowedIPs: []string{"10.0.144.0/24", "10.1.0.0/24", "10.1.32.0/24", "10.1.64.0/24"},
			want: ScanPlan{
				Direct: []string{"10.0.144.0/24", "10.1.0.0/24", "10.1.32.0/24", "10.1.64.0/24"},
			},
		},
		{
			name:       "flat setup keeps the old single /24",
			addresses:  []string{"10.0.144.1/24"},
			allowedIPs: []string{"10.0.144.2/32"},
			want:       ScanPlan{Direct: []string{"10.0.144.0/24"}},
		},
		{
			name:       "default route is skipped, falls back to local /24",
			addresses:  []string{"10.123.3.7/32"},
			allowedIPs: []string{"0.0.0.0/0"},
			fallbackIP: "10.123.3.7",
			want: ScanPlan{
				Direct:  []string{"10.123.3.0/24"},
				Skipped: []string{"0.0.0.0/0"},
			},
		},
		{
			name:       "direct networks covered by a wider one are dropped",
			addresses:  []string{"10.1.0.1/24"},
			allowedIPs: []string{"10.1.0.0/22", "10.1.1.0/24"},
			want:       ScanPlan{Direct: []string{"10.1.0.0/22"}},
		},
		{
			name:       "IPv6 is ignored",
			addresses:  []string{"10.0.144.1/24", "fd00::1/64"},
			allowedIPs: []string{"10.1.0.0/24"},
			want:       ScanPlan{Direct: []string{"10.0.144.0/24", "10.1.0.0/24"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			for _, a := range tc.addresses {
				cfg.Interface.Address = append(cfg.Interface.Address, mustCIDR(t, a))
			}
			peer := config.PeerConfig{}
			for _, a := range tc.allowedIPs {
				peer.AllowedIPs = append(peer.AllowedIPs, mustCIDR(t, a))
			}
			cfg.Peers = append(cfg.Peers, peer)

			got := PlanScan(cfg, tc.fallbackIP)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("PlanScan() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestSubnetsOf(t *testing.T) {
	subnets := subnetsOf("10.1.0.0/17")
	if len(subnets) != 128 {
		t.Fatalf("got %d subnets, want 128", len(subnets))
	}
	if subnets[0] != "10.1.0.0/24" || subnets[32] != "10.1.32.0/24" || subnets[127] != "10.1.127.0/24" {
		t.Errorf("unexpected subnets: %s ... %s ... %s", subnets[0], subnets[32], subnets[127])
	}

	if got := subnetsOf("10.1.32.0/24"); !reflect.DeepEqual(got, []string{"10.1.32.0/24"}) {
		t.Errorf("subnetsOf(/24) = %v", got)
	}
	if got := subnetsOf("10.0.0.0/8"); got != nil {
		t.Errorf("subnetsOf(/8) = %v, want nil (too many subnets)", got)
	}
}

func TestTotalHosts(t *testing.T) {
	if got := totalHosts([]string{"10.0.144.0/24", "10.1.0.0/24"}); got != 512 {
		t.Errorf("totalHosts = %d, want 512", got)
	}
}

func TestDedupeCIDRs(t *testing.T) {
	got := dedupeCIDRs([]string{"10.1.32.0/24", "10.0.144.0/24", "10.1.32.0/24"})
	want := []string{"10.0.144.0/24", "10.1.32.0/24"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dedupeCIDRs = %v, want %v", got, want)
	}
}
