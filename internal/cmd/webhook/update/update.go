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

// Options holds the command options.
type Options struct {
	Factory   *cmdutil.Factory
	Device    string
	Disable   bool
	Enable    bool
	Event     string
	Name      string
	URLs      []string
	WebhookID int
}

// NewCommand creates the webhook update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			opts.Device = args[0]
			opts.WebhookID = webhookID
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Event, "event", "", "Event type")
	cmd.Flags().StringArrayVar(&opts.URLs, "url", nil, "Webhook URL (replaces all URLs)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Webhook name")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable webhook")
	cmd.Flags().BoolVar(&opts.Disable, "disable", false, "Disable webhook")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// Validate - need at least one option
	if opts.Event == "" && len(opts.URLs) == 0 && opts.Name == "" && !opts.Enable && !opts.Disable {
		return fmt.Errorf("specify at least one option to update (--event, --url, --name, --enable, --disable)")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Build update params
	params := shelly.UpdateWebhookParams{
		Event: opts.Event,
		URLs:  opts.URLs,
		Name:  opts.Name,
	}
	if opts.Enable {
		t := true
		params.Enable = &t
	} else if opts.Disable {
		f := false
		params.Enable = &f
	}

	err := cmdutil.RunWithSpinner(ctx, ios, "Updating webhook...", func(ctx context.Context) error {
		if updateErr := svc.UpdateWebhook(ctx, opts.Device, opts.WebhookID, params); updateErr != nil {
			return fmt.Errorf("failed to update webhook: %w", updateErr)
		}
		ios.Success("Webhook %d updated on %s", opts.WebhookID, opts.Device)
		return nil
	})
	return err
}
