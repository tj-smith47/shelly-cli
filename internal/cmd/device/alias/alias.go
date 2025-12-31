// Package alias provides the device alias management subcommand.
package alias

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Alias   string
	Device  string
	List    bool
	Remove  string
}

// NewCommand creates the device alias command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "alias <device> [alias]",
		Aliases: []string{"aliases"},
		Short:   "Manage device aliases",
		Long: `Manage short aliases for devices.

Aliases allow you to reference devices by short, memorable names
in addition to their full registered names. For example, you can
create an alias "mb" for a device named "master-bathroom".

Aliases must be 1-32 characters, start with a letter or number,
and contain only letters, numbers, hyphens, and underscores.`,
		Example: `  # Add an alias to a device
  shelly device alias master-bathroom mb
  shelly device alias kitchen-light k

  # List aliases for a device
  shelly device alias master-bathroom --list

  # Remove an alias from a device
  shelly device alias master-bathroom --remove mb

  # Use an alias in commands
  shelly toggle mb
  shelly status k`,
		Args:              cobra.RangeArgs(1, 2),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			if len(args) > 1 {
				opts.Alias = args[1]
			}
			return run(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Remove, "remove", "r", "", "Remove the specified alias")
	cmd.Flags().BoolVarP(&opts.List, "list", "l", false, "List all aliases for the device")

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Verify device exists
	if _, exists := config.GetDevice(opts.Device); !exists {
		return fmt.Errorf("device %q not found", opts.Device)
	}

	// Handle --list flag
	if opts.List {
		aliases, err := config.GetDeviceAliases(opts.Device)
		if err != nil {
			return err
		}
		term.DisplayDeviceAliases(ios, opts.Device, aliases)
		return nil
	}

	// Handle --remove flag
	if opts.Remove != "" {
		if err := config.RemoveDeviceAlias(opts.Device, opts.Remove); err != nil {
			return err
		}
		term.DisplayAliasRemoved(ios, opts.Device, opts.Remove)
		return nil
	}

	// Add alias - requires the alias argument
	if opts.Alias == "" {
		return fmt.Errorf("alias argument required (or use --list to view aliases)")
	}

	// Validate alias format
	if err := config.ValidateDeviceAlias(opts.Alias); err != nil {
		return fmt.Errorf("invalid alias: %w", err)
	}

	// Check for conflicts
	if err := config.CheckAliasConflict(opts.Alias, opts.Device); err != nil {
		return err
	}

	// Add the alias
	if err := config.AddDeviceAlias(opts.Device, opts.Alias); err != nil {
		return err
	}

	term.DisplayAliasAdded(ios, opts.Device, opts.Alias)
	return nil
}
