package shelly

import (
	"net"
	"os"
)

// probeIface is a host network interface as the confirm path sees it: its name,
// whether it is wireless, and the unicast networks assigned to it. It is a plain
// value so the selection logic can be unit-tested with synthetic interfaces, no
// real NICs required.
type probeIface struct {
	Name       string
	IsWireless bool
	Nets       []*net.IPNet
}

// selectProbeBindInterfaces returns the interface names to try when confirming a
// device at target reachable, always beginning with "" — the unbound default
// route, i.e. the proven historical behaviour — followed by same-subnet
// interfaces as fallbacks.
//
// The default route is tried FIRST so a network where it already reaches the
// device (the common case) is never perturbed: the probe binds nothing, behaves
// exactly as before, and the remaining candidates are never reached. The fallbacks
// exist only for a network whose AP isolates wireless clients — there a device
// rejoined over the host's own wireless AP is reached station-to-station and may
// be dropped, while a wired interface on the same subnet still reaches it. So the
// fallbacks are ordered WIRED before WIRELESS. Interfaces whose subnet does not
// contain target are skipped (they cannot reach it on the link layer; the
// default route already covers any L3-routed path). The result is deduplicated
// and the "" entry is always present and first.
func selectProbeBindInterfaces(target net.IP, ifaces []probeIface) []string {
	if target == nil {
		return []string{""}
	}

	var wired, wireless []string
	seen := make(map[string]bool)
	add := func(dst *[]string, name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		*dst = append(*dst, name)
	}

	for _, ifc := range ifaces {
		if !ifaceContains(ifc, target) {
			continue
		}
		if ifc.IsWireless {
			add(&wireless, ifc.Name)
		} else {
			add(&wired, ifc.Name)
		}
	}

	out := make([]string, 0, len(wired)+len(wireless)+1)
	out = append(out, "") // default route first — the proven path, never perturbed when it works
	out = append(out, wired...)
	out = append(out, wireless...)
	return out
}

// ifaceContains reports whether any of the interface's assigned networks contains
// target, matching address family.
func ifaceContains(ifc probeIface, target net.IP) bool {
	for _, n := range ifc.Nets {
		if n != nil && n.Contains(target) {
			return true
		}
	}
	return false
}

// hostProbeIfaces enumerates the host's up, non-loopback interfaces with their
// unicast networks for selectProbeBindInterfaces. It is the impure adapter over
// net.Interfaces; the selection logic it feeds is pure and unit-tested.
func hostProbeIfaces() ([]probeIface, error) {
	raw, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	out := make([]probeIface, 0, len(raw))
	for _, ifc := range raw {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, addrErr := ifc.Addrs()
		if addrErr != nil {
			continue
		}
		nets := make([]*net.IPNet, 0, len(addrs))
		for _, a := range addrs {
			if ipNet, ok := a.(*net.IPNet); ok {
				nets = append(nets, ipNet)
			}
		}
		if len(nets) == 0 {
			continue
		}
		out = append(out, probeIface{
			Name:       ifc.Name,
			IsWireless: interfaceIsWireless(ifc.Name),
			Nets:       nets,
		})
	}
	return out, nil
}

// interfaceIsWireless reports whether a named interface is wireless. The kernel
// exposes a `wireless` directory under a wifi interface's sysfs node; non-Linux
// hosts have no such path, so this is false there — acceptable, as the
// AP-isolation case it informs is Linux-only in practice.
func interfaceIsWireless(name string) bool {
	if name == "" {
		return false
	}
	if _, err := os.Stat("/sys/class/net/" + name + "/wireless"); err == nil {
		return true
	}
	// Some drivers expose the 802.11 PHY link instead of a `wireless` dir.
	_, err := os.Stat("/sys/class/net/" + name + "/phy80211")
	return err == nil
}
