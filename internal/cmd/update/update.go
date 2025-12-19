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
	if currentVersion == "" || currentVersion == "dev" {
		ios.Warning("Development version detected, cannot determine update status")
		if !check {
			return fmt.Errorf("cannot update development version")
		}
	}

	// Handle rollback
	if rollback {
		previousRelease, err := ghClient.FindPreviousRelease(ctx, currentVersion, includePre)
		if err != nil {
			return fmt.Errorf("no previous version available for rollback")
		}

		ios.Printf("Rolling back from %s to %s\n", currentVersion, previousRelease.Version())

		confirm, err := f.ConfirmAction("Proceed with rollback?", yes)
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if !confirm {
			ios.Info("Rollback cancelled")
			return nil
		}

		return ghClient.InstallRelease(ctx, ios, previousRelease)
	}

	// Get target release
	var targetRelease *github.Release
	var err error

	if ver != "" {
		targetRelease, err = ghClient.FetchSpecificVersion(ctx, ver)
		if err != nil {
			return fmt.Errorf("failed to fetch version %s: %w", ver, err)
		}
	} else {
		targetRelease, err = ghClient.FetchLatestVersion(ctx, includePre || channel == "beta")
		if err != nil {
			if errors.Is(err, github.ErrNoReleases) {
				ios.Info("No releases found")
				return nil
			}
			return fmt.Errorf("failed to check for updates: %w", err)
		}
	}

	availableVersion := targetRelease.Version()

	// Compare versions
	hasUpdate := github.IsNewerVersion(currentVersion, availableVersion)

	if check {
		term.DisplayUpdateStatus(ios, currentVersion, availableVersion, hasUpdate, targetRelease.HTMLURL)
		return nil
	}

	// Check if update is needed
	if !hasUpdate && ver == "" {
		ios.Printf("Already at latest version (%s)\n", currentVersion)
		return nil
	}

	// Show update info and confirm
	ios.Printf("\nCurrent version: %s\n", currentVersion)
	ios.Printf("Available version: %s\n", availableVersion)

	if targetRelease.Body != "" {
		ios.Printf("\nRelease notes:\n%s\n", output.FormatReleaseNotes(targetRelease.Body))
	}

	confirm, err := f.ConfirmAction("\nProceed with update?", yes)
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	if !confirm {
		ios.Info("Update cancelled")
		return fmt.Errorf("update cancelled by user")
	}

	// Perform update
	return ghClient.InstallRelease(ctx, ios, targetRelease)
}
