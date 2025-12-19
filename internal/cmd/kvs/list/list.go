// Package list provides the kvs list subcommand.
package list

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Device  string
	Values  bool // Show values alongside keys
	Match   string
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls", "l"},
		Short:   "List KVS keys",
		Long: `List all keys stored in the device's Key-Value Storage (KVS).

KVS provides persistent storage on Gen2+ devices for scripts and user
data. Keys are strings and values can be strings, numbers, or booleans.
Data persists across reboots and firmware updates.

By default, only key names are listed. Use --values to also show
the stored values. Use --match for wildcard filtering.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.`,
		Example: `  # List all keys
  shelly kvs list living-room

  # List keys with values
  shelly kvs list living-room --values

  # List keys matching a pattern
  shelly kvs list living-room --match "sensor_*"

  # Output as JSON for scripting
  shelly kvs list living-room -o json

  # Get all values as JSON
  shelly kvs list living-room --values -o json

  # Export all KVS data to backup file
  shelly kvs list living-room --values -o json > kvs-backup.json

  # Find string-type keys only
  shelly kvs list living-room --values -o json | jq '.[] | select(.type == "string")'

  # Short form
  shelly kvs ls living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Values, "values", false, "Show values alongside keys")
	cmd.Flags().StringVarP(&opts.Match, "match", "m", "", "Pattern to match keys (supports * and ? wildcards)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// If values requested or pattern match, use GetMany
	if opts.Values || opts.Match != "" {
		match := opts.Match
		if match == "" {
			match = "*"
		}

		return cmdutil.RunList(ctx, ios, svc, opts.Device,
			"Getting KVS data...",
			"No keys found",
			func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.KVSItem, error) {
				return svc.GetManyKVS(ctx, device, match)
			},
			term.DisplayKVSItems)
	}

	// Default: just list keys
	return cmdutil.RunDeviceStatus(ctx, ios, svc, opts.Device,
		"Getting KVS keys...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.KVSListResult, error) {
			return svc.ListKVS(ctx, device)
		},
		term.DisplayKVSKeys)
}
