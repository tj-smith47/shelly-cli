// Package list provides the webhook list subcommand.
package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the webhook list command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <device>",
		Aliases: []string{"ls"},
		Short:   "List webhooks",
		Long: `List all configured webhooks for a device.

Webhooks are HTTP callbacks triggered by device events (button press,
switch state change, etc.). Each webhook has an event type, target URL(s),
and can be enabled or disabled.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: ID, Event, URLs (truncated if long), Enabled (Yes/No)`,
		Example: `  # List all webhooks
  shelly webhook list living-room

  # Output as JSON for scripting
  shelly webhook list living-room -o json

  # Get enabled webhooks only
  shelly webhook list living-room -o json | jq '.[] | select(.enable == true)'

  # Extract just webhook URLs
  shelly webhook list living-room -o json | jq -r '.[].urls[]'

  # Find webhooks for specific event
  shelly webhook list living-room -o json | jq '.[] | select(.event | contains("switch"))'

  # Count webhooks by event type
  shelly webhook list living-room -o json | jq 'group_by(.event) | map({event: .[0].event, count: length})'

  # Short form
  shelly webhook ls living-room`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunList(ctx, ios, svc, device,
		"Getting webhooks...",
		"No webhooks configured",
		func(ctx context.Context, svc *shelly.Service, device string) ([]shelly.WebhookInfo, error) {
			return svc.ListWebhooks(ctx, device)
		},
		displayWebhooks)
}

func displayWebhooks(ios *iostreams.IOStreams, webhooks []shelly.WebhookInfo) {
	ios.Title("Webhooks")
	ios.Println()

	table := output.NewTable("ID", "Event", "URLs", "Enabled")
	for _, w := range webhooks {
		urls := strings.Join(w.URLs, ", ")
		if len(urls) > 40 {
			urls = urls[:37] + "..."
		}
		var enabled string
		if w.Enable {
			enabled = theme.StatusOK().Render("Yes")
		} else {
			enabled = theme.StatusError().Render("No")
		}
		table.AddRow(fmt.Sprintf("%d", w.ID), w.Event, urls, enabled)
	}
	table.Print()

	ios.Printf("\n%d webhook(s) configured\n", len(webhooks))
}
