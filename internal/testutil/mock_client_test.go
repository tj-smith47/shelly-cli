package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestMockSwitchComponent_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *MockSwitchComponent
		wantNil    bool
		wantErr    bool
		wantOutput bool
	}{
		{
			name: "returns status",
			mock: &MockSwitchComponent{
				StatusResult: &model.SwitchStatus{
					ID:     0,
					Output: true,
					Source: "switch",
				},
			},
			wantNil:    false,
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "returns nil status",
			mock: &MockSwitchComponent{
				StatusResult: nil,
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "returns error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.GetStatus(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("GetStatus() status = %v, wantNil %v", status, tt.wantNil)
			}
			if status != nil && status.Output != tt.wantOutput {
				t.Errorf("GetStatus() output = %v, want %v", status.Output, tt.wantOutput)
			}
		})
	}
}

func TestMockSwitchComponent_GetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mock     *MockSwitchComponent
		wantNil  bool
		wantErr  bool
		wantName string
	}{
		{
			name: "returns config",
			mock: &MockSwitchComponent{
				ConfigResult: &model.SwitchConfig{
					ID:   0,
					Name: Ptr("test-switch"),
				},
			},
			wantNil:  false,
			wantErr:  false,
			wantName: "test-switch",
		},
		{
			name: "returns error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := tt.mock.GetConfig(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (config == nil) != tt.wantNil {
				t.Errorf("GetConfig() config = %v, wantNil %v", config, tt.wantNil)
			}
			if config != nil && config.Name != nil && *config.Name != tt.wantName {
				t.Errorf("GetConfig() name = %v, want %v", *config.Name, tt.wantName)
			}
		})
	}
}

func TestMockSwitchComponent_On(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockSwitchComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockSwitchComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.On(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("On() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.OnCalled {
				t.Error("On() should set OnCalled to true")
			}
		})
	}
}

func TestMockSwitchComponent_Off(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockSwitchComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockSwitchComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Off(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Off() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.OffCalled {
				t.Error("Off() should set OffCalled to true")
			}
		})
	}
}

func TestMockSwitchComponent_Toggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *MockSwitchComponent
		wantNil    bool
		wantErr    bool
		wantOutput bool
	}{
		{
			name: "success",
			mock: &MockSwitchComponent{
				ToggleResult: &model.SwitchStatus{
					ID:     0,
					Output: true,
				},
			},
			wantNil:    false,
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.Toggle(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("Toggle() status = %v, wantNil %v", status, tt.wantNil)
			}
			if !tt.mock.ToggleCalled {
				t.Error("Toggle() should set ToggleCalled to true")
			}
		})
	}
}

func TestMockSwitchComponent_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mock     *MockSwitchComponent
		setValue bool
		wantErr  bool
	}{
		{
			name:     "set to on",
			mock:     &MockSwitchComponent{},
			setValue: true,
			wantErr:  false,
		},
		{
			name:     "set to off",
			mock:     &MockSwitchComponent{},
			setValue: false,
			wantErr:  false,
		},
		{
			name: "error",
			mock: &MockSwitchComponent{
				Err: ErrMock,
			},
			setValue: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Set(context.Background(), tt.setValue)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.SetCalled {
				t.Error("Set() should set SetCalled to true")
			}
			if tt.mock.SetValue != tt.setValue {
				t.Errorf("Set() SetValue = %v, want %v", tt.mock.SetValue, tt.setValue)
			}
		})
	}
}

// Cover component tests

func TestMockCoverComponent_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mock      *MockCoverComponent
		wantNil   bool
		wantErr   bool
		wantState string
	}{
		{
			name: "returns status",
			mock: &MockCoverComponent{
				StatusResult: &model.CoverStatus{
					ID:     0,
					State:  "open",
					Source: "cover",
				},
			},
			wantNil:   false,
			wantErr:   false,
			wantState: "open",
		},
		{
			name: "returns error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.GetStatus(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("GetStatus() status = %v, wantNil %v", status, tt.wantNil)
			}
			if status != nil && status.State != tt.wantState {
				t.Errorf("GetStatus() state = %v, want %v", status.State, tt.wantState)
			}
		})
	}
}

func TestMockCoverComponent_GetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockCoverComponent
		wantNil bool
		wantErr bool
	}{
		{
			name: "returns config",
			mock: &MockCoverComponent{
				ConfigResult: &model.CoverConfig{
					ID:   0,
					Name: Ptr("test-cover"),
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := tt.mock.GetConfig(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (config == nil) != tt.wantNil {
				t.Errorf("GetConfig() config = %v, wantNil %v", config, tt.wantNil)
			}
		})
	}
}

func TestMockCoverComponent_Open(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockCoverComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockCoverComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			duration := 10
			err := tt.mock.Open(context.Background(), &duration)

			if (err != nil) != tt.wantErr {
				t.Errorf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.OpenCalled {
				t.Error("Open() should set OpenCalled to true")
			}
		})
	}
}

func TestMockCoverComponent_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockCoverComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockCoverComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Close(context.Background(), nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.CloseCalled {
				t.Error("Close() should set CloseCalled to true")
			}
		})
	}
}

func TestMockCoverComponent_Stop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockCoverComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockCoverComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Stop(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.StopCalled {
				t.Error("Stop() should set StopCalled to true")
			}
		})
	}
}

func TestMockCoverComponent_GoToPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		mock         *MockCoverComponent
		position     int
		wantErr      bool
		wantPosition int
	}{
		{
			name:         "set position 50",
			mock:         &MockCoverComponent{},
			position:     50,
			wantErr:      false,
			wantPosition: 50,
		},
		{
			name:         "set position 100",
			mock:         &MockCoverComponent{},
			position:     100,
			wantErr:      false,
			wantPosition: 100,
		},
		{
			name: "error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			position: 50,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.GoToPosition(context.Background(), tt.position)

			if (err != nil) != tt.wantErr {
				t.Errorf("GoToPosition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.PositionCalled {
				t.Error("GoToPosition() should set PositionCalled to true")
			}
			if tt.mock.PositionValue != tt.position {
				t.Errorf("GoToPosition() PositionValue = %v, want %v", tt.mock.PositionValue, tt.position)
			}
		})
	}
}

func TestMockCoverComponent_Calibrate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockCoverComponent
		wantErr bool
	}{
		{
			name:    "success",
			mock:    &MockCoverComponent{},
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockCoverComponent{
				Err: ErrMock,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Calibrate(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Calibrate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.CalibrateCalled {
				t.Error("Calibrate() should set CalibrateCalled to true")
			}
		})
	}
}

// Light component tests

func TestMockLightComponent_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *MockLightComponent
		wantNil    bool
		wantErr    bool
		wantOutput bool
	}{
		{
			name: "returns status",
			mock: &MockLightComponent{
				StatusResult: &model.LightStatus{
					ID:         0,
					Output:     true,
					Brightness: Ptr(75),
					Source:     "light",
				},
			},
			wantNil:    false,
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "returns error",
			mock: &MockLightComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.GetStatus(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("GetStatus() status = %v, wantNil %v", status, tt.wantNil)
			}
			if status != nil && status.Output != tt.wantOutput {
				t.Errorf("GetStatus() output = %v, want %v", status.Output, tt.wantOutput)
			}
		})
	}
}

func TestMockLightComponent_GetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockLightComponent
		wantNil bool
		wantErr bool
	}{
		{
			name: "returns config",
			mock: &MockLightComponent{
				ConfigResult: &model.LightConfig{
					ID:   0,
					Name: Ptr("test-light"),
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns error",
			mock: &MockLightComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := tt.mock.GetConfig(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (config == nil) != tt.wantNil {
				t.Errorf("GetConfig() config = %v, wantNil %v", config, tt.wantNil)
			}
		})
	}
}

func TestMockLightComponent_On(t *testing.T) {
	t.Parallel()

	mock := &MockLightComponent{}
	err := mock.On(context.Background())

	if err != nil {
		t.Errorf("On() error = %v, want nil", err)
	}
	if !mock.OnCalled {
		t.Error("On() should set OnCalled to true")
	}
}

func TestMockLightComponent_Off(t *testing.T) {
	t.Parallel()

	mock := &MockLightComponent{}
	err := mock.Off(context.Background())

	if err != nil {
		t.Errorf("Off() error = %v, want nil", err)
	}
	if !mock.OffCalled {
		t.Error("Off() should set OffCalled to true")
	}
}

func TestMockLightComponent_Toggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockLightComponent
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			mock: &MockLightComponent{
				ToggleResult: &model.LightStatus{
					ID:     0,
					Output: true,
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockLightComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.Toggle(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("Toggle() status = %v, wantNil %v", status, tt.wantNil)
			}
			if !tt.mock.ToggleCalled {
				t.Error("Toggle() should set ToggleCalled to true")
			}
		})
	}
}

func TestMockLightComponent_SetBrightnessValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mock           *MockLightComponent
		brightness     int
		wantErr        bool
		wantBrightness int
	}{
		{
			name:           "set brightness 50",
			mock:           &MockLightComponent{},
			brightness:     50,
			wantErr:        false,
			wantBrightness: 50,
		},
		{
			name:           "set brightness 100",
			mock:           &MockLightComponent{},
			brightness:     100,
			wantErr:        false,
			wantBrightness: 100,
		},
		{
			name: "error",
			mock: &MockLightComponent{
				Err: ErrMock,
			},
			brightness: 50,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.SetBrightnessValue(context.Background(), tt.brightness)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetBrightnessValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.SetCalled {
				t.Error("SetBrightnessValue() should set SetCalled to true")
			}
			if tt.mock.SetBrightness == nil {
				t.Error("SetBrightness should be set")
			} else if *tt.mock.SetBrightness != tt.brightness {
				t.Errorf("SetBrightnessValue() SetBrightness = %v, want %v", *tt.mock.SetBrightness, tt.brightness)
			}
		})
	}
}

func TestMockLightComponent_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mock           *MockLightComponent
		brightness     *int
		on             *bool
		wantErr        bool
		wantBrightness *int
		wantOn         *bool
	}{
		{
			name:           "set brightness and on",
			mock:           &MockLightComponent{},
			brightness:     Ptr(75),
			on:             Ptr(true),
			wantErr:        false,
			wantBrightness: Ptr(75),
			wantOn:         Ptr(true),
		},
		{
			name:           "set only brightness",
			mock:           &MockLightComponent{},
			brightness:     Ptr(50),
			on:             nil,
			wantErr:        false,
			wantBrightness: Ptr(50),
			wantOn:         nil,
		},
		{
			name: "error",
			mock: &MockLightComponent{
				Err: ErrMock,
			},
			brightness: Ptr(50),
			on:         Ptr(true),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Set(context.Background(), tt.brightness, tt.on)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.SetCalled {
				t.Error("Set() should set SetCalled to true")
			}
		})
	}
}

// RGB component tests

func TestMockRGBComponent_GetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *MockRGBComponent
		wantNil    bool
		wantErr    bool
		wantOutput bool
	}{
		{
			name: "returns status",
			mock: &MockRGBComponent{
				StatusResult: &model.RGBStatus{
					ID:         0,
					Output:     true,
					Brightness: Ptr(75),
					RGB: &model.RGBColor{
						Red:   255,
						Green: 128,
						Blue:  64,
					},
					Source: "rgb",
				},
			},
			wantNil:    false,
			wantErr:    false,
			wantOutput: true,
		},
		{
			name: "returns error",
			mock: &MockRGBComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.GetStatus(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("GetStatus() status = %v, wantNil %v", status, tt.wantNil)
			}
			if status != nil && status.Output != tt.wantOutput {
				t.Errorf("GetStatus() output = %v, want %v", status.Output, tt.wantOutput)
			}
		})
	}
}

func TestMockRGBComponent_GetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockRGBComponent
		wantNil bool
		wantErr bool
	}{
		{
			name: "returns config",
			mock: &MockRGBComponent{
				ConfigResult: &model.RGBConfig{
					ID:   0,
					Name: Ptr("test-rgb"),
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "returns error",
			mock: &MockRGBComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := tt.mock.GetConfig(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (config == nil) != tt.wantNil {
				t.Errorf("GetConfig() config = %v, wantNil %v", config, tt.wantNil)
			}
		})
	}
}

func TestMockRGBComponent_On(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{}
	err := mock.On(context.Background())

	if err != nil {
		t.Errorf("On() error = %v, want nil", err)
	}
	if !mock.OnCalled {
		t.Error("On() should set OnCalled to true")
	}
}

func TestMockRGBComponent_Off(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{}
	err := mock.Off(context.Background())

	if err != nil {
		t.Errorf("Off() error = %v, want nil", err)
	}
	if !mock.OffCalled {
		t.Error("Off() should set OffCalled to true")
	}
}

func TestMockRGBComponent_Toggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mock    *MockRGBComponent
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			mock: &MockRGBComponent{
				ToggleResult: &model.RGBStatus{
					ID:     0,
					Output: true,
				},
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "error",
			mock: &MockRGBComponent{
				Err: ErrMock,
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := tt.mock.Toggle(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (status == nil) != tt.wantNil {
				t.Errorf("Toggle() status = %v, wantNil %v", status, tt.wantNil)
			}
			if !tt.mock.ToggleCalled {
				t.Error("Toggle() should set ToggleCalled to true")
			}
		})
	}
}

func TestMockRGBComponent_SetBrightness(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{}
	err := mock.SetBrightness(context.Background(), 75)

	if err != nil {
		t.Errorf("SetBrightness() error = %v, want nil", err)
	}
	if !mock.SetCalled {
		t.Error("SetBrightness() should set SetCalled to true")
	}
	if mock.SetBrightnessValue == nil || *mock.SetBrightnessValue != 75 {
		t.Errorf("SetBrightnessValue = %v, want 75", mock.SetBrightnessValue)
	}
}

func TestMockRGBComponent_SetColor(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{}
	err := mock.SetColor(context.Background(), 255, 128, 64)

	if err != nil {
		t.Errorf("SetColor() error = %v, want nil", err)
	}
	if !mock.SetCalled {
		t.Error("SetColor() should set SetCalled to true")
	}
	if mock.SetRedValue == nil || *mock.SetRedValue != 255 {
		t.Errorf("SetRedValue = %v, want 255", mock.SetRedValue)
	}
	if mock.SetGreenValue == nil || *mock.SetGreenValue != 128 {
		t.Errorf("SetGreenValue = %v, want 128", mock.SetGreenValue)
	}
	if mock.SetBlueValue == nil || *mock.SetBlueValue != 64 {
		t.Errorf("SetBlueValue = %v, want 64", mock.SetBlueValue)
	}
}

func TestMockRGBComponent_SetColorAndBrightness(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{}
	err := mock.SetColorAndBrightness(context.Background(), 255, 128, 64, 80)

	if err != nil {
		t.Errorf("SetColorAndBrightness() error = %v, want nil", err)
	}
	if !mock.SetCalled {
		t.Error("SetColorAndBrightness() should set SetCalled to true")
	}
	if mock.SetRedValue == nil || *mock.SetRedValue != 255 {
		t.Errorf("SetRedValue = %v, want 255", mock.SetRedValue)
	}
	if mock.SetGreenValue == nil || *mock.SetGreenValue != 128 {
		t.Errorf("SetGreenValue = %v, want 128", mock.SetGreenValue)
	}
	if mock.SetBlueValue == nil || *mock.SetBlueValue != 64 {
		t.Errorf("SetBlueValue = %v, want 64", mock.SetBlueValue)
	}
	if mock.SetBrightnessValue == nil || *mock.SetBrightnessValue != 80 {
		t.Errorf("SetBrightnessValue = %v, want 80", mock.SetBrightnessValue)
	}
}

func TestMockRGBComponent_Set(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mock       *MockRGBComponent
		red        *int
		green      *int
		blue       *int
		brightness *int
		on         *bool
		wantErr    bool
	}{
		{
			name:       "set all values",
			mock:       &MockRGBComponent{},
			red:        Ptr(255),
			green:      Ptr(128),
			blue:       Ptr(64),
			brightness: Ptr(80),
			on:         Ptr(true),
			wantErr:    false,
		},
		{
			name:       "set only color",
			mock:       &MockRGBComponent{},
			red:        Ptr(255),
			green:      Ptr(0),
			blue:       Ptr(0),
			brightness: nil,
			on:         nil,
			wantErr:    false,
		},
		{
			name: "error",
			mock: &MockRGBComponent{
				Err: ErrMock,
			},
			red:     Ptr(255),
			green:   Ptr(128),
			blue:    Ptr(64),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mock.Set(context.Background(), tt.red, tt.green, tt.blue, tt.brightness, tt.on)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.mock.SetCalled {
				t.Error("Set() should set SetCalled to true")
			}
		})
	}
}

func TestMockRGBComponent_SetWithError(t *testing.T) {
	t.Parallel()

	mock := &MockRGBComponent{Err: ErrMock}

	// Test all methods with error
	if err := mock.On(context.Background()); !errors.Is(err, ErrMock) {
		t.Errorf("On() error = %v, want %v", err, ErrMock)
	}
	if err := mock.Off(context.Background()); !errors.Is(err, ErrMock) {
		t.Errorf("Off() error = %v, want %v", err, ErrMock)
	}
	if err := mock.SetBrightness(context.Background(), 50); !errors.Is(err, ErrMock) {
		t.Errorf("SetBrightness() error = %v, want %v", err, ErrMock)
	}
	if err := mock.SetColor(context.Background(), 255, 128, 64); !errors.Is(err, ErrMock) {
		t.Errorf("SetColor() error = %v, want %v", err, ErrMock)
	}
	if err := mock.SetColorAndBrightness(context.Background(), 255, 128, 64, 80); !errors.Is(err, ErrMock) {
		t.Errorf("SetColorAndBrightness() error = %v, want %v", err, ErrMock)
	}
}

func TestMockLightComponent_SetWithError(t *testing.T) {
	t.Parallel()

	mock := &MockLightComponent{Err: ErrMock}

	// Test all methods with error
	if err := mock.On(context.Background()); !errors.Is(err, ErrMock) {
		t.Errorf("On() error = %v, want %v", err, ErrMock)
	}
	if err := mock.Off(context.Background()); !errors.Is(err, ErrMock) {
		t.Errorf("Off() error = %v, want %v", err, ErrMock)
	}
	if err := mock.SetBrightnessValue(context.Background(), 50); !errors.Is(err, ErrMock) {
		t.Errorf("SetBrightnessValue() error = %v, want %v", err, ErrMock)
	}
}
