// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"sync"

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
	// Format is the output format (json or yaml).
	Format string
	// Parallel is the number of concurrent backup operations.
	Parallel int
	// BackupOpts are passed to the underlying CreateBackup call.
	BackupOpts backup.Options
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
// It returns results for each device, including failures.
func (e *BackupExporter) ExportAll(ctx context.Context, devices map[string]model.Device, opts BackupExportOptions) []BackupResult {
	// Cap parallelism to global rate limit (silently, no ios available)
	parallelism := opts.Parallel
	globalMax := config.GetGlobalMaxConcurrent()
	if parallelism > globalMax {
		parallelism = globalMax
	}

	var (
		mu      sync.Mutex
		results []BackupResult
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(parallelism)

	for name, device := range devices {
		deviceName := name
		deviceAddr := device.Address

		g.Go(func() error {
			result := e.exportDevice(ctx, deviceName, deviceAddr, opts)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			return nil
		})
	}

	// Wait for all goroutines; errors are tracked in results, not returned
	if err := g.Wait(); err != nil {
		// Goroutines always return nil, so this is defensive only
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

	bkp, err := e.svc.CreateBackup(ctx, addr, opts.BackupOpts)
	if err != nil {
		result.Error = err
		return result
	}

	// Write backup file
	filename := export.SanitizeFilename(name) + "." + opts.Format
	filePath := opts.Directory + "/" + filename

	if err := export.WriteBackupFile(bkp, filePath, opts.Format); err != nil {
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
