package watch

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantUse     string
		wantShort   string
		wantAliases []string
		wantHasLong bool
		wantExample bool
	}{
		{
			name:        "command properties",
			wantUse:     "watch",
			wantShort:   "Monitor alerts in real-time",
			wantAliases: []string{"monitor", "daemon", "run"},
			wantHasLong: true,
			wantExample: true,
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
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{
			name:         "interval flag",
			flagName:     "interval",
			shorthand:    "i",
			defaultValue: "30s",
		},
		{
			name:         "once flag",
			flagName:     "once",
			shorthand:    "",
			defaultValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// watch command accepts no positional args (uses flags only)
	// It should not have an Args validator that rejects zero args
	if cmd.Args != nil {
		// If Args is set, test that zero args is acceptable
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Expected no error with no args, got: %v", err)
		}
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
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning messages go to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "No alerts configured") {
		t.Errorf("expected 'No alerts configured' message in stderr, got: %s", errOutput)
	}
	// Info/hint goes to stdout
	output := out.String()
	if !strings.Contains(output, "shelly alert create") {
		t.Errorf("expected hint about creating alerts, got: %s", output)
	}
}

func TestRun_AllDisabledAlerts(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"disabled-alert": {
				Name:      "disabled-alert",
				Device:    "kitchen",
				Condition: "offline",
				Action:    "notify",
				Enabled:   false,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning messages go to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "No enabled alerts") {
		t.Errorf("expected 'No enabled alerts' message in stderr, got: %s", errOutput)
	}
}

func TestRun_AllSnoozedAlerts(t *testing.T) {
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
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Warning messages go to stderr
	errOutput := errOut.String()
	if !strings.Contains(errOutput, "No enabled alerts") || !strings.Contains(errOutput, "snoozed") {
		t.Errorf("expected message about snoozed alerts in stderr, got: %s", errOutput)
	}
}

func TestRun_OnceMode(t *testing.T) {
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
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// This should complete immediately due to --once flag
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Alert monitor started") {
		t.Errorf("expected 'Alert monitor started' message, got: %s", output)
	}
	if !strings.Contains(output, "Monitoring 1 alert") {
		t.Errorf("expected 'Monitoring 1 alert' message, got: %s", output)
	}
}

func TestRun_MultipleEnabledAlerts(t *testing.T) {
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
				Condition: "power>100",
				Action:    "notify",
				Enabled:   true,
			},
			"alert-3": {
				Name:      "alert-3",
				Device:    "device-3",
				Condition: "temperature>35",
				Action:    "notify",
				Enabled:   false, // disabled
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should count only enabled alerts (2)
	if !strings.Contains(output, "Monitoring 2 alert") {
		t.Errorf("expected 'Monitoring 2 alert' message, got: %s", output)
	}
}

func TestRun_CustomInterval(t *testing.T) {
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
	cmd.SetArgs([]string{"--once", "--interval", "1m"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "1m0s") {
		t.Errorf("expected custom interval '1m0s' in output, got: %s", output)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
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

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithCancel(context.Background())

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "100ms"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Run with the cancellable context
	cmd.SetContext(ctx)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Alert monitor stopped") {
		t.Errorf("expected 'Alert monitor stopped' message on context cancellation, got: %s", output)
	}
}

func TestRun_CtrlCMessage(t *testing.T) {
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
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Ctrl+C") {
		t.Errorf("expected 'Ctrl+C to stop' message, got: %s", output)
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Get flag values
	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 30*time.Second {
		t.Errorf("interval default = %v, want 30s", interval)
	}

	once, err := cmd.Flags().GetBool("once")
	if err != nil {
		t.Fatalf("failed to get once flag: %v", err)
	}
	if once {
		t.Error("once default should be false")
	}
}

func TestRun_MixedEnabledSnoozedDisabled(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	snoozedUntil := time.Now().Add(1 * time.Hour).Format(time.RFC3339)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"enabled": {
				Name:      "enabled",
				Device:    "device-1",
				Condition: "offline",
				Action:    "notify",
				Enabled:   true,
			},
			"disabled": {
				Name:      "disabled",
				Device:    "device-2",
				Condition: "power>100",
				Action:    "notify",
				Enabled:   false,
			},
			"snoozed": {
				Name:         "snoozed",
				Device:       "device-3",
				Condition:    "temperature>35",
				Action:       "notify",
				Enabled:      true,
				SnoozedUntil: snoozedUntil,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Should count only enabled and not snoozed (1)
	if !strings.Contains(output, "Monitoring 1 alert") {
		t.Errorf("expected 'Monitoring 1 alert' message (only enabled, not snoozed), got: %s", output)
	}
}

func TestRun_ExpiredSnoozeCountsAsEnabled(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	// Expired snooze (in the past)
	expiredSnooze := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	cfg := &config.Config{
		Alerts: map[string]config.Alert{
			"expired-snooze": {
				Name:         "expired-snooze",
				Device:       "device-1",
				Condition:    "offline",
				Action:       "notify",
				Enabled:      true,
				SnoozedUntil: expiredSnooze,
			},
		},
	}
	mgr := config.NewTestManager(cfg)

	f := cmdutil.NewFactory().SetIOStreams(ios).SetConfigManager(mgr)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--once"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	// Expired snooze should count as enabled
	if !strings.Contains(output, "Monitoring 1 alert") {
		t.Errorf("expected expired snooze to count as enabled, got: %s", output)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify Long description contains expected information
	longDesc := cmd.Long
	if !strings.Contains(longDesc, "offline") {
		t.Error("Long description should mention 'offline' condition")
	}
	if !strings.Contains(longDesc, "online") {
		t.Error("Long description should mention 'online' condition")
	}
	if !strings.Contains(longDesc, "power>N") {
		t.Error("Long description should mention 'power>N' condition")
	}
	if !strings.Contains(longDesc, "temperature>N") {
		t.Error("Long description should mention 'temperature>N' condition")
	}
	if !strings.Contains(longDesc, "notify") {
		t.Error("Long description should mention 'notify' action")
	}
	if !strings.Contains(longDesc, "webhook:URL") {
		t.Error("Long description should mention 'webhook:URL' action")
	}
	if !strings.Contains(longDesc, "command:CMD") {
		t.Error("Long description should mention 'command:CMD' action")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example
	if !strings.Contains(example, "shelly alert watch") {
		t.Error("Example should show basic usage")
	}
	if !strings.Contains(example, "--interval") {
		t.Error("Example should show --interval flag usage")
	}
	if !strings.Contains(example, "--once") {
		t.Error("Example should show --once flag usage")
	}
}
