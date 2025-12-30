// Package rotate provides the auth rotate subcommand.
package rotate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly/auth"
)

// Options holds the command options.
type Options struct {
	User       string
	Password   string
	Length     int
	Generate   bool
	ShowSecret bool
}

// NewCommand creates the auth rotate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		User:   "admin",
		Length: auth.DefaultPasswordLength,
	}

	cmd := &cobra.Command{
		Use:     "rotate <device>",
		Aliases: []string{"renew", "refresh"},
		Short:   "Rotate device credentials",
		Long: `Rotate device authentication credentials.

This command sets new authentication credentials on the device,
optionally generating a secure random password.

For security best practices:
  - Rotate credentials periodically
  - Use generated passwords (--generate)
  - Store credentials securely`,
		Example: `  # Rotate with a new password
  shelly auth rotate living-room --password newSecret123

  # Generate a random password
  shelly auth rotate living-room --generate

  # Generate and show the new password
  shelly auth rotate living-room --generate --show

  # Use specific password length
  shelly auth rotate living-room --generate --length 24`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.User, "user", "admin", "Username for authentication")
	cmd.Flags().StringVar(&opts.Password, "password", "", "New password (or use --generate)")
	cmd.Flags().IntVar(&opts.Length, "length", auth.DefaultPasswordLength, "Generated password length")
	cmd.Flags().BoolVar(&opts.Generate, "generate", false, "Generate a random password")
	cmd.Flags().BoolVar(&opts.ShowSecret, "show", false, "Show the new password in output")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Determine password
	password := opts.Password
	if opts.Generate {
		var err error
		password, err = auth.GeneratePassword(opts.Length)
		if err != nil {
			return fmt.Errorf("failed to generate password: %w", err)
		}
	}

	if password == "" {
		return fmt.Errorf("--password or --generate is required")
	}

	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	return cmdutil.RunWithSpinner(ctx, ios, "Rotating credentials...", func(ctx context.Context) error {
		if err := svc.SetAuth(ctx, device, opts.User, "", password); err != nil {
			return fmt.Errorf("failed to rotate credentials: %w", err)
		}

		ios.Success("Credentials rotated on %s", device)
		ios.Printf("  User: %s\n", opts.User)

		if opts.Generate {
			if opts.ShowSecret {
				ios.Printf("  Password: %s\n", password)
			} else {
				ios.Info("Password generated (use --show to display)")
			}
		}

		ios.Println("")
		ios.Warning("Update your stored credentials to match the new values")

		return nil
	})
}
