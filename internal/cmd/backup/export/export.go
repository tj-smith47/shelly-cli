// Package export provides the backup export subcommand.
package export

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/backup"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory   *cmdutil.Factory
	All       bool
	Directory string
	Format    string
	Parallel  int
}

// NewCommand creates the backup export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:  f,
		All:      true,
		Format:   "json",
		Parallel: 3,
	}

	cmd := &cobra.Command{
		Use:     "export <directory>",
		Aliases: []string{"save", "dump"},
		Short:   "Export backups for all registered devices",
		Long: `Export backup files for all registered devices to a directory.

Creates one backup file per device, named by device ID.
Use --format to choose JSON or YAML output.`,
		Example: `  # Export all device backups to directory
  shelly backup export ./backups

  # Export in YAML format
  shelly backup export ./backups --format yaml

  # Export with parallel processing
  shelly backup export ./backups --parallel 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Directory = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", true, "Export all registered devices")
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "json", "Output format (json, yaml)")
	cmd.Flags().IntVar(&opts.Parallel, "parallel", 3, "Number of parallel backups")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*10)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Get registered devices
	cfg := config.Get()
	if len(cfg.Devices) == 0 {
		ios.Info("No registered devices found")
		ios.Info("Use 'shelly device add' to register devices first")
		return nil
	}

	// Create directory if it doesn't exist
	if err := config.Fs().MkdirAll(opts.Directory, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ios.Info("Exporting backups for %d devices...", len(cfg.Devices))
	ios.Println()

	exporter := shelly.NewBackupExporter(opts.Factory.ShellyService())
	exportOpts := shelly.BackupExportOptions{
		Directory:  opts.Directory,
		Format:     opts.Format,
		Parallel:   opts.Parallel,
		BackupOpts: backup.Options{},
	}

	results := exporter.ExportAll(ctx, cfg.Devices, exportOpts)
	term.DisplayBackupExportResults(ios, results)

	ios.Println()

	success, failed := shelly.CountBackupResults(results)
	if success > 0 {
		ios.Success("Successfully exported %d backup(s) to %s", success, opts.Directory)
	}
	if failed > 0 {
		ios.Warning("Failed to export %d device(s):", failed)
		for _, r := range shelly.FailedBackupResults(results) {
			ios.Printf("  - %s: %v\n", r.DeviceName, r.Error)
		}
	}

	return nil
}
