// Package set provides the auth set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Password string
	Realm    string
	User     string
}

// NewCommand creates the auth set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		User:    "admin",
	}

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
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.User, "user", "admin", "Username for authentication")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Password for authentication (required)")
	cmd.Flags().StringVar(&opts.Realm, "realm", "", "Authentication realm (optional)")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	if opts.Password == "" {
		return fmt.Errorf("--password is required")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Setting authentication...", func(ctx context.Context) error {
		if err := svc.SetAuth(ctx, opts.Device, opts.User, opts.Realm, opts.Password); err != nil {
			return fmt.Errorf("failed to set authentication: %w", err)
		}
		ios.Success("Authentication enabled on %s", opts.Device)
		ios.Printf("  User: %s\n", opts.User)
		return nil
	})
}
