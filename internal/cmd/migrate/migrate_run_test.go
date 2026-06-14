package migrate

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
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
	restoreToAP  func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error)
	factoryReset func(context.Context, string) error

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
	if out := tf.OutString(); !strings.Contains(out, "static IP 10.0.0.9") {
		t.Errorf("missing static-IP notice, got %q", out)
	}
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

	err := opts.migrateViaAP(migrateTestCtx(t), stub, noWiFiBackup(), nil)
	if err == nil || !strings.Contains(err.Error(), "migration via AP failed") {
		t.Fatalf("expected AP migration failure, got %v", err)
	}
}
