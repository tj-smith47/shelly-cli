// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-cli/internal/errutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly/monitoring"
)

// Energy component type constants for auto-detection.
const (
	ComponentTypeAuto = "auto"
	ComponentTypeEM   = monitoring.EnergyTypeEM
	ComponentTypeEM1  = monitoring.EnergyTypeEM1
)

// Energy aggregation period names.
const (
	periodHour  = "hour"
	periodDay   = "day"
	periodWeek  = "week"
	periodMonth = "month"
)

// DetectEnergyComponentType finds which stored-history component, EMData or
// EM1Data, a device keeps for the given ID. Only EM and EM1 energy monitors
// keep history; for any other device, Gen1 included, the error says so and names the command
// that reads its live power and energy totals. An unreachable device returns
// the connection error.
func (s *Service) DetectEnergyComponentType(ctx context.Context, ios *iostreams.IOStreams, device string, id int) (string, error) {
	noHistory := fmt.Errorf("%s keeps no energy history: only EM and EM1 energy monitors "+
		"(such as the Pro 3EM and Pro EM) store it; for its live power and energy totals "+
		"use 'shelly energy status %s'", device, device)
	if dev, err := s.ResolveWithGeneration(ctx, device); err == nil && dev.Generation == 1 {
		return "", noHistory
	}

	emRecords, emErr := s.GetEMDataRecords(ctx, device, id, nil)
	if emErr == nil && emRecords != nil && len(emRecords.Records) > 0 {
		return ComponentTypeEM, nil
	}
	ios.DebugErr("get EMData records", emErr)

	em1Records, em1Err := s.GetEM1DataRecords(ctx, device, id, nil)
	if em1Err == nil && em1Records != nil && len(em1Records.Records) > 0 {
		return ComponentTypeEM1, nil
	}
	ios.DebugErr("get EM1Data records", em1Err)

	switch {
	case emErr == nil || em1Err == nil:
		return "", fmt.Errorf("%s has no stored energy records for component %d yet; "+
			"for its live power and energy totals use 'shelly energy status %s'", device, id, device)
	case errutil.IsNotAvailable(emErr) && errutil.IsNotAvailable(em1Err):
		return "", noHistory
	case !errutil.IsNotAvailable(em1Err):
		return "", fmt.Errorf("failed to read energy history: %w", em1Err)
	default:
		return "", fmt.Errorf("failed to read energy history: %w", emErr)
	}
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
	case periodHour:
		start = now.Add(-1 * time.Hour)
	case periodDay, "":
		start = now.Add(-24 * time.Hour)
	case periodWeek:
		start = now.Add(-7 * 24 * time.Hour)
	case periodMonth:
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
