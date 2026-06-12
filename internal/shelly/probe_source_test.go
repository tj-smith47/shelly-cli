package shelly

import (
	"net"
	"reflect"
	"testing"
)

// mustCIDR builds the *net.IPNet for a CIDR, failing the test on a bad literal.
func mustCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q): %v", cidr, err)
	}
	return n
}

func TestSelectProbeBindInterfaces(t *testing.T) {
	t.Parallel()

	target := net.ParseIP("10.23.47.227")

	tests := []struct {
		name   string
		target net.IP
		ifaces []probeIface
		want   []string
	}{
		{
			// The dual-homed sir case: wired (vmbr0) and wireless (wlp6s0) both on
			// the device's subnet. Default route first (the proven path), then the
			// wired fallback, then the wireless one.
			name:   "default route first, then wired, then wireless",
			target: target,
			ifaces: []probeIface{
				{Name: "wlp6s0", IsWireless: true, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}},
				{Name: "vmbr0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.11/23")}},
			},
			want: []string{"", "vmbr0", "wlp6s0"},
		},
		{
			// Single-homed wireless behind an isolating AP: the default route (which
			// egresses that wireless interface) is tried, then an explicit bind to
			// it. Both may fail — that is physics, and the caller fails loud (Bug A).
			name:   "wireless only",
			target: target,
			ifaces: []probeIface{
				{Name: "wlan0", IsWireless: true, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}},
			},
			want: []string{"", "wlan0"},
		},
		{
			name:   "wired only",
			target: target,
			ifaces: []probeIface{
				{Name: "eth0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}},
			},
			want: []string{"", "eth0"},
		},
		{
			// An interface on a different subnet cannot reach the target on the
			// link layer; only the default route remains.
			name:   "no same-subnet interface",
			target: target,
			ifaces: []probeIface{
				{Name: "eth0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "192.168.1.10/24")}},
			},
			want: []string{""},
		},
		{
			name:   "nil target falls back to default route",
			target: nil,
			ifaces: []probeIface{{Name: "eth0", Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}}},
			want:   []string{""},
		},
		{
			name:   "no interfaces",
			target: target,
			ifaces: nil,
			want:   []string{""},
		},
		{
			// Two wired interfaces on the subnet are both tried, in order, after the
			// default route; no duplicates.
			name:   "multiple wired preserve order, deduped",
			target: target,
			ifaces: []probeIface{
				{Name: "eth0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}},
				{Name: "eth0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.10/23")}},
				{Name: "eth1", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "10.23.46.5/23")}},
			},
			want: []string{"", "eth0", "eth1"},
		},
		{
			// An IPv6-only interface must not match an IPv4 target.
			name:   "address family must match",
			target: target,
			ifaces: []probeIface{
				{Name: "eth0", IsWireless: false, Nets: []*net.IPNet{mustCIDR(t, "fd00::1/64")}},
			},
			want: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := selectProbeBindInterfaces(tt.target, tt.ifaces)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("selectProbeBindInterfaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSelectProbeBindInterfaces_AlwaysOffersDefaultRouteFirst guards the
// invariant the confirm loop relies on: whatever the host topology, the unbound
// default route is always tried FIRST, so a network where it already reaches the
// device behaves exactly as before and the interface fallbacks never engage.
func TestSelectProbeBindInterfaces_AlwaysOffersDefaultRouteFirst(t *testing.T) {
	t.Parallel()

	cases := [][]probeIface{
		nil,
		{{Name: "eth0", Nets: []*net.IPNet{mustCIDR(t, "10.23.47.1/23")}}},
		{{Name: "wlan0", IsWireless: true, Nets: []*net.IPNet{mustCIDR(t, "10.23.47.1/23")}}},
	}
	for _, ifaces := range cases {
		got := selectProbeBindInterfaces(net.ParseIP("10.23.47.227"), ifaces)
		if len(got) == 0 || got[0] != "" {
			t.Errorf("candidates %v must begin with default route \"\"", got)
		}
	}
}
