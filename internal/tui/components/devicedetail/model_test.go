package devicedetail

import (
	"context"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNew(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)

	if m.Visible() {
		t.Error("expected detail view to be hidden on creation")
	}
}

func TestShowAndHide(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)

	device := model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
	}

	// Test Show
	m, _ = m.Show(device)
	if !m.Visible() {
		t.Error("expected detail view to be visible after Show()")
	}
	if m.device == nil || m.device.Name != "test-device" {
		t.Error("expected device to be set after Show()")
	}

	// Test Hide
	m = m.Hide()
	if m.Visible() {
		t.Error("expected detail view to be hidden after Hide()")
	}
	if m.device != nil {
		t.Error("expected device to be nil after Hide()")
	}
}

func TestSetSize(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)
	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("expected height 50, got %d", m.height)
	}
}

func TestViewWhenHidden(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)

	view := m.View()
	if view != "" {
		t.Error("expected empty view when hidden")
	}
}

func TestViewWhenLoading(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)
	m = m.SetSize(80, 40)
	m.visible = true
	m.loading = true

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view when loading")
	}
}

func TestMsgHandling(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)
	m.visible = true
	m.loading = true

	// Simulate receiving device details
	device := model.Device{
		Name:       "test-device",
		Address:    "192.168.1.100",
		Generation: 2,
		Type:       "plug",
		Model:      "SHSW-25",
	}

	status := &model.MonitoringSnapshot{
		Device:    "test-device",
		Timestamp: time.Now(),
		Online:    true,
	}

	msg := Msg{
		Device: device,
		Status: status,
		Config: map[string]any{"wifi": map[string]any{}},
	}

	m, _ = m.Update(msg)

	if m.loading {
		t.Error("expected loading to be false after receiving Msg")
	}
	if m.device == nil {
		t.Error("expected device to be set after receiving Msg")
	}
	if m.status == nil {
		t.Error("expected status to be set after receiving Msg")
	}
	if m.config == nil {
		t.Error("expected config to be set after receiving Msg")
	}
}

func TestMsgWithError(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)
	m.visible = true
	m.loading = true

	device := model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
	}

	msg := Msg{
		Device: device,
		Err:    context.DeadlineExceeded,
	}

	m, _ = m.Update(msg)

	if m.loading {
		t.Error("expected loading to be false after receiving error Msg")
	}
	if m.err == nil {
		t.Error("expected error to be set after receiving error Msg")
	}
}

func TestUpdateWhenHidden(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)

	// Updates should be no-ops when hidden
	_, cmd := m.Update(Msg{})

	if cmd != nil {
		t.Error("expected nil command when hidden")
	}
}

func TestRenderRow(t *testing.T) {
	t.Parallel()
	deps := Deps{
		Ctx: context.Background(),
		Svc: shelly.NewService(),
	}

	m := New(deps)

	// Test with label
	row := m.renderRow("Label", "Value")
	if row == "" {
		t.Error("expected non-empty row")
	}

	// Test without label (indent only)
	row = m.renderRow("", "Value")
	if row == "" {
		t.Error("expected non-empty row without label")
	}
}

func TestFormatJSON(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"key": "value",
	}

	result := FormatJSON(data)
	if result == "" {
		t.Error("expected non-empty JSON output")
	}
}
