package term

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplayScriptEvalResult displays the result of evaluating script code.
func DisplayScriptEvalResult(ios *iostreams.IOStreams, result any) {
	if result == nil {
		ios.Info("(no result)")
		return
	}

	// Try to pretty-print JSON for complex types
	switch v := result.(type) {
	case string:
		ios.Println(v)
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			ios.Printf("%d\n", int64(v))
		} else {
			ios.Printf("%v\n", v)
		}
	case bool:
		ios.Printf("%t\n", v)
	default:
		// Try to marshal as JSON for complex types
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ios.Printf("%v\n", result)
		} else {
			ios.Println(string(jsonBytes))
		}
	}
}

// DisplayScriptStatus displays detailed script status.
func DisplayScriptStatus(ios *iostreams.IOStreams, status *shelly.ScriptStatus) {
	ios.Println(theme.Bold().Render("Script Status"))
	ios.Println("")

	ios.Printf("  ID:      %d\n", status.ID)
	ios.Printf("  Status:  %s\n", output.RenderRunningState(status.Running))
	ios.Println("")

	ios.Println(theme.Bold().Render("Memory"))
	ios.Printf("  Usage:   %d bytes\n", status.MemUsage)
	ios.Printf("  Peak:    %d bytes\n", status.MemPeak)
	ios.Printf("  Free:    %d bytes\n", status.MemFree)

	if len(status.Errors) > 0 {
		ios.Println("")
		ios.Println(theme.StatusError().Render("Errors:"))
		for _, e := range status.Errors {
			ios.Printf("  - %s\n", e)
		}
	}
}

// DisplayScriptCode displays script source code.
func DisplayScriptCode(ios *iostreams.IOStreams, code string) {
	if code == "" {
		ios.Info("Script has no code")
		return
	}
	ios.Println(code)
}
