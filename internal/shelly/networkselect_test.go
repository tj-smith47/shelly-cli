package shelly

import "testing"

func TestSelectWiFiNetwork(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		readings     []wifiReading
		wantSSID     string
		wantPassword string
		wantNil      bool
	}{
		{
			name: "password recovered from another device on the same ssid",
			readings: []wifiReading{
				{ssid: "OnyxCheetah4.7", key: ""},   // masked-key bulb
				{ssid: "OnyxCheetah4.7", key: "pw"}, // relay exposes the key
			},
			wantSSID:     "OnyxCheetah4.7",
			wantPassword: "pw",
		},
		{
			name: "dominant network with a password wins",
			readings: []wifiReading{
				{ssid: "Guest", key: "g"},
				{ssid: "Home", key: "h"},
				{ssid: "Home", key: ""},
				{ssid: "Home", key: ""},
			},
			wantSSID:     "Home",
			wantPassword: "h",
		},
		{
			name: "skip dominant network when it has no recoverable password",
			readings: []wifiReading{
				{ssid: "Home", key: ""},
				{ssid: "Home", key: ""},
				{ssid: "Home", key: ""},
				{ssid: "Lab", key: "lab"},
			},
			wantSSID:     "Lab",
			wantPassword: "lab",
		},
		{
			name: "empty ssids are ignored",
			readings: []wifiReading{
				{ssid: "", key: "orphan"},
				{ssid: "Net", key: "n"},
			},
			wantSSID:     "Net",
			wantPassword: "n",
		},
		{
			name: "tie broken by ssid",
			readings: []wifiReading{
				{ssid: "Bravo", key: "b"},
				{ssid: "Alpha", key: "a"},
			},
			wantSSID:     "Alpha",
			wantPassword: "a",
		},
		{
			name: "no password recoverable",
			readings: []wifiReading{
				{ssid: "Home", key: ""},
				{ssid: "Guest", key: ""},
			},
			wantNil: true,
		},
		{
			name:    "empty input",
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := selectWiFiNetwork(tt.readings)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected %s/%s, got nil", tt.wantSSID, tt.wantPassword)
			}
			if got.SSID != tt.wantSSID || got.Password != tt.wantPassword {
				t.Errorf("got %s/%s, want %s/%s", got.SSID, got.Password, tt.wantSSID, tt.wantPassword)
			}
		})
	}
}
