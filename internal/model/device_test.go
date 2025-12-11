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
