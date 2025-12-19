package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-go/discovery"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// FilterGen1Devices filters discovered devices by generation.
// If filterOnly is true, only Gen1 devices are returned.
// If filterOnly is false, all devices are returned but enhanced with generation info.
func FilterGen1Devices(ctx context.Context, devices []discovery.DiscoveredDevice, filterOnly bool) []discovery.DiscoveredDevice {
	var filtered []discovery.DiscoveredDevice

	for _, d := range devices {
		// Detect device generation
		result, err := client.DetectGeneration(ctx, d.Address.String(), nil)
		if err != nil {
			// Can't detect, include if not filtering
			if !filterOnly {
				filtered = append(filtered, d)
			}
			continue
		}

		if result.IsGen1() {
			filtered = append(filtered, d)
		} else if !filterOnly {
			// Include non-Gen1 if not filtering
			filtered = append(filtered, d)
		}
	}

	return filtered
}
