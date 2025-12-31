// Package alias provides the device alias management subcommand.
package alias

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
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
		return listAliases(ios, opts.Device)
	}

	// Handle --remove flag
	if opts.Remove != "" {
		return removeAlias(ios, opts.Device, opts.Remove)
	}

	// Add alias - requires the alias argument
	if opts.Alias == "" {
		return fmt.Errorf("alias argument required (or use --list to view aliases)")
	}

	return addAlias(ios, opts.Device, opts.Alias)
}

func listAliases(ios *iostreams.IOStreams, deviceName string) error {
	aliases, err := config.GetDeviceAliases(deviceName)
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		ios.Info("No aliases defined for %s", deviceName)
		return nil
	}

	if output.WantsJSON() {
		return output.PrintJSON(map[string]any{
			"device":  deviceName,
			"aliases": aliases,
		})
	}
	if output.WantsYAML() {
		return output.PrintYAML(map[string]any{
			"device":  deviceName,
			"aliases": aliases,
		})
	}

	ios.Printf("Aliases for %s: %s\n", deviceName, strings.Join(aliases, ", "))
	return nil
}

func removeAlias(ios *iostreams.IOStreams, deviceName, alias string) error {
	if err := config.RemoveDeviceAlias(deviceName, alias); err != nil {
		return err
	}
	ios.Success("Removed alias %q from %s", alias, deviceName)
	return nil
}

func addAlias(ios *iostreams.IOStreams, deviceName, alias string) error {
	// Validate alias format
	if err := config.ValidateDeviceAlias(alias); err != nil {
		return fmt.Errorf("invalid alias: %w", err)
	}

	// Check for conflicts
	if err := config.CheckAliasConflict(alias, deviceName); err != nil {
		return err
	}

	// Add the alias
	if err := config.AddDeviceAlias(deviceName, alias); err != nil {
		return err
	}

	ios.Success("Added alias %q to %s", alias, deviceName)
	return nil
}
