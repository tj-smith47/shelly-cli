package create

import (
	"bytes"
	"context"
	"strings"
	"testing"

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

	if cmd.Use != "create <name> <device>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <name> <device>")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"new", "save"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Aliases count = %d, want %d", len(cmd.Aliases), len(expectedAliases))
	}

	for i, expected := range expectedAliases {
		if i < len(cmd.Aliases) && cmd.Aliases[i] != expected {
			t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], expected)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one arg",
			args:      []string{"template"},
			wantError: true,
		},
		{
			name:      "two args valid",
			args:      []string{"my-template", "device"},
			wantError: false,
		},
		{
			name:      "three args",
			args:      []string{"template", "device", "extra"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("Args() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{
			name:         "description flag",
			flagName:     "description",
			shorthand:    "d",
			defaultValue: "",
		},
		{
			name:         "include-wifi flag",
			flagName:     "include-wifi",
			shorthand:    "",
			defaultValue: "false",
		},
		{
			name:         "force flag",
			flagName:     "force",
			shorthand:    "f",
			defaultValue: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
		"shelly template create",
		"--description",
		"--include-wifi",
		"--force",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"template",
		"device",
		"WiFi",
		"configuration",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestNewCommand_ValidArgsFunction(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.ValidArgsFunction == nil {
		t.Error("ValidArgsFunction should be set for tab completion")
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory()
	opts := &Options{
		Name:        "test-template",
		Device:      "test-device",
		Description: "Test description",
		IncludeWiFi: true,
		Factory:     f,
	}
	opts.Yes = true

	if opts.Name != "test-template" {
		t.Errorf("Name = %q, want %q", opts.Name, "test-template")
	}
	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Description != "Test description" {
		t.Errorf("Description = %q, want %q", opts.Description, "Test description")
	}
	if !opts.IncludeWiFi {
		t.Error("IncludeWiFi should be true")
	}
	if !opts.Yes {
		t.Error("Yes should be true")
	}
	if opts.Factory == nil {
		t.Error("Factory should not be nil")
	}
}

func TestExecute_NoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with no args")
	}
}

func TestExecute_OneArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"my-template"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error with one arg")
	}
}

func TestExecute_InvalidTemplateName(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetContext(context.Background())
	// Invalid template name with special characters
	cmd.SetArgs([]string{"invalid/name", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid template name")
	}
	if !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error about invalid name, got: %v", err)
	}
}

func TestExecute_DeviceNotFound(t *testing.T) {
	t.Parallel()

	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{},
		},
		DeviceStates: map[string]mock.DeviceState{},
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
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"my-template", "nonexistent-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("expected error for non-existent device")
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_Success(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-template", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute output: %s", buf.String())
		t.Errorf("Execute error: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_WithDescription(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"desc-template", "test-device", "-d", "My test template description"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute output: %s", buf.String())
		t.Errorf("Execute error: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_WithIncludeWifi(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"wifi-template", "test-device", "--include-wifi"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute output: %s", buf.String())
		t.Errorf("Execute error: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_TemplateAlreadyExists(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create the first template
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"existing-template", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Try to create again without --force
	buf.Reset()
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"existing-template", "test-device"})
	cmd2.SetOut(&buf)
	cmd2.SetErr(&buf)

	err = cmd2.Execute()
	if err == nil {
		t.Error("expected error when template already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_TemplateOverwriteWithForce(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create the first template
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"overwrite-template", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("First create failed: %v", err)
	}

	// Overwrite with --force
	buf.Reset()
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"overwrite-template", "test-device", "--force"})
	cmd2.SetOut(&buf)
	cmd2.SetErr(&buf)

	err = cmd2.Execute()
	if err != nil {
		t.Errorf("Expected success with --force, got: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestExecute_AllFlags(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
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
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"all-flags-template",
		"test-device",
		"-d", "Full test description",
		"--include-wifi",
		"--force",
	})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute output: %s", buf.String())
		t.Errorf("Execute error: %v", err)
	}
}

func TestRun_InvalidTemplateName(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
		Name:    "invalid/template/name",
		Device:  "test-device",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("expected error for invalid template name")
	}
}

//nolint:paralleltest // Uses global mock server state
func TestRun_TemplateExists(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create initial template via Execute (which uses the injected factory)
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"run-test-template", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Try to create again without force
	buf.Reset()
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"run-test-template", "test-device"})
	cmd2.SetOut(&buf)
	cmd2.SetErr(&buf)

	err = cmd2.Execute()
	if err == nil {
		t.Error("expected error when template exists without --force")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

//nolint:paralleltest // Uses global mock server state
func TestRun_ForceOverwrite(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {"switch:0": map[string]any{"output": false}},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	// Create initial template via Execute
	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"force-test-template", "test-device"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Overwrite with force
	buf.Reset()
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{"force-test-template", "test-device", "--force"})
	cmd2.SetOut(&buf)
	cmd2.SetErr(&buf)

	err = cmd2.Execute()
	if err != nil {
		t.Errorf("Expected success with --force, got: %v", err)
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &Options{
		Factory: tf.Factory,
		Name:    "cancel-template",
		Device:  "test-device",
	}

	err := run(ctx, opts)
	// Should fail - either due to context cancellation or device not found
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

func TestOptions_FlagsCanBeSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []string
		flagName string
	}{
		{
			name:     "description flag",
			flags:    []string{"my-template", "device", "-d", "Test description"},
			flagName: "description",
		},
		{
			name:     "include-wifi flag",
			flags:    []string{"my-template", "device", "--include-wifi"},
			flagName: "include-wifi",
		},
		{
			name:     "force flag",
			flags:    []string{"my-template", "device", "--force"},
			flagName: "force",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			if err := cmd.ParseFlags(tt.flags); err != nil {
				t.Fatalf("ParseFlags failed: %v", err)
			}

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
		})
	}
}

func TestNewCommand_FlagsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		flagName string
	}{
		{"description"},
		{"include-wifi"},
		{"force"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.Usage == "" {
				t.Errorf("flag %q should have usage description", tt.flagName)
			}
		})
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
	}
}

func TestNewCommand_ExactArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Args should be cobra.ExactArgs(2)
	if cmd.Args == nil {
		t.Error("Args function should be set")
	}

	testCases := []struct {
		argCount  int
		wantError bool
	}{
		{0, true},
		{1, true},
		{2, false},
		{3, true},
		{4, true},
	}

	for _, tc := range testCases {
		args := make([]string, tc.argCount)
		for i := range args {
			args[i] = "arg"
		}
		err := cmd.Args(cmd, args)
		if (err != nil) != tc.wantError {
			t.Errorf("Args with %d args: error = %v, wantError = %v", tc.argCount, err, tc.wantError)
		}
	}
}
