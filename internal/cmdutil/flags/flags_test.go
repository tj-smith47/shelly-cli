package flags_test

import (
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

const (
	defValueFalse = "false"
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
	if flag.DefValue != "table" {
		t.Errorf("output default = %q, want %q", flag.DefValue, "table")
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
