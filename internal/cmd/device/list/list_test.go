package list

import (
	"context"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use != "list" {
		t.Errorf("Use = %q, want list", cmd.Use)
	}
	if len(cmd.Aliases) == 0 || cmd.Aliases[0] != "ls" {
		t.Errorf("Aliases = %v, want [ls]", cmd.Aliases)
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

	genFlag := cmd.Flags().Lookup("generation")
	if genFlag == nil {
		t.Fatal("generation flag not found")
	}
	if genFlag.Shorthand != "g" {
		t.Errorf("generation shorthand = %q, want g", genFlag.Shorthand)
	}

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("type flag not found")
	}
	if typeFlag.Shorthand != "t" {
		t.Errorf("type shorthand = %q, want t", typeFlag.Shorthand)
	}

	platformFlag := cmd.Flags().Lookup("platform")
	if platformFlag == nil {
		t.Fatal("platform flag not found")
	}
	if platformFlag.Shorthand != "p" {
		t.Errorf("platform shorthand = %q, want p", platformFlag.Shorthand)
	}
}

func TestNewCommand_VersionFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("version flag not found")
	}
}

func TestNewCommand_UpdatesFirstFlag(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	updatesFirstFlag := cmd.Flags().Lookup("updates-first")
	if updatesFirstFlag == nil {
		t.Fatal("updates-first flag not found")
	}
}

func TestNewCommand_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// List command should not have Args set (accepts no positional args by default)
	if cmd.Args != nil {
		// If Args is set, verify it accepts empty args
		err := cmd.Args(cmd, []string{})
		if err != nil {
			t.Errorf("Command should accept no args, got error: %v", err)
		}
	}
	// No Args means no positional arguments expected (which is correct for list)
}

func TestNewCommand_HasRunE(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE should be set")
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
			checkFunc: func(c *cobra.Command) bool { return c.Use != "" },
			wantOK:    true,
			errMsg:    "Use should not be empty",
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

func TestRun_EmptyDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should print info message about no devices
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output message for empty device list")
	}
}

func TestRun_WithDevices(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"living-room": {
			Name:       "living-room",
			Address:    "192.168.1.100",
			Model:      "SHSW-25",
			Type:       "relay",
			Generation: 2,
		},
		"kitchen": {
			Name:       "kitchen",
			Address:    "192.168.1.101",
			Model:      "SHPLG-S",
			Type:       "plug",
			Generation: 1,
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := &Options{
		Factory: tf.Factory,
	}

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Output should contain device information
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output with device list")
	}
}

func TestRun_WithGenerationFilter(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"gen1-device": {
			Name:       "gen1-device",
			Address:    "192.168.1.100",
			Model:      "SHSW-1",
			Generation: 1,
		},
		"gen2-device": {
			Name:       "gen2-device",
			Address:    "192.168.1.101",
			Model:      "SNSW-001X16EU",
			Generation: 2,
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := &Options{
		Factory: tf.Factory,
	}
	opts.Generation = 2

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestRun_WithDeviceTypeFilter(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"switch": {
			Name:    "switch",
			Address: "192.168.1.100",
			Model:   "SHSW-25",
			Type:    "relay",
		},
		"plug": {
			Name:    "plug",
			Address: "192.168.1.101",
			Model:   "SHPLG-S",
			Type:    "plug",
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := &Options{
		Factory: tf.Factory,
	}
	opts.DeviceType = "relay"

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestRun_WithPlatformFilter(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"shelly-device": {
			Name:     "shelly-device",
			Address:  "192.168.1.100",
			Platform: "shelly",
		},
		"tasmota-device": {
			Name:     "tasmota-device",
			Address:  "192.168.1.101",
			Platform: "tasmota",
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := &Options{
		Factory: tf.Factory,
	}
	opts.Platform = "shelly"

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}
}

func TestRun_NoMatchingFilters(t *testing.T) {
	t.Parallel()

	devices := map[string]model.Device{
		"gen2-device": {
			Name:       "gen2-device",
			Address:    "192.168.1.100",
			Generation: 2,
		},
	}

	tf := factory.NewTestFactoryWithDevices(t, devices)

	opts := &Options{
		Factory: tf.Factory,
	}
	opts.Generation = 1 // Filter for Gen1, but only Gen2 exists

	err := run(context.Background(), opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	// Should print message about no matching devices
	output := tf.OutString()
	if output == "" {
		t.Error("Expected output message for no matching filters")
	}
}

func TestOptions_DefaultValues(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Factory: tf.Factory,
	}

	// Default filter values should be zero/empty
	if opts.Generation != 0 {
		t.Errorf("Default Generation = %d, want 0", opts.Generation)
	}
	if opts.DeviceType != "" {
		t.Errorf("Default DeviceType = %q, want empty", opts.DeviceType)
	}
	if opts.Platform != "" {
		t.Errorf("Default Platform = %q, want empty", opts.Platform)
	}
	if opts.ShowVersion {
		t.Error("Default ShowVersion should be false")
	}
	if opts.UpdatesFirst {
		t.Error("Default UpdatesFirst should be false")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Parse with no flags to get defaults
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	genFlag := cmd.Flags().Lookup("generation")
	if genFlag.DefValue != "0" {
		t.Errorf("generation default = %q, want 0", genFlag.DefValue)
	}

	typeFlag := cmd.Flags().Lookup("type")
	if typeFlag.DefValue != "" {
		t.Errorf("type default = %q, want empty", typeFlag.DefValue)
	}

	platformFlag := cmd.Flags().Lookup("platform")
	if platformFlag.DefValue != "" {
		t.Errorf("platform default = %q, want empty", platformFlag.DefValue)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		checkFn func(cmd *cobra.Command) error
		wantErr bool
	}{
		{
			name: "generation flag short",
			args: []string{"-g", "2"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "generation flag long",
			args: []string{"--generation", "1"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "type flag short",
			args: []string{"-t", "relay"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "platform flag short",
			args: []string{"-p", "shelly"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "version flag",
			args: []string{"--version"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "updates-first flag",
			args: []string{"--updates-first"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "multiple flags",
			args: []string{"-g", "2", "-t", "relay", "-p", "shelly"},
			checkFn: func(cmd *cobra.Command) error {
				return nil
			},
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
			if err == nil && tt.checkFn != nil {
				if checkErr := tt.checkFn(cmd); checkErr != nil {
					t.Errorf("check failed: %v", checkErr)
				}
			}
		})
	}
}
