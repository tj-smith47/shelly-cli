// Package info provides the profile info command.
package info

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/profiles"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Model   string
	Factory *cmdutil.Factory
}

// NewCommand creates the profile info command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "info <model>",
		Aliases: []string{"show", "get", "i"},
		Short:   "Show device profile details",
		Long: `Show detailed information about a specific device model.

Displays hardware capabilities, supported protocols, components,
and resource limits for the specified device model.`,
		Example: `  # Show info for Shelly Plus 1PM
  shelly profile info SNSW-001P16EU

  # JSON output
  shelly profile info SNSW-001P16EU -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return run(opts)
		},
	}

	flags.AddOutputFlags(cmd, &opts.OutputFlags)

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	profile, ok := profiles.Get(opts.Model)
	if !ok {
		return fmt.Errorf("unknown device model: %s", opts.Model)
	}

	return cmdutil.PrintResult(ios, profile, func(ios *iostreams.IOStreams, p *profiles.Profile) {
		term.DisplayProfile(ios, p)
	})
}
