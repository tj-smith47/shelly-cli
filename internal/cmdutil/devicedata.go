package cmdutil

import (
	"context"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// DeviceData holds collected device information from config and live connection.
type DeviceData struct {
	Name       string
	Address    string
	Model      string
	Generation int
	App        string
	Online     bool
}

// CollectDeviceData gathers device information for the given device names.
// It first checks config for static data, then tries to connect to get live data.
// Devices that can't be reached use config data with Online=false.
func CollectDeviceData(ctx context.Context, svc *shelly.Service, deviceNames []string) []DeviceData {
	result := make([]DeviceData, 0, len(deviceNames))

	for _, device := range deviceNames {
		data := DeviceData{Name: device}

		// Get device config for address and static data
		deviceCfg, exists := config.GetDevice(device)
		if exists {
			data.Address = deviceCfg.Address
			data.Model = deviceCfg.Model
			data.Generation = deviceCfg.Generation
		}

		// Try to connect and get live data
		err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
			info := conn.Info()
			data.Model = info.Model
			data.Generation = info.Generation
			data.App = info.App
			data.Online = true
			return nil
		})

		if err != nil {
			// Device offline - keep config data, Online stays false
			data.Online = false
			// Only add if we have config data
			if !exists {
				continue
			}
		}

		result = append(result, data)
	}

	return result
}

// SplitDevicesAndFile splits command args into device names and an optional file path.
// If the last argument ends with one of the valid extensions, it's treated as the file path.
func SplitDevicesAndFile(args, validExtensions []string) (deviceNames []string, filePath string) {
	if len(args) <= 1 {
		return args, ""
	}

	lastArg := args[len(args)-1]
	for _, ext := range validExtensions {
		if strings.HasSuffix(lastArg, ext) {
			return args[:len(args)-1], lastArg
		}
	}

	return args, ""
}
