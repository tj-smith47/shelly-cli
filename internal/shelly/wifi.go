package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// WiFiStatusFull represents the current WiFi status with full details.
type WiFiStatusFull struct {
	Status        string
	StaIP         string
	SSID          string
	RSSI          float64
	APClientCount int
}

// WiFiConfigFull represents WiFi configuration with full details.
type WiFiConfigFull struct {
	STA  *WiFiStationFull
	STA1 *WiFiStationFull
	AP   *WiFiAPFull
}

// WiFiStationFull represents station configuration details.
type WiFiStationFull struct {
	SSID     string
	Enabled  bool
	IsOpen   bool
	IPv4Mode string
	IP       string
	Netmask  string
	Gateway  string
}

// WiFiAPFull represents access point configuration details.
type WiFiAPFull struct {
	SSID          string
	Enabled       bool
	IsOpen        bool
	RangeExtender bool
}

// WiFiNetworkFull represents a scanned network with full details.
type WiFiNetworkFull struct {
	SSID    string
	BSSID   string
	Auth    string
	Channel int
	RSSI    float64
}

// GetWiFiStatusFull gets the full WiFi status from a device.
func (s *Service) GetWiFiStatusFull(ctx context.Context, identifier string) (*WiFiStatusFull, error) {
	var result *WiFiStatusFull
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		status, err := wifi.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &WiFiStatusFull{
			Status: status.Status,
		}
		if status.StaIP != nil {
			result.StaIP = *status.StaIP
		}
		if status.SSID != nil {
			result.SSID = *status.SSID
		}
		if status.RSSI != nil {
			result.RSSI = *status.RSSI
		}
		if status.APClientCount != nil {
			result.APClientCount = *status.APClientCount
		}
		return nil
	})
	return result, err
}

// GetWiFiConfigFull gets the full WiFi configuration from a device.
func (s *Service) GetWiFiConfigFull(ctx context.Context, identifier string) (*WiFiConfigFull, error) {
	var result *WiFiConfigFull
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		cfg, err := wifi.GetConfig(ctx)
		if err != nil {
			return err
		}

		result = &WiFiConfigFull{}
		if cfg.STA != nil {
			result.STA = convertStationFull(cfg.STA)
		}
		if cfg.STA1 != nil {
			result.STA1 = convertStationFull(cfg.STA1)
		}
		if cfg.AP != nil {
			result.AP = convertAPFull(cfg.AP)
		}
		return nil
	})
	return result, err
}

func convertStationFull(sta *components.WiFiStationConfig) *WiFiStationFull {
	result := &WiFiStationFull{}
	if sta.SSID != nil {
		result.SSID = *sta.SSID
	}
	if sta.Enable != nil {
		result.Enabled = *sta.Enable
	}
	if sta.IsOpen != nil {
		result.IsOpen = *sta.IsOpen
	}
	if sta.IPv4Mode != nil {
		result.IPv4Mode = *sta.IPv4Mode
	}
	if sta.IP != nil {
		result.IP = *sta.IP
	}
	if sta.Netmask != nil {
		result.Netmask = *sta.Netmask
	}
	if sta.GW != nil {
		result.Gateway = *sta.GW
	}
	return result
}

func convertAPFull(ap *components.WiFiAPConfig) *WiFiAPFull {
	result := &WiFiAPFull{}
	if ap.SSID != nil {
		result.SSID = *ap.SSID
	}
	if ap.Enable != nil {
		result.Enabled = *ap.Enable
	}
	if ap.IsOpen != nil {
		result.IsOpen = *ap.IsOpen
	}
	if ap.RangeExtender != nil && ap.RangeExtender.Enable != nil {
		result.RangeExtender = *ap.RangeExtender.Enable
	}
	return result
}

// ScanWiFiNetworksFull scans for available WiFi networks with full details.
func (s *Service) ScanWiFiNetworksFull(ctx context.Context, identifier string) ([]WiFiNetworkFull, error) {
	var result []WiFiNetworkFull
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		scan, err := wifi.Scan(ctx)
		if err != nil {
			return err
		}

		result = make([]WiFiNetworkFull, 0, len(scan.Results))
		for _, r := range scan.Results {
			network := WiFiNetworkFull{}
			if r.SSID != nil {
				network.SSID = *r.SSID
			}
			if r.BSSID != nil {
				network.BSSID = *r.BSSID
			}
			if r.Auth != nil {
				network.Auth = *r.Auth
			}
			if r.Channel != nil {
				network.Channel = *r.Channel
			}
			if r.RSSI != nil {
				network.RSSI = *r.RSSI
			}
			result = append(result, network)
		}
		return nil
	})
	return result, err
}

// SetWiFiStation configures the primary WiFi station.
func (s *Service) SetWiFiStation(ctx context.Context, identifier, ssid, password string, enable bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		cfg := &components.WiFiConfig{
			STA: &components.WiFiStationConfig{
				SSID:   &ssid,
				Pass:   &password,
				Enable: &enable,
			},
		}
		return wifi.SetConfig(ctx, cfg)
	})
}

// SetWiFiAP configures the access point.
func (s *Service) SetWiFiAP(ctx context.Context, identifier, ssid, password string, enable bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		cfg := &components.WiFiConfig{
			AP: &components.WiFiAPConfig{
				SSID:   &ssid,
				Pass:   &password,
				Enable: &enable,
			},
		}
		return wifi.SetConfig(ctx, cfg)
	})
}
