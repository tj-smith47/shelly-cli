// Package versioncmd provides the version command for displaying CLI version info.
package versioncmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

const unknownValue = "unknown"

type options struct {
	Short       bool
	JSON        bool
	CheckUpdate bool
}

// NewCommand creates the version command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"--version"},
		Short:   "Print version information",
		Long: `Print the version of shelly CLI.

By default, shows version, commit, and build date.
Use --short for just the version number.
Use --json for machine-readable output.
Use --check to also check for available updates.`,
		Example: `  # Show version info
  shelly version

  # Short version output
  shelly version --short

  # JSON output
  shelly version --json

  # Check for updates
  shelly version --check`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Short, "short", "s", false, "Print only the version number")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output version info as JSON")
	cmd.Flags().BoolVarP(&opts.CheckUpdate, "check", "c", false, "Check for available updates")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *options) error {
	ios := f.IOStreams()
	info := version.Get()

	// Handle short output
	if opts.Short {
		ios.Printf("%s\n", info.Version)
		return nil
	}

	// Handle JSON output
	if opts.JSON {
		return outputJSON(ios, info, opts.CheckUpdate, ctx)
	}

	// Standard output
	ios.Printf("shelly version %s\n", info.Version)
	if info.Commit != "" && info.Commit != unknownValue {
		ios.Printf("  commit: %s\n", info.Commit)
	}
	if info.Date != "" && info.Date != unknownValue {
		ios.Printf("  built: %s\n", info.Date)
	}
	if info.BuiltBy != "" && info.BuiltBy != unknownValue {
		ios.Printf("  by: %s\n", info.BuiltBy)
	}
	ios.Printf("  go: %s\n", info.GoVersion)
	ios.Printf("  os/arch: %s/%s\n", info.OS, info.Arch)

	// Check for updates if requested or if cache indicates update available
	if opts.CheckUpdate {
		checkAndDisplayUpdate(ctx, ios, info.Version)
	} else {
		// Silently check cache for update availability
		displayCachedUpdateInfo(ios, info.Version)
	}

	return nil
}

type jsonOutput struct {
	Version       string  `json:"version"`
	Commit        string  `json:"commit"`
	Date          string  `json:"date"`
	BuiltBy       string  `json:"built_by"`
	GoVersion     string  `json:"go_version"`
	OS            string  `json:"os"`
	Arch          string  `json:"arch"`
	UpdateAvail   *string `json:"update_available,omitempty"`
	LatestVersion *string `json:"latest_version,omitempty"`
}

func outputJSON(ios *iostreams.IOStreams, info version.Info, checkUpdate bool, ctx context.Context) error {
	output := jsonOutput{
		Version:   info.Version,
		Commit:    info.Commit,
		Date:      info.Date,
		BuiltBy:   info.BuiltBy,
		GoVersion: info.GoVersion,
		OS:        info.OS,
		Arch:      info.Arch,
	}

	// Check for updates if requested
	if checkUpdate {
		addUpdateInfo(ctx, ios, info.Version, &output)
	}

	encoder := json.NewEncoder(ios.Out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		ios.DebugErr("encoding JSON", err)
	}
	return nil
}

func addUpdateInfo(ctx context.Context, ios *iostreams.IOStreams, currentVersion string, output *jsonOutput) {
	ghClient := github.NewClient(ios)
	release, err := ghClient.GetLatestRelease(ctx, github.DefaultOwner, github.DefaultRepo)
	if err != nil || release == nil {
		return
	}

	latestVer := release.Version()
	output.LatestVersion = &latestVer

	if github.IsNewerVersion(currentVersion, latestVer) {
		updateAvail := "yes"
		output.UpdateAvail = &updateAvail
	} else {
		updateAvail := "no"
		output.UpdateAvail = &updateAvail
	}
}

func checkAndDisplayUpdate(ctx context.Context, ios *iostreams.IOStreams, currentVersion string) {
	// Skip for dev versions
	if currentVersion == "" || currentVersion == "dev" {
		return
	}

	ios.StartProgress("Checking for updates...")

	ghClient := github.NewClient(ios)

	// Use a timeout for the update check
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	release, err := ghClient.GetLatestRelease(checkCtx, github.DefaultOwner, github.DefaultRepo)
	ios.StopProgress()

	if err != nil {
		ios.DebugErr("checking for updates", err)
		return
	}

	availableVersion := release.Version()
	if github.IsNewerVersion(currentVersion, availableVersion) {
		ios.Printf("\n")
		ios.Warning("Update available: %s -> %s", currentVersion, availableVersion)
		ios.Printf("  Run 'shelly update' to install the latest version\n")
	} else {
		ios.Printf("\n")
		ios.Success("You are using the latest version")
	}

	// Update the cache
	updateVersionCache(ios, availableVersion)
}

func displayCachedUpdateInfo(ios *iostreams.IOStreams, currentVersion string) {
	// Skip for dev versions
	if currentVersion == "" || currentVersion == "dev" {
		return
	}

	cachePath := getVersionCachePath()
	if cachePath == "" {
		return
	}

	// Check if cache exists and is recent (within 24 hours)
	info, err := os.Stat(cachePath)
	if err != nil {
		return
	}

	if time.Since(info.ModTime()) > 24*time.Hour {
		return
	}

	data, err := os.ReadFile(cachePath) //nolint:gosec // G304: cachePath is from known config directory
	if err != nil {
		return
	}

	cachedVersion := string(data)
	if github.IsNewerVersion(currentVersion, cachedVersion) {
		ios.Printf("\n")
		ios.Warning("Update available: %s -> %s", currentVersion, cachedVersion)
		ios.Printf("  Run 'shelly update' to install the latest version\n")
	}
}

func updateVersionCache(ios *iostreams.IOStreams, latestVersion string) {
	cachePath := getVersionCachePath()
	if cachePath == "" {
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o750); err != nil {
		ios.DebugErr("creating cache directory", err)
		return
	}

	if err := os.WriteFile(cachePath, []byte(latestVersion), 0o600); err != nil {
		ios.DebugErr("writing version cache", err)
	}
}

func getVersionCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "shelly", "cache", "latest-version")
}
