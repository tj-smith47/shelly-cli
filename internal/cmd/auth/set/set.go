// Package set provides the auth set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	userFlag     string
	passwordFlag string
	realmFlag    string
)

// NewCommand creates the auth set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"password", "pw"},
		Short:   "Set authentication credentials",
		Long: `Set authentication credentials for a device.

This enables authentication if not already enabled. The username defaults
to "admin" if not specified.`,
		Example: `  # Set credentials with default username
  shelly auth set living-room --password secret

  # Set credentials with custom username
  shelly auth set living-room --user myuser --password secret`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&userFlag, "user", "admin", "Username for authentication")
	cmd.Flags().StringVar(&passwordFlag, "password", "", "Password for authentication (required)")
	cmd.Flags().StringVar(&realmFlag, "realm", "", "Authentication realm (optional)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	if passwordFlag == "" {
		return fmt.Errorf("--password is required")
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Setting authentication...", func(ctx context.Context) error {
		if err := svc.SetAuth(ctx, device, userFlag, realmFlag, passwordFlag); err != nil {
			return fmt.Errorf("failed to set authentication: %w", err)
		}
		ios.Success("Authentication enabled on %s", device)
		ios.Printf("  User: %s\n", userFlag)
		return nil
	})
}
