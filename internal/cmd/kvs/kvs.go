// Package kvs provides KVS (Key-Value Storage) commands.
package kvs

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/deletecmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/export"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/get"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/importcmd"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/kvs/set"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the kvs command and its subcommands.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kvs",
		Aliases: []string{"kv", "store"},
		Short:   "Manage device key-value storage",
		Long: `Manage the Key-Value Storage (KVS) on Shelly Gen2+ devices.

KVS provides persistent storage for scripts and external applications.
Values persist across device reboots and can store strings, numbers,
booleans, and null values.

Limits:
  - Maximum 50 key-value pairs (varies by device)
  - Key length: up to 42 bytes
  - Value size: up to 256 bytes (strings)`,
		Example: `  # List all KVS keys
  shelly kvs list living-room

  # Get a specific value
  shelly kvs get living-room my_key

  # Set a value
  shelly kvs set living-room my_key "my_value"

  # Set a numeric value
  shelly kvs set living-room counter 42

  # Delete a key
  shelly kvs delete living-room my_key

  # Export all KVS data to file
  shelly kvs export living-room backup.json

  # Import KVS data from file
  shelly kvs import living-room backup.json`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(get.NewCommand(f))
	cmd.AddCommand(set.NewCommand(f))
	cmd.AddCommand(deletecmd.NewCommand(f))
	cmd.AddCommand(export.NewCommand(f))
	cmd.AddCommand(importcmd.NewCommand(f))

	return cmd
}
