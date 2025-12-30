// Package webhook provides webhook configuration commands.
package webhook

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook/server"
	"github.com/tj-smith47/shelly-cli/internal/cmd/webhook/update"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the webhook command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhook",
		Aliases: []string{"wh", "hook"},
		Short:   "Manage device webhooks",
		Long: `Manage webhooks for devices.

Webhooks allow the device to send HTTP requests when events occur,
enabling integration with external services and automation systems.`,
		Example: `  # List all webhooks
  shelly webhook list living-room

  # Create a webhook
  shelly webhook create living-room --event "switch.on" --url "http://example.com/hook"

  # Delete a webhook
  shelly webhook delete living-room 1

  # Update a webhook
  shelly webhook update living-room 1 --disable`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(update.NewCommand(f))
	cmd.AddCommand(server.NewCommand(f))

	return cmd
}
