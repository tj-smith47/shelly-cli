// Package testutil provides testing utilities.
package testutil

import (
	"context"
	"errors"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ErrMock is a standard mock error for testing.
var ErrMock = errors.New("mock error")

// MockSwitchComponent implements the SwitchOperations interface for testing.
type MockSwitchComponent struct {
	StatusResult *model.SwitchStatus
	ConfigResult *model.SwitchConfig
	ToggleResult *model.SwitchStatus
	OnCalled     bool
	OffCalled    bool
	ToggleCalled bool
	SetCalled    bool
	SetValue     bool
	Err          error
}

// GetStatus returns the mock status.
func (m *MockSwitchComponent) GetStatus(_ context.Context) (*model.SwitchStatus, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.StatusResult, nil
}

// GetConfig returns the mock config.
func (m *MockSwitchComponent) GetConfig(_ context.Context) (*model.SwitchConfig, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ConfigResult, nil
}

// On records the call and returns the mock error.
func (m *MockSwitchComponent) On(_ context.Context) error {
	m.OnCalled = true
	return m.Err
}

// Off records the call and returns the mock error.
func (m *MockSwitchComponent) Off(_ context.Context) error {
	m.OffCalled = true
	return m.Err
}

// Toggle records the call and returns the mock result.
func (m *MockSwitchComponent) Toggle(_ context.Context) (*model.SwitchStatus, error) {
	m.ToggleCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ToggleResult, nil
}

// Set records the call and returns the mock error.
func (m *MockSwitchComponent) Set(_ context.Context, on bool) error {
	m.SetCalled = true
	m.SetValue = on
	return m.Err
}

// MockCoverComponent implements the CoverOperations interface for testing.
type MockCoverComponent struct {
	StatusResult    *model.CoverStatus
	ConfigResult    *model.CoverConfig
	OpenCalled      bool
	CloseCalled     bool
	StopCalled      bool
	PositionCalled  bool
	PositionValue   int
	CalibrateCalled bool
	Err             error
}

// GetStatus returns the mock status.
func (m *MockCoverComponent) GetStatus(_ context.Context) (*model.CoverStatus, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.StatusResult, nil
}

// GetConfig returns the mock config.
func (m *MockCoverComponent) GetConfig(_ context.Context) (*model.CoverConfig, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ConfigResult, nil
}

// Open records the call and returns the mock error.
func (m *MockCoverComponent) Open(_ context.Context, _ *int) error {
	m.OpenCalled = true
	return m.Err
}

// Close records the call and returns the mock error.
func (m *MockCoverComponent) Close(_ context.Context, _ *int) error {
	m.CloseCalled = true
	return m.Err
}

// Stop records the call and returns the mock error.
func (m *MockCoverComponent) Stop(_ context.Context) error {
	m.StopCalled = true
	return m.Err
}

// GoToPosition records the call and returns the mock error.
func (m *MockCoverComponent) GoToPosition(_ context.Context, pos int) error {
	m.PositionCalled = true
	m.PositionValue = pos
	return m.Err
}

// Calibrate records the call and returns the mock error.
func (m *MockCoverComponent) Calibrate(_ context.Context) error {
	m.CalibrateCalled = true
	return m.Err
}

// MockLightComponent implements the LightOperations interface for testing.
type MockLightComponent struct {
	StatusResult  *model.LightStatus
	ConfigResult  *model.LightConfig
	ToggleResult  *model.LightStatus
	OnCalled      bool
	OffCalled     bool
	ToggleCalled  bool
	SetCalled     bool
	SetBrightness *int
	SetOn         *bool
	Err           error
}

// GetStatus returns the mock status.
func (m *MockLightComponent) GetStatus(_ context.Context) (*model.LightStatus, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.StatusResult, nil
}

// GetConfig returns the mock config.
func (m *MockLightComponent) GetConfig(_ context.Context) (*model.LightConfig, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ConfigResult, nil
}

// On records the call and returns the mock error.
func (m *MockLightComponent) On(_ context.Context) error {
	m.OnCalled = true
	return m.Err
}

// Off records the call and returns the mock error.
func (m *MockLightComponent) Off(_ context.Context) error {
	m.OffCalled = true
	return m.Err
}

// Toggle records the call and returns the mock result.
func (m *MockLightComponent) Toggle(_ context.Context) (*model.LightStatus, error) {
	m.ToggleCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ToggleResult, nil
}

// SetBrightnessValue records the call and returns the mock error.
func (m *MockLightComponent) SetBrightnessValue(_ context.Context, brightness int) error {
	m.SetCalled = true
	m.SetBrightness = &brightness
	return m.Err
}

// Set records the call and returns the mock error.
func (m *MockLightComponent) Set(_ context.Context, brightness *int, on *bool) error {
	m.SetCalled = true
	m.SetBrightness = brightness
	m.SetOn = on
	return m.Err
}

// MockRGBComponent implements the RGBOperations interface for testing.
type MockRGBComponent struct {
	StatusResult       *model.RGBStatus
	ConfigResult       *model.RGBConfig
	ToggleResult       *model.RGBStatus
	OnCalled           bool
	OffCalled          bool
	ToggleCalled       bool
	SetCalled          bool
	SetRedValue        *int
	SetGreenValue      *int
	SetBlueValue       *int
	SetBrightnessValue *int
	SetOnValue         *bool
	Err                error
}

// GetStatus returns the mock status.
func (m *MockRGBComponent) GetStatus(_ context.Context) (*model.RGBStatus, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.StatusResult, nil
}

// GetConfig returns the mock config.
func (m *MockRGBComponent) GetConfig(_ context.Context) (*model.RGBConfig, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ConfigResult, nil
}

// On records the call and returns the mock error.
func (m *MockRGBComponent) On(_ context.Context) error {
	m.OnCalled = true
	return m.Err
}

// Off records the call and returns the mock error.
func (m *MockRGBComponent) Off(_ context.Context) error {
	m.OffCalled = true
	return m.Err
}

// Toggle records the call and returns the mock result.
func (m *MockRGBComponent) Toggle(_ context.Context) (*model.RGBStatus, error) {
	m.ToggleCalled = true
	if m.Err != nil {
		return nil, m.Err
	}
	return m.ToggleResult, nil
}

// SetBrightness records the call and returns the mock error.
func (m *MockRGBComponent) SetBrightness(_ context.Context, brightness int) error {
	m.SetCalled = true
	m.SetBrightnessValue = &brightness
	return m.Err
}

// SetColor records the call and returns the mock error.
func (m *MockRGBComponent) SetColor(_ context.Context, red, green, blue int) error {
	m.SetCalled = true
	m.SetRedValue = &red
	m.SetGreenValue = &green
	m.SetBlueValue = &blue
	return m.Err
}

// SetColorAndBrightness records the call and returns the mock error.
func (m *MockRGBComponent) SetColorAndBrightness(_ context.Context, red, green, blue, brightness int) error {
	m.SetCalled = true
	m.SetRedValue = &red
	m.SetGreenValue = &green
	m.SetBlueValue = &blue
	m.SetBrightnessValue = &brightness
	return m.Err
}

// Set records the call and returns the mock error.
func (m *MockRGBComponent) Set(_ context.Context, red, green, blue, brightness *int, on *bool) error {
	m.SetCalled = true
	m.SetRedValue = red
	m.SetGreenValue = green
	m.SetBlueValue = blue
	m.SetBrightnessValue = brightness
	m.SetOnValue = on
	return m.Err
}
