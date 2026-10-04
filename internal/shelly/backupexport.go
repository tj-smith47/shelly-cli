// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"maps"
	"path/filepath"
	"slices"

	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

// BackupExportOptions configures the backup export operation.
type BackupExportOptions struct {
	// Directory is the output directory for backup files.
	Directory string
	// Parallel is the number of concurrent backup operations.
	Parallel int
	// BackupOpts are passed to the underlying CreateBackup call.
	BackupOpts backup.Options
	// Encrypt, when set, is the password each backup file is AES-encrypted with.
	Encrypt string
	// AutoName names each file like `backup create` does
	// ({device}-{mac}-{date}.json) instead of {device}.json.
	AutoName bool
}

// BackupResult represents the result of a single device backup.
type BackupResult struct {
	DeviceName string
	Address    string
	FilePath   string
	Success    bool
	Error      error
}

// BackupExporter handles exporting backups for multiple devices.
type BackupExporter struct {
	svc *Service
}

// NewBackupExporter creates a new BackupExporter.
func NewBackupExporter(svc *Service) *BackupExporter {
	return &BackupExporter{svc: svc}
}

// ExportAll exports backups for all provided devices concurrently.
// It returns one result per device, sorted by device name, including failures.
func (e *BackupExporter) ExportAll(ctx context.Context, devices map[string]model.Device, opts BackupExportOptions) []BackupResult {
	// Cap parallelism to global rate limit (silently, no ios available) and
	// floor to 1 — SetLimit(0) deadlocks errgroup permanently.
	globalMax := config.GetGlobalMaxConcurrent()
	parallelism := min(max(opts.Parallel, 1), globalMax)

	names := slices.Sorted(maps.Keys(devices))
	results := make([]BackupResult, len(names))

	var g errgroup.Group
	g.SetLimit(parallelism)

	for i, name := range names {
		g.Go(func() error {
			results[i] = e.exportDevice(ctx, name, devices[name].Address, opts)
			return nil
		})
	}

	// Every goroutine returns nil; failures are carried in results.
	if err := g.Wait(); err != nil {
		return results
	}

	return results
}

// exportDevice exports a single device and returns the result.
func (e *BackupExporter) exportDevice(ctx context.Context, name, addr string, opts BackupExportOptions) BackupResult {
	result := BackupResult{
		DeviceName: name,
		Address:    addr,
	}

	// Each device gets the same budget a single `backup create` has, so one
	// slow device cannot use up the time of the devices queued behind it.
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout*3)
	defer cancel()

	// Resolve by registry name, as `backup create <device>` does, so the
	// device's stored auth applies.
	bkp, err := e.svc.CreateBackup(ctx, name, opts.BackupOpts)
	if err != nil {
		result.Error = err
		return result
	}

	filePath := filepath.Join(opts.Directory, export.SanitizeFilename(name)+".json")
	if opts.AutoName {
		filePath, err = backup.AutoSavePathIn(opts.Directory, name, bkp, "json")
		if err != nil {
			result.Error = err
			return result
		}
	}

	if err := export.WriteBackupFile(bkp, filePath, opts.Encrypt); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.FilePath = filePath
	return result
}

// CountBackupResults returns success and failure counts from backup results.
func CountBackupResults(results []BackupResult) (success, failed int) {
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	return
}

// FailedBackupResults returns only the failed results.
func FailedBackupResults(results []BackupResult) []BackupResult {
	var failures []BackupResult
	for _, r := range results {
		if !r.Success {
			failures = append(failures, r)
		}
	}
	return failures
}
