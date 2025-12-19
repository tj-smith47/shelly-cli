// Package authstatus provides the cloud auth-status subcommand.
package authstatus

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the cloud auth-status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth-status",
		Aliases: []string{"whoami"},
		Short:   "Show cloud authentication status",
		Long: `Show the current Shelly Cloud authentication status.

Displays whether you're logged in, your email, and token validity.`,
		Example: `  # Check authentication status
  shelly cloud auth-status

  # Also available as whoami
  shelly cloud whoami`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cfg := config.Get()

	ios.Title("Cloud Authentication Status")
	ios.Println()

	// Check if logged in
	if cfg.Cloud.AccessToken == "" {
		ios.Printf("  Status: %s\n", output.RenderLoggedInState(false))
		ios.Println()
		ios.Info("Use 'shelly cloud login' to authenticate")
		return nil
	}

	// Show logged in status
	ios.Printf("  Status: %s\n", output.RenderLoggedInState(true))

	if cfg.Cloud.Email != "" {
		ios.Printf("  Email:  %s\n", cfg.Cloud.Email)
	}

	if cfg.Cloud.ServerURL != "" {
		ios.Printf("  Server: %s\n", cfg.Cloud.ServerURL)
	}

	// Check token validity and expiration
	tokenStatus := term.GetTokenStatus(cfg.Cloud.AccessToken)
	ios.Printf("  Token:  %s\n", tokenStatus.Display)

	if tokenStatus.Warning != "" {
		ios.Println()
		ios.Warning("%s", tokenStatus.Warning)
		return nil
	}

	// Show time until expiry
	remaining := shelly.TimeUntilExpiry(cfg.Cloud.AccessToken)
	if remaining > 0 {
		ios.Printf("  Expiry: %s\n", output.FormatDuration(remaining))
	}

	return nil
}
