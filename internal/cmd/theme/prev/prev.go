// Package prev provides the theme prev command.
package prev

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the options for the prev command.
type Options struct {
	Factory *cmdutil.Factory
}

// NewCommand creates the theme prev command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "prev",
		Aliases: []string{"previous", "p"},
		Short:   "Cycle to previous theme",
		Long:    `Cycle to the previous theme in the list.`,
		Example: `  # Go to previous theme
  shelly theme prev`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	theme.PrevTheme()

	current := theme.Current()
	if current == nil {
		return fmt.Errorf("failed to get current theme")
	}

	ios.Success("Theme changed to '%s'", current.ID)
	return nil
}
