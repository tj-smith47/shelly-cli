package cmdutil

import "testing"

func TestDeviceDisplayName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		explicit   string
		identifier string
		want       string
	}{
		{name: "explicit wins over alias", explicit: "Master Bath 1", identifier: "master-bath-1", want: "Master Bath 1"},
		{name: "explicit wins over ip", explicit: "Master Bath 1", identifier: "10.23.47.221", want: "Master Bath 1"},
		{name: "friendly alias used", explicit: "", identifier: "master-bath-1", want: "master-bath-1"},
		{name: "ipv4 identifier left blank", explicit: "", identifier: "10.23.47.221", want: ""},
		{name: "ipv6 identifier left blank", explicit: "", identifier: "fe80::1", want: ""},
		{name: "mac identifier left blank", explicit: "", identifier: "98:F4:AB:D1:60:D5", want: ""},
		{name: "host:port identifier left blank", explicit: "", identifier: "10.0.0.5:80", want: ""},
		{name: "empty identifier left blank", explicit: "", identifier: "", want: ""},
		{name: "mdns hostname is friendly", explicit: "", identifier: "shellyduo-ABC123.local", want: "shellyduo-ABC123.local"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := DeviceDisplayName(tt.explicit, tt.identifier); got != tt.want {
				t.Errorf("DeviceDisplayName(%q, %q) = %q, want %q", tt.explicit, tt.identifier, got, tt.want)
			}
		})
	}
}
