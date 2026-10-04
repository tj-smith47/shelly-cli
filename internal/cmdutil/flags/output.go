package flags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/output"
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

// Structured reports whether the format is json or yaml, the formats
// cmdutil.PrintStructured prints.
func (f OutputFlags) Structured() bool {
	return f.Format == string(output.FormatJSON) || f.Format == string(output.FormatYAML)
}

// AddOutputFlags adds output format flags with standard defaults (table/json/yaml).
func AddOutputFlags(cmd *cobra.Command, flags *OutputFlags) {
	AddOutputFormatFlag(cmd, &flags.Format)
}

// AddOutputFlagsCustom adds output format flags with custom default and allowed values.
// Example: AddOutputFlagsCustom(cmd, &opts.OutputFlags, "json", "json", "yaml", "text").
func AddOutputFlagsCustom(cmd *cobra.Command, flags *OutputFlags, defaultVal string, allowed ...string) {
	usage := fmt.Sprintf("Output format: %s", strings.Join(allowed, ", "))
	cmd.Flags().StringVarP(&flags.Format, formatFlag, "f", defaultVal, usage)
	if err := cmd.Flags().SetAnnotation(formatFlag, allowedFormatsAnnotation, allowed); err != nil {
		panic(fmt.Sprintf("annotate --%s on %s: %v", formatFlag, cmd.Name(), err))
	}
}

const (
	formatFlag = "format"
	// allowedFormatsAnnotation records, on a command's own --format flag, the
	// values the command can print.
	allowedFormatsAnnotation = "shelly_allowed_formats"
)

// ApplyGlobalOutputFormat makes a command that has its own --format flag honour
// the global -o/--output flag: `shelly zigbee list -o json` prints JSON exactly
// as `--format json` does. An explicit --format wins. It also rejects a format
// the command cannot print, whichever flag carried it, so an unsupported value
// is an error and never a silent fall back to text.
//
// The root command calls it before every command runs. Commands whose --format
// was not added by AddOutputFlagsCustom are left alone.
func ApplyGlobalOutputFormat(cmd *cobra.Command) error {
	local := cmd.Flags().Lookup(formatFlag)
	if local == nil {
		return validateGlobalFormat(cmd)
	}
	allowed, ok := local.Annotations[allowedFormatsAnnotation]
	if !ok {
		return validateGlobalFormat(cmd)
	}
	supports := func(v string) bool {
		for _, a := range allowed {
			if a == v {
				return true
			}
		}
		return false
	}
	unsupported := func(v string) error {
		return fmt.Errorf("%q cannot print %q output; supported formats: %s",
			cmd.CommandPath(), v, strings.Join(allowed, ", "))
	}

	if local.Changed {
		if v := local.Value.String(); !supports(v) {
			return unsupported(v)
		}
		return nil
	}

	global := cmd.Root().PersistentFlags().Lookup("output")
	if global == nil || !global.Changed {
		return nil
	}
	v := global.Value.String()
	if v == string(output.FormatTable) && !supports(v) {
		// "table" is the global spelling of the human-readable default.
		return nil
	}
	if !supports(v) {
		return unsupported(v)
	}
	return local.Value.Set(v)
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
	flags.Format = string(output.FormatTable)
}

// validateGlobalFormat rejects a -o value the CLI does not know. Without this
// check an unknown value fell back to the table, so "-o csv >> file.csv" exited
// zero and appended a human-readable table to the file.
func validateGlobalFormat(cmd *cobra.Command) error {
	global := cmd.Root().PersistentFlags().Lookup("output")
	if global == nil || !global.Changed {
		return nil
	}
	if _, err := output.ParseFormat(global.Value.String()); err != nil {
		return fmt.Errorf("unknown output format %q; supported formats: %s",
			global.Value.String(), strings.Join(output.ValidFormats(), ", "))
	}
	return nil
}
