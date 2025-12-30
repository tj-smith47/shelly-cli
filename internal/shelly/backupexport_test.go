package shelly

import (
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
)

const (
	testBackupAddress = "192.168.1.101"
)

func TestCountBackupResults(t *testing.T) {
	t.Parallel()

	results := []BackupResult{
		{Success: true},
		{Success: true},
		{Success: false},
		{Success: true},
		{Success: false},
	}

	success, failed := CountBackupResults(results)

	if success != 3 {
		t.Errorf("got success=%d, want 3", success)
	}
	if failed != 2 {
		t.Errorf("got failed=%d, want 2", failed)
	}
}

func TestFailedBackupResults(t *testing.T) {
	t.Parallel()

	results := []BackupResult{
		{DeviceName: "device1", Success: true},
		{DeviceName: "device2", Success: false},
		{DeviceName: "device3", Success: true},
		{DeviceName: "device4", Success: false},
	}

	failed := FailedBackupResults(results)

	if len(failed) != 2 {
		t.Errorf("got %d failed results, want 2", len(failed))
	}

	for _, f := range failed {
		if f.Success {
			t.Error("expected all failed results to have Success=false")
		}
	}
}

func TestBackupExportOptions_Fields(t *testing.T) {
	t.Parallel()

	opts := BackupExportOptions{
		Directory:  "/tmp/backups",
		Format:     "json",
		Parallel:   4,
		BackupOpts: backup.Options{SkipScripts: true},
	}

	if opts.Directory != "/tmp/backups" {
		t.Errorf("got Directory=%q, want %q", opts.Directory, "/tmp/backups")
	}
	if opts.Format != "json" {
		t.Errorf("got Format=%q, want %q", opts.Format, "json")
	}
	if opts.Parallel != 4 {
		t.Errorf("got Parallel=%d, want 4", opts.Parallel)
	}
	if !opts.BackupOpts.SkipScripts {
		t.Error("expected SkipScripts to be true")
	}
}

func TestBackupResult_Fields(t *testing.T) {
	t.Parallel()

	result := BackupResult{
		DeviceName: "living-room",
		Address:    testBackupAddress,
		FilePath:   "/tmp/backups/living-room.json",
		Success:    true,
		Error:      nil,
	}

	if result.DeviceName != "living-room" {
		t.Errorf("got DeviceName=%q, want %q", result.DeviceName, "living-room")
	}
	if result.Address != testBackupAddress {
		t.Errorf("got Address=%q, want %q", result.Address, testBackupAddress)
	}
	if result.FilePath != "/tmp/backups/living-room.json" {
		t.Errorf("got FilePath=%q, want %q", result.FilePath, "/tmp/backups/living-room.json")
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
	if result.Error != nil {
		t.Errorf("expected Error to be nil, got %v", result.Error)
	}
}

func TestNewBackupExporter(t *testing.T) {
	t.Parallel()

	exporter := NewBackupExporter(nil)

	if exporter == nil {
		t.Fatal("expected non-nil exporter")
	}
}
