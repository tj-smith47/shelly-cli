package github

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// defaultFs is the package-level filesystem used for install operations.
// This can be replaced in tests with an in-memory filesystem.
var (
	defaultFs   afero.Fs = afero.NewOsFs()
	defaultFsMu sync.RWMutex
)

// Function variables for testability.
// These can be replaced in tests to inject mock behavior.
var (
	osExecutable     = os.Executable
	evalSymlinks     = filepath.EvalSymlinks
	execCommandStart = defaultExecCommandStart
	runtimeGOOS      = runtime.GOOS
)

// defaultExecCommandStart is the default implementation that starts a command.
func defaultExecCommandStart(ctx context.Context, path string, args []string) error {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// SetFs sets the package-level filesystem for testing.
// Pass nil to reset to the real OS filesystem.
func SetFs(fs afero.Fs) {
	defaultFsMu.Lock()
	defer defaultFsMu.Unlock()
	if fs == nil {
		defaultFs = afero.NewOsFs()
	} else {
		defaultFs = fs
	}
}

// getFs returns the current package-level filesystem.
func getFs() afero.Fs {
	defaultFsMu.RLock()
	defer defaultFsMu.RUnlock()
	return defaultFs
}

// InstallRelease downloads and installs a release binary.
func (c *Client) InstallRelease(ctx context.Context, ios *iostreams.IOStreams, release *Release) error {
	ios.StartProgress(fmt.Sprintf("Downloading shelly %s...", release.Version()))

	// Find the appropriate asset for this platform
	asset := release.FindAssetForPlatform()
	if asset == nil {
		ios.StopProgress()
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the asset (but don't extract yet)
	downloadResult, err := c.DownloadAssetToTemp(ctx, asset)
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to download: %w", err)
	}
	defer downloadResult.Cleanup()

	ios.StopProgress()

	// Verify checksum of the archive before extraction
	checksumAsset := release.FindChecksumAsset(asset.Name)
	if checksumAsset != nil {
		ios.StartProgress("Verifying checksum...")
		if err := c.VerifyChecksum(ctx, ios, downloadResult.ArchivePath, asset.Name, checksumAsset); err != nil {
			ios.StopProgress()
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		ios.StopProgress()
		ios.Success("Checksum verified")
	}

	// Extract the binary
	ios.StartProgress("Extracting...")
	binaryPath, err := c.ExtractBinary(downloadResult.ArchivePath, downloadResult.TmpDir, "shelly")
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to extract: %w", err)
	}
	ios.StopProgress()

	// Get current executable path
	execPath, err := osExecutable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = evalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Replace the binary
	ios.StartProgress("Installing update...")
	if err := ReplaceBinary(ios, binaryPath, execPath); err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to install update: %w", err)
	}
	ios.StopProgress()

	ios.Success("Successfully updated to shelly %s", release.Version())
	ios.Info("Restart shelly to use the new version")

	// Update the version cache so notifications show the correct version
	if err := version.WriteCache(release.Version()); err != nil {
		ios.DebugErr("updating version cache", err)
	}

	return nil
}

// ReplaceBinary replaces the binary at targetPath with the one at newPath.
func ReplaceBinary(ios *iostreams.IOStreams, newPath, targetPath string) error {
	fs := getFs()

	// Read the new binary
	newBinary, err := afero.ReadFile(fs, newPath)
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}

	// Get permissions of the old binary
	info, err := fs.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("stat target: %w", err)
	}
	mode := info.Mode()

	// Create backup
	backupPath := targetPath + ".bak"
	if err := createBackup(ios, targetPath, backupPath); err != nil {
		return err
	}

	// Write new binary
	if err := afero.WriteFile(fs, targetPath, newBinary, mode); err != nil {
		return restoreFromBackup(backupPath, targetPath, err)
	}

	// Remove backup
	if rerr := fs.Remove(backupPath); rerr != nil {
		ios.DebugErr("removing backup", rerr)
	}

	return nil
}

// createBackup backs up the file at targetPath to backupPath.
func createBackup(ios *iostreams.IOStreams, targetPath, backupPath string) error {
	fs := getFs()

	if err := fs.Rename(targetPath, backupPath); err != nil {
		// On Windows, we might need to copy instead
		if runtimeGOOS != "windows" {
			return fmt.Errorf("backup failed: %w", err)
		}

		if copyErr := copyFile(ios, targetPath, backupPath); copyErr != nil {
			return fmt.Errorf("backup failed: %w", copyErr)
		}
		// Try to remove original (might fail if in use)
		if rerr := fs.Remove(targetPath); rerr != nil {
			ios.DebugErr("removing original binary", rerr)
		}
	}
	return nil
}

// restoreFromBackup restores a backup after a failed write.
func restoreFromBackup(backupPath, targetPath string, writeErr error) error {
	fs := getFs()

	if restoreErr := fs.Rename(backupPath, targetPath); restoreErr != nil {
		return fmt.Errorf("write failed (%w) and restore failed: %w", writeErr, restoreErr)
	}
	return fmt.Errorf("write failed: %w", writeErr)
}

// copyFile copies a file from src to dst.
func copyFile(ios *iostreams.IOStreams, src, dst string) error {
	fs := getFs()

	source, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := source.Close(); cerr != nil {
			ios.DebugErr("closing source file", cerr)
		}
	}()

	destination, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := destination.Close(); cerr != nil {
			ios.DebugErr("closing destination file", cerr)
		}
	}()

	_, err = io.Copy(destination, source)
	return err
}

// GetExecutablePath returns the resolved path to the current executable.
func GetExecutablePath() (string, error) {
	execPath, err := osExecutable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = evalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return execPath, nil
}

// InstallMethod represents how the CLI was installed.
type InstallMethod int

const (
	// InstallMethodUnknown is the default/unknown installation method.
	InstallMethodUnknown InstallMethod = iota
	// InstallMethodHomebrew indicates installation via Homebrew.
	InstallMethodHomebrew
	// InstallMethodDirect indicates direct download or go install (supports self-update).
	InstallMethodDirect
)

// InstallInfo contains information about how the CLI was installed.
type InstallInfo struct {
	Method        InstallMethod
	Path          string
	UpdateCommand string // Command to show user if self-update isn't supported
}

// DetectInstallMethod detects how the CLI was installed based on the executable path.
func DetectInstallMethod() InstallInfo {
	execPath, err := GetExecutablePath()
	if err != nil {
		return InstallInfo{Method: InstallMethodDirect, Path: ""}
	}

	return DetectInstallMethodFromPath(execPath)
}

// DetectInstallMethodFromPath determines installation method from a path.
func DetectInstallMethodFromPath(execPath string) InstallInfo {
	info := InstallInfo{
		Method: InstallMethodDirect,
		Path:   execPath,
	}

	// Check for Homebrew installation
	// macOS ARM: /opt/homebrew/Cellar/shelly/...
	// macOS Intel: /usr/local/Cellar/shelly/...
	// Linux: /home/linuxbrew/.linuxbrew/Cellar/shelly/...
	homebrewPaths := []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/home/linuxbrew/.linuxbrew/",
	}

	for _, prefix := range homebrewPaths {
		if len(execPath) >= len(prefix) && execPath[:len(prefix)] == prefix {
			info.Method = InstallMethodHomebrew
			info.UpdateCommand = "brew upgrade shelly"
			return info
		}
	}

	// Also check HOMEBREW_CELLAR environment variable
	if cellar := getEnv("HOMEBREW_CELLAR"); cellar != "" {
		if len(execPath) >= len(cellar) && execPath[:len(cellar)] == cellar {
			info.Method = InstallMethodHomebrew
			info.UpdateCommand = "brew upgrade shelly"
			return info
		}
	}

	// Everything else (go install, direct download, dev build) supports self-update
	return info
}

// getEnv is a variable for testing.
var getEnv = os.Getenv

// CanSelfUpdate returns true if the installation method supports self-update.
func (i InstallInfo) CanSelfUpdate() bool {
	return i.Method != InstallMethodHomebrew
}

// RestartCLI spawns a new process to replace the current one.
// This is used after self-update to run the new version.
func RestartCLI(ctx context.Context, args []string) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return err
	}

	return execCommandStart(ctx, execPath, args)
}

// ConfirmFunc is a function that asks the user for confirmation.
type ConfirmFunc func(prompt string, skipConfirm bool) (bool, error)

// PerformRollback handles the complete rollback flow.
func (c *Client) PerformRollback(ctx context.Context, ios *iostreams.IOStreams, currentVersion string, includePre bool, confirm ConfirmFunc, skipConfirm bool) error {
	release, err := c.FindPreviousRelease(ctx, currentVersion, includePre)
	if err != nil {
		return fmt.Errorf("no previous version available for rollback")
	}

	ios.RollbackInfo(currentVersion, release.Version())

	confirmed, err := confirm("Proceed with rollback?", skipConfirm)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !confirmed {
		ios.Info("Rollback cancelled")
		return nil
	}

	return c.InstallRelease(ctx, ios, release)
}

// PerformUpdate handles the complete update flow including confirmation.
func (c *Client) PerformUpdate(ctx context.Context, ios *iostreams.IOStreams, release *Release, currentVersion, releaseNotes string, confirm ConfirmFunc, skipConfirm bool) error {
	ios.UpdateInfo(currentVersion, release.Version(), releaseNotes)

	confirmed, err := confirm("\nProceed with update?", skipConfirm)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !confirmed {
		ios.Info("Update cancelled")
		return fmt.Errorf("update cancelled by user")
	}

	return c.InstallRelease(ctx, ios, release)
}
