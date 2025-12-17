// Package export provides export format builders for device data.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
	BackupOpts shelly.BackupOptions
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
	svc *shelly.Service
}

// NewBackupExporter creates a new BackupExporter.
func NewBackupExporter(svc *shelly.Service) *BackupExporter {
	return &BackupExporter{svc: svc}
}

// ExportAll exports backups for all provided devices concurrently.
// It returns results for each device, including failures.
func (e *BackupExporter) ExportAll(ctx context.Context, devices map[string]model.Device, opts BackupExportOptions) []BackupResult {
	var (
		mu      sync.Mutex
		results []BackupResult
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(opts.Parallel)

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

	backup, err := e.svc.CreateBackup(ctx, addr, opts.BackupOpts)
	if err != nil {
		result.Error = err
		return result
	}

	// Write backup file
	filename := SanitizeFilename(name) + "." + opts.Format
	filePath := filepath.Join(opts.Directory, filename)

	if err := WriteBackupFile(backup, filePath, opts.Format); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	result.FilePath = filePath
	return result
}

// SanitizeFilename replaces problematic characters in a filename.
func SanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

// WriteBackupFile writes a backup to a file in the specified format.
func WriteBackupFile(backup *shelly.DeviceBackup, filePath, format string) error {
	var data []byte
	var err error

	switch format {
	case "yaml", "yml":
		data, err = yaml.Marshal(backup)
	default:
		data, err = json.MarshalIndent(backup, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CountResults returns success and failure counts from backup results.
func CountResults(results []BackupResult) (success, failed int) {
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	return
}

// FailedResults returns only the failed results.
func FailedResults(results []BackupResult) []BackupResult {
	var failures []BackupResult
	for _, r := range results {
		if !r.Success {
			failures = append(failures, r)
		}
	}
	return failures
}

// BackupFileInfo holds information about a backup file.
type BackupFileInfo struct {
	Filename    string    `json:"filename"`
	DeviceID    string    `json:"device_id"`
	DeviceModel string    `json:"device_model"`
	FWVersion   string    `json:"fw_version"`
	CreatedAt   time.Time `json:"created_at"`
	Encrypted   bool      `json:"encrypted"`
	Size        int64     `json:"size"`
}

// ScanBackupFiles scans a directory for backup files and returns their info.
func ScanBackupFiles(dir string) ([]BackupFileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	backups := make([]BackupFileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !IsBackupFile(name) {
			continue
		}

		filePath := filepath.Join(dir, name)
		info, err := ParseBackupFile(filePath)
		if err != nil {
			continue
		}
		info.Filename = name
		backups = append(backups, info)
	}
	return backups, nil
}

// IsBackupFile checks if a filename has a backup file extension.
func IsBackupFile(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

// ParseBackupFile reads and parses a backup file, returning its metadata.
func ParseBackupFile(filePath string) (BackupFileInfo, error) {
	var info BackupFileInfo

	data, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is derived from directory listing
	if err != nil {
		return info, err
	}

	backup, err := shelly.ValidateBackup(data)
	if err != nil {
		return info, err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return info, err
	}

	info.DeviceID = backup.Device().ID
	info.DeviceModel = backup.Device().Model
	info.FWVersion = backup.Device().FWVersion
	info.CreatedAt = backup.CreatedAt
	info.Encrypted = backup.Encrypted()
	info.Size = stat.Size()

	return info, nil
}

// MarshalBackup serializes a backup to the specified format.
func MarshalBackup(backup *shelly.DeviceBackup, format string) ([]byte, error) {
	switch format {
	case "yaml", "yml":
		return yaml.Marshal(backup)
	default:
		return json.MarshalIndent(backup, "", "  ")
	}
}
