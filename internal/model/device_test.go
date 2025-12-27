package model

import "testing"

func TestDevice_HasAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dev  Device
		want bool
	}{
		{
			name: "no auth",
			dev:  Device{Name: "test"},
			want: false,
		},
		{
			name: "nil auth",
			dev:  Device{Name: "test", Auth: nil},
			want: false,
		},
		{
			name: "empty password",
			dev:  Device{Name: "test", Auth: &Auth{Username: "admin", Password: ""}},
			want: false,
		},
		{
			name: "has auth",
			dev:  Device{Name: "test", Auth: &Auth{Username: "admin", Password: "secret"}},
			want: true,
		},
		{
			name: "password only",
			dev:  Device{Name: "test", Auth: &Auth{Password: "secret"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.dev.HasAuth()
			if got != tt.want {
				t.Errorf("Device.HasAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_DisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dev  Device
		want string
	}{
		{
			name: "with name",
			dev:  Device{Name: "Living Room Switch", Address: "192.168.1.100"},
			want: "Living Room Switch",
		},
		{
			name: "empty name falls back to address",
			dev:  Device{Name: "", Address: "192.168.1.100"},
			want: "192.168.1.100",
		},
		{
			name: "no name or address",
			dev:  Device{},
			want: "",
		},
		{
			name: "address only",
			dev:  Device{Address: "shelly-1pm.local"},
			want: "shelly-1pm.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.dev.DisplayName()
			if got != tt.want {
				t.Errorf("Device.DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDevice_Fields(t *testing.T) {
	t.Parallel()

	dev := Device{
		Name:       "Kitchen Light",
		Address:    "192.168.1.50",
		Platform:   PlatformShelly,
		Generation: 2,
		Type:       "SHSW-PM",
		Model:      "Shelly Plus 1PM",
		Auth: &Auth{
			Username: "admin",
			Password: "secret123",
		},
	}

	if dev.Name != "Kitchen Light" {
		t.Errorf("Name = %q, want %q", dev.Name, "Kitchen Light")
	}
	if dev.Address != "192.168.1.50" {
		t.Errorf("Address = %q, want %q", dev.Address, "192.168.1.50")
	}
	if dev.Platform != PlatformShelly {
		t.Errorf("Platform = %q, want %q", dev.Platform, PlatformShelly)
	}
	if dev.Generation != 2 {
		t.Errorf("Generation = %d, want %d", dev.Generation, 2)
	}
	if dev.Type != "SHSW-PM" {
		t.Errorf("Type = %q, want %q", dev.Type, "SHSW-PM")
	}
	if dev.Model != "Shelly Plus 1PM" {
		t.Errorf("Model = %q, want %q", dev.Model, "Shelly Plus 1PM")
	}
	if dev.Auth.Username != "admin" {
		t.Errorf("Auth.Username = %q, want %q", dev.Auth.Username, "admin")
	}
	if dev.Auth.Password != "secret123" {
		t.Errorf("Auth.Password = %q, want %q", dev.Auth.Password, "secret123")
	}
}

func TestDevice_IsShelly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{
			name:     "empty platform defaults to Shelly",
			platform: "",
			want:     true,
		},
		{
			name:     "explicit shelly platform",
			platform: PlatformShelly,
			want:     true,
		},
		{
			name:     "tasmota platform is not Shelly",
			platform: "tasmota",
			want:     false,
		},
		{
			name:     "esphome platform is not Shelly",
			platform: "esphome",
			want:     false,
		},
		{
			name:     "custom platform is not Shelly",
			platform: "custom-plugin",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{Platform: tt.platform}
			got := dev.IsShelly()
			if got != tt.want {
				t.Errorf("Device.IsShelly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_IsPluginManaged(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{
			name:     "empty platform is not plugin-managed",
			platform: "",
			want:     false,
		},
		{
			name:     "shelly platform is not plugin-managed",
			platform: PlatformShelly,
			want:     false,
		},
		{
			name:     "tasmota platform is plugin-managed",
			platform: "tasmota",
			want:     true,
		},
		{
			name:     "esphome platform is plugin-managed",
			platform: "esphome",
			want:     true,
		},
		{
			name:     "custom platform is plugin-managed",
			platform: "custom-plugin",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{Platform: tt.platform}
			got := dev.IsPluginManaged()
			if got != tt.want {
				t.Errorf("Device.IsPluginManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDevice_GetPlatform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "empty platform defaults to shelly",
			platform: "",
			want:     PlatformShelly,
		},
		{
			name:     "explicit shelly platform",
			platform: PlatformShelly,
			want:     PlatformShelly,
		},
		{
			name:     "tasmota platform",
			platform: "tasmota",
			want:     "tasmota",
		},
		{
			name:     "custom platform",
			platform: "my-plugin",
			want:     "my-plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{Platform: tt.platform}
			got := dev.GetPlatform()
			if got != tt.want {
				t.Errorf("Device.GetPlatform() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDevice_PluginName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{
			name:     "empty platform returns empty (native Shelly)",
			platform: "",
			want:     "",
		},
		{
			name:     "shelly platform returns empty",
			platform: PlatformShelly,
			want:     "",
		},
		{
			name:     "tasmota platform returns shelly-tasmota",
			platform: "tasmota",
			want:     "shelly-tasmota",
		},
		{
			name:     "esphome platform returns shelly-esphome",
			platform: "esphome",
			want:     "shelly-esphome",
		},
		{
			name:     "custom platform returns shelly-custom",
			platform: "custom",
			want:     "shelly-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{Platform: tt.platform}
			got := dev.PluginName()
			if got != tt.want {
				t.Errorf("Device.PluginName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "lowercase colon-separated",
			input: "aa:bb:cc:dd:ee:ff",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "uppercase colon-separated",
			input: "AA:BB:CC:DD:EE:FF",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "mixed case colon-separated",
			input: "Aa:bB:Cc:dD:Ee:fF",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "dash-separated",
			input: "AA-BB-CC-DD-EE-FF",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "dot-separated (Cisco format)",
			input: "AABB.CCDD.EEFF",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "no separators",
			input: "aabbccddeeff",
			want:  "AA:BB:CC:DD:EE:FF",
		},
		{
			name:  "too short",
			input: "aabbcc",
			want:  "",
		},
		{
			name:  "too long",
			input: "aabbccddeeff00",
			want:  "",
		},
		{
			name:  "invalid characters",
			input: "gg:hh:ii:jj:kk:ll",
			want:  "",
		},
		{
			name:  "mixed valid and invalid",
			input: "aa:bb:cc:dd:ee:gg",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeMAC(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeMAC(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDevice_NormalizedMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mac  string
		want string
	}{
		{
			name: "empty MAC",
			mac:  "",
			want: "",
		},
		{
			name: "valid MAC gets normalized",
			mac:  "aa:bb:cc:dd:ee:ff",
			want: "AA:BB:CC:DD:EE:FF",
		},
		{
			name: "already normalized MAC",
			mac:  "AA:BB:CC:DD:EE:FF",
			want: "AA:BB:CC:DD:EE:FF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{MAC: tt.mac}
			got := dev.NormalizedMAC()
			if got != tt.want {
				t.Errorf("Device.NormalizedMAC() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDevice_HasAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		aliases []string
		lookup  string
		want    bool
	}{
		{
			name:    "no aliases",
			aliases: nil,
			lookup:  "mb",
			want:    false,
		},
		{
			name:    "empty aliases",
			aliases: []string{},
			lookup:  "mb",
			want:    false,
		},
		{
			name:    "exact match",
			aliases: []string{"mb", "bath"},
			lookup:  "mb",
			want:    true,
		},
		{
			name:    "case-insensitive match",
			aliases: []string{"MB", "BATH"},
			lookup:  "mb",
			want:    true,
		},
		{
			name:    "case-insensitive match reverse",
			aliases: []string{"mb", "bath"},
			lookup:  "MB",
			want:    true,
		},
		{
			name:    "no match",
			aliases: []string{"kitchen", "light"},
			lookup:  "mb",
			want:    false,
		},
		{
			name:    "partial match is not enough",
			aliases: []string{"master-bath"},
			lookup:  "master",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev := Device{Aliases: tt.aliases}
			got := dev.HasAlias(tt.lookup)
			if got != tt.want {
				t.Errorf("Device.HasAlias(%q) = %v, want %v", tt.lookup, got, tt.want)
			}
		})
	}
}
