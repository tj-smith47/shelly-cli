// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Energy component type constants for auto-detection.
const (
	ComponentTypeAuto = "auto"
	ComponentTypeEM   = "em"
	ComponentTypeEM1  = "em1"
)

// DetectEnergyComponentByID auto-detects the energy component type by checking
// which component list contains the given ID. Returns ComponentTypeAuto if no match found.
// If detection fails, a warning is logged via ios.
func (s *Service) DetectEnergyComponentByID(ctx context.Context, ios *iostreams.IOStreams, device string, id int) string {
	emIDs, emErr := s.ListEMComponents(ctx, device)
	if emErr != nil {
		ios.DebugErr("list EM components", emErr)
	}
	em1IDs, em1Err := s.ListEM1Components(ctx, device)
	if em1Err != nil {
		ios.DebugErr("list EM1 components", em1Err)
	}

	// Check if ID matches EM component
	for _, emID := range emIDs {
		if emID == id {
			return ComponentTypeEM
		}
	}

	// Check if ID matches EM1 component
	for _, em1ID := range em1IDs {
		if em1ID == id {
			return ComponentTypeEM1
		}
	}

	// Default to first available type
	if len(emIDs) > 0 {
		return ComponentTypeEM
	}
	if len(em1IDs) > 0 {
		return ComponentTypeEM1
	}

	// Detection failed - warn if both list operations returned errors
	if emErr != nil && em1Err != nil {
		ios.Warning("Could not detect energy component type: device may be offline or have no energy monitoring")
	}

	return ComponentTypeAuto
}

// DetectEnergyComponentType auto-detects whether a device uses EM or EM1 data components.
// It probes the device for EMData and EM1Data records and returns the appropriate type.
// Returns an error if no energy data components are found.
func (s *Service) DetectEnergyComponentType(ctx context.Context, ios *iostreams.IOStreams, device string, id int) (string, error) {
	// Try EMData first
	emRecords, err := s.GetEMDataRecords(ctx, device, id, nil)
	if err == nil && emRecords != nil && len(emRecords.Records) > 0 {
		return ComponentTypeEM, nil
	}
	ios.DebugErr("get EMData records", err)

	// Try EM1Data
	em1Records, err := s.GetEM1DataRecords(ctx, device, id, nil)
	if err == nil && em1Records != nil && len(em1Records.Records) > 0 {
		return ComponentTypeEM1, nil
	}
	ios.DebugErr("get EM1Data records", err)

	return "", fmt.Errorf("no energy data components found")
}

// CalculateTimeRange converts period/from/to flags to Unix timestamps.
// It supports predefined periods (hour, day, week, month) or explicit from/to times.
// Returns nil pointers if no time range is specified (empty period and no from/to).
func CalculateTimeRange(period, from, to string) (startTS, endTS *int64, err error) {
	// If explicit from/to provided, use those
	if from != "" || to != "" {
		return parseExplicitTimeRange(from, to)
	}

	// Calculate based on period
	now := time.Now()
	var start time.Time

	switch period {
	case "hour":
		start = now.Add(-1 * time.Hour)
	case "day", "":
		start = now.Add(-24 * time.Hour)
	case "week":
		start = now.Add(-7 * 24 * time.Hour)
	case "month":
		start = now.Add(-30 * 24 * time.Hour)
	default:
		return nil, nil, fmt.Errorf("invalid period: %s (use: hour, day, week, month)", period)
	}

	startUnix := start.Unix()
	endUnix := now.Unix()
	return &startUnix, &endUnix, nil
}

// parseExplicitTimeRange parses explicit from/to time strings into Unix timestamps.
func parseExplicitTimeRange(from, to string) (startTS, endTS *int64, err error) {
	if from != "" {
		t, err := ParseTime(from)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --from time: %w", err)
		}
		ts := t.Unix()
		startTS = &ts
	}
	if to != "" {
		t, err := ParseTime(to)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid --to time: %w", err)
		}
		ts := t.Unix()
		endTS = &ts
	}
	return startTS, endTS, nil
}

// ParseTime parses a time string in various formats.
// Supported formats: RFC3339, YYYY-MM-DD, YYYY-MM-DD HH:MM:SS.
func ParseTime(s string) (time.Time, error) {
	formats := []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time (use RFC3339, YYYY-MM-DD, or 'YYYY-MM-DD HH:MM:SS')")
}
