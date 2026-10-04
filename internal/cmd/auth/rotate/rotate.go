// Package rotate provides the auth rotate subcommand.
package rotate

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

// Options holds the command options.
type Options struct {
	Factory  *cmdutil.Factory
	Device   string
	Generate bool
	Length   int
	Password string
	// PasswordStdin reads Password from stdin.
	PasswordStdin bool
	ShowSecret    bool
	User          string
}

// NewCommand creates the auth rotate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Length:  auth.DefaultPasswordLength,
	}

	cmd := &cobra.Command{
		Use:     "rotate <device>",
		Aliases: []string{"renew", "refresh"},
		Short:   "Rotate device credentials",
		Long: `Rotate device authentication credentials.

This command sets a new password on the device, optionally generating a
secure random one. Gen1 devices accept any username, and without --user keep
the user stored for them; Gen2+ devices have a single user, admin.

The new credentials are saved for a registered device, and the command then
makes an authenticated request with the new password and fails if the
device does not accept it. A generated password is shown only with --show.`,
		Example: `  # Rotate with a new password
  shelly auth rotate living-room --password newSecret123

  # Read the new password from stdin, keeping it out of shell history
  shelly auth rotate living-room --password-stdin < ~/.shelly-living-room-password

  # Generate a random password
  shelly auth rotate living-room --generate

  # Generate and show the new password
  shelly auth rotate living-room --generate --show

  # Use specific password length
  shelly auth rotate living-room --generate --length 24`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.User, "user", "", "Username: any name on Gen1, where it defaults to the stored user or admin; Gen2+ devices allow only admin")
	cmdutil.AddSecretFlags(cmd, &opts.Password, &opts.PasswordStdin, "password",
		"New password (or use --password-stdin or --generate)", "Read the new password from stdin")
	cmd.Flags().IntVar(&opts.Length, "length", auth.DefaultPasswordLength, "Generated password length")
	cmd.Flags().BoolVar(&opts.Generate, "generate", false, "Generate a random password")
	cmd.Flags().BoolVar(&opts.ShowSecret, "show", false, "Show the new password in output")

	cmd.MarkFlagsMutuallyExclusive("password", "generate")
	cmd.MarkFlagsMutuallyExclusive("password-stdin", "generate")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	password := opts.Password
	if err := cmdutil.ResolveSecret(ios, &password, opts.PasswordStdin, "password", "New password"); err != nil {
		return err
	}
	if opts.Generate {
		var err error
		password, err = auth.GeneratePassword(opts.Length)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	}

	if password == "" {
		return errors.New("a new password is required: give it with --password <value>, read it from stdin with --password-stdin, or generate one with --generate")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunWithSpinner(ctx, ios, "Rotating credentials...", func(ctx context.Context) error {
		user, stored, err := svc.SetAuth(ctx, opts.Device, opts.User, password)
		if err != nil {
			return fmt.Errorf("failed to rotate credentials: %w", err)
		}

		ios.Success("Credentials rotated on %s", opts.Device)
		ios.Printf("  User: %s\n", user)

		if opts.Generate {
			if opts.ShowSecret {
				ios.Printf("  Password: %s\n", password)
			} else {
				ios.Info("Password generated (use --show to display)")
			}
		}

		term.DisplayStoredCredentials(ios, opts.Device, stored)

		return nil
	})
}
