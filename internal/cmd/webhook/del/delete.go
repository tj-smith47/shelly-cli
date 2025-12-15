// Package del provides the webhook delete subcommand.
package del

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

var yesFlag bool

// NewCommand creates the webhook delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <device> <webhook-id>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a webhook",
		Long: `Delete a webhook by ID.

Use 'shelly webhook list' to see webhook IDs.`,
		Example: `  # Delete webhook with ID 1
  shelly webhook delete living-room 1

  # Delete without confirmation
  shelly webhook delete living-room 1 --yes`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			webhookID, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid webhook ID: %s", args[1])
			}
			return run(cmd.Context(), f, args[0], webhookID)
		},
	}

	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, webhookID int) error {
	ios := f.IOStreams()

	// Confirm deletion
	confirmed, err := f.ConfirmAction(
		fmt.Sprintf("Delete webhook %d on %s?", webhookID, device),
		yesFlag,
	)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Warning("Cancelled")
		return nil
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Deleting webhook...", func(ctx context.Context) error {
		if err := svc.DeleteWebhook(ctx, device, webhookID); err != nil {
			return fmt.Errorf("failed to delete webhook: %w", err)
		}
		ios.Success("Webhook %d deleted from %s", webhookID, device)
		return nil
	})
}
