// Package plugin provides plugin management commands.
package plugin

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/exec"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/install"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/plugin/upgrade"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the plugin command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"extension", "ext"},
		Short:   "Manage CLI plugins",
		Long: `Manage CLI plugins (extensions).

Plugins are executable programs named shelly-<name> that extend the CLI.
They can be installed from local files, GitHub repositories, or URLs.

Installed plugins are stored in ~/.config/shelly/plugins/.`,
		Example: `  # List installed plugins
  shelly plugin list

  # Install from local file
  shelly plugin install ./shelly-myext

  # Install from GitHub
  shelly plugin install gh:user/shelly-myext

  # Remove a plugin
  shelly plugin remove myext

  # Run a plugin explicitly
  shelly plugin exec myext --some-flag

  # Create a new plugin scaffold
  shelly plugin create myext`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(install.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))
	cmd.AddCommand(upgrade.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(exec.NewCommand(f))

	return cmd
}
