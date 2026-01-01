package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// displayStatusFields prints a slice of status fields with aligned labels.
func displayStatusFields(ios *iostreams.IOStreams, fields []output.StatusField) {
	// Find max label width for alignment (add 1 for the colon)
	maxWidth := 0
	for _, f := range fields {
		if len(f.Label) > maxWidth {
			maxWidth = len(f.Label)
		}
	}

	// Print each field with aligned label
	for _, f := range fields {
		ios.Printf("  %-*s %s\n", maxWidth+1, f.Label+":", f.Value)
	}
}

// DisplaySwitchStatus prints switch component status.
func DisplaySwitchStatus(ios *iostreams.IOStreams, status *model.SwitchStatus) {
	ios.Title("Switch %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatSwitchStatusFields(status))
}

// DisplayLightStatus prints light component status.
func DisplayLightStatus(ios *iostreams.IOStreams, status *model.LightStatus) {
	ios.Title("Light %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatLightStatusFields(status))
}

// DisplayRGBStatus prints RGB component status.
func DisplayRGBStatus(ios *iostreams.IOStreams, status *model.RGBStatus) {
	ios.Title("RGB %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatRGBStatusFields(status))
}

// DisplayRGBWStatus prints RGBW component status.
func DisplayRGBWStatus(ios *iostreams.IOStreams, status *model.RGBWStatus) {
	ios.Title("RGBW %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatRGBWStatusFields(status))
}

// DisplayCoverStatus prints cover component status.
func DisplayCoverStatus(ios *iostreams.IOStreams, status *model.CoverStatus) {
	ios.Title("Cover %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatCoverStatusFields(status))
}

// DisplayInputStatus prints input component status.
func DisplayInputStatus(ios *iostreams.IOStreams, status *model.InputStatus) {
	ios.Title("Input %d Status", status.ID)
	ios.Println()
	displayStatusFields(ios, output.FormatInputStatusFields(status))
}

// DisplayList prints a table of components that implement the model.Listable interface.
// It uses the generic Listable interface to get headers and rows from any component type.
// If emptyMsg is provided and the list is empty, it prints the message instead of an empty table.
func DisplayList[T model.Listable](ios *iostreams.IOStreams, items []T, emptyMsg string) {
	if len(items) == 0 && emptyMsg != "" {
		ios.Info("%s", emptyMsg)
		return
	}

	// Use a zero value to get headers (works even for empty slices)
	var zero T
	headers := zero.ListHeaders()
	builder := table.NewBuilder(headers...)

	for _, item := range items {
		builder.AddRow(item.ListRow()...)
	}

	table := builder.WithModeStyle(ios).Build()
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print component list table", err)
	}
}

// DisplaySwitchList prints a table of switches.
func DisplaySwitchList(ios *iostreams.IOStreams, switches []shelly.SwitchInfo) {
	DisplayList(ios, switches, "")
}

// DisplayLightList prints a table of lights.
func DisplayLightList(ios *iostreams.IOStreams, lights []shelly.LightInfo) {
	DisplayList(ios, lights, "")
}

// DisplayRGBList prints a table of RGB components.
func DisplayRGBList(ios *iostreams.IOStreams, rgbs []shelly.RGBInfo) {
	DisplayList(ios, rgbs, "")
}

// DisplayRGBWList prints a table of RGBW components.
func DisplayRGBWList(ios *iostreams.IOStreams, rgbws []shelly.RGBWInfo) {
	DisplayList(ios, rgbws, "")
}

// DisplayCoverList prints a table of covers.
func DisplayCoverList(ios *iostreams.IOStreams, covers []shelly.CoverInfo) {
	DisplayList(ios, covers, "")
}

// DisplayInputList prints a table of inputs.
func DisplayInputList(ios *iostreams.IOStreams, inputs []shelly.InputInfo) {
	DisplayList(ios, inputs, "")
	ios.Println()
	ios.Count("input", len(inputs))
}
