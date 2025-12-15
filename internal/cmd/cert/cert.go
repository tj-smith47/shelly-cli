// Package cert provides certificate management commands.
package cert

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cert/install"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cert/show"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the cert command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cert",
		Aliases: []string{"certificate", "tls", "ssl"},
		Short:   "Manage device TLS certificates",
		Long: `Manage TLS certificates for Gen2+ Shelly devices.

Devices support custom CA certificates for secure MQTT and cloud connections.
Use these commands to view or install certificates on devices.`,
		Example: `  # Show TLS configuration
  shelly cert show kitchen

  # Install a CA certificate
  shelly cert install kitchen --ca /path/to/ca.pem`,
	}

	cmd.AddCommand(show.NewCommand(f))
	cmd.AddCommand(install.NewCommand(f))

	return cmd
}
