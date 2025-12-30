// Package testutil provides testing utilities.
package testutil

import (
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// MockDeviceResolver implements the DeviceResolver interface for testing.
type MockDeviceResolver struct {
	Device model.Device
	Err    error
}

// Resolve returns the mock device or error.
func (m *MockDeviceResolver) Resolve(_ string) (model.Device, error) {
	if m.Err != nil {
		return model.Device{}, m.Err
	}
	return m.Device, nil
}

// NewMockDevice creates a mock device for testing with common defaults.
func NewMockDevice(name, address string) model.Device {
	return model.Device{
		Name:       name,
		Address:    address,
		Generation: 2,
		Type:       "SHSW-1",
		Model:      "Shelly Plus 1",
	}
}

// NewMockDeviceWithAuth creates a mock device with authentication.
func NewMockDeviceWithAuth(name, address, username, password string) model.Device {
	dev := NewMockDevice(name, address)
	dev.Auth = &model.Auth{
		Username: username,
		Password: password,
	}
	return dev
}

// NewMockSwitchStatus creates a mock switch status for testing.
func NewMockSwitchStatus(id int, output bool, power float64) *model.SwitchStatus {
	return &model.SwitchStatus{
		ID:     id,
		Output: output,
		Source: "switch",
		Power:  &power,
	}
}

// NewMockSwitchConfig creates a mock switch config for testing.
func NewMockSwitchConfig(id int, name string) *model.SwitchConfig {
	return &model.SwitchConfig{
		ID:   id,
		Name: &name,
	}
}

// NewMockCoverStatus creates a mock cover status for testing.
func NewMockCoverStatus(id int, state string, position int) *model.CoverStatus {
	return &model.CoverStatus{
		ID:              id,
		State:           state,
		Source:          "cover",
		CurrentPosition: &position,
	}
}

// NewMockCoverConfig creates a mock cover config for testing.
func NewMockCoverConfig(id int, name string) *model.CoverConfig {
	return &model.CoverConfig{
		ID:   id,
		Name: &name,
	}
}

// NewMockLightStatus creates a mock light status for testing.
func NewMockLightStatus(id int, output bool, brightness int) *model.LightStatus {
	return &model.LightStatus{
		ID:         id,
		Output:     output,
		Brightness: &brightness,
		Source:     "light",
	}
}

// NewMockLightConfig creates a mock light config for testing.
func NewMockLightConfig(id int, name string) *model.LightConfig {
	return &model.LightConfig{
		ID:   id,
		Name: &name,
	}
}

// NewMockRGBStatus creates a mock RGB status for testing.
func NewMockRGBStatus(id int, output bool, r, g, b, brightness int) *model.RGBStatus {
	return &model.RGBStatus{
		ID:         id,
		Output:     output,
		Brightness: &brightness,
		RGB: &model.RGBColor{
			Red:   r,
			Green: g,
			Blue:  b,
		},
		Source: "rgb",
	}
}

// NewMockRGBConfig creates a mock RGB config for testing.
func NewMockRGBConfig(id int, name string) *model.RGBConfig {
	return &model.RGBConfig{
		ID:   id,
		Name: &name,
	}
}

// NewMockInputStatus creates a mock input status for testing.
func NewMockInputStatus(id int, state bool, inputType string) *model.InputStatus {
	return &model.InputStatus{
		ID:    id,
		State: state,
		Type:  inputType,
	}
}

// NewMockInputConfig creates a mock input config for testing.
func NewMockInputConfig(id int, name string) *model.InputConfig {
	return &model.InputConfig{
		ID:   id,
		Name: &name,
	}
}
