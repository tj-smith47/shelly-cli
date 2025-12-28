// Package sensoraddon provides Sensor Add-on management for Shelly devices.
package sensoraddon

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// PeripheralType represents the type of sensor add-on peripheral.
type PeripheralType = components.PeripheralType

// Peripheral type constants.
const (
	TypeDS18B20   = components.PeripheralTypeDS18B20
	TypeDHT22     = components.PeripheralTypeDHT22
	TypeDigitalIn = components.PeripheralTypeDigitalIn
	TypeAnalogIn  = components.PeripheralTypeAnalogIn
)

// ValidPeripheralTypes is the list of valid peripheral types.
var ValidPeripheralTypes = []string{
	string(TypeDS18B20),
	string(TypeDHT22),
	string(TypeDigitalIn),
	string(TypeAnalogIn),
}

// Peripheral represents a configured peripheral.
type Peripheral struct {
	Type      PeripheralType `json:"type"`
	Component string         `json:"component"`
	Addr      *string        `json:"addr,omitempty"`
}

// OneWireDevice represents a discovered OneWire device.
type OneWireDevice struct {
	Type      string  `json:"type"`
	Addr      string  `json:"addr"`
	Component *string `json:"component,omitempty"`
}

// AddOptions holds options for adding a peripheral.
type AddOptions struct {
	CID  *int
	Addr *string
}

// ConnectionProvider allows executing operations with a device connection.
type ConnectionProvider interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// Service provides Sensor Add-on operations.
type Service struct {
	provider ConnectionProvider
}

// New creates a new Sensor Add-on service.
func New(provider ConnectionProvider) *Service {
	return &Service{provider: provider}
}

// ListPeripherals returns all configured peripherals.
func (s *Service) ListPeripherals(ctx context.Context, identifier string) ([]Peripheral, error) {
	var results []Peripheral

	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		addon := components.NewSensorAddon(conn.RPCClient())
		resp, err := addon.GetPeripherals(ctx)
		if err != nil {
			return err
		}

		for pType, comps := range resp {
			for compKey, info := range comps {
				results = append(results, Peripheral{
					Type:      pType,
					Component: compKey,
					Addr:      info.Addr,
				})
			}
		}
		return nil
	})

	return results, err
}

// AddPeripheral adds a new peripheral.
func (s *Service) AddPeripheral(ctx context.Context, identifier string, pType PeripheralType, opts *AddOptions) (map[string]any, error) {
	var result map[string]any

	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		addon := components.NewSensorAddon(conn.RPCClient())

		var attrs *components.AddPeripheralAttrs
		if opts != nil && (opts.CID != nil || opts.Addr != nil) {
			attrs = &components.AddPeripheralAttrs{
				CID:  opts.CID,
				Addr: opts.Addr,
			}
		}

		resp, err := addon.AddPeripheral(ctx, pType, attrs)
		if err != nil {
			return err
		}

		// Flatten the response to just component keys
		result = make(map[string]any)
		for compKey := range resp {
			result[compKey] = struct{}{}
		}
		return nil
	})

	return result, err
}

// RemovePeripheral removes a peripheral.
func (s *Service) RemovePeripheral(ctx context.Context, identifier, component string) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		addon := components.NewSensorAddon(conn.RPCClient())
		return addon.RemovePeripheral(ctx, component)
	})
}

// ScanOneWire scans for OneWire devices.
func (s *Service) ScanOneWire(ctx context.Context, identifier string) ([]OneWireDevice, error) {
	var results []OneWireDevice

	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		addon := components.NewSensorAddon(conn.RPCClient())
		resp, err := addon.OneWireScan(ctx)
		if err != nil {
			return err
		}

		for _, dev := range resp.Devices {
			results = append(results, OneWireDevice{
				Type:      dev.Type,
				Addr:      dev.Addr,
				Component: dev.Component,
			})
		}
		return nil
	})

	return results, err
}
