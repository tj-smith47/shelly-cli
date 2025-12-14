// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// MatterConfig represents Matter configuration.
type MatterConfig struct {
	Enable bool `json:"enable"`
}

// MatterStatus represents Matter status from the device.
type MatterStatus struct {
	Commissioned bool   `json:"commissioned"`
	SetupCode    string `json:"setup_code,omitempty"`
}

// MatterEnable enables Matter on a device.
func (s *Service) MatterEnable(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Matter.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to enable Matter: %w", err)
		}
		return nil
	})
}

// MatterDisable disables Matter on a device.
func (s *Service) MatterDisable(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Matter.SetConfig", map[string]any{
			"config": map[string]any{
				"enable": false,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to disable Matter: %w", err)
		}
		return nil
	})
}

// MatterReset performs a factory reset of Matter on a device (decommissions it).
func (s *Service) MatterReset(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		_, err := conn.Call(ctx, "Matter.FactoryReset", nil)
		if err != nil {
			return fmt.Errorf("failed to reset Matter: %w", err)
		}
		return nil
	})
}

// MatterGetSetupCode gets the Matter pairing/setup code from a device.
func (s *Service) MatterGetSetupCode(ctx context.Context, identifier string) (string, error) {
	var code string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Matter.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get Matter status: %w", err)
		}

		status, ok := result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if setupCode, ok := status["setup_code"].(string); ok {
			code = setupCode
		}
		return nil
	})
	return code, err
}

// MatterGetStatus gets the full Matter status from a device.
func (s *Service) MatterGetStatus(ctx context.Context, identifier string) (map[string]any, error) {
	var status map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Matter.GetStatus", nil)
		if err != nil {
			return fmt.Errorf("failed to get Matter status: %w", err)
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

// MatterGetConfig gets the Matter configuration from a device.
func (s *Service) MatterGetConfig(ctx context.Context, identifier string) (MatterConfig, error) {
	var cfg MatterConfig
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Matter.GetConfig", nil)
		if err != nil {
			return fmt.Errorf("failed to get Matter config: %w", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if enable, ok := resultMap["enable"].(bool); ok {
			cfg.Enable = enable
		}
		return nil
	})
	return cfg, err
}

// MatterCommissioningInfo holds Matter pairing information.
type MatterCommissioningInfo struct {
	ManualCode    string `json:"manual_code,omitempty"`
	QRCode        string `json:"qr_code,omitempty"`
	Discriminator int    `json:"discriminator,omitempty"`
	SetupPINCode  int    `json:"setup_pin_code,omitempty"`
}

// MatterGetCommissioningCode gets the Matter commissioning/pairing code from a device.
func (s *Service) MatterGetCommissioningCode(ctx context.Context, identifier string) (MatterCommissioningInfo, error) {
	var info MatterCommissioningInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		result, err := conn.Call(ctx, "Matter.GetCommissioningCode", nil)
		if err != nil {
			return fmt.Errorf("failed to get Matter commissioning code: %w", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}

		if code, ok := resultMap["manual_code"].(string); ok {
			info.ManualCode = code
		}
		if qr, ok := resultMap["qr_code"].(string); ok {
			info.QRCode = qr
		}
		if disc, ok := resultMap["discriminator"].(float64); ok {
			info.Discriminator = int(disc)
		}
		if pin, ok := resultMap["setup_pin_code"].(float64); ok {
			info.SetupPINCode = int(pin)
		}
		return nil
	})
	return info, err
}

// MatterIsCommissionable checks if a device is commissionable (ready to be added to a fabric).
func (s *Service) MatterIsCommissionable(ctx context.Context, identifier string) (bool, error) {
	status, err := s.MatterGetStatus(ctx, identifier)
	if err != nil {
		return false, err
	}
	commissionable, ok := status["commissionable"].(bool)
	if !ok {
		return false, nil
	}
	return commissionable, nil
}
