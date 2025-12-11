// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
)

// OutputConfig holds output-related configuration for commands.
type OutputConfig struct {
	Format   output.Format
	Template string
}

// GetOutputConfig returns the current output configuration from viper.
func GetOutputConfig() OutputConfig {
	return OutputConfig{
		Format:   output.GetFormat(),
		Template: output.GetTemplate(),
	}
}

// FormatOutput prints data in the configured output format.
// For table format, it uses the Table type from the output package.
// For other formats (json, yaml, template), it uses the standard formatters.
func FormatOutput(ios *iostreams.IOStreams, data any) error {
	cfg := GetOutputConfig()
	formatter := output.NewFormatter(cfg.Format)
	return formatter.Format(ios.Out, data)
}

// FormatTable creates and prints a table using IOStreams.
func FormatTable(ios *iostreams.IOStreams, headers []string, rows [][]string) {
	output.PrintTableTo(ios.Out, headers, rows)
}

// PrintFormatted outputs data using the specified format.
func PrintFormatted(ios *iostreams.IOStreams, format output.Format, data any) error {
	formatter := output.NewFormatter(format)
	return formatter.Format(ios.Out, data)
}

// IsQuiet returns true if quiet mode is enabled.
func IsQuiet() bool {
	return viper.GetBool("quiet")
}

// IsVerbose returns true if verbose mode is enabled.
func IsVerbose() bool {
	return viper.GetBool("verbose")
}

// WantsJSON returns true if the output format is JSON.
func WantsJSON() bool {
	return output.GetFormat() == output.FormatJSON
}

// WantsYAML returns true if the output format is YAML.
func WantsYAML() bool {
	return output.GetFormat() == output.FormatYAML
}

// WantsTable returns true if the output format is table.
func WantsTable() bool {
	return output.GetFormat() == output.FormatTable
}

// WantsStructured returns true if the output format is JSON or YAML.
func WantsStructured() bool {
	format := output.GetFormat()
	return format == output.FormatJSON || format == output.FormatYAML
}

// ShouldShowProgress returns true if progress indicators should be shown.
// Progress is hidden in quiet mode or when outputting structured data.
func ShouldShowProgress(ios *iostreams.IOStreams) bool {
	if IsQuiet() {
		return false
	}
	if WantsStructured() {
		return false
	}
	return ios.IsStdoutTTY()
}

// ConditionalPrint prints a message only if not in quiet mode.
func ConditionalPrint(ios *iostreams.IOStreams, format string, args ...any) {
	if !IsQuiet() {
		ios.Printf(format, args...)
	}
}

// ConditionalSuccess prints a success message only if not in quiet mode.
func ConditionalSuccess(ios *iostreams.IOStreams, format string, args ...any) {
	if !IsQuiet() {
		ios.Success(format, args...)
	}
}

// ConditionalInfo prints an info message only if not in quiet mode.
func ConditionalInfo(ios *iostreams.IOStreams, format string, args ...any) {
	if !IsQuiet() {
		ios.Info(format, args...)
	}
}
