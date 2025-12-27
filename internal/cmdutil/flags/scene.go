package flags

import (
	"time"

	"github.com/spf13/cobra"
)

// SceneFlags holds common flags for scene activation operations.
// Embed this in your Options struct for commands that activate scenes.
//
// Usage:
//
//	type Options struct {
//	    flags.SceneFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddSceneFlags(cmd, &opts.SceneFlags)
//	    return cmd
//	}
type SceneFlags struct {
	Timeout    time.Duration
	Concurrent int
	DryRun     bool
}

// AddSceneFlags adds all standard scene operation flags to a command.
func AddSceneFlags(cmd *cobra.Command, flags *SceneFlags) {
	AddTimeoutFlag(cmd, &flags.Timeout)
	AddConcurrencyFlag(cmd, &flags.Concurrent)
	AddDryRunFlag(cmd, &flags.DryRun)
}

// SetSceneDefaults sets default values for scene flags.
func SetSceneDefaults(flags *SceneFlags) {
	flags.Timeout = DefaultTimeout
	flags.Concurrent = DefaultConcurrency
}
