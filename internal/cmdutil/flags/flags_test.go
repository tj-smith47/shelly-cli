package flags_test

import (
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

const (
	defValueFalse = "false"
	defValueTable = "table"
)

func newTestCommand() *cobra.Command {
	return &cobra.Command{
		Use: "test",
	}
}

func TestAddComponentIDFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var id int

	flags.AddComponentIDFlag(cmd, &id, "Light")

	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("id flag not found")
	}
	if flag.Shorthand != "i" {
		t.Errorf("id shorthand = %q, want %q", flag.Shorthand, "i")
	}
	if flag.DefValue != "0" {
		t.Errorf("id default = %q, want %q", flag.DefValue, "0")
	}
	if flag.Usage != "Light component ID (default 0)" {
		t.Errorf("id usage = %q, want %q", flag.Usage, "Light component ID (default 0)")
	}
}

func TestAddSwitchIDFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var id int

	flags.AddSwitchIDFlag(cmd, &id)

	flag := cmd.Flags().Lookup("switch")
	if flag == nil {
		t.Fatal("switch flag not found")
	}
	if flag.Shorthand != "s" {
		t.Errorf("switch shorthand = %q, want %q", flag.Shorthand, "s")
	}
	if flag.DefValue != "0" {
		t.Errorf("switch default = %q, want %q", flag.DefValue, "0")
	}
}

func TestAddTimeoutFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var timeout time.Duration

	flags.AddTimeoutFlag(cmd, &timeout)

	flag := cmd.Flags().Lookup("timeout")
	if flag == nil {
		t.Fatal("timeout flag not found")
	}
	if flag.Shorthand != "t" {
		t.Errorf("timeout shorthand = %q, want %q", flag.Shorthand, "t")
	}
	if flag.DefValue != "10s" {
		t.Errorf("timeout default = %q, want %q", flag.DefValue, "10s")
	}
}

func TestAddConcurrencyFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var concurrent int

	flags.AddConcurrencyFlag(cmd, &concurrent)

	flag := cmd.Flags().Lookup("concurrent")
	if flag == nil {
		t.Fatal("concurrent flag not found")
	}
	if flag.Shorthand != "c" {
		t.Errorf("concurrent shorthand = %q, want %q", flag.Shorthand, "c")
	}
	if flag.DefValue != "5" {
		t.Errorf("concurrent default = %q, want %q", flag.DefValue, "5")
	}
}

func TestAddOutputFormatFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var format string

	flags.AddOutputFormatFlag(cmd, &format)

	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want %q", flag.Shorthand, "o")
	}
	if flag.DefValue != defValueTable {
		t.Errorf("output default = %q, want %q", flag.DefValue, defValueTable)
	}
}

func TestAddYesFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var yes bool

	flags.AddYesFlag(cmd, &yes)

	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("yes flag not found")
	}
	if flag.Shorthand != "y" {
		t.Errorf("yes shorthand = %q, want %q", flag.Shorthand, "y")
	}
	if flag.DefValue != defValueFalse {
		t.Errorf("yes default = %q, want %q", flag.DefValue, defValueFalse)
	}
}

func TestAddConfirmFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var confirm bool

	flags.AddConfirmFlag(cmd, &confirm)

	flag := cmd.Flags().Lookup("confirm")
	if flag == nil {
		t.Fatal("confirm flag not found")
	}
	if flag.DefValue != defValueFalse {
		t.Errorf("confirm default = %q, want %q", flag.DefValue, defValueFalse)
	}
}

func TestAddDryRunFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var dryRun bool

	flags.AddDryRunFlag(cmd, &dryRun)

	flag := cmd.Flags().Lookup("dry-run")
	if flag == nil {
		t.Fatal("dry-run flag not found")
	}
	if flag.DefValue != defValueFalse {
		t.Errorf("dry-run default = %q, want %q", flag.DefValue, defValueFalse)
	}
}

func TestAddGroupFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var group string

	flags.AddGroupFlag(cmd, &group)

	flag := cmd.Flags().Lookup("group")
	if flag == nil {
		t.Fatal("group flag not found")
	}
	if flag.Shorthand != "g" {
		t.Errorf("group shorthand = %q, want %q", flag.Shorthand, "g")
	}
	if flag.DefValue != "" {
		t.Errorf("group default = %q, want %q", flag.DefValue, "")
	}
}

func TestAddAllFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var all bool

	flags.AddAllFlag(cmd, &all)

	flag := cmd.Flags().Lookup("all")
	if flag == nil {
		t.Fatal("all flag not found")
	}
	if flag.Shorthand != "a" {
		t.Errorf("all shorthand = %q, want %q", flag.Shorthand, "a")
	}
	if flag.DefValue != defValueFalse {
		t.Errorf("all default = %q, want %q", flag.DefValue, defValueFalse)
	}
}

func TestAddNameFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var name string

	flags.AddNameFlag(cmd, &name, "Override name")

	flag := cmd.Flags().Lookup("name")
	if flag == nil {
		t.Fatal("name flag not found")
	}
	if flag.Shorthand != "n" {
		t.Errorf("name shorthand = %q, want %q", flag.Shorthand, "n")
	}
	if flag.Usage != "Override name" {
		t.Errorf("name usage = %q, want %q", flag.Usage, "Override name")
	}
}

func TestAddOverwriteFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	var overwrite bool

	flags.AddOverwriteFlag(cmd, &overwrite)

	flag := cmd.Flags().Lookup("overwrite")
	if flag == nil {
		t.Fatal("overwrite flag not found")
	}
	if flag.DefValue != defValueFalse {
		t.Errorf("overwrite default = %q, want %q", flag.DefValue, defValueFalse)
	}
}

func TestAddBatchFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.BatchFlags{}

	flags.AddBatchFlags(cmd, f)

	// Check all batch flags are present
	expectedFlags := []struct {
		name      string
		shorthand string
	}{
		{"group", "g"},
		{"all", "a"},
		{"timeout", "t"},
		{"switch", "s"},
		{"concurrent", "c"},
	}

	for _, ef := range expectedFlags {
		flag := cmd.Flags().Lookup(ef.name)
		if flag == nil {
			t.Errorf("%s flag not found", ef.name)
			continue
		}
		if flag.Shorthand != ef.shorthand {
			t.Errorf("%s shorthand = %q, want %q", ef.name, flag.Shorthand, ef.shorthand)
		}
	}
}

func TestSetBatchDefaults(t *testing.T) {
	t.Parallel()

	f := &flags.BatchFlags{
		Timeout:    5 * time.Second,
		Concurrent: 10,
		SwitchID:   1,
	}

	flags.SetBatchDefaults(f)

	if f.Timeout != flags.DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", f.Timeout, flags.DefaultTimeout)
	}
	if f.Concurrent != flags.DefaultConcurrency {
		t.Errorf("Concurrent = %d, want %d", f.Concurrent, flags.DefaultConcurrency)
	}
	if f.SwitchID != flags.DefaultComponentID {
		t.Errorf("SwitchID = %d, want %d", f.SwitchID, flags.DefaultComponentID)
	}
}

func TestAddSceneFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.SceneFlags{}

	flags.AddSceneFlags(cmd, f)

	// Check all scene flags are present
	expectedFlags := []string{"timeout", "concurrent", "dry-run"}

	for _, name := range expectedFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("%s flag not found", name)
		}
	}
}

func TestSetSceneDefaults(t *testing.T) {
	t.Parallel()

	f := &flags.SceneFlags{
		Timeout:    5 * time.Second,
		Concurrent: 10,
	}

	flags.SetSceneDefaults(f)

	if f.Timeout != flags.DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", f.Timeout, flags.DefaultTimeout)
	}
	if f.Concurrent != flags.DefaultConcurrency {
		t.Errorf("Concurrent = %d, want %d", f.Concurrent, flags.DefaultConcurrency)
	}
}

func TestDefaultConstants(t *testing.T) {
	t.Parallel()

	if flags.DefaultTimeout != 10*time.Second {
		t.Errorf("DefaultTimeout = %v, want 10s", flags.DefaultTimeout)
	}
	if flags.DefaultConcurrency != 5 {
		t.Errorf("DefaultConcurrency = %d, want 5", flags.DefaultConcurrency)
	}
	if flags.DefaultComponentID != 0 {
		t.Errorf("DefaultComponentID = %d, want 0", flags.DefaultComponentID)
	}
}

func TestAddComponentFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.ComponentFlags{}

	flags.AddComponentFlags(cmd, f, "Switch")

	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("id flag not found")
	}
	if flag.Shorthand != "i" {
		t.Errorf("id shorthand = %q, want %q", flag.Shorthand, "i")
	}
}

func TestQuickComponentFlags_ComponentIDPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      int
		wantNil bool
	}{
		{"negative (all)", -1, true},
		{"zero", 0, false},
		{"positive", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := &flags.QuickComponentFlags{ID: tt.id}
			ptr := f.ComponentIDPointer()

			if tt.wantNil && ptr != nil {
				t.Errorf("ComponentIDPointer() = %v, want nil", *ptr)
			}
			if !tt.wantNil && ptr == nil {
				t.Error("ComponentIDPointer() = nil, want non-nil")
			}
			if !tt.wantNil && ptr != nil && *ptr != tt.id {
				t.Errorf("*ComponentIDPointer() = %d, want %d", *ptr, tt.id)
			}
		})
	}
}

func TestAddQuickComponentFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.QuickComponentFlags{}

	flags.AddQuickComponentFlags(cmd, f)

	flag := cmd.Flags().Lookup("id")
	if flag == nil {
		t.Fatal("id flag not found")
	}
	if flag.DefValue != "-1" {
		t.Errorf("id default = %q, want %q", flag.DefValue, "-1")
	}
}

func TestAddConfirmFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.ConfirmFlags{}

	flags.AddConfirmFlags(cmd, f)

	// Check yes flag
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("yes flag not found")
	}

	// Check confirm flag
	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag == nil {
		t.Error("confirm flag not found")
	}
}

func TestAddYesOnlyFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.ConfirmFlags{}

	flags.AddYesOnlyFlag(cmd, f)

	// Yes flag should exist
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("yes flag not found")
	}

	// Confirm flag should NOT exist
	confirmFlag := cmd.Flags().Lookup("confirm")
	if confirmFlag != nil {
		t.Error("confirm flag should not exist with AddYesOnlyFlag")
	}
}

func TestAddOutputFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.OutputFlags{}

	flags.AddOutputFlags(cmd, f)

	// Check output flag exists (via AddOutputFormatFlag)
	flag := cmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag not found")
	}
	if flag.DefValue != defValueTable {
		t.Errorf("output default = %q, want %q", flag.DefValue, defValueTable)
	}
}

func TestAddOutputFlagsCustom(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.OutputFlags{}

	flags.AddOutputFlagsCustom(cmd, f, "json", "json", "yaml", "text")

	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("format flag not found")
	}
	if flag.DefValue != "json" {
		t.Errorf("format default = %q, want %q", flag.DefValue, "json")
	}
	if flag.Shorthand != "f" {
		t.Errorf("format shorthand = %q, want %q", flag.Shorthand, "f")
	}
}

func TestAddOutputFlagsNamed(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.OutputFlags{}

	flags.AddOutputFlagsNamed(cmd, f, "out", "O", "yaml", "json", "yaml")

	flag := cmd.Flags().Lookup("out")
	if flag == nil {
		t.Fatal("out flag not found")
	}
	if flag.DefValue != "yaml" {
		t.Errorf("out default = %q, want %q", flag.DefValue, "yaml")
	}
	if flag.Shorthand != "O" {
		t.Errorf("out shorthand = %q, want %q", flag.Shorthand, "O")
	}
}

func TestSetOutputDefaults(t *testing.T) {
	t.Parallel()

	f := &flags.OutputFlags{Format: "json"}

	flags.SetOutputDefaults(f)

	if f.Format != defValueTable {
		t.Errorf("Format = %q, want %q", f.Format, defValueTable)
	}
}

func TestAddDeviceTargetFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.DeviceTargetFlags{}

	flags.AddDeviceTargetFlags(cmd, f)

	// Check group flag
	groupFlag := cmd.Flags().Lookup("group")
	if groupFlag == nil {
		t.Error("group flag not found")
	}

	// Check all flag
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("all flag not found")
	}
}

func TestAddAllOnlyFlag(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.DeviceTargetFlags{}

	flags.AddAllOnlyFlag(cmd, f)

	// All flag should exist
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Error("all flag not found")
	}

	// Group flag should NOT exist
	groupFlag := cmd.Flags().Lookup("group")
	if groupFlag != nil {
		t.Error("group flag should not exist with AddAllOnlyFlag")
	}
}

func TestAddDeviceFilterFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.DeviceFilterFlags{}

	flags.AddDeviceFilterFlags(cmd, f)

	expectedFlags := []struct {
		name      string
		shorthand string
	}{
		{"generation", "g"},
		{"type", "t"},
		{"platform", "p"},
	}

	for _, ef := range expectedFlags {
		flag := cmd.Flags().Lookup(ef.name)
		if flag == nil {
			t.Errorf("%s flag not found", ef.name)
			continue
		}
		if flag.Shorthand != ef.shorthand {
			t.Errorf("%s shorthand = %q, want %q", ef.name, flag.Shorthand, ef.shorthand)
		}
	}
}

func TestDeviceFilterFlags_HasFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		flags  flags.DeviceFilterFlags
		expect bool
	}{
		{"empty", flags.DeviceFilterFlags{}, false},
		{"generation set", flags.DeviceFilterFlags{Generation: 2}, true},
		{"type set", flags.DeviceFilterFlags{DeviceType: "switch"}, true},
		{"platform set", flags.DeviceFilterFlags{Platform: "shelly"}, true},
		{"all set", flags.DeviceFilterFlags{Generation: 1, DeviceType: "relay", Platform: "tasmota"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.flags.HasFilters()
			if got != tt.expect {
				t.Errorf("HasFilters() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestAddDeviceListFlags(t *testing.T) {
	t.Parallel()

	cmd := newTestCommand()
	f := &flags.DeviceListFlags{}

	flags.AddDeviceListFlags(cmd, f)

	// Check filter flags are added
	genFlag := cmd.Flags().Lookup("generation")
	if genFlag == nil {
		t.Error("generation flag not found")
	}

	// Check list-specific flags
	updatesFlag := cmd.Flags().Lookup("updates-first")
	if updatesFlag == nil {
		t.Fatal("updates-first flag not found")
	}
	if updatesFlag.Shorthand != "u" {
		t.Errorf("updates-first shorthand = %q, want %q", updatesFlag.Shorthand, "u")
	}

	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("version flag not found")
	}
	if versionFlag.Shorthand != "V" {
		t.Errorf("version shorthand = %q, want %q", versionFlag.Shorthand, "V")
	}
}
