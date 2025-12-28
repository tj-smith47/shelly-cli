package list

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want 'list'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Verify aliases
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases are empty")
	}

	// Verify example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept no args
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("Expected no error with no args, got: %v", err)
	}

	// Should reject args
	err = cmd.Args(cmd, []string{"extra"})
	if err == nil {
		t.Error("Expected error when args provided")
	}
}

func TestRun_NoAlerts(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "No alerts configured") {
		t.Errorf("expected 'No alerts configured' message, got: %s", output)
	}
	if !strings.Contains(output, "shelly alert create") {
		t.Errorf("expected hint about creating alerts, got: %s", output)
	}
}

func TestRun_WithAlerts(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"test-alert": {
				Name:        "test-alert",
				Device:      "living-room",
				Condition:   "offline",
				Action:      "notify",
				Enabled:     true,
				Description: "Test alert description",
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "test-alert") {
		t.Errorf("expected alert name in output, got: %s", output)
	}
	if !strings.Contains(output, "living-room") {
		t.Errorf("expected device name in output, got: %s", output)
	}
	if !strings.Contains(output, "offline") {
		t.Errorf("expected condition in output, got: %s", output)
	}
	if !strings.Contains(output, "notify") {
		t.Errorf("expected action in output, got: %s", output)
	}
	if !strings.Contains(output, "enabled") {
		t.Errorf("expected status in output, got: %s", output)
	}
	if !strings.Contains(output, "Test alert description") {
		t.Errorf("expected description in output, got: %s", output)
	}
}

func TestRun_DisabledAlert(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"disabled-alert": {
				Name:      "disabled-alert",
				Device:    "kitchen",
				Condition: "power>100",
				Action:    "webhook:http://example.com",
				Enabled:   false,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "disabled") {
		t.Errorf("expected 'disabled' status in output, got: %s", output)
	}
}

func TestRun_SnoozedAlert(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	// Set snooze time to 1 hour from now
	snoozedUntil := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"snoozed-alert": {
				Name:         "snoozed-alert",
				Device:       "bedroom",
				Condition:    "temperature>30",
				Action:       "notify",
				Enabled:      true,
				SnoozedUntil: snoozedUntil,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "snoozed until") {
		t.Errorf("expected 'snoozed until' in output, got: %s", output)
	}
}

func TestRun_ExpiredSnooze(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	// Set snooze time to 1 hour ago (expired)
	snoozedUntil := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"expired-snooze-alert": {
				Name:         "expired-snooze-alert",
				Device:       "garage",
				Condition:    "offline",
				Action:       "notify",
				Enabled:      true,
				SnoozedUntil: snoozedUntil,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Expired snooze should show as enabled, not snoozed
	if !strings.Contains(output, "enabled") {
		t.Errorf("expected 'enabled' status for expired snooze, got: %s", output)
	}
}

func TestRun_MultipleAlerts(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"alert-1": {
				Name:      "alert-1",
				Device:    "device-1",
				Condition: "offline",
				Action:    "notify",
				Enabled:   true,
			},
			"alert-2": {
				Name:      "alert-2",
				Device:    "device-2",
				Condition: "power>50",
				Action:    "command:reboot",
				Enabled:   false,
			},
			"alert-3": {
				Name:        "alert-3",
				Device:      "device-3",
				Condition:   "temperature>35",
				Action:      "webhook:http://localhost:8080/alert",
				Enabled:     true,
				Description: "High temperature warning",
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()

	// Check count message
	if !strings.Contains(output, "Configured Alerts") && !strings.Contains(output, "3") {
		t.Errorf("expected count of 3 alerts in output, got: %s", output)
	}

	// Check all alerts are listed
	if !strings.Contains(output, "alert-1") {
		t.Errorf("expected alert-1 in output, got: %s", output)
	}
	if !strings.Contains(output, "alert-2") {
		t.Errorf("expected alert-2 in output, got: %s", output)
	}
	if !strings.Contains(output, "alert-3") {
		t.Errorf("expected alert-3 in output, got: %s", output)
	}
}

func TestRun_AlertWithoutDescription(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"no-desc-alert": {
				Name:        "no-desc-alert",
				Device:      "test-device",
				Condition:   "offline",
				Action:      "notify",
				Enabled:     true,
				Description: "", // No description
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should not have "Description:" line when description is empty
	if strings.Contains(output, "Description:") {
		t.Errorf("should not show Description line when empty, got: %s", output)
	}
}

func TestRun_InvalidSnoozedUntil(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"invalid-snooze-alert": {
				Name:         "invalid-snooze-alert",
				Device:       "test-device",
				Condition:    "offline",
				Action:       "notify",
				Enabled:      true,
				SnoozedUntil: "invalid-date-format",
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// With invalid snooze date, should show as enabled (fallback)
	if !strings.Contains(output, "enabled") {
		t.Errorf("expected 'enabled' status for invalid snooze, got: %s", output)
	}
}
