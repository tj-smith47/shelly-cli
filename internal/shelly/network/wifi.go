// Package network provides network-related services for Shelly devices.
package network

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-go/gen2/components"
	"github.com/tj-smith47/shelly-go/types"

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

// ConnectionProvider allows executing operations with a device connection.
type ConnectionProvider interface {
	WithConnection(ctx context.Context, identifier string, fn func(*client.Client) error) error
}

// WiFiService provides WiFi-related operations for Shelly devices.
type WiFiService struct {
	provider ConnectionProvider
}

// NewWiFiService creates a new WiFi service.
func NewWiFiService(provider ConnectionProvider) *WiFiService {
	return &WiFiService{provider: provider}
}

// GetStatusFull gets the full WiFi status from a device.
func (s *WiFiService) GetStatusFull(ctx context.Context, identifier string) (*WiFiStatusFull, error) {
	var result *WiFiStatusFull
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// GetConfigFull gets the full WiFi configuration from a device.
func (s *WiFiService) GetConfigFull(ctx context.Context, identifier string) (*WiFiConfigFull, error) {
	var result *WiFiConfigFull
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// ScanNetworks runs Wifi.Scan on a connected device.
func ScanNetworks(ctx context.Context, conn *client.Client) ([]WiFiNetworkFull, error) {
	scan, err := components.NewWiFi(conn.RPCClient()).Scan(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]WiFiNetworkFull, 0, len(scan.Results))
	for _, r := range scan.Results {
		n := WiFiNetworkFull{}
		if r.SSID != nil {
			n.SSID = *r.SSID
		}
		if r.BSSID != nil {
			n.BSSID = *r.BSSID
		}
		if r.Auth != nil {
			n.Auth = r.Auth.String()
		}
		if r.Channel != nil {
			n.Channel = *r.Channel
		}
		if r.RSSI != nil {
			n.RSSI = *r.RSSI
		}
		result = append(result, n)
	}
	return result, nil
}

// ScanNetworksFull scans for available WiFi networks with full details.
func (s *WiFiService) ScanNetworksFull(ctx context.Context, identifier string) ([]WiFiNetworkFull, error) {
	var result []WiFiNetworkFull
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		var err error
		result, err = ScanNetworks(ctx, conn)
		return err
	})
	return result, err
}

// StationWrite is a change to a Gen2+ device's primary WiFi station.
type StationWrite struct {
	// SSID is the network; empty leaves it unchanged.
	SSID string
	// Password is written as the key; empty with Open false writes no key.
	Password string
	// Open joins a network that has no password.
	Open bool
	// Enable switches the station on or off; nil leaves it unchanged.
	Enable *bool
	// StaticIP, Gateway, Netmask and DNS switch the station to static IPv4
	// addressing when StaticIP is set; an empty DNS is left out.
	StaticIP, Gateway, Netmask, DNS string
}

// StationConfig builds the WiFi.SetConfig payload for w. A pass is sent only
// for a password or an open network, so the device otherwise keeps its key;
// is_open is never sent, as the device derives it from the key.
func StationConfig(w StationWrite) (*components.WiFiConfig, error) {
	pass, err := wifiPass(w.Password, w.Open)
	if err != nil {
		return nil, err
	}
	sta := &components.WiFiStationConfig{Pass: pass, Enable: w.Enable}
	if w.SSID != "" {
		sta.SSID = &w.SSID
	}
	if w.StaticIP != "" {
		mode := "static"
		sta.IPv4Mode, sta.IP, sta.GW, sta.Netmask = &mode, &w.StaticIP, &w.Gateway, &w.Netmask
		if w.DNS != "" {
			sta.Nameserver = &w.DNS
		}
	}
	return &components.WiFiConfig{STA: sta}, nil
}

// SetAP configures the access point. open makes it an open access point;
// otherwise an empty password keeps its current key.
func (s *WiFiService) SetAP(ctx context.Context, identifier, ssid, password string, open, enable bool) error {
	cfg, err := apConfig(ssid, password, open, enable)
	if err != nil {
		return err
	}
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return components.NewWiFi(conn.RPCClient()).SetConfig(ctx, cfg)
	})
}

// apConfig sets the access point's is_open only when its key is written: on
// the access point, unlike a station, the device takes is_open as a setting.
func apConfig(ssid, password string, open, enable bool) (*components.WiFiConfig, error) {
	pass, err := wifiPass(password, open)
	if err != nil {
		return nil, err
	}
	ap := &components.WiFiAPConfig{SSID: &ssid, Pass: pass, Enable: &enable}
	if pass != nil {
		ap.IsOpen = &open
	}
	return &components.WiFiConfig{AP: ap}, nil
}

// wifiPass returns the pass to write: empty for open, the password when one is
// given, and nil (left out, keeping the device's key) otherwise.
func wifiPass(password string, open bool) (pass *string, err error) {
	if open && password != "" {
		return nil, fmt.Errorf("%w: an open network takes no password", types.ErrInvalidParam)
	}
	if open || password != "" {
		pass = &password
	}
	return pass, nil
}
