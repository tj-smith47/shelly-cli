// Package export provides data export commands.
package export

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/export/ansible"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/csv"
	"github.com/tj-smith47/shelly-cli/internal/cmd/export/terraform"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the export command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "export",
		Aliases: []string{"exp"},
		Short:   "Export fleet data for infrastructure tools",
		Long: `Export device fleet data for infrastructure-as-code tools.

Supports exporting to CSV, Ansible inventory, and Terraform configuration
formats. Useful for documentation and fleet management workflows.

For single-device configuration export (JSON/YAML), use:
  shelly device config export <device> <file> [--format json|yaml]`,
		Example: `  # Export device list as CSV
  shelly export csv living-room bedroom kitchen devices.csv

  # Export as Ansible inventory
  shelly export ansible @all inventory.yaml

  # Export as Terraform config
  shelly export terraform @all shelly.tf`,
	}

	cmd.AddCommand(ansible.NewCommand(f))
	cmd.AddCommand(csv.NewCommand(f))
	cmd.AddCommand(terraform.NewCommand(f))

	return cmd
}
