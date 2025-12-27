// Package set provides the kvs set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
)

// Options holds command options.
type Options struct {
	Device  string
	Key     string
	Value   string
	Null    bool // Set null value
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device> <key> <value>",
		Aliases: []string{"s", "write", "put"},
		Short:   "Set a KVS value",
		Long: `Store a value in the device's Key-Value Storage.

The value is automatically parsed:
  - "true" or "false" → boolean
  - Numbers → numeric value
  - JSON arrays/objects → parsed JSON
  - Everything else → string

Use --null to set a null value.

Limits:
  - Key length: up to 42 bytes
  - Value size: up to 256 bytes (strings)`,
		Example: `  # Set a string value
  shelly kvs set living-room my_key "my_value"

  # Set a numeric value
  shelly kvs set living-room counter 42

  # Set a boolean value
  shelly kvs set living-room enabled true

  # Set a null value
  shelly kvs set living-room cleared --null

  # Set a JSON object
  shelly kvs set living-room config '{"timeout":30}'`,
		Args:              cobra.RangeArgs(2, 3),
		ValidArgsFunction: completion.DeviceThenNoComplete(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			if len(args) > 2 {
				opts.Value = args[2]
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Null, "null", false, "Set null value")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Validate arguments
	if !opts.Null && opts.Value == "" {
		return fmt.Errorf("value is required (use --null to set null)")
	}

	// Parse the value
	var value any
	if opts.Null {
		value = nil
	} else {
		value = kvs.ParseValue(opts.Value)
	}

	return cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Setting %q...", opts.Key), func(ctx context.Context) error {
		if err := svc.SetKVS(ctx, opts.Device, opts.Key, value); err != nil {
			return err
		}
		ios.Success("Key %q set to %v", opts.Key, output.FormatJSONValue(value))
		return nil
	})
}
