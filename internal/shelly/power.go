// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Power meter component type constants for auto-detection.
const (
	ComponentTypePM  = "pm"
	ComponentTypePM1 = "pm1"
)

// DetectPowerComponentType auto-detects whether a device uses PM or PM1 components.
// It checks which component types are available for the given device and ID.
// Returns ComponentTypeAuto if no power meter components are found.
func (s *Service) DetectPowerComponentType(ctx context.Context, ios *iostreams.IOStreams, device string, id int) string {
	pmIDs, err := s.ListPMComponents(ctx, device)
	ios.DebugErr("list PM components", err)
	pm1IDs, err := s.ListPM1Components(ctx, device)
	ios.DebugErr("list PM1 components", err)

	// Check if ID matches PM component
	for _, pmID := range pmIDs {
		if pmID == id {
			return ComponentTypePM
		}
	}

	// Check if ID matches PM1 component
	for _, pm1ID := range pm1IDs {
		if pm1ID == id {
			return ComponentTypePM1
		}
	}

	// Default to first available type
	if len(pmIDs) > 0 {
		return ComponentTypePM
	}
	if len(pm1IDs) > 0 {
		return ComponentTypePM1
	}

	return ComponentTypeAuto
}
