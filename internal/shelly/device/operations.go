// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"
	"encoding/json"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// Reboot reboots the device.
func (s *Service) Reboot(ctx context.Context, identifier string, delayMS int) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Reboot(ctx, delayMS)
	})
}

// FactoryReset performs a factory reset on the device.
func (s *Service) FactoryReset(ctx context.Context, identifier string) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.FactoryReset(ctx)
	})
}

// GetFullStatus returns the complete device status from Shelly.GetStatus (Gen2+).
// The returned map contains component names (e.g., "switch:0", "wifi", "sys") as keys
// and their status as json.RawMessage for flexible parsing.
func (s *Service) GetFullStatus(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	var result map[string]json.RawMessage
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		resp, err := conn.Call(ctx, "Shelly.GetStatus", nil)
		if err != nil {
			return err
		}

		// Marshal and unmarshal to convert any -> map[string]json.RawMessage
		data, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

// GetFullStatusGen1 returns the complete device status from /status endpoint (Gen1).
// The returned map contains status fields (e.g., "relays", "meters", "wifi_sta") as keys
// and their status as json.RawMessage for flexible parsing.
func (s *Service) GetFullStatusGen1(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	var result map[string]json.RawMessage
	err := s.parent.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		data, err := conn.Call(ctx, "/status")
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

// GetFullConfig returns the complete device config from Shelly.GetConfig (Gen2+).
func (s *Service) GetFullConfig(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	var result map[string]json.RawMessage
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		resp, err := conn.Call(ctx, "Shelly.GetConfig", nil)
		if err != nil {
			return err
		}
		data, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}

// GetFullConfigGen1 returns the complete device config from /settings endpoint (Gen1).
func (s *Service) GetFullConfigGen1(ctx context.Context, identifier string) (map[string]json.RawMessage, error) {
	var result map[string]json.RawMessage
	err := s.parent.WithGen1Connection(ctx, identifier, func(conn *client.Gen1Client) error {
		data, err := conn.Call(ctx, "/settings")
		if err != nil {
			return err
		}
		return json.Unmarshal(data, &result)
	})
	return result, err
}
