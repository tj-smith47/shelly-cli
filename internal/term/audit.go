package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayAuditResult prints the security audit result for a single device.
func DisplayAuditResult(ios *iostreams.IOStreams, result *model.AuditResult) {
	ios.Printf("%s %s\n",
		theme.Bold().Render(result.Device),
		theme.Dim().Render(fmt.Sprintf("(%s)", result.Address)))

	if !result.Reachable {
		ios.Printf("  %s Device unreachable - cannot audit\n", theme.StatusWarn().Render("⚠"))
		ios.Println("")
		return
	}

	// Print issues (security concerns)
	for _, issue := range result.Issues {
		ios.Printf("  %s %s\n", theme.StatusWarn().Render("⚠"), issue)
	}

	// Print warnings
	for _, warn := range result.Warnings {
		ios.Printf("  %s %s\n", theme.StatusWarn().Render("!"), warn)
	}

	// Print info items
	for _, info := range result.InfoItems {
		ios.Printf("  %s %s\n", theme.StatusOK().Render("✓"), info)
	}

	ios.Println("")
}
