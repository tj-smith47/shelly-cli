// Package create provides the webhook create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	eventFlag   string
	urlFlag     []string
	nameFlag    string
	disableFlag bool
	cidFlag     int
)

// NewCommand creates the webhook create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <device>",
		Aliases: []string{"add", "new"},
		Short:   "Create a webhook",
		Long: `Create a new webhook for a device.

Webhooks are triggered by events and send HTTP requests to the specified URLs.
Common events include "switch.on", "switch.off", "input.toggle", etc.`,
		Example: `  # Create a webhook for switch on event
  shelly webhook create living-room --event "switch.on" --url "http://example.com/hook"

  # Create with multiple URLs
  shelly webhook create living-room --event "switch.off" --url "http://a.com" --url "http://b.com"

  # Create disabled webhook
  shelly webhook create living-room --event "input.toggle" --url "http://example.com" --disable`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&eventFlag, "event", "", "Event type (e.g., switch.on, input.toggle)")
	cmd.Flags().StringArrayVar(&urlFlag, "url", nil, "Webhook URL (can be specified multiple times)")
	cmd.Flags().StringVar(&nameFlag, "name", "", "Webhook name (optional)")
	cmd.Flags().BoolVar(&disableFlag, "disable", false, "Create webhook in disabled state")
	cmd.Flags().IntVar(&cidFlag, "cid", 0, "Component ID (default: 0)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	// Validate required flags
	if eventFlag == "" {
		return fmt.Errorf("--event is required")
	}
	if len(urlFlag) == 0 {
		return fmt.Errorf("--url is required (at least one URL)")
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Creating webhook...", func(ctx context.Context) error {
		params := shelly.CreateWebhookParams{
			Event:  eventFlag,
			URLs:   urlFlag,
			Name:   nameFlag,
			Enable: !disableFlag,
			Cid:    cidFlag,
		}

		id, err := svc.CreateWebhook(ctx, device, params)
		if err != nil {
			return fmt.Errorf("failed to create webhook: %w", err)
		}

		ios.Success("Webhook created on %s", device)
		ios.Printf("  ID:    %d\n", id)
		ios.Printf("  Event: %s\n", eventFlag)
		return nil
	})
}
