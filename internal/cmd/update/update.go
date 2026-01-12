// Package update provides the update command for self-updating the CLI.
package update

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Options holds command options.
type Options struct {
	flags.ConfirmFlags
	Factory    *cmdutil.Factory
	Check      bool
	Force      bool
	Version    string
	Channel    string
	Rollback   bool
	IncludePre bool
}

// NewCommand creates the update command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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

  # Check for updates and refresh the cache
  shelly update --check --force

  # Update to latest version
  shelly update

  # Update to a specific version
  shelly update --version v1.2.0

  # Update with pre-releases
  shelly update --include-pre

  # Update without confirmation
  shelly update --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Check, "check", "c", false, "Check for updates without installing")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force refresh and update the version cache")
	cmd.Flags().StringVar(&opts.Version, "version", "", "Install a specific version")
	cmd.Flags().StringVar(&opts.Channel, "channel", "stable", "Release channel (stable, beta)")
	cmd.Flags().BoolVar(&opts.Rollback, "rollback", false, "Rollback to previous version")
	flags.AddYesOnlyFlag(cmd, &opts.ConfirmFlags)
	cmd.Flags().BoolVar(&opts.IncludePre, "include-pre", false, "Include pre-release versions")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	ghClient := github.NewClient(ios)
	currentVersion := version.Version

	// Detect installation method
	installInfo := github.DetectInstallMethod()

	// For non-check operations, verify self-update is supported
	if !opts.Check && !installInfo.CanSelfUpdate() {
		ios.Warning("Homebrew installation detected")
		ios.Printf("  Update using: %s\n", installInfo.UpdateCommand)
		return fmt.Errorf("self-update not supported for Homebrew installations")
	}

	if opts.Rollback {
		return ghClient.PerformRollback(ctx, ios, currentVersion, opts.IncludePre, opts.Factory.ConfirmAction, opts.Yes)
	}

	release, err := ghClient.GetTargetRelease(ctx, opts.Version, opts.IncludePre || opts.Channel == "beta")
	if err != nil {
		if errors.Is(err, github.ErrNoReleases) {
			ios.Info("No releases found")
			return nil
		}
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	availableVersion := release.Version()
	hasUpdate := version.IsNewerVersion(currentVersion, availableVersion)

	if opts.Check {
		term.DisplayUpdateCheckInfo(ios, term.UpdateCheckInfo{
			CurrentVersion:   currentVersion,
			AvailableVersion: availableVersion,
			HasUpdate:        hasUpdate,
			ReleaseURL:       release.HTMLURL,
			CanSelfUpdate:    installInfo.CanSelfUpdate(),
			UpdateCommand:    installInfo.UpdateCommand,
		})
		// Update cache when --force is used
		if opts.Force {
			if err := version.WriteCache(availableVersion); err != nil {
				ios.DebugErr("updating version cache", err)
			}
		}
		return nil
	}

	if !hasUpdate && opts.Version == "" {
		ios.Success("Already at latest version (%s)", currentVersion)
		return nil
	}

	return ghClient.PerformUpdate(ctx, ios, release, currentVersion, output.FormatReleaseNotes(release.Body), opts.Factory.ConfirmAction, opts.Yes)
}
