// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// LoRaConfig represents LoRa configuration.
type LoRaConfig struct {
	ID   int   `json:"id"`
	Freq int64 `json:"freq"`
	BW   int   `json:"bw"`
	DR   int   `json:"dr"`
	TxP  int   `json:"txp"`
}

// LoRaStatus represents LoRa status.
type LoRaStatus struct {
	ID   int     `json:"id"`
	RSSI int     `json:"rssi"`
	SNR  float64 `json:"snr"`
}

// LoRaSendBytes sends data over LoRa.
func (s *Service) LoRaSendBytes(ctx context.Context, identifier string, componentID int, data string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"id":   componentID,
			"data": data,
		}
		_, err := conn.Call(ctx, "LoRa.SendBytes", params)
		if err != nil {
			return fmt.Errorf("failed to send LoRa data: %w", err)
		}
		return nil
	})
}

// LoRaSetConfig updates LoRa configuration.
func (s *Service) LoRaSetConfig(ctx context.Context, identifier string, componentID int, config map[string]any) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"id":     componentID,
			"config": config,
		}
		_, err := conn.Call(ctx, "LoRa.SetConfig", params)
		if err != nil {
			return fmt.Errorf("failed to set LoRa config: %w", err)
		}
		return nil
	})
}

// LoRaGetConfig gets LoRa configuration.
func (s *Service) LoRaGetConfig(ctx context.Context, identifier string, componentID int) (map[string]any, error) {
	var config map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{"id": componentID}
		result, err := conn.Call(ctx, "LoRa.GetConfig", params)
		if err != nil {
			return fmt.Errorf("failed to get LoRa config: %w", err)
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

// LoRaGetStatus gets LoRa status.
func (s *Service) LoRaGetStatus(ctx context.Context, identifier string, componentID int) (map[string]any, error) {
	var status map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{"id": componentID}
		result, err := conn.Call(ctx, "LoRa.GetStatus", params)
		if err != nil {
			return fmt.Errorf("failed to get LoRa status: %w", err)
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
