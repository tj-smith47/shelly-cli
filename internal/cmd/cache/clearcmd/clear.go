// Package clearcmd provides the cache clear command.
package clearcmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the cache clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "clear",
		Short:   "Clear the discovery cache",
		Long:    "Clear all cached device discovery results.",
		Aliases: []string{"c", "rm"},
		Example: `  # Clear the discovery cache
  shelly cache clear`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}
	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	fs := config.Fs()

	cacheDir, err := config.CacheDir()
	if err != nil {
		return err
	}

	// Check if cache directory exists
	if _, err := fs.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache is already empty")
		return nil
	}

	// Remove all files in cache directory
	entries, err := afero.ReadDir(fs, cacheDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		ios.Info("Cache is already empty")
		return nil
	}

	for _, entry := range entries {
		path := filepath.Join(cacheDir, entry.Name())
		if err := fs.RemoveAll(path); err != nil {
			ios.DebugErr("remove cache entry", err)
		}
	}

	ios.Success("Cache cleared")
	return nil
}
