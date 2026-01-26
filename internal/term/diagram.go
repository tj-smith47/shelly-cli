package term

import (
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// DisplayDiagram prints a rendered wiring diagram to the terminal.
func DisplayDiagram(ios *iostreams.IOStreams, rendered string) {
	ios.Println(rendered)
}

// DisplayDiagramGenerationNote prints a note when a device slug matches
// multiple generations, showing which was selected and listing alternatives.
func DisplayDiagramGenerationNote(ios *iostreams.IOStreams, selected int, allGens []int) {
	var others []string
	for _, g := range allGens {
		if g != selected {
			others = append(others, fmt.Sprintf("Gen%d", g))
		}
	}
	if len(others) == 0 {
		return
	}
	ios.Info("Showing Gen%d (latest). Also available: %s. Use -g <gen> to select.",
		selected, strings.Join(others, ", "))
}
