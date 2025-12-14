// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
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
	var config map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Zigbee.GetConfig", nil)
		if err != nil {
			return fmt.Errorf("failed to get Zigbee config: %w", err)
		}

		var ok bool
		config, ok = result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		return nil
	})
	return config, err
}
