package ui

import (
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MrVasquez96/go-wg/wg/config"

	scanner "github.com/MrVasquez96/go-ipscan"
)

const (
	// maxScanHosts is the largest network scanned address by address (a /20).
	maxScanHosts = 4096
	// maxSweepSubnets caps how many /24s a wide route may be probed for
	// (1024 == a /14). Anything wider is left alone.
	maxSweepSubnets = 1024
)

// subnetProbes lists the host addresses probed when looking for active /24s,
// in the order they are tried. Sites hand out addresses from .1 upwards, but
// the lowest ones are not always online, so a single .1 probe misses whole
// networks - the first few addresses get a TCP fallback for hosts that drop
// ICMP, the rest are ping-only to keep a wide sweep affordable.
var subnetProbes = []struct {
	host byte
	tcp  bool
}{
	{1, true}, {2, true}, {3, true},
	{4, false}, {5, false}, {6, false}, {7, false},
	{8, false}, {9, false}, {10, false},
	{254, true},
}

// subnetProbePorts are the ports tried when ICMP is filtered.
var subnetProbePorts = []int{22, 80, 443}

// ScanPlan describes what can be scanned through a tunnel.
type ScanPlan struct {
	// Direct networks are scanned address by address.
	Direct []string
	// Wide networks are first probed for active /24s, which are then scanned.
	Wide []string
	// Skipped networks are too large to handle at all (e.g. 0.0.0.0/0).
	Skipped []string
}

// IsEmpty reports whether the plan has nothing to scan.
func (p ScanPlan) IsEmpty() bool { return len(p.Direct) == 0 && len(p.Wide) == 0 }

// PlanScan derives what is reachable through a tunnel from its config: the
// interface's own addresses plus every peer's AllowedIPs. In a hub-and-spoke
// setup the remote sites live outside the interface's own /24, so the routed
// networks - not the local address - decide what is scannable.
//
// Networks longer than /24 (including the /32 addresses wg-quick configs
// commonly use) are widened to their containing /24. When nothing usable is
// found, the fallback IP's /24 is scanned as before.
func PlanScan(cfg *config.Config, fallbackIP string) ScanPlan {
	var nets []*net.IPNet

	if cfg != nil {
		for i := range cfg.Interface.Address {
			nets = append(nets, &cfg.Interface.Address[i])
		}
		for _, peer := range cfg.Peers {
			for i := range peer.AllowedIPs {
				nets = append(nets, &peer.AllowedIPs[i])
			}
		}
	}

	var plan ScanPlan
	seen := make(map[string]bool)
	var direct []*net.IPNet

	for _, n := range nets {
		ip := n.IP.To4()
		if ip == nil {
			continue // IPv6 is out of scope for the scanner
		}
		ones, _ := n.Mask.Size()
		if ones > 24 {
			ones = 24 // widen host routes to their /24 so neighbours show up too
		}
		mask := net.CIDRMask(ones, 32)
		norm := &net.IPNet{IP: ip.Mask(mask), Mask: mask}
		key := norm.String()
		if seen[key] {
			continue
		}
		seen[key] = true

		switch size := uint64(1) << uint(32-ones); {
		case size <= maxScanHosts:
			direct = append(direct, norm)
		case size/256 <= maxSweepSubnets:
			plan.Wide = append(plan.Wide, key)
		default:
			plan.Skipped = append(plan.Skipped, key) // 0.0.0.0/0 and friends
		}
	}

	sort.Slice(direct, func(i, j int) bool {
		return ipToUint32(direct[i].IP.String()) < ipToUint32(direct[j].IP.String())
	})
	for i, n := range direct {
		if coveredBy(n, direct, i) {
			continue
		}
		plan.Direct = append(plan.Direct, n.String())
	}

	if plan.IsEmpty() {
		if fallback := scanner.DeriveCIDR24(fallbackIP); fallback != "" {
			plan.Direct = append(plan.Direct, fallback)
		}
	}

	sort.Strings(plan.Wide)
	sort.Strings(plan.Skipped)
	return plan
}

// coveredBy reports whether nets[skip] is contained in a wider network in nets.
func coveredBy(n *net.IPNet, nets []*net.IPNet, skip int) bool {
	nOnes, _ := n.Mask.Size()
	for i, other := range nets {
		if i == skip {
			continue
		}
		otherOnes, _ := other.Mask.Size()
		if otherOnes < nOnes && other.Contains(n.IP) {
			return true
		}
	}
	return false
}

// subnetsOf splits a CIDR into the /24s it contains.
func subnetsOf(cidr string) []string {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	ip := ipnet.IP.To4()
	if ip == nil {
		return nil
	}
	ones, _ := ipnet.Mask.Size()
	if ones >= 24 {
		return []string{ipnet.String()}
	}
	count := 1 << uint(24-ones)
	if count > maxSweepSubnets {
		return nil
	}

	subnets := make([]string, 0, count)
	base := uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8
	for i := 0; i < count; i++ {
		addr := base + uint32(i)<<8
		subnets = append(subnets, net.IPv4(byte(addr>>24), byte(addr>>16), byte(addr>>8), 0).String()+"/24")
	}
	return subnets
}

// discoverSubnets probes a handful of likely host addresses in every /24 of
// cidr and returns the /24s that answered, so a wide route can be narrowed to
// the sites that actually exist. Probing runs one address offset at a time
// across all subnets, so subnets that answer early drop out of the remaining
// rounds. onProgress reports probes completed out of the planned total.
func discoverSubnets(cidr string, workers int, timeout time.Duration, stopped *atomic.Bool, onProgress func(done, total int)) []string {
	subnets := subnetsOf(cidr)
	if len(subnets) == 0 {
		return nil
	}
	if workers < 1 {
		workers = 1
	}

	var (
		mu     sync.Mutex
		found  = make(map[string]bool)
		active []string
		done   int
	)
	total := len(subnets) * len(subnetProbes)

	for _, probe := range subnetProbes {
		if stopped.Load() {
			break
		}

		// Subnets already found skip this round; count their probes as done.
		var pending []string
		mu.Lock()
		for _, subnet := range subnets {
			if found[subnet] {
				done++
			} else {
				pending = append(pending, subnet)
			}
		}
		progress := done
		mu.Unlock()
		onProgress(progress, total)

		if len(pending) == 0 {
			break
		}

		jobs := make(chan string)
		var wg sync.WaitGroup
		for i := 0; i < workers && i < len(pending); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for subnet := range jobs {
					alive := !stopped.Load() && probeSubnet(subnet, probe.host, probe.tcp, timeout)

					mu.Lock()
					if alive && !found[subnet] {
						found[subnet] = true
						active = append(active, subnet)
					}
					done++
					progress := done
					mu.Unlock()

					onProgress(progress, total)
				}
			}()
		}

		for _, subnet := range pending {
			if stopped.Load() {
				break
			}
			jobs <- subnet
		}
		close(jobs)
		wg.Wait()
	}

	sort.Slice(active, func(i, j int) bool {
		return ipToUint32(strings.TrimSuffix(active[i], "/24")) < ipToUint32(strings.TrimSuffix(active[j], "/24"))
	})
	return active
}

// probeSubnet reports whether the given host address in the /24 responds to
// ping, or - when tcp is set - to a connection on a common port.
func probeSubnet(subnet string, host byte, tcp bool, timeout time.Duration) bool {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}
	base := ipnet.IP.To4()
	if base == nil {
		return false
	}

	ip := net.IPv4(base[0], base[1], base[2], host).String()
	if scanner.Ping(ip) {
		return true
	}
	if !tcp {
		return false
	}
	for _, port := range subnetProbePorts {
		if scanner.IsPortOpen(ip, port, timeout) {
			return true
		}
	}
	return false
}

// totalHosts returns how many addresses the given CIDRs cover in total.
func totalHosts(cidrs []string) uint32 {
	var total uint32
	for _, c := range cidrs {
		_, ipnet, err := net.ParseCIDR(c)
		if err != nil {
			continue
		}
		ones, bits := ipnet.Mask.Size()
		total += 1 << uint(bits-ones)
	}
	return total
}
