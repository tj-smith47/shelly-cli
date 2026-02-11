package backup

import (
	"encoding/json"
	"testing"

	"github.com/tj-smith47/shelly-go/gen1"
)

func TestMarshalGen1WiFi(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *gen1.Settings
		wantNil  bool
		wantKeys []string
	}{
		{
			name:     "empty settings",
			settings: &gen1.Settings{},
			wantNil:  true,
		},
		{
			name: "with station",
			settings: &gen1.Settings{
				WiFiSta: &gen1.WiFiStaSettings{SSID: "TestNetwork", Key: "pass123"},
			},
			wantKeys: []string{"sta"},
		},
		{
			name: "with station and AP",
			settings: &gen1.Settings{
				WiFiSta: &gen1.WiFiStaSettings{SSID: "TestNetwork"},
				WiFiAp:  &gen1.WiFiApSettings{SSID: "ShellyAP", Enabled: true},
			},
			wantKeys: []string{"sta", "ap"},
		},
		{
			name: "all WiFi fields",
			settings: &gen1.Settings{
				WiFiSta:   &gen1.WiFiStaSettings{SSID: "Net1"},
				WiFiSta1:  &gen1.WiFiStaSettings{SSID: "Net2"},
				WiFiAp:    &gen1.WiFiApSettings{SSID: "AP"},
				ApRoaming: &gen1.ApRoamingSettings{Enabled: true},
			},
			wantKeys: []string{"sta", "sta1", "ap", "ap_roaming"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := marshalGen1WiFi(tt.settings)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %s", string(result))
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			var parsed map[string]json.RawMessage
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Fatalf("failed to parse result: %v", err)
			}

			for _, key := range tt.wantKeys {
				if _, ok := parsed[key]; !ok {
					t.Errorf("missing key %q in WiFi result", key)
				}
			}
		})
	}
}

func TestMarshalGen1Components(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *gen1.Settings
		wantNil  bool
		wantKeys []string
	}{
		{
			name:     "empty settings",
			settings: &gen1.Settings{},
			wantNil:  true,
		},
		{
			name: "with relays",
			settings: &gen1.Settings{
				Relays: []gen1.RelaySettings{
					{Name: "relay0"},
				},
			},
			wantKeys: []string{"relays"},
		},
		{
			name: "with lights and relays",
			settings: &gen1.Settings{
				Lights: []gen1.LightSettings{
					{Name: "light0"},
				},
				Relays: []gen1.RelaySettings{
					{Name: "relay0"},
				},
			},
			wantKeys: []string{"lights", "relays"},
		},
		{
			name: "all component types",
			settings: &gen1.Settings{
				Lights:  []gen1.LightSettings{{Name: "light0"}},
				Relays:  []gen1.RelaySettings{{Name: "relay0"}},
				Rollers: []gen1.RollerSettings{{DefaultState: "stop"}},
				Meters:  []gen1.MeterSettings{{PowerLimit: 100}},
				EMeters: []gen1.EMeterSettings{{CTType: 1}},
			},
			wantKeys: []string{"lights", "relays", "rollers", "meters", "emeters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := marshalGen1Components(tt.settings)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %d keys", len(result))
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			for _, key := range tt.wantKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("missing key %q in components", key)
				}
			}
		})
	}
}

func TestMarshalGen1Schedules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *gen1.Settings
		wantNil  bool
		wantLen  int
	}{
		{
			name:     "no schedules",
			settings: &gen1.Settings{},
			wantNil:  true,
		},
		{
			name: "relay schedules",
			settings: &gen1.Settings{
				Relays: []gen1.RelaySettings{
					{Schedule: true, ScheduleRules: []string{"0 0 8 * * *"}},
					{Schedule: false, ScheduleRules: []string{}},
				},
			},
			wantLen: 1,
		},
		{
			name: "light schedules",
			settings: &gen1.Settings{
				Lights: []gen1.LightSettings{
					{Schedule: true, ScheduleRules: []string{"0 0 20 * * *", "0 0 22 * * *"}},
				},
			},
			wantLen: 1,
		},
		{
			name: "relay and light schedules",
			settings: &gen1.Settings{
				Relays: []gen1.RelaySettings{
					{Schedule: true, ScheduleRules: []string{"0 0 8 * * *"}},
				},
				Lights: []gen1.LightSettings{
					{Schedule: true, ScheduleRules: []string{"0 0 20 * * *"}},
				},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := marshalGen1Schedules(tt.settings)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %s", string(result))
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			var entries []json.RawMessage
			if err := json.Unmarshal(result, &entries); err != nil {
				t.Fatalf("failed to parse schedules: %v", err)
			}

			if len(entries) != tt.wantLen {
				t.Errorf("got %d schedule entries, want %d", len(entries), tt.wantLen)
			}
		})
	}
}

func TestMustMarshal(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	result := mustMarshal(data)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if parsed["key"] != "value" {
		t.Errorf("got %q, want %q", parsed["key"], "value")
	}
}

func TestMustMarshal_Panics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unmarshalable value")
		}
	}()

	// Channels cannot be marshaled to JSON
	mustMarshal(make(chan int))
}

func TestAddWarning(t *testing.T) {
	t.Parallel()

	result := &RestoreResult{}

	addWarning(result, "error %d: %s", 1, "test")

	if len(result.Warnings) != 1 {
		t.Fatalf("got %d warnings, want 1", len(result.Warnings))
	}

	if result.Warnings[0] != "error 1: test" {
		t.Errorf("got %q, want %q", result.Warnings[0], "error 1: test")
	}

	addWarning(result, "second warning")

	if len(result.Warnings) != 2 {
		t.Fatalf("got %d warnings, want 2", len(result.Warnings))
	}
}

func TestSanitizeForPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"Living Room", "Living_Room"},
		{"path/to/thing", "path_to_thing"},
		{"back\\slash", "back_slash"},
		{"colon:name", "colon_name"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := sanitizeForPath(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeForPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
