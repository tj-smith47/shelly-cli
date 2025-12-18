// Package clear provides the cache clear command.
package clear

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// NewCommand creates the cache clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clear",
		Short:   "Clear the discovery cache",
		Long:    "Clear all cached device discovery results.",
		Aliases: []string{"c", "rm"},
		Example: `  # Clear the discovery cache
  shelly cache clear`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}
	return cmd
}

func run(_ context.Context, f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cacheDir, err := config.CacheDir()
	if err != nil {
		return err
	}

	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache is already empty")
		return nil
	}

	// Remove all files in cache directory
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		ios.Info("Cache is already empty")
		return nil
	}

	for _, entry := range entries {
		path := filepath.Join(cacheDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			ios.DebugErr("remove cache entry", err)
		}
	}

	ios.Success("Cache cleared")
	return nil
}
