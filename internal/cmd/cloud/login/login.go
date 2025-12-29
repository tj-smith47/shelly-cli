// Package login provides the cloud login subcommand.
package login

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/network"
)

// Options holds command options.
type Options struct {
	Factory  *cmdutil.Factory
	Email    string
	Password string
}

// NewCommand creates the cloud login command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "login",
		Aliases: []string{"auth", "signin"},
		Short:   "Authenticate with Shelly Cloud",
		Long: `Authenticate with the Shelly Cloud API.

This command authenticates you with the Shelly Cloud service using your
email and password. The access token is stored locally for future use.

You can provide credentials via:
  1. Command flags (--email, --password)
  2. Interactive prompts (if TTY available)
  3. Environment variables (SHELLY_CLOUD_EMAIL, SHELLY_CLOUD_PASSWORD)`,
		Example: `  # Interactive login
  shelly cloud login

  # Login with flags
  shelly cloud login --email user@example.com --password mypassword

  # Login with environment variables
  SHELLY_CLOUD_EMAIL=user@example.com SHELLY_CLOUD_PASSWORD=mypassword shelly cloud login`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Email, "email", "", "Shelly Cloud email")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Shelly Cloud password")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()

	// Get email: flag > env var > interactive prompt
	email := opts.Email
	if email == "" {
		email = os.Getenv("SHELLY_CLOUD_EMAIL")
	}
	if email == "" {
		if !ios.CanPrompt() {
			return errors.New("email required (use --email flag or SHELLY_CLOUD_EMAIL env var)")
		}
		var err error
		email, err = iostreams.InputRequired("Email")
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
	}

	// Get password: flag > env var > interactive prompt
	password := opts.Password
	if password == "" {
		password = os.Getenv("SHELLY_CLOUD_PASSWORD")
	}
	if password == "" {
		if !ios.CanPrompt() {
			return errors.New("password required (use --password flag or SHELLY_CLOUD_PASSWORD env var)")
		}
		var err error
		password, err = iostreams.Password("Password")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	if email == "" || password == "" {
		return errors.New("email and password are required")
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Authenticating with Shelly Cloud...", func(ctx context.Context) error {
		_, result, err := network.NewCloudClientWithCredentials(ctx, email, password)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Save credentials to config
		cfg := config.Get()
		cfg.Cloud.Enabled = true
		cfg.Cloud.Email = email
		cfg.Cloud.AccessToken = result.Token
		cfg.Cloud.ServerURL = result.UserAPIURL

		if err := config.Save(); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		ios.Success("Logged in as %s", email)
		if !result.Expiry.IsZero() {
			ios.Info("Token expires: %s", result.Expiry.Format("2006-01-02 15:04:05"))
		}
		return nil
	})
}
