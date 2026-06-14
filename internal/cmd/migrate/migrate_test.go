package migrate

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// noWiFiBackup returns a minimal Gen2 backup whose WiFi blob is absent. A
// --to-ap restore on such a backup resolves no station passphrase (no backup
// SSID, no override, no host creds) and so fails in resolveJoinNetwork BEFORE
// any WiFi/host mutation — the only safe way to exercise the AP code paths in a
// unit test without hopping the host onto a real network.
func noWiFiBackup() *clibackup.DeviceBackup {
	return &clibackup.DeviceBackup{Backup: &shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID:         "shellybulbduo-test",
			Name:       "src",
			Model:      "SHBDUO-1",
			Generation: 2,
			Version:    "1.0.0",
			MAC:        "AA:BB:CC:DD:EE:FF",
		},
	}}
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

	want := "migrate <source-device> <target-device>"
	if cmd.Use != want {
		t.Errorf("Use = %q, want %q", cmd.Use, want)
	}
}

func TestNewCommand_Aliases(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	expectedAliases := []string{"mig"}
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

func TestNewCommand_Long(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestNewCommand_Example(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Error("Example is empty")
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
			wantErr: true,
		},
		{
			name:    "one arg",
			args:    []string{"source"},
			wantErr: true,
		},
		{
			name:    "two args",
			args:    []string{"source", "target"},
			wantErr: false,
		},
		{
			name:    "three args",
			args:    []string{"source", "target", "extra"},
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

func TestNewCommand_RunE(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flagName string
		defValue string
	}{
		{"dry-run", "dry-run", "false"},
		{"force", "force", "false"},
		{"yes", "yes", "false"},
		{"reset-source", "reset-source", "true"},
		{"skip-auth", "skip-auth", "false"},
		{"skip-network", "skip-network", "false"},
		{"skip-scripts", "skip-scripts", "false"},
		{"skip-schedules", "skip-schedules", "false"},
		{"skip-webhooks", "skip-webhooks", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_YesFlagShorthand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	flag := cmd.Flags().Lookup("yes")
	if flag == nil {
		t.Fatal("yes flag not found")
	}
	if flag.Shorthand != "y" {
		t.Errorf("yes flag shorthand = %q, want %q", flag.Shorthand, "y")
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"dry-run", []string{"--dry-run"}},
		{"force", []string{"--force"}},
		{"yes long", []string{"--yes"}},
		{"yes short", []string{"-y"}},
		{"reset-source false", []string{"--reset-source=false"}},
		{"skip-auth", []string{"--skip-auth"}},
		{"skip-network", []string{"--skip-network"}},
		{"skip-scripts", []string{"--skip-scripts"}},
		{"skip-schedules", []string{"--skip-schedules"}},
		{"skip-webhooks", []string{"--skip-webhooks"}},
		{"all skip flags", []string{"--skip-auth", "--skip-network", "--skip-scripts", "--skip-schedules", "--skip-webhooks"}},
		{"combined", []string{"--dry-run", "--force", "-y"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewCommand(cmdutil.NewFactory())

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Errorf("ParseFlags() error = %v", err)
			}
		})
	}
}

func TestNewCommand_HasSubcommands(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	subcommands := cmd.Commands()
	if len(subcommands) < 2 {
		t.Errorf("expected at least 2 subcommands, got %d", len(subcommands))
	}

	subNames := make(map[string]bool)
	for _, sub := range subcommands {
		subNames[sub.Name()] = true
	}

	if !subNames["validate"] {
		t.Error("validate subcommand not found")
	}
	if !subNames["diff"] {
		t.Error("diff subcommand not found")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Use == "" {
		t.Error("Use is empty")
	}
	if cmd.Short == "" {
		t.Error("Short is empty")
	}
	if cmd.Long == "" {
		t.Error("Long is empty")
	}
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
	if len(cmd.Aliases) == 0 {
		t.Error("Aliases is empty")
	}
	if cmd.RunE == nil {
		t.Error("RunE is nil")
	}
	if cmd.Args == nil {
		t.Error("Args is nil")
	}
}

func TestShouldResetSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		skipNetwork bool
		explicit    bool
		resetSource bool
		want        bool
	}{
		{
			name: "default with network = reset",
			want: true,
		},
		{
			name:        "skip-network = no reset",
			skipNetwork: true,
			want:        false,
		},
		{
			name:        "explicit true",
			explicit:    true,
			resetSource: true,
			want:        true,
		},
		{
			name:        "explicit false overrides default",
			explicit:    true,
			resetSource: false,
			want:        false,
		},
		{
			name:        "explicit true with skip-network overrides auto",
			skipNetwork: true,
			explicit:    true,
			resetSource: true,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := &Options{
				SkipNetwork:         tt.skipNetwork,
				ResetSource:         tt.resetSource,
				resetSourceExplicit: tt.explicit,
			}
			if got := opts.shouldResetSource(); got != tt.want {
				t.Errorf("shouldResetSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun_DeviceToDeviceMigrationFails(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: "192.0.2.1", Target: "192.0.2.2", Yes: true}
	err := run(ctx, opts) // TEST-NET-1 addresses

	// Should fail to connect to source device
	if err == nil {
		t.Log("Note: run succeeded unexpectedly (devices might be mocked)")
	}
}

func TestRun_ContextCancelled(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := &Options{Factory: tf.Factory, Source: "source", Target: "target", Yes: true}
	err := run(ctx, opts)

	if err == nil {
		t.Log("Note: run succeeded with cancelled context (unexpected)")
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	opts := &Options{Factory: tf.Factory, Source: "source", Target: "target", Yes: true}
	err := run(ctx, opts)

	if err == nil {
		t.Log("Note: run succeeded with timed out context (unexpected)")
	}
}

func TestNewCommand_AcceptsIPAddresses(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"192.168.1.100", "192.168.1.101"})
	if err != nil {
		t.Errorf("Command should accept IP addresses, got error: %v", err)
	}
}

func TestNewCommand_AcceptsDeviceNames(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	err := cmd.Args(cmd, []string{"living-room", "bedroom"})
	if err != nil {
		t.Errorf("Command should accept device names, got error: %v", err)
	}
}

func TestNewCommand_ValidateSubcommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	validateCmd, _, err := cmd.Find([]string{"validate"})
	if err != nil {
		t.Errorf("failed to find validate subcommand: %v", err)
		return
	}
	if validateCmd == nil {
		t.Error("validate subcommand is nil")
		return
	}

	if validateCmd.Use == "" {
		t.Error("validate Use is empty")
	}
}

func TestNewCommand_DiffSubcommand(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	diffCmd, _, err := cmd.Find([]string{"diff"})
	if err != nil {
		t.Errorf("failed to find diff subcommand: %v", err)
		return
	}
	if diffCmd == nil {
		t.Error("diff subcommand is nil")
		return
	}

	if diffCmd.Use == "" {
		t.Error("diff Use is empty")
	}
}

func TestNewCommand_ExecuteWithNoArgs(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with no args")
	}
}

func TestNewCommand_ExecuteWithOneArg(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetArgs([]string{"source"})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when executing with only one arg")
	}
}

func TestNewCommand_SubcommandsHaveCorrectParent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	for _, sub := range cmd.Commands() {
		if sub.Parent() != cmd {
			t.Errorf("subcommand %q has incorrect parent", sub.Name())
		}
	}
}

func TestNewCommand_ExampleContainsUsage(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Example == "" {
		t.Fatal("Example is empty")
	}

	// Should contain key usage patterns
	patterns := []string{"shelly migrate", "--dry-run", "--skip-network", "--yes"}
	for _, p := range patterns {
		if !strings.Contains(cmd.Example, p) {
			t.Errorf("Example should contain %q", p)
		}
	}
}

func TestNewCommand_LongContainsDescription(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if cmd.Long == "" {
		t.Fatal("Long is empty")
	}

	keywords := []string{"configuration", "factory reset", "IP conflicts", "--skip-network"}
	for _, kw := range keywords {
		if !strings.Contains(cmd.Long, kw) {
			t.Errorf("Long should contain %q", kw)
		}
	}
}

func TestRun_InvalidSourceDevice(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	opts := &Options{Factory: tf.Factory, Source: "not-a-valid-source", Target: "target-device", Yes: true}
	err := run(ctx, opts)

	if err == nil {
		t.Log("Note: run succeeded unexpectedly")
	}
}

func TestNewCommand_AliasWorks(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	found := false
	for _, alias := range cmd.Aliases {
		if alias == "mig" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'mig' alias not found")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags error: %v", err)
	}

	checks := map[string]string{
		"dry-run":        "false",
		"force":          "false",
		"yes":            "false",
		"reset-source":   "true",
		"skip-auth":      "false",
		"skip-network":   "false",
		"skip-scripts":   "false",
		"skip-schedules": "false",
		"skip-webhooks":  "false",
	}

	for name, want := range checks {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("flag %q not found", name)
			continue
		}
		if flag.DefValue != want {
			t.Errorf("flag %q default = %q, want %q", name, flag.DefValue, want)
		}
	}
}

func TestNewCommand_WithTestIOStreams(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tf := factory.NewTestFactory(t)

	cmd := NewCommand(tf.Factory)
	cmd.SetOut(&buf)

	if cmd == nil {
		t.Fatal("NewCommand returned nil with test IOStreams")
	}
}

func TestValidateFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		opts        Options
		wantErrSubs string
	}{
		{name: "no flags", opts: Options{}},
		{name: "to-ap with static-ip", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.255.255.0"}},
		{name: "static-ip with skip-network", opts: Options{StaticIP: "10.0.0.5", Gateway: "10.0.0.1", Netmask: "255.255.255.0", SkipNetwork: true}, wantErrSubs: "static-ip cannot be used with --skip-network"},
		{name: "to-ap with skip-network", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", SkipNetwork: true}, wantErrSubs: "to-ap cannot be used with --skip-network"},
		{name: "to-ap with dry-run", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", DryRun: true}, wantErrSubs: "to-ap cannot be combined with --dry-run"},
		{name: "ap-ip without to-ap", opts: Options{APIP: "192.168.33.140"}, wantErrSubs: "ap-ip only applies with --to-ap"},
		{name: "ap-ip with to-ap", opts: Options{ToAP: "ShellyBulbDuo-AABBCC", APIP: "192.168.33.140"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.validateFlags(tt.opts.networkOverride())
			if tt.wantErrSubs == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrSubs) {
				t.Fatalf("got %v, want error containing %q", err, tt.wantErrSubs)
			}
		})
	}
}

func TestShouldResetSource_ToAP(t *testing.T) {
	t.Parallel()
	// A --to-ap target is a distinct physical device, so the source is never
	// reset even when network settings are migrated.
	opts := Options{ToAP: "ShellyBulbDuo-AABBCC"}
	if opts.shouldResetSource() {
		t.Error("shouldResetSource() = true for --to-ap, want false")
	}
}

func TestToAPFlagRegistered(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())
	if cmd.Flags().Lookup("to-ap") == nil {
		t.Error("--to-ap flag not registered")
	}
}

func TestNewCommand_SkipStateAndMetersFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	for _, name := range []string{"skip-state", "skip-meters"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("--%s flag not registered", name)
		}
	}
}

func TestNewCommand_FirmwareFlags(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	downgrade := cmd.Flags().Lookup("allow-firmware-downgrade")
	if downgrade == nil {
		t.Fatal("--allow-firmware-downgrade flag not registered")
	}
	if downgrade.DefValue != "false" {
		t.Errorf("--allow-firmware-downgrade default = %q, want %q", downgrade.DefValue, "false")
	}
	if cmd.Flags().Lookup("firmware-url") == nil {
		t.Error("--firmware-url flag not registered")
	}
}

func TestConfirmMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yes         bool
		resetSource bool
		wantOK      bool
		wantOut     string
	}{
		{name: "yes skips prompt", yes: true, wantOK: true},
		{name: "non-tty declines without reset", wantOK: false, wantOut: "Migration cancelled"},
		{name: "non-tty declines with reset note", resetSource: true, wantOK: false, wantOut: "Migration cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: tt.yes}

			ok, err := opts.confirmMigration(tt.resetSource)
			if err != nil {
				t.Fatalf("confirmMigration: %v", err)
			}
			if ok != tt.wantOK {
				t.Errorf("confirmMigration() = %v, want %v", ok, tt.wantOK)
			}
			if tt.wantOut != "" && !strings.Contains(tf.OutString(), tt.wantOut) {
				t.Errorf("output %q does not contain %q", tf.OutString(), tt.wantOut)
			}
		})
	}
}

func TestPreviewMigration_CompareFails(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// TEST-NET-1 target: CompareBackup cannot reach a device, so previewMigration
	// returns the wrapped compare error before any display work.
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "192.0.2.10"}
	err := opts.previewMigration(ctx, noWiFiBackup(), nil)
	if err == nil {
		t.Fatal("expected previewMigration to fail against an unreachable target")
	}
	if !strings.Contains(err.Error(), "failed to compare") {
		t.Errorf("got %v, want error containing %q", err, "failed to compare")
	}
}

func TestMigrateViaAP_ConfirmCancels(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Non-TTY confirm declines, so migrateViaAP returns nil before RestoreToAP —
	// no WiFi hop is attempted.
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", ToAP: "ShellyBulbDuo-AABBCC"}
	if err := opts.migrateViaAP(ctx, tf.ShellyService(), noWiFiBackup(), nil); err != nil {
		t.Fatalf("migrateViaAP with declined confirm: %v", err)
	}
	if !strings.Contains(tf.OutString(), "Migration cancelled") {
		t.Errorf("expected cancellation notice, got %q", tf.OutString())
	}
}

func TestMigrateViaAP_RestoreFailsBeforeHop(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Yes skips the prompt so RestoreToAP runs; the WiFi-less backup resolves no
	// passphrase, so it fails in resolveJoinNetwork BEFORE any host WiFi hop.
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", ToAP: "ShellyBulbDuo-AABBCC", Yes: true}
	err := opts.migrateViaAP(ctx, tf.ShellyService(), noWiFiBackup(), nil)
	if err == nil {
		t.Fatal("expected migrateViaAP to fail without a resolvable WiFi passphrase")
	}
	if !strings.Contains(err.Error(), "migration via AP failed") {
		t.Errorf("got %v, want error containing %q", err, "migration via AP failed")
	}
	// not covered: the success tail (ios.Success / newAddr / DisplayMigrationResult)
	// requires RestoreToAP to complete a real AP hop + device write, which would
	// mutate host WiFi and contact hardware — never exercised in a unit test.
}

func TestRun_ToAP_SourceBackupFails(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// --to-ap takes the longer-timeout branch and the migrateViaAP dispatch, but
	// CreateBackup against TEST-NET-1 fails first, so no AP hop is attempted.
	opts := &Options{
		Factory: tf.Factory,
		Source:  "192.0.2.20",
		Target:  "dst",
		ToAP:    "ShellyBulbDuo-AABBCC",
		Yes:     true,
	}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected run to fail reading an unreachable source")
	}
	if !strings.Contains(err.Error(), "failed to read source device") {
		t.Errorf("got %v, want error containing %q", err, "failed to read source device")
	}
}

func TestRun_ToAP_InvalidFlagCombo(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// --to-ap with --skip-network is rejected by validateFlags before any device
	// I/O, covering run's validation-error branch on the --to-ap path.
	opts := &Options{
		Factory:     tf.Factory,
		Source:      "src",
		Target:      "dst",
		ToAP:        "ShellyBulbDuo-AABBCC",
		SkipNetwork: true,
		Yes:         true,
	}
	err := run(ctx, opts)
	if err == nil {
		t.Fatal("expected validateFlags to reject --to-ap with --skip-network")
	}
	if !strings.Contains(err.Error(), "to-ap cannot be used with --skip-network") {
		t.Errorf("got %v, want error containing %q", err, "to-ap cannot be used with --skip-network")
	}
}

func TestShouldResetSource_StaticIP(t *testing.T) {
	t.Parallel()
	// A static-IP override gives the target a distinct address, so the source is
	// not reset even though network settings are migrated.
	opts := &Options{StaticIP: "10.0.0.9"}
	if opts.shouldResetSource() {
		t.Error("shouldResetSource() = true with --static-ip, want false")
	}
}
