// Package authstatus provides the cloud auth-status subcommand.
package authstatus

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cloud auth-status command.
func NewCommand() *cobra.Command {
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
			return run()
		},
	}

	return cmd
}

func run() error {
	ios := iostreams.System()

	cfg := config.Get()

	ios.Title("Cloud Authentication Status")
	ios.Println()

	// Check if logged in
	if cfg.Cloud.AccessToken == "" {
		ios.Printf("  Status: %s\n", theme.StatusError().Render("Not logged in"))
		ios.Println()
		ios.Info("Use 'shelly cloud login' to authenticate")
		return nil
	}

	// Show logged in status
	ios.Printf("  Status: %s\n", theme.StatusOK().Render("Logged in"))

	if cfg.Cloud.Email != "" {
		ios.Printf("  Email:  %s\n", cfg.Cloud.Email)
	}

	if cfg.Cloud.ServerURL != "" {
		ios.Printf("  Server: %s\n", cfg.Cloud.ServerURL)
	}

	// Check token validity and expiration
	tokenStatus := getTokenStatus(cfg.Cloud.AccessToken)
	ios.Printf("  Token:  %s\n", tokenStatus.display)

	if tokenStatus.warning != "" {
		ios.Println()
		ios.Warning("%s", tokenStatus.warning)
		return nil
	}

	// Show time until expiry
	remaining := shelly.TimeUntilExpiry(cfg.Cloud.AccessToken)
	if remaining > 0 {
		ios.Printf("  Expiry: %s\n", formatDuration(remaining))
	}

	return nil
}

type tokenStatusInfo struct {
	display string
	warning string
}

func getTokenStatus(token string) tokenStatusInfo {
	if err := shelly.ValidateToken(token); err != nil {
		return tokenStatusInfo{
			display: theme.StatusError().Render("Invalid"),
			warning: "Token is invalid. Please run 'shelly cloud login' to re-authenticate.",
		}
	}

	if shelly.IsTokenExpired(token) {
		return tokenStatusInfo{
			display: theme.StatusError().Render("Expired"),
			warning: "Token has expired. Please run 'shelly cloud login' to re-authenticate.",
		}
	}

	return tokenStatusInfo{
		display: theme.StatusOK().Render("Valid"),
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
