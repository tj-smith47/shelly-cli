// Package export provides business logic for exporting Shelly device data.
package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/tj-smith47/shelly-go/gen2/components"
)

// EMDataCSVHeaders defines the CSV header row for 3-phase energy meter data.
var EMDataCSVHeaders = []string{
	"timestamp",
	"a_voltage", "a_current", "a_act_power", "a_aprt_power", "a_pf", "a_freq",
	"b_voltage", "b_current", "b_act_power", "b_aprt_power", "b_pf", "b_freq",
	"c_voltage", "c_current", "c_act_power", "c_aprt_power", "c_pf", "c_freq",
	"total_current", "total_act_power", "total_aprt_power",
	"total_act_energy", "total_act_ret_energy",
	"n_current",
}

// EM1DataCSVHeaders defines the CSV header row for single-phase energy meter data.
var EM1DataCSVHeaders = []string{
	"timestamp",
	"voltage", "current", "act_power", "aprt_power", "pf", "freq",
	"act_energy", "act_ret_energy",
}

// dataBlock represents a generic data block with timestamp and period.
type dataBlock[V any] struct {
	TS     int64
	Period int
	Values []V
}

// formatDataCSV is a generic CSV formatter for energy data.
// It extracts the common logic for formatting timestamped data blocks.
func formatDataCSV[V any](
	headers []string,
	blocks []dataBlock[V],
	rowFormatter func(timestamp string, v V) []string,
) ([]byte, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

	if err := csvWriter.Write(headers); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for _, block := range blocks {
		period := int64(block.Period)
		for i, v := range block.Values {
			measurementTS := block.TS + int64(i)*period
			timeStr := time.Unix(measurementTS, 0).UTC().Format(time.RFC3339)

			if err := csvWriter.Write(rowFormatter(timeStr, v)); err != nil {
				return nil, fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("CSV write error: %w", err)
	}

	return buf.Bytes(), nil
}

// FormatEMDataCSV converts 3-phase EMData to CSV format.
// It returns the CSV data as bytes, ready to be written to a file or stdout.
func FormatEMDataCSV(data *components.EMDataGetDataResult) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("no data to export")
	}

	blocks := make([]dataBlock[components.EMDataValues], len(data.Data))
	for i, b := range data.Data {
		blocks[i] = dataBlock[components.EMDataValues]{
			TS:     b.TS,
			Period: b.Period,
			Values: b.Values,
		}
	}

	return formatDataCSV(EMDataCSVHeaders, blocks, formatEMDataRow)
}

// FormatEM1DataCSV converts single-phase EM1Data to CSV format.
// It returns the CSV data as bytes, ready to be written to a file or stdout.
func FormatEM1DataCSV(data *components.EM1DataGetDataResult) ([]byte, error) {
	if data == nil {
		return nil, fmt.Errorf("no data to export")
	}

	blocks := make([]dataBlock[components.EM1DataValues], len(data.Data))
	for i, b := range data.Data {
		blocks[i] = dataBlock[components.EM1DataValues]{
			TS:     b.TS,
			Period: b.Period,
			Values: b.Values,
		}
	}

	return formatDataCSV(EM1DataCSVHeaders, blocks, formatEM1DataRow)
}

// formatEMDataRow formats a single 3-phase measurement as a CSV row.
func formatEMDataRow(timestamp string, v components.EMDataValues) []string {
	return []string{
		timestamp,
		FormatFloat(v.AVoltage),
		FormatFloat(v.ACurrent),
		FormatFloat(v.AActivePower),
		FormatFloat(v.AApparentPower),
		FormatFloatPtr(v.APowerFactor),
		FormatFloatPtr(v.AFreq),
		FormatFloat(v.BVoltage),
		FormatFloat(v.BCurrent),
		FormatFloat(v.BActivePower),
		FormatFloat(v.BApparentPower),
		FormatFloatPtr(v.BPowerFactor),
		FormatFloatPtr(v.BFreq),
		FormatFloat(v.CVoltage),
		FormatFloat(v.CCurrent),
		FormatFloat(v.CActivePower),
		FormatFloat(v.CApparentPower),
		FormatFloatPtr(v.CPowerFactor),
		FormatFloatPtr(v.CFreq),
		FormatFloat(v.TotalCurrent),
		FormatFloat(v.TotalActivePower),
		FormatFloat(v.TotalAprtPower),
		FormatFloatPtr(v.TotalActEnergy),
		FormatFloatPtr(v.TotalActRetEnergy),
		FormatFloatPtr(v.NCurrent),
	}
}

// formatEM1DataRow formats a single single-phase measurement as a CSV row.
func formatEM1DataRow(timestamp string, v components.EM1DataValues) []string {
	return []string{
		timestamp,
		FormatFloat(v.Voltage),
		FormatFloat(v.Current),
		FormatFloat(v.ActivePower),
		FormatFloat(v.ApparentPower),
		FormatFloatPtr(v.PowerFactor),
		FormatFloatPtr(v.Freq),
		FormatFloatPtr(v.ActEnergy),
		FormatFloatPtr(v.ActRetEnergy),
	}
}

// FormatFloat formats a float64 value for CSV export.
// It uses automatic precision to avoid unnecessary trailing zeros.
func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// FormatFloatPtr formats a *float64 value for CSV export.
// It returns an empty string for nil values.
func FormatFloatPtr(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}
