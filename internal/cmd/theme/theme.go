// Package theme provides CLI commands for managing themes.
package theme

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/current"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/exportcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/next"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/prev"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/preview"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/semantic"
	"github.com/tj-smith47/shelly-cli/internal/cmd/theme/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the theme command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "theme",
		Aliases: []string{"themes", "t"},
		Short:   "Manage CLI color themes",
		Long: `Manage CLI color themes using bubbletint.

The Shelly CLI supports 280+ built-in terminal color themes. Themes affect
all CLI output including tables, status indicators, and the TUI dashboard.

Popular themes include:
  - dracula (default)
  - nord
  - tokyo-night
  - github-dark
  - gruvbox
  - catppuccin
  - one-dark
  - solarized`,
		Example: `  # List all available themes
  shelly theme list

  # Set a theme
  shelly theme set nord

  # Preview a theme
  shelly theme preview tokyo-night

  # Show current theme
  shelly theme current

  # Cycle through themes
  shelly theme next
  shelly theme prev`,
	}

	// Add subcommands
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(preview.NewCommand(f))
	cmd.AddCommand(current.NewCommand(f))
	cmd.AddCommand(next.NewCommand(f))
	cmd.AddCommand(prev.NewCommand(f))
	cmd.AddCommand(exportcmd.NewCommand(f))
	cmd.AddCommand(importcmd.NewCommand(f))
	cmd.AddCommand(semantic.NewCommand(f))

	return cmd
}
