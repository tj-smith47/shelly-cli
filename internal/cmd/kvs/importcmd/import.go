// Package importcmd provides the kvs import subcommand.
// Named importcmd to avoid conflict with the "import" keyword.
package importcmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Device    string
	File      string
	Overwrite bool
	DryRun    bool
	Yes       bool
	Factory   *cmdutil.Factory
}

// NewCommand creates the kvs import command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "import <device> <file>",
		Aliases: []string{"load", "restore"},
		Short:   "Import KVS data from file",
		Long: `Import key-value pairs from a file to the device.

By default, existing keys are skipped. Use --overwrite to replace them.
Use --dry-run to see what would be imported without making changes.`,
		Example: `  # Import from JSON file
  shelly kvs import living-room kvs-backup.json

  # Import with overwrite
  shelly kvs import living-room kvs-backup.json --overwrite

  # Dry run to see what would be imported
  shelly kvs import living-room kvs-backup.json --dry-run

  # Import without confirmation
  shelly kvs import living-room kvs-backup.json --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.File = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Overwrite, "overwrite", false, "Overwrite existing keys")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be imported without making changes")
	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithTimeout(ctx, 20*time.Second) // Allow more time for imports
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Parse the import file
	data, err := kvs.ParseImportFile(opts.File)
	if err != nil {
		return err
	}

	if len(data.Items) == 0 {
		ios.NoResults("No keys to import")
		return nil
	}

	// Show preview
	term.DisplayKVSImportPreview(ios, data)

	// Handle dry run
	if opts.DryRun {
		term.DisplayKVSDryRun(ios, len(data.Items), opts.Overwrite)
		return nil
	}

	// Confirm import
	action := "Import"
	if opts.Overwrite {
		action = "Import and overwrite"
	}
	msg := fmt.Sprintf("%s %d key(s) to %s?", action, len(data.Items), opts.Device)
	confirmed, err := opts.Factory.ConfirmAction(msg, opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Aborted")
		return nil
	}

	// Execute import
	var imported, skipped int
	err = cmdutil.RunWithSpinner(ctx, ios, "Importing KVS data...", func(ctx context.Context) error {
		var importErr error
		imported, skipped, importErr = svc.ImportKVS(ctx, opts.Device, data, opts.Overwrite)
		return importErr
	})
	if err != nil {
		return err
	}

	term.DisplayKVSImportResults(ios, imported, skipped)
	return nil
}
