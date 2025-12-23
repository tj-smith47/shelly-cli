package shelly

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// WifiStatus represents the current WiFi status.
type WifiStatus struct {
	Status        string
	StaIP         string
	SSID          string
	RSSI          float64
	APClientCount int
}

// WifiConfig represents WiFi configuration.
type WifiConfig struct {
	STA  *WifiStationConfig
	STA1 *WifiStationConfig
	AP   *WifiAPConfig
}

// WifiStationConfig represents station configuration.
type WifiStationConfig struct {
	SSID     string
	Enabled  bool
	IsOpen   bool
	IPv4Mode string
	IP       string
	Netmask  string
	Gateway  string
}

// WifiAPConfig represents access point configuration.
type WifiAPConfig struct {
	SSID          string
	Enabled       bool
	IsOpen        bool
	RangeExtender bool
}

// WifiNetwork represents a scanned network.
type WifiNetwork struct {
	SSID    string
	BSSID   string
	Auth    string
	Channel int
	RSSI    float64
}

// GetWifiStatus gets the WiFi status from a device.
func (s *Service) GetWifiStatus(ctx context.Context, identifier string) (*WifiStatus, error) {
	var result *WifiStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		status, err := wifi.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &WifiStatus{
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

// GetWifiConfig gets the WiFi configuration from a device.
func (s *Service) GetWifiConfig(ctx context.Context, identifier string) (*WifiConfig, error) {
	var result *WifiConfig
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		cfg, err := wifi.GetConfig(ctx)
		if err != nil {
			return err
		}

		result = &WifiConfig{}
		if cfg.STA != nil {
			result.STA = convertStationConfig(cfg.STA)
		}
		if cfg.STA1 != nil {
			result.STA1 = convertStationConfig(cfg.STA1)
		}
		if cfg.AP != nil {
			result.AP = convertAPConfig(cfg.AP)
		}
		return nil
	})
	return result, err
}

func convertStationConfig(sta *components.WiFiStationConfig) *WifiStationConfig {
	result := &WifiStationConfig{}
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

func convertAPConfig(ap *components.WiFiAPConfig) *WifiAPConfig {
	result := &WifiAPConfig{}
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

// ScanWifiNetworks scans for available WiFi networks.
func (s *Service) ScanWifiNetworks(ctx context.Context, identifier string) ([]WifiNetwork, error) {
	var result []WifiNetwork
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		scan, err := wifi.Scan(ctx)
		if err != nil {
			return err
		}

		result = make([]WifiNetwork, 0, len(scan.Results))
		for _, r := range scan.Results {
			network := WifiNetwork{}
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

// SetWifiStation configures the primary WiFi station.
func (s *Service) SetWifiStation(ctx context.Context, identifier, ssid, password string, enable bool) error {
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

// SetWifiAP configures the access point.
func (s *Service) SetWifiAP(ctx context.Context, identifier, ssid, password string, enable bool) error {
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
