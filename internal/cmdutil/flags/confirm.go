package flags

import "github.com/spf13/cobra"

// ConfirmFlags holds flags for confirmation handling in destructive operations.
// Embed this in your Options struct for commands that require confirmation.
//
// Usage:
//
//	type Options struct {
//	    flags.ConfirmFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddConfirmFlags(cmd, &opts.ConfirmFlags)
//	    return cmd
//	}
type ConfirmFlags struct {
	Yes     bool // Skip confirmation prompt
	Confirm bool // Double-confirm for destructive operations
}

// AddConfirmFlags adds confirmation-related flags to a command.
func AddConfirmFlags(cmd *cobra.Command, flags *ConfirmFlags) {
	AddYesFlag(cmd, &flags.Yes)
	AddConfirmFlag(cmd, &flags.Confirm)
}

// AddYesOnlyFlag adds only the --yes/-y flag (no double-confirm).
// Use this for operations that are recoverable but still need confirmation.
func AddYesOnlyFlag(cmd *cobra.Command, flags *ConfirmFlags) {
	AddYesFlag(cmd, &flags.Yes)
}
