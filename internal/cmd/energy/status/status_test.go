package status

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "status [device] [id]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status [device] [id]")
	}

	// Test Aliases
	wantAliases := []string{"st"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Args(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		// Zero args pass the arity check because --all takes none; RunE
		// rejects zero args without --all (TestAll_ArgsAndFlagConflicts).
		{"no args", []string{}, false},
		{"one arg valid", []string{"device"}, false},
		{"two args valid", []string{"device", "0"}, false},
		{"three args", []string{"device", "0", "extra"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

	// Test type flag
	flag := cmd.Flags().Lookup("type")
	if flag == nil {
		t.Fatal("--type flag not found")
	}
	if flag.DefValue != "auto" {
		t.Errorf("--type default = %q, want %q", flag.DefValue, "auto")
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
		"shelly energy status",
		"--type",
		"-o json",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestNewCommand_InvalidComponentID(t *testing.T) {
	t.Parallel()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f := cmdutil.NewFactory().SetIOStreams(ios)

	cmd := NewCommand(f)
	cmd.SetArgs([]string{"device", "not-a-number"})
	cmd.SetOut(out)
	cmd.SetErr(errOut)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid component ID")
	}

	if !strings.Contains(err.Error(), "invalid component ID") {
		t.Errorf("expected 'invalid component ID' error, got: %v", err)
	}
}

func runOne(t *testing.T, state mock.DeviceState, gen int, args ...string) (*factory.TestFactory, error) {
	t.Helper()
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{Name: "dev", Address: "192.0.2.10", MAC: "AA:BB:CC:DD:EE:FF", Type: "X", Model: "X", Generation: gen},
		}},
		DeviceStates: map[string]mock.DeviceState{"dev": state},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs(append([]string{"dev"}, args...))
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	return tf, cmd.Execute()
}

//nolint:paralleltest // uses global mock config manager
func TestRun_EveryCarrierTable(t *testing.T) {
	tests := []struct {
		name  string
		state mock.DeviceState
		gen   int
		args  []string
		want  []string
	}{
		{"em", mock.DeviceState{"em:0": map[string]any{
			"id": 0, "a_voltage": 230.0, "a_act_power": 345.0, "total_current": 4.5, "total_act_power": 1035.0,
		}}, 2, []string{"0", "--type", "em"}, []string{"1035.0"}},
		{"em1", mock.DeviceState{"em1:0": map[string]any{
			"id": 0, "voltage": 230.0, "current": 2.5, "act_power": 575.0,
		}}, 2, []string{"--type", "em1"}, []string{"575.0"}},
		{"switch", mock.DeviceState{
			"switch:0": map[string]any{"id": 0, "apower": 1.0},
			"switch:1": map[string]any{"id": 1, "apower": 48.5, "voltage": 121.0, "aenergy": map[string]any{"total": 300.0}},
		}, 2, []string{"1"}, []string{"Switch #1", "48.50 W", "121.00 V", "300.00 Wh"}},
		{"gen1 meter", mock.DeviceState{"meters": []any{map[string]any{"power": 60.5, "total": 600}}}, 1, nil,
			[]string{"Power Meter (Meter)", "60.50 W", "10.00 Wh"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf, err := runOne(t, tt.state, tt.gen, tt.args...)
			if err != nil {
				t.Fatalf("Execute: %v\nstderr: %s", err, tf.ErrString())
			}
			for _, w := range tt.want {
				if !strings.Contains(tf.OutString(), w) {
					t.Errorf("output missing %q:\n%s", w, tf.OutString())
				}
			}
		})
	}
}

//nolint:paralleltest // uses global mock config manager
func TestRun_NothingMetersPower(t *testing.T) {
	_, err := runOne(t, mock.DeviceState{"switch:0": map[string]any{"id": 0, "output": true}}, 2)
	if err == nil || !strings.Contains(err.Error(), "dev: no component on this device reports power") {
		t.Errorf("err = %v", err)
	}
}

func TestNewCommand_LongDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())
	for _, pattern := range []string{"EM (3-phase)", "switch, cover", "Gen1", "--all", "stderr"} {
		if !strings.Contains(cmd.Long, pattern) {
			t.Errorf("Long description should contain %q", pattern)
		}
	}
}
