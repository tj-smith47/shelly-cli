package disable

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func(*cobra.Command) bool
		wantOK    bool
		errMsg    string
	}{
		{
			name:      "has use",
			checkFunc: func(c *cobra.Command) bool { return c.Use == "disable <device>" },
			wantOK:    true,
			errMsg:    "Use should be 'disable <device>'",
		},
		{
			name:      "has short",
			checkFunc: func(c *cobra.Command) bool { return c.Short != "" },
			wantOK:    true,
			errMsg:    "Short should not be empty",
		},
		{
			name:      "has long",
			checkFunc: func(c *cobra.Command) bool { return c.Long != "" },
			wantOK:    true,
			errMsg:    "Long should not be empty",
		},
		{
			name:      "has example",
			checkFunc: func(c *cobra.Command) bool { return c.Example != "" },
			wantOK:    true,
			errMsg:    "Example should not be empty",
		},
		{
			name:      "has aliases",
			checkFunc: func(c *cobra.Command) bool { return len(c.Aliases) > 0 },
			wantOK:    true,
			errMsg:    "Aliases should not be empty",
		},
		{
			name:      "has RunE",
			checkFunc: func(c *cobra.Command) bool { return c.RunE != nil },
			wantOK:    true,
			errMsg:    "RunE should be set",
		},
		{
			name:      "has ValidArgsFunction",
			checkFunc: func(c *cobra.Command) bool { return c.ValidArgsFunction != nil },
			wantOK:    true,
			errMsg:    "ValidArgsFunction should be set for completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			if tt.checkFunc(cmd) != tt.wantOK {
				t.Error(tt.errMsg)
			}
		})
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"off"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
		return
	}
	for i, alias := range expectedAliases {
		if cmd.Aliases[i] != alias {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no args", []string{}, true},
		{"one arg valid", []string{"device"}, false},
		{"two args", []string{"device", "extra"}, true},
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Fatal("--id flag not found")
	}
	if idFlag.DefValue != "0" {
		t.Errorf("id default = %q, want 0", idFlag.DefValue)
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly thermostat schedule disable",
		"--id",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with no arguments")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "id flag",
			args:    []string{"--id", "1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			err := cmd.ParseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithMockGen1Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "gen1-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SHSW-1",
					Model:      "Shelly 1",
					Generation: 1,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"gen1-device": {"relay": map[string]any{"ison": true}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"gen1-device", "--id", "1"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for Gen1 device")
	}

	if !strings.Contains(err.Error(), "Gen2+") {
		t.Errorf("Expected error mentioning Gen2+, got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestExecute_WithMockGen2Device(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
				},
				"schedule:1": map[string]any{
					"id":       float64(1),
					"enable":   true,
					"timespec": "0 0 8 * *",
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"thermostat-device", "--id", "1"})

	err = cmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error for Gen2 device schedule disable: %v", err)
	}

	output := tf.OutString()
	if !strings.Contains(output, "disabled") {
		t.Logf("Output: %s", output)
	}
}

//nolint:paralleltest // Uses global config.SetDefaultManager via demo.InjectIntoFactory
func TestNewCommand_ExecuteWithIDFlag(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "thermostat-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:01",
					Type:       "SNSN-0043X",
					Model:      "Shelly Wall Display",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"thermostat-device": {
				"thermostat:0": map[string]any{
					"id":        float64(0),
					"enable":    true,
					"target_C":  float64(22.0),
					"current_C": float64(21.5),
					"output":    true,
				},
				"schedule:1": map[string]any{
					"id":       float64(1),
					"enable":   true,
					"timespec": "0 0 8 * *",
				},
			},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"thermostat-device", "--id", "1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v", err)
	}
}
