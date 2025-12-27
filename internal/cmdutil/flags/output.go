package flags

import "github.com/spf13/cobra"

// OutputFlags holds flags for controlling command output format.
// Embed this in your Options struct for commands with customizable output.
//
// Usage:
//
//	type Options struct {
//	    flags.OutputFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddOutputFlags(cmd, &opts.OutputFlags)
//	    return cmd
//	}
type OutputFlags struct {
	Format string
}

// AddOutputFlags adds output format flags to a command.
func AddOutputFlags(cmd *cobra.Command, flags *OutputFlags) {
	AddOutputFormatFlag(cmd, &flags.Format)
}

// SetOutputDefaults sets default values for output flags.
func SetOutputDefaults(flags *OutputFlags) {
	flags.Format = "table"
}
