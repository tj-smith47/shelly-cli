// Package del provides the kvs delete subcommand.
package del

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds command options.
type Options struct {
	Device  string
	Key     string
	Yes     bool // Skip confirmation
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
		ValidArgsFunction: completeDeviceThenKey(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

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

	return cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Deleting %q...", opts.Key), func(ctx context.Context) error {
		if err := svc.DeleteKVS(ctx, opts.Device, opts.Key); err != nil {
			return err
		}
		ios.Success("Key %q deleted", opts.Key)
		return nil
	})
}

// completeDeviceThenKey provides completion for device and key arguments.
func completeDeviceThenKey() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: complete device names
			devices := config.ListDevices()
			var completions []string
			for name := range devices {
				completions = append(completions, name)
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		// Second argument: key (no completion - would require device query)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
