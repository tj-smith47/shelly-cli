// Package template provides configuration template commands.
package template

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/template/apply"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/diff"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/template/list"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the template command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"tpl", "templates"},
		Short:   "Manage device configuration templates",
		Long: `Manage configuration templates for Shelly devices.

Templates capture device configuration that can be applied to other
devices of the same model. This enables consistent configuration
across multiple devices and easy provisioning.

Templates store:
  - Device configuration (WiFi excluded by default)
  - Component settings (switches, lights, covers, etc.)
  - Schedules, scripts, and webhooks
  - Network settings (optional)`,
		Example: `  # List all templates
  shelly template list

  # Create a template from a device
  shelly template create my-config living-room

  # Apply a template to a device
  shelly template apply my-config bedroom

  # Preview changes without applying
  shelly template diff my-config bedroom

  # Export template to file
  shelly template export my-config template.yaml

  # Import template from file
  shelly template import template.yaml`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(apply.NewCommand(f))
	cmd.AddCommand(diff.NewCommand(f))
	cmd.AddCommand(export.NewCommand(f))
	cmd.AddCommand(importcmd.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))

	return cmd
}
