// Package export provides data export commands.
package export

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/export/ansible"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/csv"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/jsoncmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/terraform"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/yamlcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the export command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export device data in various formats",
		Long: `Export device configuration and status data in various formats.

Supports exporting to JSON, YAML, CSV, Ansible inventory, and Terraform
configuration formats. Useful for backup, documentation, and infrastructure
as code workflows.`,
		Example: `  # Export device config as JSON
  shelly export json living-room

  # Export device config as YAML to file
  shelly export yaml living-room device.yaml

  # Export device list as CSV
  shelly export csv living-room bedroom kitchen devices.csv

  # Export as Ansible inventory
  shelly export ansible @all inventory.yaml

  # Export as Terraform config
  shelly export terraform @all shelly.tf`,
	}

	cmd.AddCommand(jsoncmd.NewCommand(f))
	cmd.AddCommand(yamlcmd.NewCommand(f))
	cmd.AddCommand(csv.NewCommand(f))
	cmd.AddCommand(ansible.NewCommand(f))
	cmd.AddCommand(terraform.NewCommand(f))

	return cmd
}
