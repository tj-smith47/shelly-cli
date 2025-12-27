// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import "github.com/spf13/cobra"

// AddCommandsToGroup adds multiple commands to a root command and assigns them to a group.
// This is a convenience function for organizing help output.
func AddCommandsToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.GroupID = groupID
		root.AddCommand(cmd)
	}
}
