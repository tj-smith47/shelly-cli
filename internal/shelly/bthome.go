// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// BTHomeDeviceStatus represents BTHome device status from Shelly.GetStatus.
type BTHomeDeviceStatus struct {
	ID         int     `json:"id"`
	RSSI       *int    `json:"rssi"`
	Battery    *int    `json:"battery"`
	LastUpdate float64 `json:"last_updated_ts"`
}

// CollectBTHomeDevices collects BTHomeDevice components from device status.
// Returns a slice of BTHomeDeviceStatus for all bthomedevice:* keys in the status map.
func CollectBTHomeDevices(status map[string]json.RawMessage, ios *iostreams.IOStreams) []BTHomeDeviceStatus {
	devices := make([]BTHomeDeviceStatus, 0, len(status))

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomedevice:") {
			continue
		}

		var devStatus BTHomeDeviceStatus
		if err := json.Unmarshal(raw, &devStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome device %s: %v", key, err)
			}
			continue
		}

		devices = append(devices, devStatus)
	}

	return devices
}

// BTHomeSensorStatus represents BTHome sensor status from Shelly.GetStatus.
type BTHomeSensorStatus struct {
	ID         int     `json:"id"`
	RSSI       *int    `json:"rssi"`
	Battery    *int    `json:"battery"`
	LastUpdate float64 `json:"last_updated_ts"`
}

// CollectBTHomeSensors collects BTHomeSensor components from device status.
// Returns a slice of BTHomeSensorStatus for all bthomesensor:* keys in the status map.
func CollectBTHomeSensors(status map[string]json.RawMessage, ios *iostreams.IOStreams) []BTHomeSensorStatus {
	sensors := make([]BTHomeSensorStatus, 0, len(status))

	for key, raw := range status {
		if !strings.HasPrefix(key, "bthomesensor:") {
			continue
		}

		var sensorStatus BTHomeSensorStatus
		if err := json.Unmarshal(raw, &sensorStatus); err != nil {
			if ios != nil {
				ios.Debug("failed to unmarshal BTHome sensor %s: %v", key, err)
			}
			continue
		}

		sensors = append(sensors, sensorStatus)
	}

	return sensors
}

// BTHomeAddDeviceResult represents the result of adding a BTHome device.
type BTHomeAddDeviceResult struct {
	Key string `json:"key"`
}

// BTHomeAddDevice adds a BTHome device to a Shelly gateway.
func (s *Service) BTHomeAddDevice(ctx context.Context, identifier, addr, name string) (BTHomeAddDeviceResult, error) {
	var result BTHomeAddDeviceResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		config := map[string]any{
			"addr": addr,
		}
		if name != "" {
			config["name"] = name
		}
		params := map[string]any{
			"config": config,
		}

		resultAny, err := conn.Call(ctx, "BTHome.AddDevice", params)
		if err != nil {
			return fmt.Errorf("failed to add BTHome device: %w", err)
		}

		resultMap, ok := resultAny.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if key, ok := resultMap["key"].(string); ok {
			result.Key = key
		}
		return nil
	})
	return result, err
}

// BTHomeStartDiscovery starts BTHome device discovery on a gateway.
func (s *Service) BTHomeStartDiscovery(ctx context.Context, identifier string, duration int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"duration": duration,
		}
		_, err := conn.Call(ctx, "BTHome.StartDeviceDiscovery", params)
		if err != nil {
			return fmt.Errorf("failed to start BTHome discovery: %w", err)
		}
		return nil
	})
}

// BTHomeRemoveDevice removes a BTHome device from a gateway.
func (s *Service) BTHomeRemoveDevice(ctx context.Context, identifier string, deviceID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"id": deviceID,
		}
		_, err := conn.Call(ctx, "BTHome.DeleteDevice", params)
		if err != nil {
			return fmt.Errorf("failed to remove BTHome device: %w", err)
		}
		return nil
	})
}
