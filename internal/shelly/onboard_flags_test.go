package shelly

import "testing"

func TestOnboardWiFiConfig_IsStatic(t *testing.T) {
	t.Parallel()
	if (&OnboardWiFiConfig{SSID: "n"}).IsStatic() {
		t.Error("DHCP config (no StaticIP) should not be static")
	}
	static := &OnboardWiFiConfig{SSID: "n", StaticIP: "10.23.47.227", Gateway: "10.23.47.1", Netmask: "255.255.254.0"}
	if !static.IsStatic() {
		t.Error("config with StaticIP should be static")
	}
	var nilCfg *OnboardWiFiConfig
	if nilCfg.IsStatic() {
		t.Error("nil config should not be static")
	}
}

func TestFindByAP(t *testing.T) {
	t.Parallel()
	devices := []OnboardDevice{
		{Name: "a", SSID: "shellycolorbulb-AABBCC"},
		{Name: "b", SSID: "shellycolorbulb-DDEEFF"},
		{Name: "no-ssid", SSID: ""},
	}
	tests := []struct {
		name     string
		target   string
		wantName string
		wantOK   bool
	}{
		{"exact", "shellycolorbulb-DDEEFF", "b", true},
		{"case-insensitive substring", "ddeeff", "b", true},
		{"uppercase substring", "AABBCC", "a", true},
		{"no match", "shellyplus1-000000", "", false},
		{"empty target never matches empty ssid", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := FindByAP(devices, tt.target)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v (got %+v)", ok, tt.wantOK, got)
			}
			if ok && got.Name != tt.wantName {
				t.Errorf("matched %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}
