// Package backup provides backup and restore commands.
package backup

import (
	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"

	"github.com/tj-smith47/shelly-cli/internal/cmd/backup/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/backup/restore"
)

// NewCommand creates the backup command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backup",
		Aliases: []string{"bak"},
		Short:   "Backup and restore device configurations",
		Long: `Create, restore, and manage device backups.

Backups include device configuration, scripts, schedules, and webhooks.
Use encryption to protect sensitive data in backups.`,
		Example: `  # Create a backup
  shelly backup create living-room backup.json

  # Restore from backup
  shelly backup restore living-room backup.json

  # List existing backups
  shelly backup list

  # Export all device backups
  shelly backup export ./backups`,
	}

	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(restore.NewCommand(f))
	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(export.NewCommand(f))

	return cmd
}
