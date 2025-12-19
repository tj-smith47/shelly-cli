// Package mock provides the mock command for mock device mode.
package mock

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/mock/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mock/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mock/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/mock/scenario"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the mock command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mock",
		Aliases: []string{"simulate", "test"},
		Short:   "Mock device mode for testing",
		Long: `Mock device mode for testing without real hardware.

Create and manage mock devices for testing CLI commands
and automation scripts without physical Shelly devices.

Subcommands:
  create    - Create a new mock device
  list      - List mock devices
  delete    - Delete a mock device
  scenario  - Load a test scenario`,
		Example: `  # Create a mock device
  shelly mock create kitchen-light --model "Plus 1PM"

  # List mock devices
  shelly mock list

  # Load test scenario
  shelly mock scenario home-setup`,
	}

	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(scenario.NewCommand(f))

	return cmd
}
