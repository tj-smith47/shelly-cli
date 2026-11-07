// Package clearcmd provides the cache clear command.
package clearcmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

const commandUse = "clear"

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Type    string
	Expired bool
	All     bool
	Yes     bool
}

// NewCommand creates the cache clear command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:   commandUse,
		Short: "Clear the cache",
		Long: `Clear cached device data.

Use flags to specify what to clear:
  --all      Clear all cached data
  --device   Clear cache for a specific device
  --type     Clear cache for a specific data type (requires --device)
  --expired  Clear only expired entries

Each form asks for confirmation; pass --yes to skip the prompt in scripts.`,
		Aliases: []string{"c", "rm", "clean"},
		Example: `  # Clear all cache
  shelly cache clear --all

  # Clear cache for a specific device
  shelly cache clear --device kitchen

  # Clear cache for specific device and type
  shelly cache clear --device kitchen --type firmware

  # Clear only expired entries
  shelly cache clear --expired

  # Skip the confirmation prompt (scripts)
  shelly cache clear --all --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "Clear all cached data")
	cmd.Flags().StringVarP(&opts.Device, "device", "d", "", "Clear cache for specific device")
	cmd.Flags().StringVarP(&opts.Type, "type", "t", "", "Clear cache for specific data type (requires --device)")
	cmd.Flags().BoolVarP(&opts.Expired, "expired", "e", false, "Clear only expired entries")
	flags.AddYesFlag(cmd, &opts.Yes)

	return cmd
}

//nolint:gocyclo // Sequential flag handling with early returns is clear despite complexity score
func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	fc := opts.Factory.FileCache()

	if fc == nil {
		ios.Info("Cache not available")
		return nil
	}

	// Flag validation happens before the prompt so a bad invocation never asks.
	if opts.Type != "" && opts.Device == "" {
		return fmt.Errorf("--type requires --device flag")
	}
	if !opts.Expired && opts.Device == "" && !opts.All {
		return fmt.Errorf("specify --all to clear all cache, --device for specific device, or --expired for cleanup")
	}

	var what string
	switch {
	case opts.Expired:
		what = "Clear all expired cache entries?"
	case opts.Device != "" && opts.Type != "":
		what = fmt.Sprintf("Clear the %s cache for %s?", opts.Type, opts.Device)
	case opts.Device != "":
		what = fmt.Sprintf("Clear all cached data for %s?", opts.Device)
	default:
		stats, err := fc.Stats()
		if err != nil {
			return err
		}
		if stats.TotalEntries == 0 {
			ios.Info("Cache is already empty")
			return nil
		}
		what = fmt.Sprintf("Clear all %d cache entries?", stats.TotalEntries)
	}

	confirmed, err := opts.Factory.ConfirmAction(what, opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Cancelled")
		return nil
	}

	// Handle --expired flag
	if opts.Expired {
		removed, err := fc.Cleanup()
		if err != nil {
			return err
		}
		switch removed {
		case 0:
			ios.Info("No expired entries to clean up")
		case 1:
			ios.Success("Removed 1 expired entry")
		default:
			ios.Success("Removed %d expired entries", removed)
		}
		return nil
	}

	// Handle --device + --type
	if opts.Device != "" && opts.Type != "" {
		if err := fc.Invalidate(opts.Device, opts.Type); err != nil {
			return err
		}
		ios.Success("Cleared %s cache for %s", opts.Type, opts.Device)
		return nil
	}

	// Handle --device only
	if opts.Device != "" {
		if err := fc.InvalidateDevice(opts.Device); err != nil {
			return err
		}
		ios.Success("Cleared all cache for %s", opts.Device)
		return nil
	}

	if err := fc.InvalidateAll(); err != nil {
		return err
	}

	ios.Success("Cache cleared")
	return nil
}
