package term

import (
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// PrintMapSection prints a map section with indentation recursively.
func PrintMapSection(ios *iostreams.IOStreams, m map[string]any, indent string) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			ios.Println(indent + k + ":")
			PrintMapSection(ios, val, indent+"  ")
		default:
			ios.Printf("%s%s: %v\n", indent, k, v)
		}
	}
}

// PrintJSONResult prints a JSON result with a styled header.
func PrintJSONResult(ios *iostreams.IOStreams, header string, result any) {
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		ios.DebugErr("failed to marshal result", err)
		return
	}
	ios.Println("  " + theme.Highlight().Render(header))
	ios.Println(string(jsonBytes))
	ios.Println()
}
