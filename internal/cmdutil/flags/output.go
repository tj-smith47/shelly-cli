package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

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
//	    flags.AddOutputFlags(cmd, &opts.OutputFlags)  // default: table, json, yaml
//	    // OR with custom values:
//	    flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "json", "json", "yaml")
//	    return cmd
//	}
type OutputFlags struct {
	Format string
}

// AddOutputFlags adds output format flags with standard defaults (table/json/yaml).
func AddOutputFlags(cmd *cobra.Command, flags *OutputFlags) {
	AddOutputFormatFlag(cmd, &flags.Format)
}

// AddOutputFlagsCustom adds output format flags with custom default and allowed values.
// Example: AddOutputFlagsCustom(cmd, &opts.OutputFlags, "json", "json", "yaml", "text").
func AddOutputFlagsCustom(cmd *cobra.Command, flags *OutputFlags, defaultVal string, allowed ...string) {
	usage := fmt.Sprintf("Output format: %s", strings.Join(allowed, ", "))
	cmd.Flags().StringVarP(&flags.Format, "format", "f", defaultVal, usage)
}

// AddOutputFlagsNamed adds output format flags with a custom flag name.
// Use this when the flag should be named something other than "format" (e.g., "output").
// Example: AddOutputFlagsNamed(cmd, &opts.OutputFlags, "output", "o", "json", "json", "yaml").
func AddOutputFlagsNamed(cmd *cobra.Command, flags *OutputFlags, name, shorthand, defaultVal string, allowed ...string) {
	usage := fmt.Sprintf("Output format: %s", strings.Join(allowed, ", "))
	cmd.Flags().StringVarP(&flags.Format, name, shorthand, defaultVal, usage)
}

// SetOutputDefaults sets default values for output flags.
func SetOutputDefaults(flags *OutputFlags) {
	flags.Format = "table"
}
