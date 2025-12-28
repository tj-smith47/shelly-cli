// Package provision provides device provisioning for Shelly devices.
package provision

import (
	"context"
	"encoding/json"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// DeviceInfo holds basic device information for provisioning.
type DeviceInfo struct {
	Model string `json:"model"`
	MAC   string `json:"mac"`
	ID    string `json:"id"`
}

// BTHomeDiscovery holds BTHome discovery status.
type BTHomeDiscovery struct {
	Active    bool  `json:"active"`
	StartedAt int64 `json:"started_at,omitempty"`
	Duration  int   `json:"duration,omitempty"`
}

// ConnectionProvider allows executing operations with a device connection.
type ConnectionProvider interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// Service provides provisioning-related operations for Shelly devices.
type Service struct {
	provider ConnectionProvider
}

// New creates a new provisioning service.
func New(provider ConnectionProvider) *Service {
	return &Service{provider: provider}
}

// GetDeviceInfoByAddress attempts to connect to a device at the given address
// and retrieve its basic info. Used for provisioning.
func (s *Service) GetDeviceInfoByAddress(ctx context.Context, address string) (*DeviceInfo, error) {
	var result *DeviceInfo
	err := s.provider.WithConnection(ctx, address, func(conn *client.Client) error {
		info := conn.Info()
		result = &DeviceInfo{
			Model: info.Model,
			MAC:   info.MAC,
			ID:    info.ID,
		}
		return nil
	})
	return result, err
}

// ConfigureWiFi configures a device's WiFi station settings.
// Used during provisioning to set up the device's network connection.
func (s *Service) ConfigureWiFi(ctx context.Context, address, ssid, password string) error {
	return s.provider.WithConnection(ctx, address, func(conn *client.Client) error {
		params := map[string]any{
			"config": map[string]any{
				"sta": map[string]any{
					"ssid":   ssid,
					"pass":   password,
					"enable": true,
				},
			},
		}
		_, err := conn.Call(ctx, "WiFi.SetConfig", params)
		return err
	})
}

// GetBTHomeStatus returns the BTHome status including discovery.
func (s *Service) GetBTHomeStatus(ctx context.Context, identifier string) (*BTHomeDiscovery, error) {
	var result *BTHomeDiscovery
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		bthome := components.NewBTHome(conn.RPCClient())
		status, err := bthome.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &BTHomeDiscovery{}
		if status.Discovery != nil {
			result.Active = true
			result.StartedAt = int64(status.Discovery.StartedAt)
			result.Duration = status.Discovery.Duration
		}
		return nil
	})
	return result, err
}

// StartBTHomeDiscovery starts BTHome device discovery.
func (s *Service) StartBTHomeDiscovery(ctx context.Context, identifier string, duration int) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		bthome := components.NewBTHome(conn.RPCClient())
		return bthome.StartDeviceDiscovery(ctx, &duration)
	})
}

// ExtractWiFiSSID extracts the station SSID from a raw WiFi.GetConfig result.
func ExtractWiFiSSID(rawResult any) string {
	wifiBytes, err := json.Marshal(rawResult)
	if err != nil {
		return ""
	}
	var wifiConfig struct {
		Sta struct {
			SSID string `json:"ssid"`
		} `json:"sta"`
	}
	if err := json.Unmarshal(wifiBytes, &wifiConfig); err != nil {
		return ""
	}
	return wifiConfig.Sta.SSID
}
