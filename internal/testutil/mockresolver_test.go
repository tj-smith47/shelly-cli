package testutil

import (
	"testing"
)

func TestMockDeviceResolver_Resolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolver *MockDeviceResolver
		input    string
		wantErr  bool
		wantName string
		wantAddr string
	}{
		{
			name: "returns device",
			resolver: &MockDeviceResolver{
				Device: NewMockDevice("kitchen-light", "192.168.1.100"),
			},
			input:    "kitchen-light",
			wantErr:  false,
			wantName: "kitchen-light",
			wantAddr: "192.168.1.100",
		},
		{
			name: "returns error",
			resolver: &MockDeviceResolver{
				Err: ErrMock,
			},
			input:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			device, err := tt.resolver.Resolve(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if device.Name != tt.wantName {
					t.Errorf("Resolve() device.Name = %v, want %v", device.Name, tt.wantName)
				}
				if device.Address != tt.wantAddr {
					t.Errorf("Resolve() device.Address = %v, want %v", device.Address, tt.wantAddr)
				}
			}
		})
	}
}

func TestNewMockDevice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deviceName string
		address    string
	}{
		{
			name:       "basic device",
			deviceName: "kitchen-light",
			address:    "192.168.1.100",
		},
		{
			name:       "different device",
			deviceName: "living-room-switch",
			address:    "10.0.0.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			device := NewMockDevice(tt.deviceName, tt.address)

			if device.Name != tt.deviceName {
				t.Errorf("NewMockDevice() Name = %v, want %v", device.Name, tt.deviceName)
			}
			if device.Address != tt.address {
				t.Errorf("NewMockDevice() Address = %v, want %v", device.Address, tt.address)
			}
			if device.Generation != 2 {
				t.Errorf("NewMockDevice() Generation = %v, want 2", device.Generation)
			}
			if device.Type != "SHSW-1" {
				t.Errorf("NewMockDevice() Type = %v, want SHSW-1", device.Type)
			}
			if device.Model != "Shelly Plus 1" {
				t.Errorf("NewMockDevice() Model = %v, want 'Shelly Plus 1'", device.Model)
			}
			if device.Auth != nil {
				t.Error("NewMockDevice() Auth should be nil")
			}
		})
	}
}

func TestNewMockDeviceWithAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deviceName string
		address    string
		username   string
		password   string
	}{
		{
			name:       "basic auth",
			deviceName: "kitchen-light",
			address:    "192.168.1.100",
			username:   "admin",
			password:   "secret123",
		},
		{
			name:       "different credentials",
			deviceName: "office-switch",
			address:    "10.0.0.50",
			username:   "user",
			password:   "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			device := NewMockDeviceWithAuth(tt.deviceName, tt.address, tt.username, tt.password)

			if device.Name != tt.deviceName {
				t.Errorf("NewMockDeviceWithAuth() Name = %v, want %v", device.Name, tt.deviceName)
			}
			if device.Address != tt.address {
				t.Errorf("NewMockDeviceWithAuth() Address = %v, want %v", device.Address, tt.address)
			}
			if device.Auth == nil {
				t.Fatal("NewMockDeviceWithAuth() Auth should not be nil")
			}
			if device.Auth.Username != tt.username {
				t.Errorf("NewMockDeviceWithAuth() Username = %v, want %v", device.Auth.Username, tt.username)
			}
			if device.Auth.Password != tt.password {
				t.Errorf("NewMockDeviceWithAuth() Password = %v, want %v", device.Auth.Password, tt.password)
			}
			// Verify it inherits from NewMockDevice
			if device.Generation != 2 {
				t.Errorf("NewMockDeviceWithAuth() Generation = %v, want 2", device.Generation)
			}
		})
	}
}

func TestNewMockSwitchStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		output     bool
		power      float64
		wantSource string
	}{
		{
			name:       "on with power",
			id:         0,
			output:     true,
			power:      100.5,
			wantSource: "switch",
		},
		{
			name:       "off with zero power",
			id:         1,
			output:     false,
			power:      0.0,
			wantSource: "switch",
		},
		{
			name:       "high power",
			id:         2,
			output:     true,
			power:      2500.75,
			wantSource: "switch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := NewMockSwitchStatus(tt.id, tt.output, tt.power)

			if status == nil {
				t.Fatal("NewMockSwitchStatus() returned nil")
			}
			if status.ID != tt.id {
				t.Errorf("NewMockSwitchStatus() ID = %v, want %v", status.ID, tt.id)
			}
			if status.Output != tt.output {
				t.Errorf("NewMockSwitchStatus() Output = %v, want %v", status.Output, tt.output)
			}
			if status.Power == nil {
				t.Fatal("NewMockSwitchStatus() Power should not be nil")
			}
			if *status.Power != tt.power {
				t.Errorf("NewMockSwitchStatus() Power = %v, want %v", *status.Power, tt.power)
			}
			if status.Source != tt.wantSource {
				t.Errorf("NewMockSwitchStatus() Source = %v, want %v", status.Source, tt.wantSource)
			}
		})
	}
}

func TestNewMockSwitchConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		configName string
	}{
		{
			name:       "basic config",
			id:         0,
			configName: "kitchen-switch",
		},
		{
			name:       "different id",
			id:         1,
			configName: "living-room-switch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewMockSwitchConfig(tt.id, tt.configName)

			if config == nil {
				t.Fatal("NewMockSwitchConfig() returned nil")
			}
			if config.ID != tt.id {
				t.Errorf("NewMockSwitchConfig() ID = %v, want %v", config.ID, tt.id)
			}
			if config.Name == nil {
				t.Fatal("NewMockSwitchConfig() Name should not be nil")
			}
			if *config.Name != tt.configName {
				t.Errorf("NewMockSwitchConfig() Name = %v, want %v", *config.Name, tt.configName)
			}
		})
	}
}

func TestNewMockCoverStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		state      string
		position   int
		wantSource string
	}{
		{
			name:       "open position",
			id:         0,
			state:      "open",
			position:   100,
			wantSource: "cover",
		},
		{
			name:       "closed position",
			id:         0,
			state:      "closed",
			position:   0,
			wantSource: "cover",
		},
		{
			name:       "half open",
			id:         1,
			state:      "stopped",
			position:   50,
			wantSource: "cover",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := NewMockCoverStatus(tt.id, tt.state, tt.position)

			if status == nil {
				t.Fatal("NewMockCoverStatus() returned nil")
			}
			if status.ID != tt.id {
				t.Errorf("NewMockCoverStatus() ID = %v, want %v", status.ID, tt.id)
			}
			if status.State != tt.state {
				t.Errorf("NewMockCoverStatus() State = %v, want %v", status.State, tt.state)
			}
			if status.CurrentPosition == nil {
				t.Fatal("NewMockCoverStatus() CurrentPosition should not be nil")
			}
			if *status.CurrentPosition != tt.position {
				t.Errorf("NewMockCoverStatus() CurrentPosition = %v, want %v", *status.CurrentPosition, tt.position)
			}
			if status.Source != tt.wantSource {
				t.Errorf("NewMockCoverStatus() Source = %v, want %v", status.Source, tt.wantSource)
			}
		})
	}
}

func TestNewMockCoverConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		configName string
	}{
		{
			name:       "basic config",
			id:         0,
			configName: "bedroom-blinds",
		},
		{
			name:       "different id",
			id:         1,
			configName: "office-curtains",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewMockCoverConfig(tt.id, tt.configName)

			if config == nil {
				t.Fatal("NewMockCoverConfig() returned nil")
			}
			if config.ID != tt.id {
				t.Errorf("NewMockCoverConfig() ID = %v, want %v", config.ID, tt.id)
			}
			if config.Name == nil {
				t.Fatal("NewMockCoverConfig() Name should not be nil")
			}
			if *config.Name != tt.configName {
				t.Errorf("NewMockCoverConfig() Name = %v, want %v", *config.Name, tt.configName)
			}
		})
	}
}

func TestNewMockLightStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		output     bool
		brightness int
		wantSource string
	}{
		{
			name:       "on full brightness",
			id:         0,
			output:     true,
			brightness: 100,
			wantSource: "light",
		},
		{
			name:       "off",
			id:         0,
			output:     false,
			brightness: 0,
			wantSource: "light",
		},
		{
			name:       "dimmed",
			id:         1,
			output:     true,
			brightness: 50,
			wantSource: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := NewMockLightStatus(tt.id, tt.output, tt.brightness)

			if status == nil {
				t.Fatal("NewMockLightStatus() returned nil")
			}
			if status.ID != tt.id {
				t.Errorf("NewMockLightStatus() ID = %v, want %v", status.ID, tt.id)
			}
			if status.Output != tt.output {
				t.Errorf("NewMockLightStatus() Output = %v, want %v", status.Output, tt.output)
			}
			if status.Brightness == nil {
				t.Fatal("NewMockLightStatus() Brightness should not be nil")
			}
			if *status.Brightness != tt.brightness {
				t.Errorf("NewMockLightStatus() Brightness = %v, want %v", *status.Brightness, tt.brightness)
			}
			if status.Source != tt.wantSource {
				t.Errorf("NewMockLightStatus() Source = %v, want %v", status.Source, tt.wantSource)
			}
		})
	}
}

func TestNewMockLightConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		configName string
	}{
		{
			name:       "basic config",
			id:         0,
			configName: "kitchen-dimmer",
		},
		{
			name:       "different id",
			id:         1,
			configName: "living-room-light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewMockLightConfig(tt.id, tt.configName)

			if config == nil {
				t.Fatal("NewMockLightConfig() returned nil")
			}
			if config.ID != tt.id {
				t.Errorf("NewMockLightConfig() ID = %v, want %v", config.ID, tt.id)
			}
			if config.Name == nil {
				t.Fatal("NewMockLightConfig() Name should not be nil")
			}
			if *config.Name != tt.configName {
				t.Errorf("NewMockLightConfig() Name = %v, want %v", *config.Name, tt.configName)
			}
		})
	}
}

func TestNewMockRGBStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		output     bool
		r, g, b    int
		brightness int
		wantSource string
	}{
		{
			name:       "red at full brightness",
			id:         0,
			output:     true,
			r:          255,
			g:          0,
			b:          0,
			brightness: 100,
			wantSource: "rgb",
		},
		{
			name:       "white dimmed",
			id:         0,
			output:     true,
			r:          255,
			g:          255,
			b:          255,
			brightness: 50,
			wantSource: "rgb",
		},
		{
			name:       "off",
			id:         1,
			output:     false,
			r:          0,
			g:          0,
			b:          0,
			brightness: 0,
			wantSource: "rgb",
		},
		{
			name:       "custom color",
			id:         0,
			output:     true,
			r:          128,
			g:          64,
			b:          192,
			brightness: 75,
			wantSource: "rgb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := NewMockRGBStatus(tt.id, tt.output, tt.r, tt.g, tt.b, tt.brightness)

			if status == nil {
				t.Fatal("NewMockRGBStatus() returned nil")
			}
			if status.ID != tt.id {
				t.Errorf("NewMockRGBStatus() ID = %v, want %v", status.ID, tt.id)
			}
			if status.Output != tt.output {
				t.Errorf("NewMockRGBStatus() Output = %v, want %v", status.Output, tt.output)
			}
			if status.Brightness == nil {
				t.Fatal("NewMockRGBStatus() Brightness should not be nil")
			}
			if *status.Brightness != tt.brightness {
				t.Errorf("NewMockRGBStatus() Brightness = %v, want %v", *status.Brightness, tt.brightness)
			}
			if status.RGB == nil {
				t.Fatal("NewMockRGBStatus() RGB should not be nil")
			}
			if status.RGB.Red != tt.r {
				t.Errorf("NewMockRGBStatus() RGB.Red = %v, want %v", status.RGB.Red, tt.r)
			}
			if status.RGB.Green != tt.g {
				t.Errorf("NewMockRGBStatus() RGB.Green = %v, want %v", status.RGB.Green, tt.g)
			}
			if status.RGB.Blue != tt.b {
				t.Errorf("NewMockRGBStatus() RGB.Blue = %v, want %v", status.RGB.Blue, tt.b)
			}
			if status.Source != tt.wantSource {
				t.Errorf("NewMockRGBStatus() Source = %v, want %v", status.Source, tt.wantSource)
			}
		})
	}
}

func TestNewMockRGBConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		configName string
	}{
		{
			name:       "basic config",
			id:         0,
			configName: "rgb-strip",
		},
		{
			name:       "different id",
			id:         1,
			configName: "ambient-light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewMockRGBConfig(tt.id, tt.configName)

			if config == nil {
				t.Fatal("NewMockRGBConfig() returned nil")
			}
			if config.ID != tt.id {
				t.Errorf("NewMockRGBConfig() ID = %v, want %v", config.ID, tt.id)
			}
			if config.Name == nil {
				t.Fatal("NewMockRGBConfig() Name should not be nil")
			}
			if *config.Name != tt.configName {
				t.Errorf("NewMockRGBConfig() Name = %v, want %v", *config.Name, tt.configName)
			}
		})
	}
}

func TestNewMockInputStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        int
		state     bool
		inputType string
	}{
		{
			name:      "button pressed",
			id:        0,
			state:     true,
			inputType: "button",
		},
		{
			name:      "switch off",
			id:        0,
			state:     false,
			inputType: "switch",
		},
		{
			name:      "different id",
			id:        1,
			state:     true,
			inputType: "button",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := NewMockInputStatus(tt.id, tt.state, tt.inputType)

			if status == nil {
				t.Fatal("NewMockInputStatus() returned nil")
			}
			if status.ID != tt.id {
				t.Errorf("NewMockInputStatus() ID = %v, want %v", status.ID, tt.id)
			}
			if status.State != tt.state {
				t.Errorf("NewMockInputStatus() State = %v, want %v", status.State, tt.state)
			}
			if status.Type != tt.inputType {
				t.Errorf("NewMockInputStatus() Type = %v, want %v", status.Type, tt.inputType)
			}
		})
	}
}

func TestNewMockInputConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         int
		configName string
	}{
		{
			name:       "basic config",
			id:         0,
			configName: "wall-button",
		},
		{
			name:       "different id",
			id:         1,
			configName: "door-sensor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewMockInputConfig(tt.id, tt.configName)

			if config == nil {
				t.Fatal("NewMockInputConfig() returned nil")
			}
			if config.ID != tt.id {
				t.Errorf("NewMockInputConfig() ID = %v, want %v", config.ID, tt.id)
			}
			if config.Name == nil {
				t.Fatal("NewMockInputConfig() Name should not be nil")
			}
			if *config.Name != tt.configName {
				t.Errorf("NewMockInputConfig() Name = %v, want %v", *config.Name, tt.configName)
			}
		})
	}
}
