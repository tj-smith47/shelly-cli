package term

import (
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		totalEnergy, _, _, _ := shelly.CalculateEMMetrics(data)
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
		totalEnergy, _, _, _ := shelly.CalculateEM1Metrics(data)
		ios.Printf("\nEstimated total energy consumption: %.2f kWh\n", totalEnergy)
	}
}
