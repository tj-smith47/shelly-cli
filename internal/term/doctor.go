package term

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PrintDoctorHeader prints the doctor command header.
func PrintDoctorHeader(ios *iostreams.IOStreams) {
	ios.Println()
	ios.Println(theme.Title().Render("Shelly CLI Doctor"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println()
}

// PrintDoctorSummary prints the doctor command summary.
func PrintDoctorSummary(ios *iostreams.IOStreams, issues, warnings int) {
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))

	switch {
	case issues == 0 && warnings == 0:
		ios.Success("No issues found. Your Shelly CLI setup looks healthy!")
	case issues == 0:
		ios.Success("No critical issues found.")
		if warnings > 0 {
			ios.Info("%d warning(s) - see above for details", warnings)
		}
	default:
		WarnStdout(ios, "%d issue(s) found - see above for details", issues)
		if warnings > 0 {
			ios.Info("%d additional warning(s)", warnings)
		}
	}

	ios.Println()
	ios.Println(fmt.Sprintf("Run %s for all diagnostics including device tests.",
		theme.Code().Render("shelly doctor --full")))
	ios.Println()
}

// WarnStdout prints a warning message to stdout (not stderr) for consistent diagnostic output ordering.
func WarnStdout(ios *iostreams.IOStreams, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ios.Println(theme.StatusWarn().Render("⚠") + " " + msg)
}
