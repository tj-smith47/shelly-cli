package cmdutil

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

// stubAPRestorer drives RestoreAtAP without hopping the host's WiFi.
type stubAPRestorer struct {
	result  *backup.RestoreResult
	newAddr string
	err     error
}

func (s stubAPRestorer) RestoreToAP(_ context.Context, _, _, _ string, _ *backup.DeviceBackup, _ backup.RestoreOptions) (*backup.RestoreResult, string, error) {
	return s.result, s.newAddr, s.err
}

func newTestIOS() (*iostreams.IOStreams, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return iostreams.Test(&bytes.Buffer{}, out, &bytes.Buffer{}), out
}

// okReport reports success (nil) for any result; partialReport rejects.
func okReport(*iostreams.IOStreams, string, *backup.RestoreResult) error { return nil }

var errSectionRejected = errors.New("cloud section rejected")

func rejectReport(*iostreams.IOStreams, string, *backup.RestoreResult) error {
	return errSectionRejected
}

// TestRestoreAtAP_Success: a clean restore returns nil and surfaces the new LAN
// address so the user knows where the device landed.
func TestRestoreAtAP_Success(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: true}, newAddr: "10.23.47.227"}

	err := RestoreAtAP(context.Background(), ios, svc, "ShellyBulbDuo-AABBCC", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", okReport)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if o := out.String(); !strings.Contains(o, "10.23.47.227") {
		t.Errorf("new LAN address should be surfaced, got %q", o)
	}
}

// TestRestoreAtAP_PartialRejection: a reporter error (a section rejected with no
// top-level transport error) propagates so the exit code is non-zero, yet the new
// address is STILL surfaced so a partial failure can be finished by hand.
func TestRestoreAtAP_PartialRejection(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: false}, newAddr: "10.23.47.227"}

	err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", rejectReport)
	if !errors.Is(err, errSectionRejected) {
		t.Fatalf("reporter error must propagate, got %v", err)
	}
	if o := out.String(); !strings.Contains(o, "10.23.47.227") {
		t.Errorf("address must still be surfaced on partial failure, got %q", o)
	}
}

// TestRestoreAtAP_TransportFailure: a RestoreToAP error is wrapped with the
// caller's prefix and reported before any result reporting.
func TestRestoreAtAP_TransportFailure(t *testing.T) {
	t.Parallel()
	ios, _ := newTestIOS()
	svc := stubAPRestorer{err: errors.New("connection refused")}
	reportCalled := false
	report := func(*iostreams.IOStreams, string, *backup.RestoreResult) error {
		reportCalled = true
		return nil
	}

	err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "migration via AP failed", report)
	if err == nil || !strings.Contains(err.Error(), "migration via AP failed") {
		t.Fatalf("expected wrapped transport failure, got %v", err)
	}
	if reportCalled {
		t.Error("reporter must not run when the restore call itself failed")
	}
}

// TestRestoreAtAP_NoAddress: when the device did not rejoin (empty newAddr), no
// "is live at" line is printed.
func TestRestoreAtAP_NoAddress(t *testing.T) {
	t.Parallel()
	ios, out := newTestIOS()
	svc := stubAPRestorer{result: &backup.RestoreResult{Success: true}, newAddr: ""}

	if err := RestoreAtAP(context.Background(), ios, svc, "ap", "192.168.33.1", "fr",
		&backup.DeviceBackup{}, backup.RestoreOptions{}, "restore via AP failed", okReport); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if o := out.String(); strings.Contains(o, "is live at") {
		t.Errorf("no address line should print when the device did not rejoin, got %q", o)
	}
}
