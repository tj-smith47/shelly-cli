// Package deletecmd provides the kvs delete subcommand.
package deletecmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Device  string
	Key     string
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "delete <device> <key>",
		Aliases: []string{"del", "rm", "remove"},
		Short:   "Delete a KVS key",
		Long: `Remove a key-value pair from the device's Key-Value Storage.

This operation is irreversible. Use --yes to skip the confirmation prompt.`,
		Example: `  # Delete a key (with confirmation)
  shelly kvs delete living-room my_key

  # Delete without confirmation
  shelly kvs delete living-room my_key --yes`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenNoComplete(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	kvsSvc := opts.Factory.KVSService()

	// Confirm deletion
	msg := fmt.Sprintf("Delete key %q from %s?", opts.Key, opts.Device)
	confirmed, err := opts.Factory.ConfirmAction(msg, opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Aborted")
		return nil
	}

	err = cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Deleting %q...", opts.Key), func(ctx context.Context) error {
		if deleteErr := kvsSvc.Delete(ctx, opts.Device, opts.Key); deleteErr != nil {
			return deleteErr
		}
		ios.Success("Key %q deleted", opts.Key)
		return nil
	})
	return err
}
