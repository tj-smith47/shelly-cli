package restore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	shellybackup "github.com/tj-smith47/shelly-go/backup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	clibackup "github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// stubRestoreService is an in-memory restoreService for driving run() and
// restoreViaAP without reaching a device or hopping WiFi.
type stubRestoreService struct {
	restore     func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error)
	restoreToAP func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error)
}

func (s *stubRestoreService) RestoreBackup(ctx context.Context, id string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
	if s.restore != nil {
		return s.restore(ctx, id, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, nil
}

func (s *stubRestoreService) RestoreToAP(ctx context.Context, ssid, apIP, name string, bkp *clibackup.DeviceBackup, opts clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
	if s.restoreToAP != nil {
		return s.restoreToAP(ctx, ssid, apIP, name, bkp, opts)
	}
	return &clibackup.RestoreResult{Success: true}, "10.0.0.50", nil
}

var errStubRestore = errors.New("stub restore failure")

const testBackupPath = "/b.json"

// validBackupData marshals a minimal valid Gen2 backup the restore path accepts.
func validBackupData(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(shellybackup.Backup{
		Version: 1,
		DeviceInfo: &shellybackup.DeviceInfo{
			ID: "shellyplus1-test", Name: "Test Device", Model: "SNSW-001X16EU",
			Generation: 2, Version: "1.0.0", MAC: "AA:BB:CC:DD:EE:FF",
		},
		Config:    json.RawMessage(`{"sys":{"device":{"name":"Test Device"}}}`),
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("marshal backup: %v", err)
	}
	return data
}

// setupRestore installs a memmap FS holding a valid backup at testBackupPath and
// returns the captured output buffers plus a Factory bound to them.
func setupRestore(t *testing.T) (out, errOut *bytes.Buffer, f *cmdutil.Factory) {
	t.Helper()
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })

	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	ios := iostreams.Test(nil, out, errOut)
	f = cmdutil.NewFactory().SetIOStreams(ios)

	if err := afero.WriteFile(config.Fs(), testBackupPath, validBackupData(t), 0o600); err != nil {
		t.Fatalf("write backup file: %v", err)
	}
	return out, errOut, f
}

func restoreTestCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestRun_Restore_Success drives the on-LAN restore path through the service stub.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_Success(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	if combined := out.String() + errOut.String(); !strings.Contains(combined, "Backup restored to dev") {
		t.Errorf("missing success message, got %q", combined)
	}
}

// TestRun_Restore_Fails covers the on-LAN restore failure path.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_Fails(t *testing.T) {
	_, _, f := setupRestore(t)
	stub := &stubRestoreService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return nil, errStubRestore
		},
	}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "failed to restore backup") {
		t.Fatalf("expected restore failure, got %v", err)
	}
}

// TestRun_Restore_PartialFailure covers B4: shelly-go reports a per-section
// rejection as Success=false with a nil top-level error. run must NOT print a
// success line, must surface the rejected section, and must return an error so
// the exit code is non-zero.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_Restore_PartialFailure(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{
		restore: func(context.Context, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, error) {
			return &clibackup.RestoreResult{Success: false, Errors: []string{"wifi section rejected"}}, nil
		},
	}
	opts := &Options{Factory: f, Device: "dev", FilePath: testBackupPath, svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil {
		t.Fatal("a partial restore failure must return a non-nil error")
	}
	combined := out.String() + errOut.String()
	if strings.Contains(combined, "Backup restored to") {
		t.Errorf("must not print a success line on partial failure, got %q", combined)
	}
	if !strings.Contains(combined, "wifi section rejected") {
		t.Errorf("the rejected section must be surfaced, got %q", combined)
	}
}

// TestRun_RestoreViaAP_Success drives the --to-ap dispatch through restoreViaAP to
// a successful at-AP restore that reports the device's new LAN address.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreViaAP_Success(t *testing.T) {
	out, errOut, f := setupRestore(t)
	stub := &stubRestoreService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return &clibackup.RestoreResult{Success: true}, "10.23.47.227", nil
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	if err := run(restoreTestCtx(t), opts); err != nil {
		t.Fatalf("run: %v", err)
	}
	combined := out.String() + errOut.String()
	if !strings.Contains(combined, "Backup restored to fr") || !strings.Contains(combined, "10.23.47.227") {
		t.Errorf("missing AP success / new address, got %q", combined)
	}
}

// TestRun_RestoreViaAP_Fails covers the at-AP restore failure path.
//
//nolint:paralleltest // Test modifies global state via config.SetFs
func TestRun_RestoreViaAP_Fails(t *testing.T) {
	_, _, f := setupRestore(t)
	stub := &stubRestoreService{
		restoreToAP: func(context.Context, string, string, string, *clibackup.DeviceBackup, clibackup.RestoreOptions) (*clibackup.RestoreResult, string, error) {
			return nil, "", errStubRestore
		},
	}
	opts := &Options{Factory: f, Device: "fr", FilePath: testBackupPath, ToAP: "ShellyBulbDuo-AABBCC", svc: stub}

	err := run(restoreTestCtx(t), opts)
	if err == nil || !strings.Contains(err.Error(), "failed to restore via AP") {
		t.Fatalf("expected AP restore failure, got %v", err)
	}
}
