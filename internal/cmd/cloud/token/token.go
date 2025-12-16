// Package token provides the cloud token subcommand.
package token

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cloud token command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "token",
		Aliases: []string{"tok", "key"},
		Short:   "Show or manage cloud token",
		Long: `Show the current Shelly Cloud access token.

This command displays the access token for debugging purposes.
Be careful not to share or expose your token.`,
		Example: `  # Show the current token
  shelly cloud token

  # Copy token to clipboard (Linux)
  shelly cloud token | xclip -selection clipboard`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	cfg := config.Get()

	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Parse the token to get details
	tokenInfo, err := shelly.ParseToken(cfg.Cloud.AccessToken)
	if err != nil {
		ios.Error("Token is invalid: %v", err)
		return err
	}

	ios.Title("Cloud Token")
	ios.Println()

	// Show token (truncated in TTY, full in pipe)
	if ios.IsStdoutTTY() {
		// Show truncated token in terminal
		token := cfg.Cloud.AccessToken
		if len(token) > 50 {
			ios.Printf("  Token:      %s...%s\n", token[:20], token[len(token)-10:])
		} else {
			ios.Printf("  Token:      %s\n", token)
		}
		ios.Printf("  %s\n", theme.Dim().Render("(pipe to clipboard for full token)"))
	} else {
		// Output full token when piped
		ios.Println(cfg.Cloud.AccessToken)
		return nil
	}

	ios.Println()

	// Show token details
	if tokenInfo.UserAPIURL != "" {
		ios.Printf("  Server:     %s\n", tokenInfo.UserAPIURL)
	}

	if !tokenInfo.Expiry.IsZero() {
		ios.Printf("  Expires:    %s\n", tokenInfo.Expiry.Format("2006-01-02 15:04:05"))

		remaining := shelly.TimeUntilExpiry(cfg.Cloud.AccessToken)
		if remaining > 0 {
			ios.Printf("  Remaining:  %s\n", formatDuration(remaining))
		} else {
			ios.Printf("  Status:     %s\n", output.RenderTokenValidity(true, true))
		}
	}

	return nil
}

func formatDuration(d any) string {
	dur, ok := d.(interface{ Hours() float64 })
	if !ok {
		return "unknown"
	}

	hours := dur.Hours()
	if hours < 1 {
		mins := hours * 60
		return fmt.Sprintf("%.0fm", mins)
	}
	if hours < 24 {
		return fmt.Sprintf("%.0fh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%.0fd %.0fh", days, float64(int(hours)%24))
}
