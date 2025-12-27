// Package wireless provides wireless protocol operations for Shelly devices.
package wireless

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// FetchMatterStatus fetches combined Matter config and status.
func (s *Service) FetchMatterStatus(ctx context.Context, device string, ios *iostreams.IOStreams) (model.MatterStatus, error) {
	var status model.MatterStatus

	// Get Matter config
	cfg, err := s.MatterGetConfig(ctx, device)
	if err != nil {
		ios.Debug("Matter.GetConfig failed: %v", err)
		return status, fmt.Errorf("matter not available on this device: %w", err)
	}
	status.Enabled = cfg.Enable

	// Get Matter status
	statusMap, err := s.MatterGetStatus(ctx, device)
	if err != nil {
		ios.Debug("Matter.GetStatus failed: %v", err)
		// Config succeeded, return partial info
		return status, nil
	}

	if commissionable, ok := statusMap["commissionable"].(bool); ok {
		status.Commissionable = commissionable
	}
	if fabricsCount, ok := statusMap["fabrics_count"].(float64); ok {
		status.FabricsCount = int(fabricsCount)
	}

	return status, nil
}

// MatterConfig represents Matter configuration.
type MatterConfig struct {
	Enable bool `json:"enable"`
}

// MatterEnable enables Matter on a device.
func (s *Service) MatterEnable(ctx context.Context, identifier string) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// MatterGetCommissioningCode gets the Matter commissioning/pairing code from a device.
func (s *Service) MatterGetCommissioningCode(ctx context.Context, identifier string) (model.CommissioningInfo, error) {
	var info model.CommissioningInfo
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
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
		info.Available = info.ManualCode != ""
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
