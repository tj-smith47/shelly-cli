// Package cache provides the cache command for managing CLI cache.
package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the cache command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage CLI cache",
		Long: `Manage the Shelly CLI cache directory.

The cache stores:
  - Device discovery results
  - Firmware update information
  - Version check data`,
	}

	cmd.AddCommand(newClearCommand(f))
	cmd.AddCommand(newShowCommand(f))

	return cmd
}

func newClearCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "clear",
		Aliases: []string{"clean", "purge"},
		Short:   "Clear the cache",
		Long:    `Clear all cached data from the CLI cache directory.`,
		Example: `  # Clear all cached data
  shelly cache clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClear(f)
		},
	}
}

func newShowCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "show",
		Aliases: []string{"info", "path"},
		Short:   "Show cache information",
		Long:    `Show the cache directory path and its contents.`,
		Example: `  # Show cache info
  shelly cache show`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(f)
		},
	}
}

func runClear(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache directory does not exist: %s", cacheDir)
		return nil
	}

	// Remove cache directory contents
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		path := filepath.Join(cacheDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			ios.Warning("Failed to remove %s: %v", entry.Name(), err)
		} else {
			count++
		}
	}

	ios.Success("Cleared %d item(s) from cache", count)
	return nil
}

func runShow(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	ios.Printf("Cache directory: %s\n", cacheDir)
	ios.Println("")

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache directory does not exist (no cached data)")
		return nil
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	if len(entries) == 0 {
		ios.Info("Cache is empty")
		return nil
	}

	ios.Printf("Contents:\n")
	var totalSize int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		size := info.Size()
		totalSize += size
		ios.Printf("  %s (%d bytes)\n", entry.Name(), size)
	}

	ios.Println("")
	ios.Printf("Total: %d item(s), %d bytes\n", len(entries), totalSize)

	return nil
}

func getCacheDir() (string, error) {
	configDir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "cache"), nil
}
