package create

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// Test constants for repeated strings.
const (
	conditionOffline = "offline"
	conditionOnline  = "online"
	actionNotify     = "notify"
)

// setupTest initializes the test environment with isolated filesystem.
func setupTest(t *testing.T) {
	t.Helper()
	factory.SetupTestFs(t)
}

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

func TestNewCommand_Use(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "create <name>" {
		t.Errorf("Use = %q, want %q", cmd.Use, "create <name>")
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"add", "new"}
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

func TestNewCommand_Short(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expected := "Create a monitoring alert"
	if cmd.Short != expected {
		t.Errorf("Short = %q, want %q", cmd.Short, expected)
	}
}

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	wantPatterns := []string{
		"offline",
		"online",
		"power>N",
		"temperature>N",
		"notify",
		"webhook",
		"command",
	}
	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long should contain %q", pattern)
		}
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
	}

	wantPatterns := []string{
		"shelly alert create",
		"--device",
		"--condition",
		"--action",
		"offline",
		"power>",
		"webhook:",
	}
	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("Example should contain %q", pattern)
		}
	}
}

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"device", "d", ""},
		{"condition", "c", ""},
		{"action", "a", "notify"},
		{"description", "", ""},
	}

	for _, tt := range tests {
		flag := cmd.Flags().Lookup(tt.name)
		if flag == nil {
			t.Errorf("flag %q not found", tt.name)
			continue
		}
		if flag.Shorthand != tt.shorthand {
			t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
		}
		if flag.DefValue != tt.defValue {
			t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
		}
	}
}

func TestNewCommand_RequiredFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test that required flags are actually required
	requiredFlags := []string{"device", "condition"}
	for _, name := range requiredFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("required flag %q not found", name)
			continue
		}
		// Check annotations for required flag
		annotations := flag.Annotations
		if _, ok := annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
			t.Errorf("flag %q should be marked as required", name)
		}
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	// Test requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("expected error with no args")
	}
	if err := cmd.Args(cmd, []string{"test-alert"}); err != nil {
		t.Errorf("unexpected error with 1 arg: %v", err)
	}
	if err := cmd.Args(cmd, []string{"alert1", "alert2"}); err == nil {
		t.Error("expected error with 2 args")
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Device:      "test-device",
		Condition:   conditionOffline,
		Action:      actionNotify,
		Description: "Test alert description",
	}

	if opts.Device != "test-device" {
		t.Errorf("Device = %q, want %q", opts.Device, "test-device")
	}
	if opts.Condition != conditionOffline {
		t.Errorf("Condition = %q, want %q", opts.Condition, conditionOffline)
	}
	if opts.Action != actionNotify {
		t.Errorf("Action = %q, want %q", opts.Action, actionNotify)
	}
	if opts.Description != "Test alert description" {
		t.Errorf("Description = %q, want %q", opts.Description, "Test alert description")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_Success(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"test-alert",
		"--device", "kitchen",
		"--condition", conditionOffline,
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify alert was created in config
	alert, exists := tf.Config.Alerts["test-alert"]
	if !exists {
		t.Fatal("Alert was not created in config")
	}
	if alert.Name != "test-alert" {
		t.Errorf("Alert name = %q, want %q", alert.Name, "test-alert")
	}
	if alert.Device != "kitchen" {
		t.Errorf("Alert device = %q, want %q", alert.Device, "kitchen")
	}
	if alert.Condition != conditionOffline {
		t.Errorf("Alert condition = %q, want %q", alert.Condition, conditionOffline)
	}
	if alert.Action != actionNotify {
		t.Errorf("Alert action = %q, want %q", alert.Action, actionNotify)
	}
	if !alert.Enabled {
		t.Error("Alert should be enabled by default")
	}
	if alert.CreatedAt == "" {
		t.Error("Alert CreatedAt should be set")
	}

	// Verify success output
	output := tf.OutString()
	if !strings.Contains(output, "Created alert") {
		t.Errorf("Output should contain 'Created alert', got: %q", output)
	}
	if !strings.Contains(output, "test-alert") {
		t.Errorf("Output should contain alert name, got: %q", output)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_WithAllOptions(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"power-alert",
		"--device", "heater",
		"--condition", "power>2000",
		"--action", "webhook:http://example.com/alert",
		"--description", "Alert for high power consumption",
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify alert was created with all options
	alert, exists := tf.Config.Alerts["power-alert"]
	if !exists {
		t.Fatal("Alert was not created in config")
	}
	if alert.Device != "heater" {
		t.Errorf("Alert device = %q, want %q", alert.Device, "heater")
	}
	if alert.Condition != "power>2000" {
		t.Errorf("Alert condition = %q, want %q", alert.Condition, "power>2000")
	}
	if alert.Action != "webhook:http://example.com/alert" {
		t.Errorf("Alert action = %q, want %q", alert.Action, "webhook:http://example.com/alert")
	}
	if alert.Description != "Alert for high power consumption" {
		t.Errorf("Alert description = %q, want %q", alert.Description, "Alert for high power consumption")
	}

	// Verify output includes all details
	output := tf.OutString()
	if !strings.Contains(output, "heater") {
		t.Errorf("Output should contain device name, got: %q", output)
	}
	if !strings.Contains(output, "power>2000") {
		t.Errorf("Output should contain condition, got: %q", output)
	}
	if !strings.Contains(output, "webhook:") {
		t.Errorf("Output should contain action, got: %q", output)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_AlertAlreadyExists(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Pre-populate an existing alert
	tf.Config.Alerts["existing-alert"] = config.Alert{
		Name:      "existing-alert",
		Device:    "kitchen",
		Condition: conditionOffline,
		Action:    actionNotify,
		Enabled:   true,
	}

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"existing-alert",
		"--device", "bedroom",
		"--condition", conditionOnline,
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when alert already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention 'already exists', got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_MissingDeviceFlag(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"test-alert",
		"--condition", conditionOffline,
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when device flag is missing")
	}
	if !strings.Contains(err.Error(), "device") {
		t.Errorf("Error should mention 'device', got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_MissingConditionFlag(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"test-alert",
		"--device", "kitchen",
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when condition flag is missing")
	}
	if !strings.Contains(err.Error(), "condition") {
		t.Errorf("Error should mention 'condition', got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_MissingName(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"--device", "kitchen",
		"--condition", conditionOffline,
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when alert name is missing")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_TemperatureCondition(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"temp-alert",
		"--device", "sensor",
		"--condition", "temperature>30",
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	alert := tf.Config.Alerts["temp-alert"]
	if alert.Condition != "temperature>30" {
		t.Errorf("Alert condition = %q, want %q", alert.Condition, "temperature>30")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_OnlineCondition(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"online-alert",
		"--device", "garage",
		"--condition", conditionOnline,
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	alert := tf.Config.Alerts["online-alert"]
	if alert.Condition != conditionOnline {
		t.Errorf("Alert condition = %q, want %q", alert.Condition, conditionOnline)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_CommandAction(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{
		"cmd-alert",
		"--device", "switch",
		"--condition", conditionOffline,
		"--action", "command:echo device offline",
	})
	cmd.SetOut(tf.TestIO.Out)
	cmd.SetErr(tf.TestIO.ErrOut)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	alert := tf.Config.Alerts["cmd-alert"]
	if alert.Action != "command:echo device offline" {
		t.Errorf("Alert action = %q, want %q", alert.Action, "command:echo device offline")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_Success(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:     tf.Factory,
		Name:        "kitchen-offline",
		Device:      "kitchen",
		Condition:   conditionOffline,
		Action:      actionNotify,
		Description: "Kitchen offline alert",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	// Verify alert in config
	alert, exists := tf.Config.Alerts["kitchen-offline"]
	if !exists {
		t.Fatal("Alert was not created")
	}
	if alert.Name != "kitchen-offline" {
		t.Errorf("Alert name = %q, want %q", alert.Name, "kitchen-offline")
	}
	if alert.Device != "kitchen" {
		t.Errorf("Alert device = %q, want %q", alert.Device, "kitchen")
	}
	if alert.Condition != conditionOffline {
		t.Errorf("Alert condition = %q, want %q", alert.Condition, conditionOffline)
	}
	if alert.Action != actionNotify {
		t.Errorf("Alert action = %q, want %q", alert.Action, actionNotify)
	}
	if alert.Description != "Kitchen offline alert" {
		t.Errorf("Alert description = %q, want %q", alert.Description, "Kitchen offline alert")
	}
	if !alert.Enabled {
		t.Error("Alert should be enabled")
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_AlertExists(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Pre-create alert
	tf.Config.Alerts["dup-alert"] = config.Alert{
		Name:      "dup-alert",
		Device:    "existing",
		Condition: conditionOffline,
		Action:    actionNotify,
		Enabled:   true,
	}

	opts := &Options{
		Factory:   tf.Factory,
		Name:      "dup-alert",
		Device:    "new-device",
		Condition: conditionOnline,
		Action:    actionNotify,
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for duplicate alert name")
	}
	if !strings.Contains(err.Error(), "dup-alert") {
		t.Errorf("Error should contain alert name, got: %v", err)
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Error should mention 'already exists', got: %v", err)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_OutputFormat(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:   tf.Factory,
		Name:      "output-test",
		Device:    "light",
		Condition: conditionOffline,
		Action:    "webhook:http://test.com",
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	output := tf.OutString()

	// Check all output lines
	expectedLines := []string{
		"Created alert",
		"output-test",
		"Device: light",
		"Condition: " + conditionOffline,
		"Action: webhook:http://test.com",
	}
	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Output should contain %q, got: %q", line, output)
		}
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_EmptyDescription(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory:     tf.Factory,
		Name:        "no-desc-alert",
		Device:      "switch",
		Condition:   "power>1000",
		Action:      actionNotify,
		Description: "", // Empty description is allowed
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v", err)
	}

	alert := tf.Config.Alerts["no-desc-alert"]
	if alert.Description != "" {
		t.Errorf("Alert description = %q, want empty", alert.Description)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestRun_DefaultAction(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Verify default action is "notify"
	cmd := NewCommand(tf.Factory)
	actionFlag := cmd.Flags().Lookup("action")
	if actionFlag.DefValue != actionNotify {
		t.Errorf("Default action = %q, want %q", actionFlag.DefValue, actionNotify)
	}
}

//nolint:paralleltest // Uses global config.SetFs which cannot be parallelized
func TestExecute_MultipleAlerts(t *testing.T) {
	setupTest(t)
	tf := factory.NewTestFactory(t)

	// Create first alert
	cmd1 := NewCommand(tf.Factory)
	cmd1.SetContext(context.Background())
	cmd1.SetArgs([]string{
		"alert-1",
		"--device", "device1",
		"--condition", conditionOffline,
	})
	cmd1.SetOut(tf.TestIO.Out)
	cmd1.SetErr(tf.TestIO.ErrOut)

	if err := cmd1.Execute(); err != nil {
		t.Errorf("First alert creation failed: %v", err)
	}

	// Reset output buffers
	tf.Reset()

	// Create second alert
	cmd2 := NewCommand(tf.Factory)
	cmd2.SetContext(context.Background())
	cmd2.SetArgs([]string{
		"alert-2",
		"--device", "device2",
		"--condition", conditionOnline,
	})
	cmd2.SetOut(tf.TestIO.Out)
	cmd2.SetErr(tf.TestIO.ErrOut)

	if err := cmd2.Execute(); err != nil {
		t.Errorf("Second alert creation failed: %v", err)
	}

	// Verify both alerts exist
	if _, exists := tf.Config.Alerts["alert-1"]; !exists {
		t.Error("First alert not found")
	}
	if _, exists := tf.Config.Alerts["alert-2"]; !exists {
		t.Error("Second alert not found")
	}
	if len(tf.Config.Alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(tf.Config.Alerts))
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		checkFunc func() bool
		errMsg    string
	}{
		{
			name: "has use",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Use != ""
			},
			errMsg: "Use should not be empty",
		},
		{
			name: "has short",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Short != ""
			},
			errMsg: "Short should not be empty",
		},
		{
			name: "has long",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Long != ""
			},
			errMsg: "Long should not be empty",
		},
		{
			name: "has example",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.Example != ""
			},
			errMsg: "Example should not be empty",
		},
		{
			name: "has aliases",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return len(cmd.Aliases) > 0
			},
			errMsg: "Aliases should not be empty",
		},
		{
			name: "has RunE",
			checkFunc: func() bool {
				cmd := NewCommand(cmdutil.NewFactory())
				return cmd.RunE != nil
			},
			errMsg: "RunE should be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !tt.checkFunc() {
				t.Error(tt.errMsg)
			}
		})
	}
}
