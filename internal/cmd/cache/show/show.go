// Package show provides the cache show command.
package show

import (
	"context"
	"fmt"
	"sort"

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
		Use:   "show",
		Short: "Show cache statistics",
		Long: `Display detailed information about the file cache.

Shows cache statistics including:
  - Total entries and size
  - Entries by data type
  - Device count
  - Expired entry count
  - Oldest and newest entries`,
		Aliases: []string{"s", "stats", "status"},
		Example: `  # Show cache statistics
  shelly cache show

  # Show cache stats in JSON format
  shelly cache show -o json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}
	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	fc := opts.Factory.FileCache()

	cacheDir, err := config.CacheDir()
	if err != nil {
		return err
	}

	if fc == nil {
		ios.Info("Cache not available")
		return nil
	}

	stats, err := fc.Stats()
	if err != nil {
		return err
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, map[string]any{
			"location":        cacheDir,
			"total_entries":   stats.TotalEntries,
			"total_size":      stats.TotalSize,
			"device_count":    stats.DeviceCount,
			"expired_entries": stats.ExpiredEntries,
			"oldest_entry":    stats.OldestEntry,
			"newest_entry":    stats.NewestEntry,
			"type_counts":     stats.TypeCounts,
		})
	}

	// Summary table
	builder := table.NewBuilder("Property", "Value")
	builder.AddRow("Location", cacheDir)
	builder.AddRow("Total Entries", fmt.Sprintf("%d", stats.TotalEntries))
	builder.AddRow("Total Size", output.FormatSize(stats.TotalSize))
	builder.AddRow("Devices", fmt.Sprintf("%d", stats.DeviceCount))
	builder.AddRow("Expired", fmt.Sprintf("%d", stats.ExpiredEntries))
	if !stats.OldestEntry.IsZero() {
		builder.AddRow("Oldest Entry", output.FormatAge(stats.OldestEntry))
	}
	if !stats.NewestEntry.IsZero() {
		builder.AddRow("Newest Entry", output.FormatAge(stats.NewestEntry))
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	// Type breakdown
	if len(stats.TypeCounts) == 0 {
		return nil
	}

	ios.Println("\nEntries by Type:")
	types := make([]string, 0, len(stats.TypeCounts))
	for t := range stats.TypeCounts {
		types = append(types, t)
	}
	sort.Strings(types)

	typeBuilder := table.NewBuilder("Type", "Count")
	for _, t := range types {
		typeBuilder.AddRow(t, fmt.Sprintf("%d", stats.TypeCounts[t]))
	}

	typeTbl := typeBuilder.WithModeStyle(ios).Build()
	if err := typeTbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print type table", err)
	}

	return nil
}
