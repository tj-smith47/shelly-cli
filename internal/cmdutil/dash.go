package cmdutil

import "github.com/spf13/cobra"

// AnnotationDashIsOutput marks a command whose bare "-" argument means "write to
// stdout". The root command replaces a bare "-" with piped stdin for every other
// command, so a command that documents "-" as its stdout target must carry this
// annotation to receive the argument as typed.
const AnnotationDashIsOutput = "shelly.dash-is-output"

// DashIsOutputAnnotation returns the Annotations map for a command that reads a
// bare "-" argument as "write to stdout".
func DashIsOutputAnnotation() map[string]string {
	return map[string]string{AnnotationDashIsOutput: "true"}
}

// DashIsOutput reports whether cmd reads a bare "-" argument as "write to stdout".
func DashIsOutput(cmd *cobra.Command) bool {
	return cmd != nil && cmd.Annotations[AnnotationDashIsOutput] == "true"
}
