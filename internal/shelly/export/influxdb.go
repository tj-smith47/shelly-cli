// Package export provides business logic for exporting Shelly device data.
package export

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// InfluxDBPoint represents a single InfluxDB line protocol point.
type InfluxDBPoint struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]float64
	Timestamp   time.Time
}

// ReadingsToInfluxDBPoints converts ComponentReadings to InfluxDB points.
func ReadingsToInfluxDBPoints(readings []model.ComponentReading, timestamp time.Time) []InfluxDBPoint {
	points := make([]InfluxDBPoint, 0, len(readings))
	for _, r := range readings {
		tags := map[string]string{"device": r.Device, "component": r.Type, "component_id": fmt.Sprintf("%d", r.ID)}
		if r.Phase != "" {
			tags["phase"] = r.Phase
		}
		fields := map[string]float64{"power": r.Power, "voltage": r.Voltage, "current": r.Current}
		if r.Energy != nil {
			fields["energy"] = *r.Energy
		}
		if r.Freq != nil {
			fields["frequency"] = *r.Freq
		}
		points = append(points, InfluxDBPoint{Measurement: "shelly", Tags: tags, Fields: fields, Timestamp: timestamp})
	}
	return points
}

// FormatInfluxDBLineProtocol formats points as InfluxDB line protocol.
// Each line follows: measurement,tag1=value1,tag2=value2 field1=value1,field2=value2 timestamp.
func FormatInfluxDBLineProtocol(points []InfluxDBPoint) string {
	var result string
	for _, p := range points {
		result += FormatInfluxDBPoint(p) + "\n"
	}
	return result
}

// FormatInfluxDBPoint formats a single point as InfluxDB line protocol.
// Returns the line without a trailing newline.
func FormatInfluxDBPoint(p InfluxDBPoint) string {
	// Build tags string (sorted for consistent output)
	tagParts := make([]string, 0, len(p.Tags))
	for k, v := range p.Tags {
		tagParts = append(tagParts, fmt.Sprintf("%s=%s", EscapeInfluxTag(k), EscapeInfluxTag(v)))
	}
	sort.Strings(tagParts)
	tagsStr := strings.Join(tagParts, ",")

	// Build fields string (sorted for consistent output)
	fieldParts := make([]string, 0, len(p.Fields))
	for k, v := range p.Fields {
		fieldParts = append(fieldParts, fmt.Sprintf("%s=%g", EscapeInfluxTag(k), v))
	}
	sort.Strings(fieldParts)
	fieldsStr := strings.Join(fieldParts, ",")

	// Format: measurement,tags fields timestamp
	if tagsStr != "" {
		return fmt.Sprintf("%s,%s %s %d", p.Measurement, tagsStr, fieldsStr, p.Timestamp.UnixNano())
	}
	return fmt.Sprintf("%s %s %d", p.Measurement, fieldsStr, p.Timestamp.UnixNano())
}

// EscapeInfluxTag escapes special characters in InfluxDB tag keys/values.
func EscapeInfluxTag(s string) string {
	s = strings.ReplaceAll(s, " ", "\\ ")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

// ParseTags converts key=value pairs to a tag map.
func ParseTags(tagPairs []string) map[string]string {
	tags := make(map[string]string)
	for _, pair := range tagPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			tags[parts[0]] = parts[1]
		}
	}
	return tags
}

// EMReadingsToInfluxDBPoints converts EM readings to InfluxDB points with phase labels.
func EMReadingsToInfluxDBPoints(emStatuses []*model.EMStatus, device string, timestamp time.Time) []InfluxDBPoint {
	points := make([]InfluxDBPoint, 0, len(emStatuses)*4)
	for _, status := range emStatuses {
		compID := fmt.Sprintf("%d", status.ID)
		points = append(points,
			// Phase A
			InfluxDBPoint{
				Measurement: "shelly_power",
				Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "a"},
				Fields:      map[string]float64{"power": status.AActivePower, "voltage": status.AVoltage, "current": status.ACurrent},
				Timestamp:   timestamp,
			},
			// Phase B
			InfluxDBPoint{
				Measurement: "shelly_power",
				Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "b"},
				Fields:      map[string]float64{"power": status.BActivePower, "voltage": status.BVoltage, "current": status.BCurrent},
				Timestamp:   timestamp,
			},
			// Phase C
			InfluxDBPoint{
				Measurement: "shelly_power",
				Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "c"},
				Fields:      map[string]float64{"power": status.CActivePower, "voltage": status.CVoltage, "current": status.CCurrent},
				Timestamp:   timestamp,
			},
			// Total
			InfluxDBPoint{
				Measurement: "shelly_power",
				Tags:        map[string]string{"device": device, "component": "em", "component_id": compID, "phase": "total"},
				Fields:      map[string]float64{"power": status.TotalActivePower, "current": status.TotalCurrent},
				Timestamp:   timestamp,
			},
		)
	}
	return points
}

// CollectMeterReadings is a generic collector for single-phase meter types.
func CollectMeterReadings[T model.MeterReading](
	device, compType string,
	ids []int,
	getFunc func(id int) (T, error),
) []model.ComponentReading {
	readings := make([]model.ComponentReading, 0, len(ids))
	for _, id := range ids {
		status, err := getFunc(id)
		if err != nil {
			continue
		}
		readings = append(readings, model.ComponentReading{
			Device:  device,
			Type:    compType,
			ID:      id,
			Power:   status.GetPower(),
			Voltage: status.GetVoltage(),
			Current: status.GetCurrent(),
			Energy:  status.GetEnergy(),
			Freq:    status.GetFreq(),
		})
	}
	return readings
}

// CollectEMReadings collects 3-phase EM readings (each phase as separate reading).
func CollectEMReadings(device string, emStatuses []*model.EMStatus) []model.ComponentReading {
	readings := make([]model.ComponentReading, 0, len(emStatuses)*4) // 4 readings per EM (3 phases + total)
	for _, status := range emStatuses {
		base := model.ComponentReading{Device: device, Type: "em", ID: status.ID}
		readings = append(readings,
			model.ComponentReading{Device: base.Device, Type: base.Type, ID: base.ID, Phase: "a", Power: status.AActivePower, Voltage: status.AVoltage, Current: status.ACurrent, Freq: status.AFreq},
			model.ComponentReading{Device: base.Device, Type: base.Type, ID: base.ID, Phase: "b", Power: status.BActivePower, Voltage: status.BVoltage, Current: status.BCurrent, Freq: status.BFreq},
			model.ComponentReading{Device: base.Device, Type: base.Type, ID: base.ID, Phase: "c", Power: status.CActivePower, Voltage: status.CVoltage, Current: status.CCurrent, Freq: status.CFreq},
			model.ComponentReading{Device: base.Device, Type: base.Type, ID: base.ID, Phase: "total", Power: status.TotalActivePower, Current: status.TotalCurrent},
		)
	}
	return readings
}
