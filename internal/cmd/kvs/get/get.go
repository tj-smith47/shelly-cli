// Package get provides the kvs get subcommand.
package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	Key     string
	Raw     bool // Output raw value only
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "get <device> <key>",
		Aliases: []string{"g", "read"},
		Short:   "Get a KVS value",
		Long: `Retrieve a value from the device's Key-Value Storage.

Returns the value, its type, and the etag (version identifier).
Use --raw to output only the value without formatting.`,
		Example: `  # Get a value
  shelly kvs get living-room my_key

  # Get raw value only (for scripting)
  shelly kvs get living-room my_key --raw

  # Output as JSON
  shelly kvs get living-room my_key -o json`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completion.DeviceThenNoComplete(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Raw, "raw", "r", false, "Output raw value only")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, opts.Device,
		fmt.Sprintf("Getting key %q...", opts.Key),
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.KVSGetResult, error) {
			return svc.GetKVS(ctx, device, opts.Key)
		},
		func(ios *iostreams.IOStreams, result *shelly.KVSGetResult) {
			if opts.Raw {
				cmdutil.DisplayKVSRaw(ios, result)
			} else {
				cmdutil.DisplayKVSResult(ios, opts.Key, result)
			}
		})
}
