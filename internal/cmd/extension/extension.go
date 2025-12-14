// Package extension provides extension management commands.
package extension

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/exec"
	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/install"
	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/remove"
	"github.com/tj-smith47/shelly-cli/internal/cmd/extension/upgrade"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the extension command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Aliases: []string{"ext", "plugin"},
		Short:   "Manage CLI extensions",
		Long: `Manage CLI extensions (plugins).

Extensions are executable programs named shelly-<name> that extend the CLI.
They can be installed from local files, GitHub repositories, or URLs.

Installed extensions are stored in ~/.config/shelly/plugins/.`,
		Example: `  # List installed extensions
  shelly extension list

  # Install from local file
  shelly extension install ./shelly-myext

  # Install from GitHub
  shelly extension install gh:user/shelly-myext

  # Remove an extension
  shelly extension remove myext

  # Run an extension explicitly
  shelly extension exec myext --some-flag

  # Create a new extension scaffold
  shelly extension create myext`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(install.NewCommand(f))
	cmd.AddCommand(remove.NewCommand(f))
	cmd.AddCommand(upgrade.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(exec.NewCommand(f))

	return cmd
}
