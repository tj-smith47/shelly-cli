// Package update provides the update command for self-updating the CLI.
package update

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// NewCommand creates the update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		check      bool
		ver        string
		channel    string
		rollback   bool
		yes        bool
		includePre bool
	)

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade", "u"},
		Short:   "Update shelly to the latest version",
		Long: `Update shelly to the latest version.

By default, downloads and installs the latest stable release from GitHub.
Use --check to only check for updates without installing.
Use --version to install a specific version.`,
		Example: `  # Check for updates
  shelly update --check

  # Update to latest version
  shelly update

  # Update to a specific version
  shelly update --version v1.2.0

  # Update with pre-releases
  shelly update --include-pre

  # Update without confirmation
  shelly update --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, check, ver, channel, rollback, yes, includePre)
		},
	}

	cmd.Flags().BoolVarP(&check, "check", "c", false, "Check for updates without installing")
	cmd.Flags().StringVar(&ver, "version", "", "Install a specific version")
	cmd.Flags().StringVar(&channel, "channel", "stable", "Release channel (stable, beta)")
	cmd.Flags().BoolVar(&rollback, "rollback", false, "Rollback to previous version")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&includePre, "include-pre", false, "Include pre-release versions")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, check bool, ver, channel string, rollback, yes, includePre bool) error {
	ios := f.IOStreams()
	ghClient := github.NewClient(ios)
	currentVersion := version.Version

	if version.IsDevelopment() {
		ios.Warning("Development version detected, cannot determine update status")
		if !check {
			return fmt.Errorf("cannot update development version")
		}
	}

	if rollback {
		return ghClient.PerformRollback(ctx, ios, currentVersion, includePre, f.ConfirmAction, yes)
	}

	release, err := ghClient.GetTargetRelease(ctx, ver, includePre || channel == "beta")
	if err != nil {
		if errors.Is(err, github.ErrNoReleases) {
			ios.Info("No releases found")
			return nil
		}
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	availableVersion := release.Version()
	hasUpdate := github.IsNewerVersion(currentVersion, availableVersion)

	if check {
		term.DisplayUpdateStatus(ios, currentVersion, availableVersion, hasUpdate, release.HTMLURL)
		return nil
	}

	if !hasUpdate && ver == "" {
		ios.Printf("Already at latest version (%s)\n", currentVersion)
		return nil
	}

	return ghClient.PerformUpdate(ctx, ios, release, currentVersion, output.FormatReleaseNotes(release.Body), f.ConfirmAction, yes)
}
