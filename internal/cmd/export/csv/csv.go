// Package csv provides the export csv subcommand.
package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Devices []string
	File    string
	NoHead  bool
	Factory *cmdutil.Factory
}

// NewCommand creates the export csv command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   "csv <devices...> [file]",
		Short: "Export device list as CSV",
		Long: `Export multiple devices as a CSV file.

The CSV includes device name, address, model, generation, and online status.
Use @all to export all registered devices.

If the last argument ends in .csv, it's treated as the output file.
Otherwise outputs to stdout.`,
		Example: `  # Export to stdout
  shelly export csv living-room bedroom

  # Export all devices to file
  shelly export csv @all devices.csv

  # Export specific devices
  shelly export csv living-room bedroom kitchen devices.csv

  # Without header row
  shelly export csv @all --no-header`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: cmdutil.CompleteDevicesWithGroups(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if last arg is a file
			if len(args) > 1 && strings.HasSuffix(args[len(args)-1], ".csv") {
				opts.File = args[len(args)-1]
				opts.Devices = args[:len(args)-1]
			} else {
				opts.Devices = args
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.NoHead, "no-header", false, "Omit CSV header row")

	return cmd
}

// DeviceRow represents a single CSV row.
type DeviceRow struct {
	Name       string
	Address    string
	Model      string
	Generation int
	App        string
	Online     bool
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout) // Allow more time for multiple devices
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Expand @all to all registered devices
	devices := cmdutil.ExpandDeviceArgs(opts.Devices)
	if len(devices) == 0 {
		return fmt.Errorf("no devices specified")
	}

	// Collect device data
	var rows []DeviceRow
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching device data...", func(ctx context.Context) error {
		for _, device := range devices {
			row := DeviceRow{Name: device}

			// Get device config for address
			deviceCfg, exists := config.GetDevice(device)
			if exists {
				row.Address = deviceCfg.Address
				row.Model = deviceCfg.Model
				row.Generation = deviceCfg.Generation
			}

			err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
				info := conn.Info()
				row.Address = deviceCfg.Address // Use config address
				row.Model = info.Model
				row.Generation = info.Generation
				row.App = info.App
				row.Online = true
				return nil
			})

			if err != nil {
				// Device offline - use config data (already populated above)
				row.Online = false
			}

			rows = append(rows, row)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Write CSV
	var writer *csv.Writer
	var file *os.File
	if opts.File == "" {
		writer = csv.NewWriter(ios.Out)
	} else {
		var createErr error
		file, createErr = os.Create(opts.File)
		if createErr != nil {
			return fmt.Errorf("failed to create file: %w", createErr)
		}
		defer iostreams.CloseWithDebug("closing csv export file", file)
		writer = csv.NewWriter(file)
	}

	// Write header
	if !opts.NoHead {
		if err := writer.Write([]string{"name", "address", "model", "generation", "app", "online"}); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Write rows
	for _, row := range rows {
		online := "no"
		if row.Online {
			online = "yes"
		}
		record := []string{
			row.Name,
			row.Address,
			row.Model,
			fmt.Sprintf("%d", row.Generation),
			row.App,
			online,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("csv write error: %w", err)
	}

	if opts.File != "" {
		ios.Success("Exported %d devices to %s", len(rows), opts.File)
	}

	return nil
}
