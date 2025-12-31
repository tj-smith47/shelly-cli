// Package version provides the version command for displaying CLI version info.
package version

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/github"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	CheckUpdate bool
	JSON        bool
	Short       bool
}

// NewCommand creates the version command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Short, "short", "s", false, "Print only the version number")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output version info as JSON")
	cmd.Flags().BoolVarP(&opts.CheckUpdate, "check", "c", false, "Check for available updates")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	info := version.Get()
	checker := func(c context.Context) (*version.UpdateResult, error) {
		return github.CheckForUpdates(c, ios, info.Version)
	}

	if opts.Short {
		ios.Printf("%s\n", info.Version)
		return nil
	}
	if opts.JSON {
		if err := version.WriteJSONOutput(ctx, ios.Out, info, opts.CheckUpdate, github.ReleaseFetcher(ios), github.IsNewerVersion); err != nil {
			ios.DebugErr("encoding JSON", err)
		}
		return nil
	}

	term.DisplayVersionInfo(ios, info.Version, info.Commit, info.Date, info.BuiltBy, info.GoVersion, info.OS, info.Arch)

	if opts.CheckUpdate {
		term.RunUpdateCheck(ctx, ios, checker)
	} else if cached := version.ReadCachedVersion(); cached != "" && github.IsNewerVersion(info.Version, cached) {
		term.DisplayUpdateAvailable(ios, info.Version, cached)
	}
	return nil
}
