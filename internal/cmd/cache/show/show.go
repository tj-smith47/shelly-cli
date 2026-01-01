// Package show provides the cache show command.
package show

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
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
	fs := config.Fs()

	cacheDir, err := config.CacheDir()
	if err != nil {
		return err
	}

	if _, err := fs.Stat(cacheDir); os.IsNotExist(err) {
		ios.Info("Cache directory does not exist")
		return nil
	}

	var totalSize int64
	var fileCount int

	err = afero.Walk(fs, cacheDir, func(_ string, info os.FileInfo, err error) error {
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

	builder := table.NewBuilder("Property", "Value")
	builder.AddRow("Location", cacheDir)
	builder.AddRow("Files", fmt.Sprintf("%d", fileCount))
	builder.AddRow("Size", output.FormatSize(totalSize))

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	return nil
}
