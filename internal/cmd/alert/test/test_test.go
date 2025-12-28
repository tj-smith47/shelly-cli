package test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		wantUse      string
		wantShort    string
		wantAliases  []string
		wantHasLong  bool
		wantExample  bool
		wantArgsFunc bool
	}{
		{
			name:         "command properties",
			wantUse:      "test <name>",
			wantShort:    "Test an alert by triggering it",
			wantAliases:  []string{"trigger", "fire"},
			wantHasLong:  true,
			wantExample:  true,
			wantArgsFunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			if cmd == nil {
				t.Fatal("NewCommand returned nil")
			}

			if cmd.Use != tt.wantUse {
				t.Errorf("Use = %q, want %q", cmd.Use, tt.wantUse)
			}

			if cmd.Short != tt.wantShort {
				t.Errorf("Short = %q, want %q", cmd.Short, tt.wantShort)
			}

			if len(cmd.Aliases) != len(tt.wantAliases) {
				t.Errorf("Aliases length = %d, want %d", len(cmd.Aliases), len(tt.wantAliases))
			}
			for i, alias := range tt.wantAliases {
				if i < len(cmd.Aliases) && cmd.Aliases[i] != alias {
					t.Errorf("Alias[%d] = %q, want %q", i, cmd.Aliases[i], alias)
				}
			}

			if tt.wantHasLong && cmd.Long == "" {
				t.Error("Long description is empty")
			}

			if tt.wantExample && cmd.Example == "" {
				t.Error("Example is empty")
			}

			if tt.wantArgsFunc && cmd.Args == nil {
				t.Error("Args function is nil")
			}
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args returns error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "one arg accepted",
			args:    []string{"test-alert"},
			wantErr: false,
		},
		{
			name:    "two args returns error",
			args:    []string{"alert1", "alert2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRun_AlertNotFound(t *testing.T) {
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
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent alert")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should contain 'not found', got: %v", err)
	}
}

func TestRun_NotifyAction(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"test-alert": {
				Name:      "test-alert",
				Device:    "kitchen",
				Condition: "offline",
				Action:    "notify",
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"test-alert"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Testing alert") {
		t.Errorf("expected 'Testing alert' in output, got: %s", output)
	}
	if !strings.Contains(output, "kitchen") {
		t.Errorf("expected device 'kitchen' in output, got: %s", output)
	}
	if !strings.Contains(output, "offline") {
		t.Errorf("expected condition 'offline' in output, got: %s", output)
	}
	if !strings.Contains(output, "[TEST]") {
		t.Errorf("expected '[TEST]' marker in output, got: %s", output)
	}
	if !strings.Contains(output, "Alert test completed") {
		t.Errorf("expected 'Alert test completed' in output, got: %s", output)
	}
}

func TestRun_WebhookAction(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"webhook-alert": {
				Name:      "webhook-alert",
				Device:    "living-room",
				Condition: "power>100",
				Action:    "webhook:http://example.com/alert",
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"webhook-alert"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Would send webhook") {
		t.Errorf("expected webhook message in output, got: %s", output)
	}
	if !strings.Contains(output, "http://example.com/alert") {
		t.Errorf("expected webhook URL in output, got: %s", output)
	}
}

func TestRun_CommandAction(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"command-alert": {
				Name:      "command-alert",
				Device:    "garage",
				Condition: "temperature>30",
				Action:    "command:echo hello",
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"command-alert"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Would execute command") {
		t.Errorf("expected command message in output, got: %s", output)
	}
	if !strings.Contains(output, "echo hello") {
		t.Errorf("expected command in output, got: %s", output)
	}
}

func TestRun_UnknownAction(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"unknown-action": {
				Name:      "unknown-action",
				Device:    "bedroom",
				Condition: "offline",
				Action:    "unknown-action-type",
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"unknown-action"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning messages go to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "Unknown action type") {
		t.Errorf("expected warning for unknown action type in stderr, got: %s", errOutput)
	}
}

func TestRun_DisplaysAlertInfo(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"info-test": {
				Name:      "info-test",
				Device:    "study",
				Condition: "power<10",
				Action:    "notify",
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"info-test"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Check that device, condition, and action are all displayed
	if !strings.Contains(output, "Device: study") {
		t.Errorf("expected 'Device: study' in output, got: %s", output)
	}
	if !strings.Contains(output, "Condition: power<10") {
		t.Errorf("expected 'Condition: power<10' in output, got: %s", output)
	}
	if !strings.Contains(output, "Action: notify") {
		t.Errorf("expected 'Action: notify' in output, got: %s", output)
	}
}

func TestRun_ShortWebhookAction(t *testing.T) {
	t.Parallel()

	// Test webhook action that is exactly 8 characters (edge case)
	// The code checks len > 8, so "webhook:" (exactly 8 chars) falls through to unknown action
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"short-action": {
				Name:      "short-action",
				Device:    "test",
				Condition: "offline",
				Action:    "webhook:", // 8 chars exactly, empty URL - falls through to unknown
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"short-action"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Due to len > 8 check, "webhook:" (8 chars) triggers unknown action warning
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "Unknown action type") {
		t.Errorf("expected unknown action warning for 'webhook:' (8 chars), got stderr: %s", errOutput)
	}
}

func TestRun_ShortCommandAction(t *testing.T) {
	t.Parallel()

	// Test command action that is exactly 8 characters (edge case)
	// The code checks len > 8, so "command:" (exactly 8 chars) falls through to unknown action
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"short-cmd": {
				Name:      "short-cmd",
				Device:    "test",
				Condition: "offline",
				Action:    "command:", // 8 chars exactly, empty command - falls through to unknown
				Enabled:   true,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"short-cmd"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Due to len > 8 check, "command:" (8 chars) triggers unknown action warning
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "Unknown action type") {
		t.Errorf("expected unknown action warning for 'command:' (8 chars), got stderr: %s", errOutput)
	}
}
