package term

import (
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
)

// DisplayEMDataHistory shows 3-phase energy meter history data.
func DisplayEMDataHistory(ios *iostreams.IOStreams, data *components.EMDataGetDataResult, id int, startTS, endTS *int64, limit int) {
	if output.WantsStructured() {
		if err := output.FormatOutput(ios.Out, data); err != nil {
			ios.DebugErr("format output", err)
		}
		return
	}

	// Human-readable format
	ios.Printf("Energy History (EM) #%d\n", id)
	if startTS != nil {
		ios.Printf("From: %s\n", time.Unix(*startTS, 0).Format(time.RFC3339))
	}
	if endTS != nil {
		ios.Printf("To:   %s\n", time.Unix(*endTS, 0).Format(time.RFC3339))
	}
	ios.Printf("\n")

	totalRecords := 0
	for _, block := range data.Data {
		totalRecords += len(block.Values)
	}

	if totalRecords == 0 {
		ios.Warning("No data available for the specified time range")
		return
	}

	ios.Printf("Total data points: %d\n", totalRecords)
	ios.Printf("Data blocks: %d\n\n", len(data.Data))

	count := 0
	for _, block := range data.Data {
		blockTime := time.Unix(block.TS, 0)
		for i, values := range block.Values {
			if limit > 0 && count >= limit {
				ios.Printf("\n(showing first %d of %d records, use --limit to see more)\n", limit, totalRecords)
				displayEMMetricsSummary(ios, data, count)
				return
			}

			recordTime := blockTime.Add(time.Duration(i*block.Period) * time.Second)
			ios.Printf("[%s] Total: %.2fW (A: %.2fW, B: %.2fW, C: %.2fW) | Voltage: A=%.1fV B=%.1fV C=%.1fV\n",
				recordTime.Format("2006-01-02 15:04:05"),
				values.TotalActivePower,
				values.AActivePower,
				values.BActivePower,
				values.CActivePower,
				values.AVoltage,
				values.BVoltage,
				values.CVoltage,
			)
			count++
		}
	}

	displayEMMetricsSummary(ios, data, count)
}

func displayEMMetricsSummary(ios *iostreams.IOStreams, data *components.EMDataGetDataResult, count int) {
	if count > 0 {
		totalEnergy, _, _, _ := monitoring.CalculateEMMetrics(data)
		ios.Printf("\nEstimated total energy consumption: %.2f kWh\n", totalEnergy)
	}
}

// DisplayEM1DataHistory shows single-phase energy meter history data.
func DisplayEM1DataHistory(ios *iostreams.IOStreams, data *components.EM1DataGetDataResult, id int, startTS, endTS *int64, limit int) {
	if output.WantsStructured() {
		if err := output.FormatOutput(ios.Out, data); err != nil {
			ios.DebugErr("format output", err)
		}
		return
	}

	// Human-readable format
	ios.Printf("Energy History (EM1) #%d\n", id)
	if startTS != nil {
		ios.Printf("From: %s\n", time.Unix(*startTS, 0).Format(time.RFC3339))
	}
	if endTS != nil {
		ios.Printf("To:   %s\n", time.Unix(*endTS, 0).Format(time.RFC3339))
	}
	ios.Printf("\n")

	totalRecords := 0
	for _, block := range data.Data {
		totalRecords += len(block.Values)
	}

	if totalRecords == 0 {
		ios.Warning("No data available for the specified time range")
		return
	}

	ios.Printf("Total data points: %d\n", totalRecords)
	ios.Printf("Data blocks: %d\n\n", len(data.Data))

	count := 0
	for _, block := range data.Data {
		blockTime := time.Unix(block.TS, 0)
		for i, values := range block.Values {
			if limit > 0 && count >= limit {
				ios.Printf("\n(showing first %d of %d records, use --limit to see more)\n", limit, totalRecords)
				displayEM1MetricsSummary(ios, data, count)
				return
			}

			recordTime := blockTime.Add(time.Duration(i*block.Period) * time.Second)
			pf := ""
			if values.PowerFactor != nil {
				pf = fmt.Sprintf(" | PF: %.3f", *values.PowerFactor)
			}
			ios.Printf("[%s] Power: %.2fW | Voltage: %.1fV | Current: %.2fA%s\n",
				recordTime.Format("2006-01-02 15:04:05"),
				values.ActivePower,
				values.Voltage,
				values.Current,
				pf,
			)
			count++
		}
	}

	displayEM1MetricsSummary(ios, data, count)
}

func displayEM1MetricsSummary(ios *iostreams.IOStreams, data *components.EM1DataGetDataResult, count int) {
	if count > 0 {
		totalEnergy, _, _, _ := monitoring.CalculateEM1Metrics(data)
		ios.Printf("\nEstimated total energy consumption: %.2f kWh\n", totalEnergy)
	}
}

// DisplayEMStatus shows 3-phase energy monitor status.
func DisplayEMStatus(ios *iostreams.IOStreams, status *model.EMStatus) {
	if output.WantsStructured() {
		if err := output.FormatOutput(ios.Out, status); err != nil {
			ios.DebugErr("format output", err)
		}
		return
	}

	// Human-readable format with bordered table
	ios.Printf("Energy Monitor (EM) #%d\n\n", status.ID)

	builder := table.NewBuilder("Metric", "Phase A", "Phase B", "Phase C", "Total")

	// Voltage row
	builder.AddRow("Voltage",
		fmt.Sprintf("%.2f V", status.AVoltage),
		fmt.Sprintf("%.2f V", status.BVoltage),
		fmt.Sprintf("%.2f V", status.CVoltage),
		"-")

	// Current row
	builder.AddRow("Current",
		fmt.Sprintf("%.2f A", status.ACurrent),
		fmt.Sprintf("%.2f A", status.BCurrent),
		fmt.Sprintf("%.2f A", status.CCurrent),
		fmt.Sprintf("%.2f A", status.TotalCurrent))

	// Active Power row
	builder.AddRow("Active Power",
		fmt.Sprintf("%.2f W", status.AActivePower),
		fmt.Sprintf("%.2f W", status.BActivePower),
		fmt.Sprintf("%.2f W", status.CActivePower),
		fmt.Sprintf("%.2f W", status.TotalActivePower))

	// Apparent Power row
	builder.AddRow("Apparent Power",
		fmt.Sprintf("%.2f VA", status.AApparentPower),
		fmt.Sprintf("%.2f VA", status.BApparentPower),
		fmt.Sprintf("%.2f VA", status.CApparentPower),
		fmt.Sprintf("%.2f VA", status.TotalAprtPower))

	// Power Factor row (optional)
	aPF, bPF, cPF := "-", "-", "-"
	if status.APowerFactor != nil {
		aPF = fmt.Sprintf("%.3f", *status.APowerFactor)
	}
	if status.BPowerFactor != nil {
		bPF = fmt.Sprintf("%.3f", *status.BPowerFactor)
	}
	if status.CPowerFactor != nil {
		cPF = fmt.Sprintf("%.3f", *status.CPowerFactor)
	}
	builder.AddRow("Power Factor", aPF, bPF, cPF, "-")

	// Frequency row (optional)
	aFreq, bFreq, cFreq := "-", "-", "-"
	if status.AFreq != nil {
		aFreq = fmt.Sprintf("%.2f Hz", *status.AFreq)
	}
	if status.BFreq != nil {
		bFreq = fmt.Sprintf("%.2f Hz", *status.BFreq)
	}
	if status.CFreq != nil {
		cFreq = fmt.Sprintf("%.2f Hz", *status.CFreq)
	}
	builder.AddRow("Frequency", aFreq, bFreq, cFreq, "-")

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print EM status table", err)
	}

	if status.NCurrent != nil {
		ios.Printf("\nNeutral Current: %.2f A\n", *status.NCurrent)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}
}

// DisplayEM1Status shows single-phase energy monitor status.
func DisplayEM1Status(ios *iostreams.IOStreams, status *model.EM1Status) {
	if output.WantsStructured() {
		if err := output.FormatOutput(ios.Out, status); err != nil {
			ios.DebugErr("format output", err)
		}
		return
	}

	// Human-readable format with bordered table
	ios.Printf("Energy Monitor (EM1) #%d\n\n", status.ID)

	builder := table.NewBuilder("Metric", "Value")
	builder.AddRow("Voltage", fmt.Sprintf("%.2f V", status.Voltage))
	builder.AddRow("Current", fmt.Sprintf("%.2f A", status.Current))
	builder.AddRow("Active Power", fmt.Sprintf("%.2f W", status.ActPower))
	builder.AddRow("Apparent Power", fmt.Sprintf("%.2f VA", status.AprtPower))
	if status.PF != nil {
		builder.AddRow("Power Factor", fmt.Sprintf("%.3f", *status.PF))
	}
	if status.Freq != nil {
		builder.AddRow("Frequency", fmt.Sprintf("%.2f Hz", *status.Freq))
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print EM1 status table", err)
	}

	if len(status.Errors) > 0 {
		ios.Printf("\nErrors: %v\n", status.Errors)
	}
}
