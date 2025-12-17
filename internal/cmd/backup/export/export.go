// Package export provides the backup export subcommand.
package export

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/shelly/export"
)

var (
	allFlag      bool
	formatFlag   string
	parallelFlag int
)

// NewCommand creates the backup export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", true, "Export all registered devices")
	cmd.Flags().StringVarP(&formatFlag, "format", "f", "json", "Output format (json, yaml)")
	cmd.Flags().IntVar(&parallelFlag, "parallel", 3, "Number of parallel backups")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, dir string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*10)
	defer cancel()

	ios := f.IOStreams()

	// Get registered devices
	cfg := config.Get()
	if len(cfg.Devices) == 0 {
		ios.Info("No registered devices found")
		ios.Info("Use 'shelly device add' to register devices first")
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ios.Info("Exporting backups for %d devices...", len(cfg.Devices))
	ios.Println()

	exporter := export.NewBackupExporter(f.ShellyService())
	opts := export.BackupExportOptions{
		Directory:  dir,
		Format:     formatFlag,
		Parallel:   parallelFlag,
		BackupOpts: shelly.BackupOptions{},
	}

	results := exporter.ExportAll(ctx, cfg.Devices, opts)
	cmdutil.DisplayBackupExportResults(ios, results)

	ios.Println()

	success, failed := export.CountResults(results)
	if success > 0 {
		ios.Success("Successfully exported %d backup(s) to %s", success, dir)
	}
	if failed > 0 {
		ios.Warning("Failed to export %d device(s):", failed)
		for _, r := range export.FailedResults(results) {
			ios.Printf("  - %s: %v\n", r.DeviceName, r.Error)
		}
	}

	return nil
}
