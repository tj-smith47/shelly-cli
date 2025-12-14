// Package next provides the theme next command.
package next

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the theme next command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "next",
		Aliases: []string{"n"},
		Short:   "Cycle to next theme",
		Long:    `Cycle to the next theme in the list.`,
		Example: `  # Go to next theme
  shelly theme next`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(f)
		},
	}

	return cmd
}

func run(f *cmdutil.Factory) error {
	ios := f.IOStreams()

	theme.NextTheme()

	current := theme.Current()
	if current == nil {
		return fmt.Errorf("failed to get current theme")
	}

	ios.Success("Theme changed to '%s'", current.ID)
	return nil
}
