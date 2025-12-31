// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// AddCommandsToGroup adds multiple commands to a root command and assigns them to a group.
// This is a convenience function for organizing help output.
func AddCommandsToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.GroupID = groupID
		root.AddCommand(cmd)
	}
}

// SafeConfig returns the config, logging errors to debug if load fails.
// Use this when config is optional and the command can continue without it.
// Returns nil if config loading fails.
func SafeConfig(f *Factory) *config.Config {
	cfg, err := f.Config()
	if err != nil {
		f.IOStreams().DebugErr("load config", err)
		return nil
	}
	return cfg
}
