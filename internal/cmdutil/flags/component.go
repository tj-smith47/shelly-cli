package flags

import "github.com/spf13/cobra"

// ComponentFlags holds flags for targeting specific device components.
// Embed this in your Options struct for commands that operate on components.
//
// Usage:
//
//	type Options struct {
//	    flags.ComponentFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Switch")
//	    return cmd
//	}
type ComponentFlags struct {
	ID int
}

// AddComponentFlags adds a component ID flag (--id/-i) to a command.
// The componentName is used in the help text (e.g., "Switch", "Light", "Cover").
func AddComponentFlags(cmd *cobra.Command, flags *ComponentFlags, componentName string) {
	AddComponentIDFlag(cmd, &flags.ID, componentName)
}
