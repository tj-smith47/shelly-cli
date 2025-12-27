// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

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

// FetchLoRaFullStatus fetches combined LoRa config and status.
func (s *Service) FetchLoRaFullStatus(ctx context.Context, device string, componentID int, ios *iostreams.IOStreams) (model.LoRaFullStatus, error) {
	var full model.LoRaFullStatus

	// Get config
	cfgMap, err := s.LoRaGetConfig(ctx, device, componentID)
	if err != nil {
		ios.Debug("LoRa.GetConfig failed: %v", err)
		return full, fmt.Errorf("LoRa not available on this device: %w", err)
	}

	full.Config = parseLoRaConfig(cfgMap)

	// Get status
	statusMap, err := s.LoRaGetStatus(ctx, device, componentID)
	if err != nil {
		ios.Debug("LoRa.GetStatus failed: %v", err)
		// Config succeeded, status failed - still return partial info
		return full, nil
	}
	full.Status = parseLoRaStatus(statusMap)

	return full, nil
}

func parseLoRaConfig(cfgMap map[string]any) *model.LoRaConfig {
	cfg := &model.LoRaConfig{}
	if id, ok := cfgMap["id"].(float64); ok {
		cfg.ID = int(id)
	}
	if freq, ok := cfgMap["freq"].(float64); ok {
		cfg.Freq = int64(freq)
	}
	if bw, ok := cfgMap["bw"].(float64); ok {
		cfg.BW = int(bw)
	}
	if dr, ok := cfgMap["dr"].(float64); ok {
		cfg.DR = int(dr)
	}
	if txp, ok := cfgMap["txp"].(float64); ok {
		cfg.TxP = int(txp)
	}
	return cfg
}

func parseLoRaStatus(statusMap map[string]any) *model.LoRaStatus {
	st := &model.LoRaStatus{}
	if id, ok := statusMap["id"].(float64); ok {
		st.ID = int(id)
	}
	if rssi, ok := statusMap["rssi"].(float64); ok {
		st.RSSI = int(rssi)
	}
	if snr, ok := statusMap["snr"].(float64); ok {
		st.SNR = snr
	}
	return st
}
