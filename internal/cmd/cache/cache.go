// Package cache provides the cache command for managing CLI cache.
package cache

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/cache/clearcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/cache/show"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the cache command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage CLI cache",
		Long: `Manage the Shelly CLI cache directory.

The cache stores:
  - Device discovery results
  - Firmware update information
  - Version check data`,
		Aliases: []string{"ca"},
		Example: `  # Show cache statistics
  shelly cache show

  # Clear the cache
  shelly cache clear`,
	}

	cmd.AddCommand(clearcmd.NewCommand(f))
	cmd.AddCommand(show.NewCommand(f))

	return cmd
}
