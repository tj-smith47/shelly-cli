package migrate

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/ratelimit"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/testutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// stubMigrateService is an in-memory migrateService for driving run(),
// migrateViaAP, and previewMigration without reaching a device or hopping WiFi.
// Each hook, when nil, returns a benign default so a test sets only the behavior
// it asserts on.
type stubMigrateService struct {
	createBackup func(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error)
	checkCompat  func(context.Context, *clibackup.DeviceBackup, string, bool) error
	compare      func(context.Context, string, *clibackup.DeviceBackup) (*model.BackupDiff, error)
	restore      func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error)
	plan         func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error)
	apPlanErr    error
	restoreToAP  func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error)
	factoryReset func(context.Context, string) error
	targetMAC    string

	factoryResetCalls int
}

func (s *stubMigrateService) CreateBackup(ctx context.Context, id string, opts clibackup.Options) (*clibackup.DeviceBackup, error) {
	if s.createBackup != nil {
		return s.createBackup(ctx, id, opts)
	}
	return noWiFiBackup(), nil
}

func (s *stubMigrateService) CheckMigrationCompatibility(ctx context.Context, bkp *clibackup.DeviceBackup, target string, force bool) error {
	if s.checkCompat != nil {
		return s.checkCompat(ctx, bkp, target, force)
	}
	return nil
}

func (s *stubMigrateService) CompareBackup(ctx context.Context, id string, bkp *clibackup.DeviceBackup) (*model.BackupDiff, error) {
	if s.compare != nil {
		return s.compare(ctx, id, bkp)
	}
	return &model.BackupDiff{}, nil
}

func (s *stubMigrateService) RestoreBackup(ctx context.Context, id string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
	if s.restore != nil {
		return s.restore(ctx, id, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, nil
}

func (s *stubMigrateService) RestoreToAP(ctx context.Context, ssid, apIP, name string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
	if s.restoreToAP != nil {
		return s.restoreToAP(ctx, ssid, apIP, name, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, "10.0.0.50", nil
}

func (s *stubMigrateService) DeviceFactoryReset(ctx context.Context, id string) error {
	s.factoryResetCalls++
	if s.factoryReset != nil {
		return s.factoryReset(ctx, id)
	}
	return nil
}

func migrateTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

var errStubMigrate = errors.New("stub migrate failure")

// TestRun_FullMigration_ResetsSource drives the on-LAN happy path: with network
// migrated and no static-IP/skip-network override, the source is factory reset.
func TestRun_FullMigration_ResetsSource(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stub.factoryResetCalls != 1 {
		t.Errorf("expected source to be factory reset once, got %d", stub.factoryResetCalls)
	}
	if out := tf.OutString(); !strings.Contains(out, "Migration completed") {
		t.Errorf("missing success message, got %q", out)
	}
}

// TestRun_FailedRestore_NoReset proves the destructive guard: when the target
// rejects the restore (Success=false with a nil top-level error, the shape
// shelly-go returns on a partial failure), run must abort BEFORE the source
// factory-reset and return an error — never wipe the source after a bad restore.
func TestRun_FailedRestore_NoReset(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return &clibackup.RestoreResult{Success: false, Errors: []string{"mqtt rejected"}}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	err := run(migrateTestCtx(t), opts)
	if err == nil {
		t.Fatal("a rejected restore must make run return an error so the exit code is non-zero")
	}
	if stub.factoryResetCalls != 0 {
		t.Errorf("source must NOT be factory reset after a failed restore, got %d calls", stub.factoryResetCalls)
	}
	if out := tf.OutString() + tf.ErrString(); strings.Contains(out, "Migration completed") {
		t.Errorf("must not report completion on a failed restore, got %q", out)
	}
}

// TestRun_SkipNetwork_NoReset confirms --skip-network leaves the source online.
func TestRun_SkipNetwork_NoReset(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, SkipNetwork: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stub.factoryResetCalls != 0 {
		t.Errorf("source must not be reset with --skip-network, got %d calls", stub.factoryResetCalls)
	}
}

// TestRun_ResetFails_Warns proves a failed source reset is a warning, not fatal:
// the migration already succeeded, so run returns nil.
func TestRun_ResetFails_Warns(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{factoryReset: func(context.Context, string) error { return errStubMigrate }}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("reset failure must not be fatal, got %v", err)
	}
	if errOut := tf.ErrString(); !strings.Contains(errOut, "factory reset of source failed") {
		t.Errorf("missing reset-failure warning, got %q", errOut)
	}
}

// TestRun_NetworkWithoutReset_Warns covers the IP-conflict warning when network is
// migrated but the source is explicitly not reset.
func TestRun_NetworkWithoutReset_Warns(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{}
	opts := &Options{
		Factory: tf.Factory, Source: "src", Target: "dst", Yes: true,
		ResetSource: false, resetSourceExplicit: true, svc: stub,
	}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if stub.factoryResetCalls != 0 {
		t.Errorf("source must not be reset, got %d calls", stub.factoryResetCalls)
	}
	if errOut := tf.ErrString(); !strings.Contains(errOut, "without factory-resetting") {
		t.Errorf("missing IP-conflict warning, got %q", errOut)
	}
}

// TestRun_DryRun_Default exercises previewMigration with no override: the reset
// notice fires and the empty diff renders "No differences found".
func TestRun_DryRun_Default(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	out := tf.OutString()
	if !strings.Contains(out, "Migration Preview") {
		t.Errorf("missing preview header, got %q", out)
	}
	if errOut := tf.ErrString(); !strings.Contains(errOut, "will be factory reset") {
		t.Errorf("missing reset notice, got %q", errOut)
	}
}

// TestRun_DryRun_StaticIP exercises previewMigration's override branch: a static-IP
// target keeps the source online, so the static-IP notice (not the reset notice) is
// shown.
func TestRun_DryRun_StaticIP(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{}
	opts := &Options{
		Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true,
		StaticIP: "10.0.0.9", Gateway: "10.0.0.1", Netmask: "255.255.254.0", svc: stub,
	}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if out := tf.OutString(); !strings.Contains(out, "static IP 10.0.0.9 (gateway 10.0.0.1, netmask 255.255.254.0)") {
		t.Errorf("missing static-IP notice, got %q", out)
	}
}

// staticSourceBackup is a Gen2 source whose station has a static address.
func staticSourceBackup() *clibackup.DeviceBackup {
	bkp := noWiFiBackup()
	bkp.WiFi = []byte(`{"sta":{"ssid":"home","enable":true,"ipv4mode":"static","ip":"10.0.0.8",` +
		`"gw":"10.0.0.1","netmask":"255.255.255.0","nameserver":"10.0.0.53"}}`)
	return bkp
}

// TestRun_DryRun_StaticIPTakesSourceAddressing checks that the preview shows the
// gateway and netmask the migration writes, taken from the source device when
// only --static-ip is given, and that an address the source cannot complete is
// refused before the target is checked.
func TestRun_DryRun_StaticIPTakesSourceAddressing(t *testing.T) {
	t.Parallel()

	t.Run("source has static settings", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		stub := &stubMigrateService{createBackup: func(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error) {
			return staticSourceBackup(), nil
		}}
		opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true, StaticIP: "10.0.0.9", svc: stub}

		if err := run(migrateTestCtx(t), opts); err != nil {
			t.Fatalf("run: %v", err)
		}
		want := `Target "dst" will get static IP 10.0.0.9 (gateway 10.0.0.1, netmask 255.255.255.0); source "src" keeps its own address`
		if out := tf.OutString(); !strings.Contains(out, want) {
			t.Errorf("preview lacks %q, got %q", want, out)
		}
	})

	t.Run("source uses DHCP", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		stub := &stubMigrateService{checkCompat: func(context.Context, *clibackup.DeviceBackup, string, bool) error {
			t.Error("the target was checked before the address was refused")
			return nil
		}}
		opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true, StaticIP: "10.0.0.9", svc: stub}

		err := run(migrateTestCtx(t), opts)
		want := "static address 10.0.0.9 needs a gateway and a netmask, and the backup has none — pass --gateway and --netmask"
		if err == nil || err.Error() != want {
			t.Errorf("err = %v, want %q", err, want)
		}
	})
}

// TestRun_ToAP_Success drives the --to-ap dispatch through migrateViaAP to a
// successful at-AP restore that returns the device's new LAN address.
func TestRun_ToAP_Success(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return &clibackup.RestoreResult{Success: true}, "10.23.47.227", nil
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "fr", Yes: true, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	out := tf.OutString()
	if !strings.Contains(out, "Migration completed") || !strings.Contains(out, "10.23.47.227") {
		t.Errorf("missing AP success / new address, got %q", out)
	}
}

// TestRun_CreateBackupFails covers the source-read failure path.
func TestRun_CreateBackupFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		createBackup: func(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error) {
			return nil, errStubMigrate
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	err := run(migrateTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "failed to read source device") {
		t.Fatalf("expected source-read failure, got %v", err)
	}
}

// TestRun_CompatibilityFails covers the on-LAN compatibility-check failure path.
func TestRun_CompatibilityFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		checkCompat: func(context.Context, *clibackup.DeviceBackup, string, bool) error { return errStubMigrate },
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err == nil {
		t.Fatal("expected compatibility failure to propagate")
	}
}

// TestRun_RestoreBackupFails covers the on-LAN restore failure path.
func TestRun_RestoreBackupFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return nil, errStubMigrate
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}

	err := run(migrateTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "migration failed") {
		t.Fatalf("expected migration failure, got %v", err)
	}
}

// TestMigrateViaAP_RestoreToAPFails covers the at-AP restore failure path directly.
func TestMigrateViaAP_RestoreToAPFails(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return nil, "", errStubMigrate
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	err := opts.migrateViaAP(migrateTestCtx(t), stub, noWiFiBackup(), shellybackup.StaticNetwork{}, nil)
	if err == nil || !strings.Contains(err.Error(), "migration via AP failed") {
		t.Fatalf("expected AP migration failure, got %v", err)
	}
}

// TestMigrateViaAP_PartialRestore proves the --to-ap path does not report a
// false success: a Success=false result with a nil top-level error (a section
// rejection at the AP) must surface an error and never print "Migration
// completed!". The device's new LAN address is still reported for recovery.
func TestMigrateViaAP_PartialRestore(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	stub := &stubMigrateService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return &clibackup.RestoreResult{Success: false, Errors: []string{"cloud section rejected"}}, "10.23.47.227", nil
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "fr", Yes: true, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	err := opts.migrateViaAP(migrateTestCtx(t), stub, noWiFiBackup(), shellybackup.StaticNetwork{}, nil)
	if err == nil {
		t.Fatal("a rejected at-AP restore must return a non-nil error")
	}
	out := tf.OutString() + tf.ErrString()
	if strings.Contains(out, "Migration completed") {
		t.Errorf("must not claim completion on partial failure, got %q", out)
	}
	if !strings.Contains(out, "10.23.47.227") {
		t.Errorf("the device's new LAN address should still be surfaced, got %q", out)
	}
	if !strings.Contains(out, "cloud section rejected") {
		t.Errorf("the rejected section must be surfaced, got %q", out)
	}
}

// TestRun_ToAP_StaticIPWithoutGateway checks that a --to-ap migration accepts
// --static-ip alone and hands the empty gateway and netmask through, so the
// device takes them from the source's static settings.
func TestRun_ToAP_StaticIPWithoutGateway(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var got *clibackup.NetworkOverride
	stub := &stubMigrateService{
		createBackup: func(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error) {
			return staticSourceBackup(), nil
		},
		restoreToAP: func(_ context.Context, _, _, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			got = o.NetworkOverride
			return &clibackup.RestoreResult{Success: true}, "10.23.47.227", nil
		},
	}
	opts := &Options{
		Factory: tf.Factory, Source: "src", Target: "fr", Yes: true,
		ToAP: "ShellyBulbDuo-AABBCC", StaticIP: "10.23.47.227", svc: stub,
	}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || got.StaticIP != "10.23.47.227" || got.Gateway != "" || got.Netmask != "" {
		t.Errorf("override = %+v, want the static IP with no gateway or netmask", got)
	}
}

// TestRun_StaticIPWithoutGateway checks that a LAN migration accepts --static-ip
// alone and hands the empty gateway and netmask to the restore, which takes them
// from the source's static settings.
func TestRun_StaticIPWithoutGateway(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var got *clibackup.NetworkOverride
	stub := &stubMigrateService{
		createBackup: func(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error) {
			return staticSourceBackup(), nil
		},
		restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			got = o.NetworkOverride
			return &clibackup.RestoreResult{Success: true}, nil
		},
	}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, StaticIP: "10.0.0.9", svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || got.StaticIP != "10.0.0.9" || got.Gateway != "" || got.Netmask != "" {
		t.Errorf("override = %+v, want the static IP with no gateway or netmask", got)
	}
}

// TestNewCommand_StaticIPAlone runs the cobra command with --static-ip and no
// gateway or netmask against an in-process device whose station uses DHCP: cobra
// accepts the flag on its own, and the run reaches the static address
// resolution, which refuses the address for want of a gateway and netmask.
//
//nolint:paralleltest // demo.InjectIntoFactory sets the global config manager
func TestNewCommand_StaticIPAlone(t *testing.T) {
	demo, err := mock.StartWithFixtures(&mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{Devices: []mock.DeviceFixture{
			{Name: "src", Address: "192.168.1.100", MAC: "AA:BB:CC:DD:EE:01", Type: "SNSW-001X16EU", Model: "Shelly Plus 1", Generation: 2},
			{Name: "dst", Address: "192.168.1.101", MAC: "AA:BB:CC:DD:EE:02", Type: "SNSW-001X16EU", Model: "Shelly Plus 1", Generation: 2},
		}},
	})
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)
	tf := factory.NewTestFactory(t)
	demo.InjectIntoFactory(tf.Factory)

	cmd := NewCommand(tf.Factory)
	cmd.SetContext(migrateTestCtx(t))
	cmd.SetArgs([]string{"src", "dst", "--static-ip", "10.0.0.9", "--dry-run"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	want := "static address 10.0.0.9 needs a gateway and a netmask, and the backup has none — pass --gateway and --netmask"
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
	if !errors.Is(err, shellybackup.ErrIncompleteStaticNetwork) {
		t.Errorf("errors.Is(ErrIncompleteStaticNetwork) = false for %v", err)
	}
}

// TestRun_OpenReachesRestore checks that --open reaches the restore as an open
// network override.
func TestRun_OpenReachesRestore(t *testing.T) {
	t.Parallel()
	tf := factory.NewTestFactory(t)
	var got *clibackup.NetworkOverride
	stub := &stubMigrateService{restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
		got = o.NetworkOverride
		return &clibackup.RestoreResult{Success: true}, nil
	}}
	opts := &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, SSID: "Guest", Open: true, svc: stub}

	if err := run(migrateTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if got == nil || *got != (clibackup.NetworkOverride{SSID: "Guest", Open: true}) {
		t.Errorf("override = %+v, want an open Guest network", got)
	}
}

func (s *stubMigrateService) DescribeRestoreName(_ context.Context, _ string, bkp *clibackup.DeviceBackup, o clibackup.RestoreOptions) (string, error) {
	return clibackup.DescribeName(o.Name, o.AliasName, bkp.RecordedName(), bkp.Device().MAC, s.targetMAC), nil
}

func (s *stubMigrateService) PlanLANStation(_ context.Context, _ string, _ *clibackup.DeviceBackup, ov *clibackup.NetworkOverride, _ bool) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
	if s.plan != nil {
		return s.plan(ov)
	}
	return ov, shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyKept}, nil
}

func (s *stubMigrateService) PlanAPStation(context.Context, *clibackup.DeviceBackup, *clibackup.NetworkOverride) (shelly.LANStationPlan, error) {
	return shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyGiven}, s.apPlanErr
}

// TestRun_StationPlan covers the LAN station decision: a refusal stops the run
// before any write, the planned override reaches the restore, and a dry run
// prints the decision.
func TestRun_StationPlan(t *testing.T) {
	t.Parallel()
	refuse := func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
		return nil, shelly.LANStationPlan{}, shelly.PassphraseError("home")
	}
	t.Run("refused before any write", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		restored := false
		stub := &stubMigrateService{plan: refuse, restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			restored = true
			return &clibackup.RestoreResult{Success: true}, nil
		}}
		err := run(migrateTestCtx(t), &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub})
		if err == nil || !strings.Contains(err.Error(), `no WiFi passphrase for "home"`) {
			t.Fatalf("err = %v, want the passphrase error", err)
		}
		if restored || stub.factoryResetCalls != 0 {
			t.Error("a device was written after the refusal")
		}
	})
	t.Run("dry run refuses too", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		err := run(migrateTestCtx(t), &Options{Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true, svc: &stubMigrateService{plan: refuse}})
		if err == nil {
			t.Fatal("dry run accepted a network it could not join")
		}
	})
	t.Run("host passphrase reaches the restore", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		var got *clibackup.NetworkOverride
		stub := &stubMigrateService{
			plan: func(*clibackup.NetworkOverride) (*clibackup.NetworkOverride, shelly.LANStationPlan, error) {
				return &clibackup.NetworkOverride{Password: "stored"}, shelly.LANStationPlan{SSID: "home", Key: shelly.StationKeyHost}, nil
			},
			restore: func(_ context.Context, _ string, _ *clibackup.DeviceBackup, o clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
				got = o.NetworkOverride
				return &clibackup.RestoreResult{Success: true}, nil
			},
		}
		if err := run(migrateTestCtx(t), &Options{Factory: tf.Factory, Source: "src", Target: "dst", Yes: true, svc: stub}); err != nil {
			t.Fatalf("run: %v", err)
		}
		if got == nil || got.Password != "stored" {
			t.Errorf("override = %+v, want the host passphrase", got)
		}
	})
	t.Run("dry run prints the decision", func(t *testing.T) {
		t.Parallel()
		tf := factory.NewTestFactory(t)
		if err := run(migrateTestCtx(t), &Options{Factory: tf.Factory, Source: "src", Target: "dst", DryRun: true, svc: &stubMigrateService{}}); err != nil {
			t.Fatalf("run: %v", err)
		}
		if out := tf.OutString() + tf.ErrString(); !strings.Contains(out, `WiFi station "home": same network as the device, which keeps its current password`) {
			t.Errorf("plan not printed: %q", out)
		}
	})
}

// TestMigrate_DryRunName drives `migrate --dry-run` against loopback Gen2
// devices and asserts the name line: the source's backup records no name, so
// the target keeps its own name when it is the source device, takes its alias
// when it is another device, and takes --name when given. No Set method is
// called on either device.
//
//nolint:paralleltest // SetupTestFs swaps the process-global config
func TestMigrate_DryRunName(t *testing.T) {
	const srcMAC = "AABBCCDDEEFF"
	tests := []struct {
		name      string
		targetMAC string
		args      []string
		want      string
	}{
		{name: "target is the source device", targetMAC: srcMAC,
			want: "name: left unchanged (the backup records none)"},
		{name: "target is another device", targetMAC: "112233445566",
			want: `name: "dst" from the alias (different device)`},
		{name: "--name", targetMAC: "112233445566", args: []string{"--name", "Kitchen"},
			want: `name: "Kitchen" from --name`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory.SetupTestFs(t)
			src := testutil.NewGen2NameDevice(t, srcMAC)
			dst := testutil.NewGen2NameDevice(t, tt.targetMAC)
			tf := factory.NewTestFactory(t)
			tf.SetShellyService(shelly.New(&testutil.Resolver{Devices: map[string]model.Device{
				"src": {Address: src.Addr, Generation: 2},
				"dst": {Address: dst.Addr, Generation: 2},
			}},
				shelly.WithRateLimiter(ratelimit.New(ratelimit.WithGen2MinInterval(0)))))

			cmd := NewCommand(tf.Factory)
			cmd.SetArgs(append([]string{"src", "dst", "--dry-run", "--skip-network"}, tt.args...))
			cmd.SetContext(migrateTestCtx(t))
			if err := cmd.Execute(); err != nil {
				t.Fatalf("migrate --dry-run: %v", err)
			}
			if out := tf.OutString() + tf.ErrString(); !strings.Contains(out, tt.want) {
				t.Errorf("output %q, want the line %q", out, tt.want)
			}
			if sets := append(src.SetCalls(), dst.SetCalls()...); len(sets) != 0 {
				t.Errorf("dry run called %q", sets)
			}
		})
	}
}

// apDryRunService is the real service, its WiFi scanner and target device
// fakes, with the source backup read from a fixture.
type apDryRunService struct {
	*shelly.Service
	bkp *clibackup.DeviceBackup
}

func (s apDryRunService) CreateBackup(context.Context, string, clibackup.Options) (*clibackup.DeviceBackup, error) {
	return s.bkp, nil
}

// TestRun_ToAPDryRun drives `migrate --to-ap --dry-run` through the real
// service: it prints the plan a hop would carry out, reading this host's WiFi
// state but never scanning, joining or leaving a network or writing to any
// device.
func TestRun_ToAPDryRun(t *testing.T) {
	t.Parallel()
	gen1 := &clibackup.DeviceBackup{Backup: &shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID: "shellybulbduo-ddeeff", Name: "sr", Model: "SHBDUO-1", Generation: 1,
			Version: "20230913-111821/v1.14.0-gcb84623", MAC: "AA:BB:CC:DD:EE:FF",
		},
	}}
	const hopLine = "nothing was written"
	tests := []struct {
		name    string
		bkp     *clibackup.DeviceBackup
		opts    Options
		want    []string
		notWant []string
	}{
		{
			name: "gen1 onto another device takes the alias",
			bkp:  gen1,
			opts: Options{ToAP: "ShellyBulbDuo-0A0B0C", SSID: "home", Password: "pw",
				StaticIP: "10.0.0.9", Gateway: "10.0.0.1", Netmask: "255.255.255.0"},
			want: []string{
				"Migration Preview (dry run)",
				`name: "fr" from the alias (different device)`,
				`WiFi station "home": joined with the password from --password; ` +
					"static address 10.0.0.9 (gateway 10.0.0.1, netmask 255.255.255.0) from --static-ip",
				`Target "fr" will get static IP 10.0.0.9`,
				"Would update Gen1 firmware to match the backup (20230913-111821/v1.14.0-gcb84623) " +
					"if the device runs older firmware",
				`Would hop onto AP "ShellyBulbDuo-0A0B0C" from this host; ` + hopLine,
			},
			notWant: []string{"factory reset"},
		},
		{
			name: "gen2 onto its own device, host password",
			bkp:  noWiFiBackup(),
			opts: Options{ToAP: "ShellyPlus1-AABBCCDDEEFF", SSID: "home", Name: "Kitchen"},
			want: []string{
				`name: "Kitchen" from --name`,
				`WiFi station "home": joined with the password stored on this host`,
				hopLine,
			},
			notWant: []string{"firmware", "factory reset"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tf := factory.NewTestFactory(t)
			dev := testutil.NewGen2NameDevice(t, "112233445566")
			scanner := &testutil.CountingWiFiScanner{CurrentSSID: "home", Passwords: map[string]string{"home": "hpw"}}
			svc := apDryRunService{
				Service: shelly.New(testutil.Gen2At(dev.Addr), shelly.WithWiFiScanner(scanner)),
				bkp:     tt.bkp,
			}
			opts := tt.opts
			opts.Factory, opts.Source, opts.Target, opts.DryRun, opts.svc = tf.Factory, "sr", "fr", true, svc

			if err := run(migrateTestCtx(t), &opts); err != nil {
				t.Fatalf("run: %v", err)
			}
			if n := scanner.Moves(); n != 0 {
				t.Errorf("dry run made %d WiFi scans, joins or leaves, want 0", n)
			}
			if sets := dev.SetCalls(); len(sets) != 0 {
				t.Errorf("dry run wrote to the device: %q", sets)
			}
			combined := tf.OutString() + tf.ErrString()
			for _, w := range tt.want {
				if !strings.Contains(combined, w) {
					t.Errorf("output missing %q:\n%s", w, combined)
				}
			}
			for _, w := range tt.notWant {
				if strings.Contains(combined, w) {
					t.Errorf("output has %q:\n%s", w, combined)
				}
			}
		})
	}
}

// TestRun_ToAPRefusals covers --to-ap migrations refused before the prompt and
// any hop, alike and with the same error by the dry run and the real run:
// network flags with no network named, and a network this host has no stored
// passphrase for (named, or the host's own when none is named); and, on a dry
// run, a static address the source cannot complete.
func TestRun_ToAPRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		opts    Options
		dryOnly bool
		wantErr string
	}{
		{name: "--password with no network named", opts: Options{Password: "pw"}, wantErr: "need a network"},
		{name: "--open with no network named", opts: Options{Open: true}, wantErr: "need a network"},
		{name: "incomplete static address", opts: Options{SSID: "home", StaticIP: "10.0.0.9"}, dryOnly: true,
			wantErr: "needs a gateway and a netmask"},
		{name: "no stored password for the named network", opts: Options{SSID: "home"},
			wantErr: `no WiFi passphrase for "home"`},
		{name: "no stored password for the host's network", opts: Options{},
			wantErr: `no WiFi passphrase for "home"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var errs []string
			for _, dryRun := range []bool{true, false} {
				if !dryRun && tt.dryOnly {
					continue
				}
				tf := factory.NewTestFactory(t)
				dev := testutil.NewGen2NameDevice(t, "112233445566")
				scanner := &testutil.CountingWiFiScanner{CurrentSSID: "home"}
				svc := apDryRunService{
					Service: shelly.New(testutil.Gen2At(dev.Addr), shelly.WithWiFiScanner(scanner)),
					bkp:     noWiFiBackup(),
				}
				opts := tt.opts
				opts.Factory, opts.Source, opts.Target, opts.svc = tf.Factory, "sr", "fr", svc
				opts.ToAP, opts.DryRun = "ShellyBulbDuo-0A0B0C", dryRun

				err := run(migrateTestCtx(t), &opts)
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("dry run %v: err = %v, want %q", dryRun, err, tt.wantErr)
				}
				errs = append(errs, err.Error())
				if n := scanner.Moves(); n != 0 {
					t.Errorf("dry run %v: %d WiFi scans, joins or leaves, want 0", dryRun, n)
				}
				if sets := dev.SetCalls(); len(sets) != 0 {
					t.Errorf("dry run %v: wrote to the device: %q", dryRun, sets)
				}
				if strings.Contains(tf.OutString()+tf.ErrString(), "Migrate \"sr\"") {
					t.Errorf("dry run %v: prompted before refusing", dryRun)
				}
			}
			if len(errs) == 2 && errs[0] != errs[1] {
				t.Errorf("dry run refused with %q, real run with %q", errs[0], errs[1])
			}
		})
	}
}
