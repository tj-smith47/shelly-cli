package provisioning

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

const (
	testSSID     = "MyNetwork"
	testPassword = "secret"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}

	m := New(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.step != StepInstructions {
		t.Errorf("step = %d, want %d", m.step, StepInstructions)
	}
}

func TestNew_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := Deps{Ctx: nil, Svc: &shelly.Service{}}
	New(deps)
}

func TestNew_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := Deps{Ctx: context.Background(), Svc: nil}
	New(deps)
}

func TestDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    Deps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    Deps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    Deps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    Deps{Ctx: context.Background(), Svc: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.deps.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModel_Reset(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepCredentials
	m.ssid = testSSID
	m.password = testPassword
	m.err = errors.New("some error")

	updated := m.Reset()

	if updated.step != StepInstructions {
		t.Errorf("step = %d, want %d", updated.step, StepInstructions)
	}
	if updated.ssid != "" {
		t.Error("ssid should be empty")
	}
	if updated.password != "" {
		t.Error("password should be empty")
	}
	if updated.err != nil {
		t.Error("err should be nil")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetSize(100, 50)

	if updated.Width != 100 {
		t.Errorf("width = %d, want 100", updated.Width)
	}
	if updated.Height != 50 {
		t.Errorf("height = %d, want 50", updated.Height)
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.focused {
		t.Error("should not be focused initially")
	}

	updated := m.SetFocused(true)

	if !updated.focused {
		t.Error("should be focused after SetFocused(true)")
	}
}

func TestModel_Update_DeviceFound(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepWaiting
	m.polling = true
	msg := DeviceFoundMsg{
		DeviceInfo: &shelly.ProvisioningDeviceInfo{
			Model: "ShellyPlus1",
			MAC:   "AA:BB:CC:DD:EE:FF",
		},
	}

	updated, _ := m.Update(msg)

	if updated.polling {
		t.Error("should not be polling after device found")
	}
	if updated.step != StepCredentials {
		t.Errorf("step = %d, want %d", updated.step, StepCredentials)
	}
	if updated.deviceInfo == nil {
		t.Error("deviceInfo should be set")
	}
}

func TestModel_Update_DeviceFoundError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepWaiting
	m.polling = true
	msg := DeviceFoundMsg{
		Err: errors.New("connection failed"),
	}

	updated, cmd := m.Update(msg)

	if updated.polling {
		t.Error("should not be polling after error")
	}
	if updated.step != StepWaiting {
		t.Error("should remain on waiting step to retry")
	}
	if cmd == nil {
		t.Error("should return poll command to retry")
	}
}

func TestModel_Update_Configured(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepConfiguring
	msg := ConfiguredMsg{}

	updated, _ := m.Update(msg)

	if updated.step != StepSuccess {
		t.Errorf("step = %d, want %d", updated.step, StepSuccess)
	}
}

func TestModel_Update_ConfiguredError(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepConfiguring
	testErr := errors.New("config failed")
	msg := ConfiguredMsg{Err: testErr}

	updated, _ := m.Update(msg)

	if updated.step != StepError {
		t.Errorf("step = %d, want %d", updated.step, StepError)
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestModel_Update_Poll(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepWaiting
	m.focused = true

	updated, cmd := m.Update(PollMsg{})

	if cmd == nil {
		t.Error("should return check device command")
	}
	_ = updated // Just check no panic
}

func TestModel_Update_PollNotWaiting(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepInstructions

	_, cmd := m.Update(PollMsg{})

	if cmd != nil {
		t.Error("should not poll when not in waiting step")
	}
}

func TestModel_HandleKey_Instructions(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepInstructions

	// Enter moves to waiting
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if updated.step != StepWaiting {
		t.Errorf("step = %d, want %d", updated.step, StepWaiting)
	}
	if cmd == nil {
		t.Error("should return check device command")
	}

	// Esc resets
	m.step = StepInstructions
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if updated.step != StepInstructions {
		t.Error("should reset on Esc")
	}
}

func TestModel_HandleKey_Credentials(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials

	// Tab switches field
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if updated.inputField != 1 {
		t.Errorf("inputField = %d, want 1", updated.inputField)
	}

	// Tab again switches back
	updated, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if updated.inputField != 0 {
		t.Errorf("inputField = %d, want 0", updated.inputField)
	}
}

func TestModel_HandleKey_Credentials_TypeSSID(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials
	m.inputField = 0

	// Type characters
	m, _ = m.Update(tea.KeyPressMsg{Code: 'M'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'y'})

	if m.ssid != "My" {
		t.Errorf("ssid = %q, want %q", m.ssid, "My")
	}
}

func TestModel_HandleKey_Credentials_TypePassword(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials
	m.inputField = 1

	// Type characters
	m, _ = m.Update(tea.KeyPressMsg{Code: 'p'})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'w'})

	if m.password != "pw" {
		t.Errorf("password = %q, want %q", m.password, "pw")
	}
}

func TestModel_HandleKey_Credentials_Backspace(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials
	m.inputField = 0
	m.ssid = "ABC"

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if updated.ssid != "AB" {
		t.Errorf("ssid = %q, want %q", updated.ssid, "AB")
	}
}

func TestModel_HandleKey_Credentials_Enter(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials
	m.ssid = testSSID
	m.password = testPassword

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.step != StepConfiguring {
		t.Errorf("step = %d, want %d", updated.step, StepConfiguring)
	}
	if cmd == nil {
		t.Error("should return configure command")
	}
}

func TestModel_HandleKey_Credentials_EnterNoSSID(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepCredentials
	m.ssid = "" // Empty

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.step != StepCredentials {
		t.Error("should not proceed without SSID")
	}
	if cmd != nil {
		t.Error("should not return command without SSID")
	}
}

func TestModel_HandleKey_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepSuccess

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.step != StepInstructions {
		t.Error("should reset on Enter")
	}
}

func TestModel_HandleKey_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.step = StepError

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'r'})

	if updated.step != StepInstructions {
		t.Error("should reset on R")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.step = StepInstructions

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if updated.step != StepInstructions {
		t.Error("should not respond when not focused")
	}
	if cmd != nil {
		t.Error("should not return command when not focused")
	}
}

func TestModel_View_Instructions(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepInstructions
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_Waiting(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepWaiting
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_Credentials(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepCredentials
	m.deviceInfo = &shelly.ProvisioningDeviceInfo{
		Model: "ShellyPlus1",
		MAC:   "AA:BB:CC:DD:EE:FF",
	}
	m.ssid = testSSID
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_Configuring(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepConfiguring
	m.ssid = testSSID
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_Success(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepSuccess
	m.ssid = testSSID
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_View_Error(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepError
	m.err = errors.New("connection failed")
	m = m.SetSize(80, 40)

	view := m.View()

	if view == "" {
		t.Error("View() should not be empty")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.step = StepCredentials
	m.deviceInfo = &shelly.ProvisioningDeviceInfo{Model: "ShellyPlus1"}
	m.ssid = testSSID
	m.err = errors.New("test error")
	m.polling = true

	if m.Step() != StepCredentials {
		t.Errorf("Step() = %d, want %d", m.Step(), StepCredentials)
	}
	if m.DeviceInfo() == nil {
		t.Error("DeviceInfo() should not be nil")
	}
	if m.SSID() != testSSID {
		t.Errorf("SSID() = %q, want %q", m.SSID(), testSSID)
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if !m.Polling() {
		t.Error("Polling() should be true")
	}
}

func TestModel_GetStepText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		step Step
		want string
	}{
		{StepInstructions, "Step 1 of 4: Connect to Device"},
		{StepWaiting, "Step 2 of 4: Detecting Device"},
		{StepCredentials, "Step 3 of 4: Enter WiFi Credentials"},
		{StepConfiguring, "Step 4 of 4: Configuring Device"},
		{StepSuccess, "Setup Complete"},
		{StepError, "Setup Failed"},
	}

	for _, tt := range tests {
		m := newTestModel()
		m.step = tt.step
		got := m.getStepText()
		if got != tt.want {
			t.Errorf("getStepText() for step %d = %q, want %q", tt.step, got, tt.want)
		}
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Title.Render("test")
	_ = styles.Step.Render("test")
	_ = styles.Text.Render("test")
	_ = styles.Highlight.Render("test")
	_ = styles.Input.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Success.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Warning.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
