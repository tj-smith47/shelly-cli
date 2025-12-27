// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ZigbeeEnable enables Zigbee on a device.
func (s *Service) ZigbeeEnable(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Zigbee.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to enable Zigbee: %w", err)
		}
		return nil
	})
}

// ZigbeeDisable disables Zigbee on a device.
func (s *Service) ZigbeeDisable(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Zigbee.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": false,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to disable Zigbee: %w", err)
		}
		return nil
	})
}

// ZigbeeStartNetworkSteering starts Zigbee network steering (pairing mode).
func (s *Service) ZigbeeStartNetworkSteering(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Zigbee.StartNetworkSteering", nil)
		if err != nil {
			return fmt.Errorf("failed to start network steering: %w", err)
		}
		return nil
	})
}

// ZigbeeGetStatus gets Zigbee status from a device.
func (s *Service) ZigbeeGetStatus(ctx context.Context, identifier string) (map[string]any, error) {
	var status map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Zigbee.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get Zigbee status: %w", err)
		}

		var ok bool
		status, ok = result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		return nil
	})
	return status, err
}

// ZigbeeGetConfig gets Zigbee configuration from a device.
func (s *Service) ZigbeeGetConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var cfg map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Zigbee.GetConfig", nil)
		if err != nil {
			return fmt.Errorf("failed to get Zigbee config: %w", err)
		}

		var ok bool
		cfg, ok = result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		return nil
	})
	return cfg, err
}

// ScanZigbeeDevices scans configured devices for Zigbee support.
func (s *Service) ScanZigbeeDevices(ctx context.Context, ios *iostreams.IOStreams) ([]model.ZigbeeDevice, error) {
	cfg := config.Get()
	if cfg == nil {
		return nil, fmt.Errorf("no configuration found; run 'shelly init' first")
	}

	devices := make([]model.ZigbeeDevice, 0, len(cfg.Devices))

	for name, dev := range cfg.Devices {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		device, ok := s.checkZigbeeSupport(ctx, name, dev, ios)
		if !ok {
			continue
		}

		if device.Enabled {
			s.enrichZigbeeStatus(ctx, dev.Address, &device)
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func (s *Service) checkZigbeeSupport(ctx context.Context, name string, dev model.Device, ios *iostreams.IOStreams) (model.ZigbeeDevice, bool) {
	result, err := s.RawRPC(ctx, dev.Address, "Zigbee.GetConfig", nil)
	if err != nil {
		ios.Debug("device %s does not support Zigbee: %v", name, err)
		return model.ZigbeeDevice{}, false
	}

	var cfg struct {
		Enable bool `json:"enable"`
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		ios.Debug("failed to marshal result for %s: %v", name, err)
		return model.ZigbeeDevice{}, false
	}
	if json.Unmarshal(jsonBytes, &cfg) != nil {
		return model.ZigbeeDevice{}, false
	}

	return model.ZigbeeDevice{
		Name:    name,
		Address: dev.Address,
		Model:   dev.Model,
		Enabled: cfg.Enable,
	}, true
}

func (s *Service) enrichZigbeeStatus(ctx context.Context, address string, device *model.ZigbeeDevice) {
	statusResult, err := s.RawRPC(ctx, address, "Zigbee.GetStatus", nil)
	if err != nil {
		return
	}

	var status struct {
		NetworkState string `json:"network_state"`
		EUI64        string `json:"eui64"`
	}
	statusBytes, err := json.Marshal(statusResult)
	if err != nil {
		return
	}
	if json.Unmarshal(statusBytes, &status) == nil {
		device.NetworkState = status.NetworkState
		device.EUI64 = status.EUI64
	}
}

// FetchZigbeeStatus fetches the Zigbee status for a specific device.
func (s *Service) FetchZigbeeStatus(ctx context.Context, device string, ios *iostreams.IOStreams) (model.ZigbeeStatus, error) {
	var status model.ZigbeeStatus

	// Get config
	cfg, err := s.ZigbeeGetConfig(ctx, device)
	if err != nil {
		return status, fmt.Errorf("zigbee not available on this device: %w", err)
	}
	if enable, ok := cfg["enable"].(bool); ok {
		status.Enabled = enable
	}

	// Get status
	st, err := s.ZigbeeGetStatus(ctx, device)
	if err != nil {
		ios.Debug("Zigbee.GetStatus failed: %v", err)
		return status, nil // Config succeeded, return partial info
	}

	if networkState, ok := st["network_state"].(string); ok {
		status.NetworkState = networkState
	}
	if eui64, ok := st["eui64"].(string); ok {
		status.EUI64 = eui64
	}
	if panID, ok := st["pan_id"].(float64); ok {
		status.PANID = uint16(panID)
	}
	if channel, ok := st["channel"].(float64); ok {
		status.Channel = int(channel)
	}
	if coordinatorEUI64, ok := st["coordinator_eui64"].(string); ok {
		status.CoordinatorEUI64 = coordinatorEUI64
	}

	return status, nil
}
