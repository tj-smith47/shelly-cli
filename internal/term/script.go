package term

import (
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly/automation"
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
func DisplayScriptStatus(ios *iostreams.IOStreams, status *automation.ScriptStatus) {
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

// DisplayScriptTemplateList displays a list of script templates.
func DisplayScriptTemplateList(ios *iostreams.IOStreams, templates []config.ScriptTemplate) {
	table := output.NewTable("Name", "Category", "Description", "Source")

	for _, tpl := range templates {
		source := "user"
		if tpl.BuiltIn {
			source = "built-in"
		}
		table.AddRow(tpl.Name, tpl.Category, tpl.Description, source)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Count("template", len(templates))
}

// DisplayScriptTemplate displays detailed script template information.
func DisplayScriptTemplate(ios *iostreams.IOStreams, tpl config.ScriptTemplate) {
	ios.Println(theme.Bold().Render("Script Template: " + tpl.Name))
	ios.Println()

	// Metadata
	if tpl.Description != "" {
		ios.Printf("  Description:  %s\n", tpl.Description)
	}
	if tpl.Category != "" {
		ios.Printf("  Category:     %s\n", tpl.Category)
	}
	if tpl.Author != "" {
		ios.Printf("  Author:       %s\n", tpl.Author)
	}
	if tpl.Version != "" {
		ios.Printf("  Version:      %s\n", tpl.Version)
	}
	if tpl.MinGen > 0 {
		ios.Printf("  Min Gen:      %d\n", tpl.MinGen)
	}
	source := "user-defined"
	if tpl.BuiltIn {
		source = "built-in"
	}
	ios.Printf("  Source:       %s\n", source)

	// Variables
	if len(tpl.Variables) > 0 {
		ios.Println()
		ios.Println(theme.Bold().Render("Variables"))
		for _, v := range tpl.Variables {
			required := ""
			if v.Required {
				required = " (required)"
			}
			defaultVal := ""
			if v.Default != nil {
				defaultVal = fmt.Sprintf(" [default: %v]", v.Default)
			}
			ios.Printf("  %s (%s)%s%s\n", v.Name, v.Type, required, defaultVal)
			if v.Description != "" {
				ios.Printf("    %s\n", v.Description)
			}
		}
	}

	// Code
	ios.Println()
	ios.Println(theme.Bold().Render("Code"))
	ios.Println(theme.Dim().Render("─────────────────────────────────────────"))
	ios.Println(tpl.Code)
}
