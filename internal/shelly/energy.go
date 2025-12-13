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
	// Try RFC3339 first
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}

	// Try date-only format (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}

	// Try datetime format (YYYY-MM-DD HH:MM:SS)
	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time (use RFC3339, YYYY-MM-DD, or 'YYYY-MM-DD HH:MM:SS')")
}
