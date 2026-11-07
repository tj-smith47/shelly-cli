// Package logout provides the cloud logout subcommand.
package logout

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	Yes     bool
}

// NewCommand creates the cloud logout command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "logout",
		Aliases: []string{"signout"},
		Short:   "Clear cloud credentials",
		Long: `Clear the stored Shelly Cloud credentials.

This removes your access token and email from the local configuration.
You will need to login again to use cloud commands.`,
		Example: `  # Logout from Shelly Cloud
  shelly cloud logout

  # Skip the confirmation prompt (scripts)
  shelly cloud logout --yes`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	flags.AddYesFlag(cmd, &opts.Yes)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	cfg, err := opts.Factory.Config()
	if err != nil {
		return err
	}

	// Check if we're logged in
	if cfg.Cloud.AccessToken == "" {
		ios.Info("Not logged in to Shelly Cloud")
		return nil
	}

	email := cfg.Cloud.Email

	what := "Log out from Shelly Cloud and forget the stored token?"
	if email != "" {
		what = fmt.Sprintf("Log out %s from Shelly Cloud and forget the stored token?", email)
	}
	confirmed, err := opts.Factory.ConfirmAction(what, opts.Yes)
	if err != nil {
		return err
	}
	if !confirmed {
		ios.Info("Cancelled")
		return nil
	}

	// Clear credentials
	cfg.Cloud.Enabled = false
	cfg.Cloud.Email = ""
	cfg.Cloud.AccessToken = ""
	cfg.Cloud.RefreshToken = ""
	cfg.Cloud.ServerURL = ""

	if err := config.Save(); err != nil {
		ios.Error("Failed to save configuration: %v", err)
		return err
	}

	if email != "" {
		ios.Success("Logged out from %s", email)
	} else {
		ios.Success("Logged out from Shelly Cloud")
	}
	return nil
}
