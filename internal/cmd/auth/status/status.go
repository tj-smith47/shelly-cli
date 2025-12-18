// Package status provides the auth status subcommand.
package status

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the auth status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st"},
		Short:   "Show authentication status",
		Long: `Show the authentication status for a device.

Displays whether authentication is enabled for the device.`,
		Example: `  # Show auth status
  shelly auth status living-room

  # Output as JSON
  shelly auth status living-room -o json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Getting authentication status...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.AuthStatus, error) {
			return svc.GetAuthStatus(ctx, device)
		},
		cmdutil.DisplayAuthStatus)
}
