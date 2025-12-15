// Package update provides the webhook update subcommand.
package update

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	eventFlag   string
	urlFlag     []string
	nameFlag    string
	enableFlag  bool
	disableFlag bool
)

// NewCommand creates the webhook update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <device> <webhook-id>",
		Aliases: []string{"edit", "modify"},
		Short:   "Update a webhook",
		Long: `Update an existing webhook.

Only specified fields will be updated. Use --enable or --disable to
change the webhook's active state.`,
		Example: `  # Change webhook URL
  shelly webhook update living-room 1 --url "http://new-url.com"

  # Disable a webhook
  shelly webhook update living-room 1 --disable

  # Enable and change event
  shelly webhook update living-room 1 --enable --event "switch.off"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[1])
			}
			return run(cmd.Context(), f, args[0], webhookID)
		},
	}

	cmd.Flags().StringVar(&eventFlag, "event", "", "Event type")
	cmd.Flags().StringArrayVar(&urlFlag, "url", nil, "Webhook URL (replaces all URLs)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Webhook name")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable webhook")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Disable webhook")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, webhookID int) error {
	// Validate - need at least one option
	if eventFlag == "" && len(urlFlag) == 0 && nameFlag == "" && !enableFlag && !disableFlag {
		return fmt.Errorf("specify at least one option to update (--event, --url, --name, --enable, --disable)")
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Build update params
	params := shelly.UpdateWebhookParams{
		Event: eventFlag,
		URLs:  urlFlag,
		Name:  nameFlag,
	}
	if enableFlag {
		t := true
		params.Enable = &t
	} else if disableFlag {
		f := false
		params.Enable = &f
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Updating webhook...", func(ctx context.Context) error {
		if err := svc.UpdateWebhook(ctx, device, webhookID, params); err != nil {
			return fmt.Errorf("failed to update webhook: %w", err)
		}
		ios.Success("Webhook %d updated on %s", webhookID, device)
		return nil
	})
}
