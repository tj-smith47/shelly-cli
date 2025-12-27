package flags

import (
	"time"

	"github.com/spf13/cobra"
)

// BatchFlags holds common flags for batch operations targeting multiple devices.
// Embed this in your Options struct for commands that operate on device groups.
//
// Usage:
//
//	type Options struct {
//	    flags.BatchFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddBatchFlags(cmd, &opts.BatchFlags)
//	    return cmd
//	}
type BatchFlags struct {
	GroupName  string
	All        bool
	Timeout    time.Duration
	SwitchID   int
	Concurrent int
}

// AddBatchFlags adds all standard batch operation flags to a command.
// This is a convenience function for commands that support batch operations
// targeting multiple devices via groups or the --all flag.
func AddBatchFlags(cmd *cobra.Command, flags *BatchFlags) {
	AddGroupFlag(cmd, &flags.GroupName)
	AddAllFlag(cmd, &flags.All)
	AddTimeoutFlag(cmd, &flags.Timeout)
	AddSwitchIDFlag(cmd, &flags.SwitchID)
	AddConcurrencyFlag(cmd, &flags.Concurrent)
}

// SetBatchDefaults sets default values for batch flags.
// Call this after AddBatchFlags if the flags struct was not zero-initialized.
func SetBatchDefaults(flags *BatchFlags) {
	flags.Timeout = DefaultTimeout
	flags.Concurrent = DefaultConcurrency
	flags.SwitchID = DefaultComponentID
}
