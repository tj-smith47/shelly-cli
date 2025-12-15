// Package export provides the kvs export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	File    string
	Format  string // json or yaml
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs export command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "export <device> [file]",
		Aliases: []string{"exp", "save", "dump"},
		Short:   "Export KVS data to file",
		Long: `Export all key-value pairs from the device to a file.

If no file is specified, output is written to stdout.
The export format can be JSON (default) or YAML.`,
		Example: `  # Export to JSON file
  shelly kvs export living-room kvs-backup.json

  # Export to YAML file
  shelly kvs export living-room kvs-backup.yaml --format yaml

  # Export to stdout
  shelly kvs export living-room

  # Export to stdout as YAML
  shelly kvs export living-room --format yaml`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				opts.File = args[1]
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Format, "format", "f", "json", "Output format (json, yaml)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate format
	switch opts.Format {
	case "json", "yaml", "yml":
	default:
		return fmt.Errorf("unsupported format: %s (use json or yaml)", opts.Format)
	}

	// Export data
	var data *shelly.KVSExport
	err := cmdutil.RunWithSpinner(ctx, ios, "Exporting KVS data...", func(ctx context.Context) error {
		var exportErr error
		data, exportErr = svc.ExportKVS(ctx, opts.Device)
		return exportErr
	})
	if err != nil {
		return err
	}

	// Encode data
	var encoded []byte
	switch opts.Format {
	case "yaml", "yml":
		encoded, err = yaml.Marshal(data)
	default:
		encoded, err = json.MarshalIndent(data, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	// Write output
	if opts.File != "" {
		if err := os.WriteFile(opts.File, encoded, 0o600); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		ios.Success("Exported %d key(s) to %s", len(data.Items), opts.File)
	} else {
		ios.Println(string(encoded))
	}

	return nil
}
