// Package set provides the auth set subcommand.
package set

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// flagPassword is the name of the password flag and alias.
const flagPassword = "password"

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Password string
	// PasswordStdin reads Password from stdin.
	PasswordStdin bool
	User          string
}

// NewCommand creates the auth set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		User:    auth.DefaultUser,
	}

	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{flagPassword, "pw"},
		Short:   "Set authentication credentials",
		Long: `Set the password a device requires, turning authentication on.

Gen1 devices accept any username; without --user they get the user stored
for them, or admin. Gen2+ devices have a single user, admin, so --user can
only be left out or set to admin.

Once the device has the new password, every request to it needs that
password, so the new credentials are saved for a registered device. The
command then makes an authenticated request with the new password and fails
if the device does not accept it.`,
		Example: `  # Set the password (user admin)
  shelly auth set living-room --password secret

  # Read the password from stdin, keeping it out of shell history
  shelly auth set living-room --password-stdin < ~/.shelly-living-room-password

  # Set a custom username (Gen1 devices only)
  shelly auth set garage-gen1 --user myuser --password secret`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.User, "user", "", "Username: any name on Gen1, where it defaults to the stored user or admin; Gen2+ devices allow only admin")
	cmdutil.AddSecretFlags(cmd, &opts.Password, &opts.PasswordStdin, flagPassword,
		"New device password (or use --password-stdin)", "Read the new device password from stdin")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	if err := cmdutil.ResolveSecret(ios, &opts.Password, opts.PasswordStdin, flagPassword, "New password"); err != nil {
		return err
	}
	if opts.Password == "" {
		return errors.New("a password is required: give it with --password <value>, or read it from stdin with --password-stdin")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Setting authentication...", func(ctx context.Context) error {
		user, stored, err := svc.SetAuth(ctx, opts.Device, opts.User, opts.Password)
		if err != nil {
			return fmt.Errorf("failed to set authentication: %w", err)
		}
		ios.Success("Authentication enabled on %s", opts.Device)
		ios.Printf("  User: %s\n", user)
		term.DisplayStoredCredentials(ios, opts.Device, stored)
		return nil
	})
}
