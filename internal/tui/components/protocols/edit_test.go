package protocols

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/tui/components/editmodal"
)

func TestNewMQTTEditModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)

	testutil.AssertFalse(t, model.IsVisible(), "should not be visible initially")
}

func TestMQTTEditModel_ShowHide(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)

	// Show with data
	data := &MQTTData{
		Enable:      true,
		Server:      "mqtt.example.com:1883",
		User:        "testuser",
		ClientID:    "test-client",
		TopicPrefix: "shelly/test",
		SSLCA:       "",
		Connected:   true,
	}

	model = model.Show("192.168.1.100", data)
	testutil.AssertTrue(t, model.IsVisible(), "should be visible after Show")

	// Hide
	model = model.Hide()
	testutil.AssertFalse(t, model.IsVisible(), "should not be visible after Hide")
}

func TestMQTTEditModel_SetSize(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.SetSize(100, 50)

	testutil.AssertEqual(t, 100, model.Width)
	testutil.AssertEqual(t, 50, model.Height)
}

func TestMQTTEditModel_Init(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestMQTTEditModel_Update_NotVisible(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)

	// Should not process keys when not visible
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := model.Update(keyMsg)

	testutil.AssertFalse(t, updated.IsVisible(), "should remain hidden")
	if cmd != nil {
		t.Error("cmd should be nil when not visible")
	}
}

func TestMQTTEditModel_Update_EscClose(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.Show("192.168.1.100", &MQTTData{})

	// Press Esc to close
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEscape}
	updated, cmd := model.Update(keyMsg)

	testutil.AssertFalse(t, updated.IsVisible(), "should be hidden after Esc")
	if cmd == nil {
		t.Fatal("should return close message cmd")
	}

	// Execute cmd and check message type
	msg := cmd()
	_, ok := msg.(MQTTEditClosedMsg)
	testutil.AssertTrue(t, ok, "should return MQTTEditClosedMsg")
}

func TestMQTTEditModel_Update_TabNavigation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.Show("192.168.1.100", &MQTTData{})

	// Start at enable field
	testutil.AssertEqual(t, int(MQTTFieldEnable), model.Cursor)

	// Tab to next field
	keyMsg := tea.KeyPressMsg{Code: tea.KeyTab}
	updated, _ := model.Update(keyMsg)
	testutil.AssertEqual(t, int(MQTTFieldServer), updated.Cursor)

	// Tab again
	updated, _ = updated.Update(keyMsg)
	testutil.AssertEqual(t, int(MQTTFieldUser), updated.Cursor)
}

func TestMQTTEditModel_Update_ShiftTabNavigation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.Show("192.168.1.100", &MQTTData{})

	// Move to second field first
	tabMsg := tea.KeyPressMsg{Code: tea.KeyTab}
	model, _ = model.Update(tabMsg)
	testutil.AssertEqual(t, int(MQTTFieldServer), model.Cursor)

	// Shift+Tab back
	shiftTabMsg := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	updated, _ := model.Update(shiftTabMsg)
	testutil.AssertEqual(t, int(MQTTFieldEnable), updated.Cursor)
}

func TestMQTTEditModel_ValidationServerRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.Show("192.168.1.100", &MQTTData{Enable: false})

	// Enable MQTT but don't set server
	model.enableToggle = model.enableToggle.SetValue(true)

	// Try to save
	updated, _ := model.save()

	testutil.AssertError(t, updated.Err)
	testutil.AssertContains(t, updated.Err.Error(), "server address is required")
}

func TestMQTTEditModel_ValidationTopicPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		topicPrefix string
		wantErr     bool
		errContains string
	}{
		{"valid prefix", "shelly/test", false, ""},
		{"dollar prefix", "$shelly/test", true, "$"},
		{"hash char", "shelly/#/test", true, "#"},
		{"plus char", "shelly/+/test", true, "+"},
		{"percent char", "shelly/%/test", true, "%"},
		{"question char", "shelly/?/test", true, "?"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := &shelly.Service{}

			model := NewMQTTEditModel(ctx, svc)
			model = model.Show("192.168.1.100", &MQTTData{Enable: false})

			// Set server and topic prefix
			model.serverInput = model.serverInput.SetValue("mqtt.example.com")
			model.topicPrefixInput = model.topicPrefixInput.SetValue(tc.topicPrefix)

			updated, _ := model.save()

			if tc.wantErr {
				testutil.AssertError(t, updated.Err)
				testutil.AssertContains(t, updated.Err.Error(), tc.errContains)
			} else if updated.Err != nil && !updated.Saving {
				// Note: saving will still fail because we're using a nil service
				// but the validation should pass
				t.Errorf("unexpected validation error: %v", updated.Err)
			}
		})
	}
}

func TestMQTTEditModel_TLSDropdownMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		sslca    string
		expected int
	}{
		{"no TLS", "", 0},
		{"TLS no verify", "*", 1},
		{"TLS default CA", "ca.pem", 2},
		{"TLS user CA", "user_ca.pem", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := &shelly.Service{}

			model := NewMQTTEditModel(ctx, svc)
			data := &MQTTData{SSLCA: tc.sslca}
			model = model.Show("192.168.1.100", data)

			testutil.AssertEqual(t, tc.expected, model.tlsDropdown.Selected())
		})
	}
}

func TestMQTTEditModel_GetSSLCAFromDropdown(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		selected int
		expected string
	}{
		{0, ""},
		{1, "*"},
		{2, "ca.pem"},
		{3, "user_ca.pem"},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		svc := &shelly.Service{}

		model := NewMQTTEditModel(ctx, svc)
		model.tlsDropdown = model.tlsDropdown.SetSelected(tc.selected)

		testutil.AssertEqual(t, tc.expected, model.getSSLCAFromDropdown())
	}
}

func TestMQTTEditModel_View_NotVisible(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	view := model.View()

	testutil.AssertEqual(t, "", view)
}

func TestMQTTEditModel_View_Visible(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.SetSize(80, 40)
	model = model.Show("192.168.1.100", &MQTTData{
		Enable:    true,
		Server:    "mqtt.example.com:1883",
		Connected: true,
	})

	view := model.View()

	testutil.AssertContains(t, view, "MQTT Configuration")
	testutil.AssertContains(t, view, "Server:")
	testutil.AssertContains(t, view, "Connected")
}

func TestMQTTEditModel_SaveResultMsg(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := &shelly.Service{}

	model := NewMQTTEditModel(ctx, svc)
	model = model.Show("192.168.1.100", &MQTTData{})
	model.Saving = true

	// Simulate successful save result
	saveMsg := MQTTEditSaveResultMsg{Err: nil}
	updated, cmd := model.Update(saveMsg)

	testutil.AssertFalse(t, updated.IsVisible(), "should be hidden after save")
	testutil.AssertFalse(t, updated.Saving, "saving should be false")
	if cmd == nil {
		t.Error("should return close cmd")
	}
}

func TestMQTTEditStyles(t *testing.T) {
	t.Parallel()

	styles := editmodal.DefaultStyles()

	// Verify styles can be created (they're lipgloss.Style which are structs, not pointers)
	// Just verify they have some expected characteristics
	_ = styles.Modal.Render("test")
	_ = styles.Title.Render("test")
	_ = styles.Label.Render("test")
}
