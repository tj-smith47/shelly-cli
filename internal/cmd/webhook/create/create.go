// Package create provides the webhook create subcommand.
package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Cid     int
	Device  string
	Disable bool
	Event   string
	Name    string
	URLs    []string
}

// NewCommand creates the webhook create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Event, "event", "", "Event type (e.g., switch.on, input.toggle)")
	cmd.Flags().StringArrayVar(&opts.URLs, "url", nil, "Webhook URL (can be specified multiple times)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Webhook name (optional)")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Create webhook in disabled state")
	cmd.Flags().IntVar(&opts.Cid, "cid", 0, "Component ID (default: 0)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// Validate required flags
	if opts.Event == "" {
		return fmt.Errorf("--event is required")
	}
	if len(opts.URLs) == 0 {
		return fmt.Errorf("--url is required (at least one URL)")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	err := cmdutil.RunWithSpinner(ctx, ios, "Creating webhook...", func(ctx context.Context) error {
		params := shelly.CreateWebhookParams{
			Event:  opts.Event,
			URLs:   opts.URLs,
			Name:   opts.Name,
			Enable: !opts.Disable,
			Cid:    opts.Cid,
		}

		id, createErr := svc.CreateWebhook(ctx, opts.Device, params)
		if createErr != nil {
			return fmt.Errorf("failed to create webhook: %w", createErr)
		}

		ios.Success("Webhook created on %s", opts.Device)
		ios.Printf("  ID:    %d\n", id)
		ios.Printf("  Event: %s\n", opts.Event)
		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cached webhook list
	cmdutil.InvalidateCache(opts.Factory, opts.Device, cache.TypeWebhooks)
	return nil
}
