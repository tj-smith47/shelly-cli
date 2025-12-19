package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
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
