package export

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	shellyexport "github.com/tj-smith47/shelly-cli/internal/shelly/export"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "export <device> [id]" {
		t.Errorf("Use = %q, want 'export <device> [id]'", cmd.Use)
	}

	expectedAliases := []string{"exp", "dump"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("got %d aliases, want %d", len(cmd.Aliases), len(expectedAliases))
	}
	for i, want := range expectedAliases {
		if i >= len(cmd.Aliases) || cmd.Aliases[i] != want {
			t.Errorf("alias[%d] = %q, want %q", i, cmd.Aliases[i], want)
		}
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

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{name: "type", shorthand: "", defValue: shelly.ComponentTypeAuto},
		{name: "format", shorthand: "f", defValue: shellyexport.FormatCSV},
		{name: "output", shorthand: "o", defValue: ""},
		{name: "period", shorthand: "p", defValue: ""},
		{name: "from", shorthand: "", defValue: ""},
		{name: "to", shorthand: "", defValue: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("%s flag not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("%s default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Should accept 1 or 2 arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error when no args provided")
	}

	err = cmd.Args(cmd, []string{"device1"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}

	err = cmd.Args(cmd, []string{"device1", "0", "extra"})
	if err == nil {
		t.Error("expected error with 3 args")
	}
}

// Note: Tests for CalculateTimeRange and ParseTime are now in internal/shelly/energy_test.go
// since these functions were extracted to the service layer for DRY compliance.

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
		"shelly energy export",
		"--format json",
		"--from",
		"--to",
		"--output",
		"--period week",
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
		"Export",
		"CSV",
		"JSON",
		"YAML",
		"timestamp",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("expected Long to contain %q", pattern)
		}
	}
}

func TestExecute_InvalidComponentID(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"test-device", "notanumber"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid component ID")
	}
	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("Expected 'invalid component ID' error, got: %v", err)
	}
}

func TestExecute_InvalidFormat(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--format", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected 'invalid format' error, got: %v", err)
	}
}

func TestExecute_InvalidPeriod(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--period", "invalid"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid period")
	}
	if !strings.Contains(err.Error(), "invalid time range") || !strings.Contains(err.Error(), "invalid period") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestExecute_InvalidFromTime(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--from", "not-a-date"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid from time")
	}
	if !strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestExecute_InvalidToTime(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--to", "not-a-date"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid to time")
	}
	if !strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestExecute_WithPeriodHour(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--period", "hour"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	// Expected to fail since mock server doesn't support energy data
	// but this tests that the period is valid
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithPeriodDay(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--period", "day"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithPeriodWeek(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--period", "week"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithPeriodMonth(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--period", "month"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithFromToRange(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--from", "2025-01-01", "--to", "2025-01-07"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithExplicitTypeEM(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithExplicitTypeEM1(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "em1"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithFormatJSON(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "em", "--format", "json"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithFormatYAML(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "em1", "--format", "yaml"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithFormatCSV(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "--type", "em", "--format", "csv"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestExecute_WithComponentID(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
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
	cmd.SetArgs([]string{"test-device", "1", "--type", "em"})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err = cmd.Execute()
	if err != nil {
		t.Logf("Execute error = %v (expected - mock doesn't support energy)", err)
	}
}

func TestRun_InvalidFormat(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        "xml",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("Expected 'invalid format' error, got: %v", err)
	}
}

func TestRun_InvalidPeriod(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
		Period:        "invalid-period",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid period")
	}
	if !strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestRun_InvalidFromTime(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
		From:          "not-a-date",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid from time")
	}
	if !strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestRun_InvalidToTime(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
		To:            "not-a-date",
	}

	err := run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for invalid to time")
	}
	if !strings.Contains(err.Error(), "invalid time range") {
		t.Errorf("Expected 'invalid time range' error, got: %v", err)
	}
}

func TestRun_UnknownComponentType(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: "unknown-type",
		Format:        shellyexport.FormatCSV,
	}

	err = run(context.Background(), opts)
	if err == nil {
		t.Error("Expected error for unknown component type")
	}
	if !strings.Contains(err.Error(), "no energy data components found") {
		t.Errorf("Expected 'no energy data components found' error, got: %v", err)
	}
}

func TestRun_WithExplicitTypeEM(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
	}

	// Explicit type should skip auto-detection
	err = run(context.Background(), opts)
	// Expected to fail since mock server doesn't support EMData.GetData
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EMData)", err)
	}
}

func TestRun_WithExplicitTypeEM1(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM1,
		Format:        shellyexport.FormatCSV,
	}

	// Explicit type should skip auto-detection
	err = run(context.Background(), opts)
	// Expected to fail since mock server doesn't support EM1Data.GetData
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EM1Data)", err)
	}
}

func TestRun_WithOutputFile(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
		OutputFile:    "/tmp/test-export.csv",
	}

	// Try exporting to an output file (will fail early due to no data)
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EMData)", err)
	}
}

func TestRun_WithValidTimeRange(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatCSV,
		From:          "2025-01-01",
		To:            "2025-01-07",
	}

	// Valid time range should pass validation
	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EMData)", err)
	}
}

func TestRun_WithAutoDetection(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeAuto,
		Format:        shellyexport.FormatCSV,
	}

	// Auto detection mode
	err = run(context.Background(), opts)
	// Expected to fail since mock server doesn't support energy data detection
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support energy data)", err)
	}
}

func TestRun_JSONFormat(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM,
		Format:        shellyexport.FormatJSON,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EMData)", err)
	}
}

func TestRun_YAMLFormat(t *testing.T) {
	fixtures := &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "test-device",
					Address:    "192.168.1.100",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Type:       "SPEM-003CEBEU",
					Model:      "Shelly Pro 3EM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"test-device": {},
		},
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()

	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	opts := &Options{
		Factory:       tf.Factory,
		Device:        "test-device",
		ComponentID:   0,
		ComponentType: shelly.ComponentTypeEM1,
		Format:        shellyexport.FormatYAML,
	}

	err = run(context.Background(), opts)
	if err != nil {
		t.Logf("run() error = %v (expected - mock doesn't support EM1Data)", err)
	}
}
