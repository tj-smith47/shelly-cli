package github

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// InstallRelease downloads and installs a release binary.
func (c *Client) InstallRelease(ctx context.Context, ios *iostreams.IOStreams, release *Release) error {
	ios.StartProgress(fmt.Sprintf("Downloading shelly %s...", release.Version()))

	// Find the appropriate asset for this platform
	asset := release.FindAssetForPlatform()
	if asset == nil {
		ios.StopProgress()
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download and extract
	binaryPath, cleanup, err := c.DownloadAndExtract(ctx, asset, "shelly")
	if err != nil {
		ios.StopProgress()
		return fmt.Errorf("failed to download: %w", err)
	}
	defer cleanup()

	ios.StopProgress()

	// Verify checksum if available
	checksumAsset := release.FindChecksumAsset(asset.Name)
	if checksumAsset != nil {
		ios.StartProgress("Verifying checksum...")
		if err := c.VerifyChecksum(ctx, ios, binaryPath, asset.Name, checksumAsset); err != nil {
			ios.StopProgress()
			return fmt.Errorf("checksum verification failed: %w", err)
		}
		ios.StopProgress()
		ios.Success("Checksum verified")
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
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

	return nil
}

// ReplaceBinary replaces the binary at targetPath with the one at newPath.
func ReplaceBinary(ios *iostreams.IOStreams, newPath, targetPath string) error {
	// Read the new binary
	newBinary, err := os.ReadFile(newPath) //nolint:gosec // G304: newPath is from controlled temp directory
	if err != nil {
		return fmt.Errorf("read new binary: %w", err)
	}

	// Get permissions of the old binary
	info, err := os.Stat(targetPath)
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
	if err := os.WriteFile(targetPath, newBinary, mode); err != nil {
		return restoreFromBackup(backupPath, targetPath, err)
	}

	// Remove backup
	if rerr := os.Remove(backupPath); rerr != nil {
		ios.DebugErr("removing backup", rerr)
	}

	return nil
}

// createBackup backs up the file at targetPath to backupPath.
func createBackup(ios *iostreams.IOStreams, targetPath, backupPath string) error {
	if err := os.Rename(targetPath, backupPath); err != nil {
		// On Windows, we might need to copy instead
		if runtime.GOOS != "windows" {
			return fmt.Errorf("backup failed: %w", err)
		}

		if copyErr := copyFile(ios, targetPath, backupPath); copyErr != nil {
			return fmt.Errorf("backup failed: %w", copyErr)
		}
		// Try to remove original (might fail if in use)
		if rerr := os.Remove(targetPath); rerr != nil {
			ios.DebugErr("removing original binary", rerr)
		}
	}
	return nil
}

// restoreFromBackup restores a backup after a failed write.
func restoreFromBackup(backupPath, targetPath string, writeErr error) error {
	if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
		return fmt.Errorf("write failed (%w) and restore failed: %w", writeErr, restoreErr)
	}
	return fmt.Errorf("write failed: %w", writeErr)
}

// copyFile copies a file from src to dst.
func copyFile(ios *iostreams.IOStreams, src, dst string) error {
	source, err := os.Open(src) //nolint:gosec // G304: src is the current executable path
	if err != nil {
		return err
	}
	defer func() {
		if cerr := source.Close(); cerr != nil {
			ios.DebugErr("closing source file", cerr)
		}
	}()

	destination, err := os.Create(dst) //nolint:gosec // G304: dst is backup path derived from executable
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
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return execPath, nil
}

// RestartCLI spawns a new process to replace the current one.
// This is used after self-update to run the new version.
func RestartCLI(ctx context.Context, args []string) error {
	execPath, err := GetExecutablePath()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, execPath, args...) //nolint:gosec // G204: execPath is the current executable
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
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
