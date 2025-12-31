// Package show provides the cache show command.
package show

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the cache show command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "show",
		Short:   "Show cache statistics",
		Long:    "Display information about the discovery cache.",
		Aliases: []string{"s", "stats"},
		Example: `  # Show cache statistics
  shelly cache show`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}
	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	cacheDir, err := config.CacheDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache directory does not exist")
		return nil
	}

	var totalSize int64
	var fileCount int

	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return err
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, map[string]any{
			"location": cacheDir,
			"files":    fileCount,
			"size":     totalSize,
		})
	}

	table := output.NewTable("Property", "Value")
	table.AddRow("Location", cacheDir)
	table.AddRow("Files", fmt.Sprintf("%d", fileCount))
	table.AddRow("Size", output.FormatSize(totalSize))

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	return nil
}
