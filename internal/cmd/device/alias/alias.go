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

// NewCommand creates the device alias command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var remove string
	var list bool

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
			return run(f, args, remove, list)
		},
	}

	cmd.Flags().StringVarP(&remove, "remove", "r", "", "Remove the specified alias")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List all aliases for the device")

	return cmd
}

func run(f *cmdutil.Factory, args []string, remove string, list bool) error {
	ios := f.IOStreams()
	deviceName := args[0]

	// Verify device exists
	if _, exists := config.GetDevice(deviceName); !exists {
		return fmt.Errorf("device %q not found", deviceName)
	}

	// Handle --list flag
	if list {
		return listAliases(ios, deviceName)
	}

	// Handle --remove flag
	if remove != "" {
		return removeAlias(ios, deviceName, remove)
	}

	// Add alias - requires the alias argument
	if len(args) < 2 {
		return fmt.Errorf("alias argument required (or use --list to view aliases)")
	}

	return addAlias(ios, deviceName, args[1])
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
