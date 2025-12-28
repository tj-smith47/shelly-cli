// Package modbus provides Modbus configuration for Shelly devices.
package modbus

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// Status holds Modbus status information.
type Status struct {
	Enabled bool `json:"enabled"`
}

// Config holds Modbus configuration.
type Config struct {
	Enable bool `json:"enable"`
}

// ConnectionProvider allows executing operations with a device connection.
type ConnectionProvider interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// Service provides Modbus-related operations for Shelly devices.
type Service struct {
	provider ConnectionProvider
}

// New creates a new Modbus service.
func New(provider ConnectionProvider) *Service {
	return &Service{provider: provider}
}

// GetStatus returns the Modbus status.
func (s *Service) GetStatus(ctx context.Context, identifier string) (*Status, error) {
	var result *Status
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		modbus := components.NewModbus(conn.RPCClient())
		status, err := modbus.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &Status{
			Enabled: status.Enabled,
		}
		return nil
	})
	return result, err
}

// GetConfig returns the Modbus configuration.
func (s *Service) GetConfig(ctx context.Context, identifier string) (*Config, error) {
	var result *Config
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		modbus := components.NewModbus(conn.RPCClient())
		config, err := modbus.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = &Config{
			Enable: config.Enable,
		}
		return nil
	})
	return result, err
}

// SetConfig updates the Modbus configuration.
func (s *Service) SetConfig(ctx context.Context, identifier string, enable bool) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		modbus := components.NewModbus(conn.RPCClient())
		return modbus.SetConfig(ctx, &components.ModbusConfig{
			Enable: enable,
		})
	})
}
