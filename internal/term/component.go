package term

import (
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// DisplaySwitchStatus prints switch component status.
func DisplaySwitchStatus(ios *iostreams.IOStreams, status *model.SwitchStatus) {
	ios.Title("Switch %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:   %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	DisplayPowerMetrics(ios, status.Power, status.Voltage, status.Current)
	if status.Energy != nil {
		ios.Printf("  Energy:  %.2f Wh\n", status.Energy.Total)
	}
}

// DisplayLightStatus prints light component status.
func DisplayLightStatus(ios *iostreams.IOStreams, status *model.LightStatus) {
	ios.Title("Light %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

// DisplayRGBStatus prints RGB component status.
func DisplayRGBStatus(ios *iostreams.IOStreams, status *model.RGBStatus) {
	ios.Title("RGB %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.RGB != nil {
		ios.Printf("  Color:      R:%d G:%d B:%d\n",
			status.RGB.Red, status.RGB.Green, status.RGB.Blue)
	}
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

// DisplayRGBWStatus prints RGBW component status.
func DisplayRGBWStatus(ios *iostreams.IOStreams, status *model.RGBWStatus) {
	ios.Title("RGBW %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:      %s\n", output.RenderOnOff(status.Output, output.CaseUpper, theme.FalseError))
	if status.RGB != nil {
		ios.Printf("  Color:      R:%d G:%d B:%d\n",
			status.RGB.Red, status.RGB.Green, status.RGB.Blue)
	}
	if status.White != nil {
		ios.Printf("  White:      %d\n", *status.White)
	}
	if status.Brightness != nil {
		ios.Printf("  Brightness: %d%%\n", *status.Brightness)
	}
	if status.Power != nil {
		ios.Printf("  Power:      %.1f W\n", *status.Power)
	}
	if status.Voltage != nil {
		ios.Printf("  Voltage:    %.1f V\n", *status.Voltage)
	}
	if status.Current != nil {
		ios.Printf("  Current:    %.3f A\n", *status.Current)
	}
}

// DisplayCoverStatus prints cover component status.
func DisplayCoverStatus(ios *iostreams.IOStreams, status *model.CoverStatus) {
	ios.Title("Cover %d Status", status.ID)
	ios.Println()

	ios.Printf("  State:    %s\n", output.RenderCoverState(status.State))
	if status.CurrentPosition != nil {
		ios.Printf("  Position: %d%%\n", *status.CurrentPosition)
	}
	DisplayPowerMetricsWide(ios, status.Power, status.Voltage, status.Current)
}

// DisplayInputStatus prints input component status.
func DisplayInputStatus(ios *iostreams.IOStreams, status *model.InputStatus) {
	ios.Title("Input %d Status", status.ID)
	ios.Println()

	ios.Printf("  State: %s\n", output.RenderActive(status.State, output.CaseLower, theme.FalseError))
	if status.Type != "" {
		ios.Printf("  Type:  %s\n", status.Type)
	}
}

// DisplaySwitchList prints a table of switches.
func DisplaySwitchList(ios *iostreams.IOStreams, switches []shelly.SwitchInfo) {
	t := output.NewTable("ID", "Name", "State", "Power")
	for _, sw := range switches {
		name := output.FormatComponentName(sw.Name, "switch", sw.ID)
		state := output.RenderOnOff(sw.Output, output.CaseUpper, theme.FalseError)
		power := output.FormatPowerTableValue(sw.Power)
		t.AddRow(fmt.Sprintf("%d", sw.ID), name, state, power)
	}
	printTable(ios, t)
}

// DisplayLightList prints a table of lights.
func DisplayLightList(ios *iostreams.IOStreams, lights []shelly.LightInfo) {
	t := output.NewTable("ID", "Name", "State", "Brightness", "Power")
	for _, lt := range lights {
		name := output.FormatComponentName(lt.Name, "light", lt.ID)
		state := output.RenderOnOff(lt.Output, output.CaseUpper, theme.FalseError)

		brightness := "-"
		if lt.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", lt.Brightness)
		}

		power := output.FormatPowerTableValue(lt.Power)
		t.AddRow(fmt.Sprintf("%d", lt.ID), name, state, brightness, power)
	}
	printTable(ios, t)
}

// DisplayRGBList prints a table of RGB components.
func DisplayRGBList(ios *iostreams.IOStreams, rgbs []shelly.RGBInfo) {
	t := output.NewTable("ID", "Name", "State", "Color", "Brightness", "Power")
	for _, rgb := range rgbs {
		name := output.FormatComponentName(rgb.Name, "rgb", rgb.ID)
		state := output.RenderOnOff(rgb.Output, output.CaseUpper, theme.FalseError)
		color := fmt.Sprintf("R:%d G:%d B:%d", rgb.Red, rgb.Green, rgb.Blue)

		brightness := "-"
		if rgb.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", rgb.Brightness)
		}

		power := output.FormatPowerTableValue(rgb.Power)
		t.AddRow(fmt.Sprintf("%d", rgb.ID), name, state, color, brightness, power)
	}
	printTable(ios, t)
}

// DisplayRGBWList prints a table of RGBW components.
func DisplayRGBWList(ios *iostreams.IOStreams, rgbws []shelly.RGBWInfo) {
	t := output.NewTable("ID", "Name", "State", "Color", "White", "Brightness", "Power")
	for _, rgbw := range rgbws {
		name := output.FormatComponentName(rgbw.Name, "rgbw", rgbw.ID)
		state := output.RenderOnOff(rgbw.Output, output.CaseUpper, theme.FalseError)
		color := fmt.Sprintf("R:%d G:%d B:%d", rgbw.Red, rgbw.Green, rgbw.Blue)

		white := "-"
		if rgbw.White >= 0 {
			white = fmt.Sprintf("%d", rgbw.White)
		}

		brightness := "-"
		if rgbw.Brightness >= 0 {
			brightness = fmt.Sprintf("%d%%", rgbw.Brightness)
		}

		power := output.FormatPowerTableValue(rgbw.Power)
		t.AddRow(fmt.Sprintf("%d", rgbw.ID), name, state, color, white, brightness, power)
	}
	printTable(ios, t)
}

// DisplayCoverList prints a table of covers.
func DisplayCoverList(ios *iostreams.IOStreams, covers []shelly.CoverInfo) {
	t := output.NewTable("ID", "Name", "State", "Position", "Power")
	for _, cover := range covers {
		name := output.FormatComponentName(cover.Name, "cover", cover.ID)
		state := output.RenderCoverState(cover.State)

		position := "-"
		if cover.Position >= 0 {
			position = fmt.Sprintf("%d%%", cover.Position)
		}

		power := output.FormatPowerTableValue(cover.Power)
		t.AddRow(fmt.Sprintf("%d", cover.ID), name, state, position, power)
	}
	printTable(ios, t)
}

// DisplayInputList prints a table of inputs.
func DisplayInputList(ios *iostreams.IOStreams, inputs []shelly.InputInfo) {
	table := output.NewTable("ID", "Name", "Type", "State")

	for _, input := range inputs {
		name := input.Name
		if name == "" {
			name = theme.Dim().Render("-")
		}

		state := output.RenderActive(input.State, output.CaseLower, theme.FalseError)

		table.AddRow(
			fmt.Sprintf("%d", input.ID),
			name,
			input.Type,
			state,
		)
	}

	printTable(ios, table)
	ios.Println()
	ios.Count("input", len(inputs))
}
