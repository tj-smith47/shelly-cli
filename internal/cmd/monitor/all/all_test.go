package all

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
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
			wantUse:     "all",
			wantShort:   "Monitor all registered devices",
			wantAliases: []string{"overview", "summary"},
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

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"overview", "summary"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
		flagType     string
	}{
		{
			name:         "interval flag",
			flagName:     "interval",
			shorthand:    "i",
			defaultValue: (5 * time.Second).String(),
			flagType:     "duration",
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

			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}

			if flag.Value.Type() != tt.flagType {
				t.Errorf("flag %q type = %q, want %q", tt.flagName, flag.Value.Type(), tt.flagType)
			}
		})
	}
}

func TestNewCommand_FlagUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	intervalFlag := cmd.Flags().Lookup("interval")
	if intervalFlag == nil {
		t.Fatal("interval flag not found")
	}

	if intervalFlag.Usage == "" {
		t.Error("interval flag usage is empty")
	}

	// Usage should mention refresh
	if !strings.Contains(strings.ToLower(intervalFlag.Usage), "refresh") {
		t.Errorf("interval flag usage should mention 'refresh', got: %q", intervalFlag.Usage)
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
			name:    "no args",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "one arg",
			args:    []string{"device1"},
			wantErr: true,
		},
		{
			name:    "multiple args",
			args:    []string{"device1", "device2"},
			wantErr: true,
		},
		{
			name:    "flag-like arg",
			args:    []string{"--some-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Args(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	longDesc := cmd.Long

	// Should mention monitoring all devices
	if !strings.Contains(strings.ToLower(longDesc), "all devices") {
		t.Error("Long description should mention 'all devices'")
	}

	// Should mention power consumption or status
	if !strings.Contains(strings.ToLower(longDesc), "power") && !strings.Contains(strings.ToLower(longDesc), "status") {
		t.Error("Long description should mention power consumption or status")
	}

	// Should mention Ctrl+C
	if !strings.Contains(longDesc, "Ctrl+C") {
		t.Error("Long description should mention 'Ctrl+C'")
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	example := cmd.Example

	// Should show basic usage
	if !strings.Contains(example, "shelly monitor all") {
		t.Error("Example should show basic 'shelly monitor all' usage")
	}

	// Should show interval flag usage
	if !strings.Contains(example, "--interval") || !strings.Contains(example, "-i") {
		t.Error("Example should demonstrate --interval flag usage")
	}
}

func TestOptions_Defaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 5*time.Second {
		t.Errorf("interval default = %v, want 5s", interval)
	}
}

func TestNewCommand_IntervalFlagGetDuration(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Simulate setting a custom interval
	if err := cmd.Flags().Set("interval", "10s"); err != nil {
		t.Fatalf("failed to set interval flag: %v", err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval flag: %v", err)
	}
	if interval != 10*time.Second {
		t.Errorf("interval = %v, want 10s", interval)
	}
}

func TestNewCommand_IntervalFlagInvalidValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Invalid duration should fail to parse
	err := cmd.Flags().Set("interval", "not-a-duration")
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestNewCommand_IntervalFlagValidValue(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Valid duration should parse successfully
	err := cmd.Flags().Set("interval", "30s")
	if err != nil {
		t.Errorf("unexpected error for valid duration: %v", err)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "1s"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetContext(ctx)

	// Run with the cancelled context - should return immediately
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The command should have handled cancellation gracefully
	// Even if no devices registered message appears or context is cancelled
	// Both are valid outcomes
}

func TestRun_QuickContextCancellation(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)

	f := cmdutil.NewFactory().SetIOStreams(ios)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithCancel(context.Background())

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"--interval", "100ms"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	// Cancel context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	cmd.SetContext(ctx)
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCommand_MultipleFactoryCalls(t *testing.T) {
	t.Parallel()

	// Verify that calling NewCommand multiple times with the same factory works
	f := cmdutil.NewFactory()

	cmd1 := NewCommand(f)
	cmd2 := NewCommand(f)

	if cmd1 == nil || cmd2 == nil {
		t.Fatal("NewCommand returned nil")
	}

	// Both commands should have same properties
	if cmd1.Use != cmd2.Use {
		t.Error("Commands should have same Use")
	}

	if cmd1.Short != cmd2.Short {
		t.Error("Commands should have same Short")
	}
}

func TestNewCommand_FlagCanBeSet(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Verify flag can be set to a custom value
	if err := cmd.Flags().Set("interval", "15s"); err != nil {
		t.Fatalf("failed to set interval: %v", err)
	}

	interval, err := cmd.Flags().GetDuration("interval")
	if err != nil {
		t.Fatalf("failed to get interval: %v", err)
	}
	if interval != 15*time.Second {
		t.Errorf("interval = %v, want 15s", interval)
	}
}
