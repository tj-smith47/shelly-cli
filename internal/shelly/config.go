// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// GetConfig returns the full device configuration.
func (s *Service) GetConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var result map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		config, err := conn.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = config
		return nil
	})
	return result, err
}

// LoadConfig loads a configuration from either a device or a file.
// It auto-detects based on whether the source is a file path or device identifier.
// Returns the config map and a display name for the source.
func (s *Service) LoadConfig(ctx context.Context, source string) (cfg map[string]any, name string, err error) {
	if IsConfigFile(source) {
		return LoadConfigFromFile(source)
	}
	return s.loadConfigFromDevice(ctx, source)
}

// loadConfigFromDevice fetches configuration from a live device using raw RPC.
func (s *Service) loadConfigFromDevice(ctx context.Context, device string) (cfg map[string]any, name string, err error) {
	conn, err := s.Connect(ctx, device)
	if err != nil {
		return nil, "", err
	}
	defer iostreams.CloseWithDebug("closing config connection", conn)

	rawResult, err := conn.Call(ctx, "Shelly.GetConfig", nil)
	if err != nil {
		return nil, "", err
	}

	// Convert to map
	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal config: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, "", fmt.Errorf("failed to parse config: %w", err)
	}

	return result, device, nil
}

// LoadConfigFromFile reads and parses a JSON configuration file.
func LoadConfigFromFile(filePath string) (cfg map[string]any, name string, err error) {
	data, err := os.ReadFile(filePath) //nolint:gosec // User-provided file path is intentional
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	return config, filePath, nil
}

// IsConfigFile checks if the given path is a configuration file.
// Returns true if the path has a config file extension or if the file exists.
func IsConfigFile(path string) bool {
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return true
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// SetConfig updates device configuration.
// The config parameter should be a map of component keys to configuration
// objects. Only specified components will be updated.
func (s *Service) SetConfig(ctx context.Context, identifier string, config map[string]any) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.SetConfig(ctx, config)
	})
}

// SetComponentConfig updates a specific component's configuration.
func (s *Service) SetComponentConfig(ctx context.Context, identifier, component string, config map[string]any) error {
	fullConfig := map[string]any{
		component: config,
	}
	return s.SetConfig(ctx, identifier, fullConfig)
}

// WiFiStatus holds WiFi status information.
type WiFiStatus struct {
	StaIP   string `json:"sta_ip"`
	Status  string `json:"status"`
	SSID    string `json:"ssid"`
	RSSI    int    `json:"rssi"`
	APCount int    `json:"ap_client_count,omitempty"`
}

// GetWiFiStatus returns the WiFi status.
func (s *Service) GetWiFiStatus(ctx context.Context, identifier string) (*WiFiStatus, error) {
	var result *WiFiStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		status, err := wifi.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &WiFiStatus{
			Status: status.Status,
		}
		if status.StaIP != nil {
			result.StaIP = *status.StaIP
		}
		if status.SSID != nil {
			result.SSID = *status.SSID
		}
		if status.RSSI != nil {
			result.RSSI = int(*status.RSSI)
		}
		if status.APClientCount != nil {
			result.APCount = *status.APClientCount
		}
		return nil
	})
	return result, err
}

// GetWiFiConfig returns the WiFi configuration.
func (s *Service) GetWiFiConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var result map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		config, err := wifi.GetConfig(ctx)
		if err != nil {
			return err
		}
		// Convert to map for flexibility
		result = map[string]any{
			"sta":  config.STA,
			"sta1": config.STA1,
			"ap":   config.AP,
			"roam": config.Roam,
		}
		return nil
	})
	return result, err
}

// SetWiFiConfig updates the WiFi configuration.
func (s *Service) SetWiFiConfig(ctx context.Context, identifier, ssid, password string, enable *bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())

		// Build the station config
		staConfig := &components.WiFiStationConfig{}
		if ssid != "" {
			staConfig.SSID = &ssid
		}
		if password != "" {
			staConfig.Pass = &password
		}
		if enable != nil {
			staConfig.Enable = enable
		}

		return wifi.SetConfig(ctx, &components.WiFiConfig{
			STA: staConfig,
		})
	})
}

// WiFiScanResult holds a WiFi scan result.
type WiFiScanResult struct {
	SSID    string `json:"ssid"`
	BSSID   string `json:"bssid"`
	RSSI    int    `json:"rssi"`
	Channel int    `json:"channel"`
	Auth    string `json:"auth"`
}

// DedupeWiFiNetworks deduplicates WiFi scan results by SSID, keeping the
// strongest signal for each network, and returns them sorted by signal strength.
func DedupeWiFiNetworks(results []WiFiScanResult) []WiFiScanResult {
	// Dedupe by SSID, keeping strongest signal
	seen := make(map[string]WiFiScanResult)
	for _, r := range results {
		if r.SSID == "" {
			continue
		}
		existing, exists := seen[r.SSID]
		if !exists || r.RSSI > existing.RSSI {
			seen[r.SSID] = r
		}
	}

	// Convert to slice
	networks := make([]WiFiScanResult, 0, len(seen))
	for _, n := range seen {
		networks = append(networks, n)
	}

	// Sort by signal strength (strongest first)
	sort.Slice(networks, func(i, j int) bool {
		return networks[i].RSSI > networks[j].RSSI
	})

	return networks
}

// ScanWiFi scans for available WiFi networks.
func (s *Service) ScanWiFi(ctx context.Context, identifier string) ([]WiFiScanResult, error) {
	var results []WiFiScanResult
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		scanResults, err := wifi.Scan(ctx)
		if err != nil {
			return err
		}
		for _, r := range scanResults.Results {
			result := WiFiScanResult{}
			if r.SSID != nil {
				result.SSID = *r.SSID
			}
			if r.BSSID != nil {
				result.BSSID = *r.BSSID
			}
			if r.RSSI != nil {
				result.RSSI = int(*r.RSSI)
			}
			if r.Channel != nil {
				result.Channel = *r.Channel
			}
			if r.Auth != nil {
				result.Auth = *r.Auth
			}
			results = append(results, result)
		}
		return nil
	})
	return results, err
}

// CloudStatus holds cloud connection status.
type CloudStatus struct {
	Connected bool `json:"connected"`
}

// GetCloudStatus returns the cloud connection status.
func (s *Service) GetCloudStatus(ctx context.Context, identifier string) (*CloudStatus, error) {
	var result *CloudStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		cloud := components.NewCloud(conn.RPCClient())
		status, err := cloud.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &CloudStatus{
			Connected: status.Connected,
		}
		return nil
	})
	return result, err
}

// GetCloudConfig returns the cloud configuration.
func (s *Service) GetCloudConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var result map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		cloud := components.NewCloud(conn.RPCClient())
		config, err := cloud.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = map[string]any{
			"enable": config.Enable,
			"server": config.Server,
		}
		return nil
	})
	return result, err
}

// SetCloudEnabled enables or disables cloud connection.
func (s *Service) SetCloudEnabled(ctx context.Context, identifier string, enable bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		cloud := components.NewCloud(conn.RPCClient())
		return cloud.SetConfig(ctx, &components.CloudConfig{
			Enable: &enable,
		})
	})
}

// SetWiFiAPConfig updates the WiFi access point configuration.
func (s *Service) SetWiFiAPConfig(ctx context.Context, identifier, ssid, password string, enable *bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())

		apConfig := &components.WiFiAPConfig{}
		if ssid != "" {
			apConfig.SSID = &ssid
		}
		if password != "" {
			apConfig.Pass = &password
		}
		if enable != nil {
			apConfig.Enable = enable
		}

		return wifi.SetConfig(ctx, &components.WiFiConfig{
			AP: apConfig,
		})
	})
}

// WiFiAPClient holds information about a client connected to the AP.
type WiFiAPClient struct {
	MAC   string `json:"mac"`
	IP    string `json:"ip"`
	Since int64  `json:"since"`
}

// ListWiFiAPClients lists clients connected to the device's access point.
func (s *Service) ListWiFiAPClients(ctx context.Context, identifier string) ([]WiFiAPClient, error) {
	var results []WiFiAPClient
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		wifi := components.NewWiFi(conn.RPCClient())
		response, err := wifi.ListAPClients(ctx)
		if err != nil {
			return err
		}
		for _, c := range response.APClients {
			client := WiFiAPClient{
				MAC: c.MAC,
			}
			if c.IP != nil {
				client.IP = *c.IP
			}
			if c.Since != nil {
				client.Since = *c.Since
			}
			results = append(results, client)
		}
		return nil
	})
	return results, err
}

// WebhookInfo holds webhook information.
type WebhookInfo struct {
	ID     int      `json:"id"`
	Name   string   `json:"name,omitempty"`
	Event  string   `json:"event"`
	Enable bool     `json:"enable"`
	URLs   []string `json:"urls"`
	Cid    int      `json:"cid"`
}

// ListWebhooks returns all configured webhooks for a device.
func (s *Service) ListWebhooks(ctx context.Context, identifier string) ([]WebhookInfo, error) {
	var results []WebhookInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		webhook := components.NewWebhook(conn.RPCClient())
		resp, err := webhook.List(ctx)
		if err != nil {
			return err
		}
		for _, h := range resp.Hooks {
			info := WebhookInfo{
				Event:  h.Event,
				Enable: h.Enable,
				URLs:   h.URLs,
				Cid:    h.Cid,
			}
			if h.ID != nil {
				info.ID = *h.ID
			}
			if h.Name != nil {
				info.Name = *h.Name
			}
			results = append(results, info)
		}
		return nil
	})
	return results, err
}

// CreateWebhookParams holds parameters for creating a webhook.
type CreateWebhookParams struct {
	Event  string
	URLs   []string
	Name   string
	Enable bool
	Cid    int
}

// CreateWebhook creates a new webhook.
func (s *Service) CreateWebhook(ctx context.Context, identifier string, params CreateWebhookParams) (int, error) {
	var id int
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		webhook := components.NewWebhook(conn.RPCClient())

		cfg := &components.WebhookConfig{
			Event:  params.Event,
			URLs:   params.URLs,
			Enable: params.Enable,
			Cid:    params.Cid,
		}
		if params.Name != "" {
			cfg.Name = &params.Name
		}

		resp, err := webhook.Create(ctx, cfg)
		if err != nil {
			return err
		}
		id = resp.ID
		return nil
	})
	return id, err
}

// DeleteWebhook deletes a webhook by ID.
func (s *Service) DeleteWebhook(ctx context.Context, identifier string, webhookID int) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		webhook := components.NewWebhook(conn.RPCClient())
		_, err := webhook.Delete(ctx, webhookID)
		return err
	})
}

// UpdateWebhookParams holds parameters for updating a webhook.
type UpdateWebhookParams struct {
	Event  string
	URLs   []string
	Name   string
	Enable *bool
}

// UpdateWebhook updates an existing webhook.
func (s *Service) UpdateWebhook(ctx context.Context, identifier string, webhookID int, params UpdateWebhookParams) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		webhook := components.NewWebhook(conn.RPCClient())

		cfg := &components.WebhookConfig{
			Event: params.Event,
			URLs:  params.URLs,
		}
		if params.Name != "" {
			cfg.Name = &params.Name
		}
		if params.Enable != nil {
			cfg.Enable = *params.Enable
		}

		_, err := webhook.Update(ctx, webhookID, cfg)
		return err
	})
}

// ListSupportedWebhookEvents returns supported webhook event types.
func (s *Service) ListSupportedWebhookEvents(ctx context.Context, identifier string) ([]string, error) {
	var events []string
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		webhook := components.NewWebhook(conn.RPCClient())
		resp, err := webhook.ListSupported(ctx)
		if err != nil {
			return err
		}
		events = resp.HookTypes
		return nil
	})
	return events, err
}

// BLEConfig holds BLE configuration.
type BLEConfig struct {
	Enable       bool `json:"enable"`
	RPCEnabled   bool `json:"rpc_enabled"`
	ObserverMode bool `json:"observer_mode"`
}

// GetBLEConfig returns the BLE configuration.
func (s *Service) GetBLEConfig(ctx context.Context, identifier string) (*BLEConfig, error) {
	var result *BLEConfig
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		ble := components.NewBLE(conn.RPCClient())
		config, err := ble.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = &BLEConfig{}
		if config.Enable != nil {
			result.Enable = *config.Enable
		}
		if config.RPC != nil && config.RPC.Enable != nil {
			result.RPCEnabled = *config.RPC.Enable
		}
		if config.Observer != nil && config.Observer.Enable != nil {
			result.ObserverMode = *config.Observer.Enable
		}
		return nil
	})
	return result, err
}

// SetBLEConfig updates the BLE configuration.
func (s *Service) SetBLEConfig(ctx context.Context, identifier string, enable, rpcEnabled, observerMode *bool) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		ble := components.NewBLE(conn.RPCClient())

		config := &components.BLEConfig{}
		if enable != nil {
			config.Enable = enable
		}
		if rpcEnabled != nil {
			config.RPC = &components.BLERPCConfig{Enable: rpcEnabled}
		}
		if observerMode != nil {
			config.Observer = &components.BLEObserverConfig{Enable: observerMode}
		}

		return ble.SetConfig(ctx, config)
	})
}

// TUIMatterStatus holds Matter status information for the TUI.
type TUIMatterStatus struct {
	Enabled        bool `json:"enabled"`
	Commissionable bool `json:"commissionable"`
	FabricsCount   int  `json:"fabrics_count"`
}

// GetTUIMatterStatus returns the Matter status for the TUI.
func (s *Service) GetTUIMatterStatus(ctx context.Context, identifier string) (*TUIMatterStatus, error) {
	var result *TUIMatterStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Matter uses a different import path - construct RPC call manually
		matterResult, err := conn.Call(ctx, "Matter.GetConfig", nil)
		if err != nil {
			return err
		}
		configData, err := json.Marshal(matterResult)
		if err != nil {
			return err
		}
		var config struct {
			Enable bool `json:"enable"`
		}
		if err := json.Unmarshal(configData, &config); err != nil {
			return err
		}

		statusResult, err := conn.Call(ctx, "Matter.GetStatus", nil)
		if err != nil {
			return err
		}
		statusData, err := json.Marshal(statusResult)
		if err != nil {
			return err
		}
		var status struct {
			Commissionable bool `json:"commissionable"`
			FabricsCount   int  `json:"fabrics_count"`
		}
		if err := json.Unmarshal(statusData, &status); err != nil {
			return err
		}

		result = &TUIMatterStatus{
			Enabled:        config.Enable,
			Commissionable: status.Commissionable,
			FabricsCount:   status.FabricsCount,
		}
		return nil
	})
	return result, err
}

// TUIZigbeeStatus holds Zigbee status information for the TUI.
type TUIZigbeeStatus struct {
	Enabled      bool   `json:"enabled"`
	NetworkState string `json:"network_state"`
	Channel      int    `json:"channel,omitempty"`
	PANID        uint16 `json:"pan_id,omitempty"`
}

// GetTUIZigbeeStatus returns the Zigbee status for the TUI.
func (s *Service) GetTUIZigbeeStatus(ctx context.Context, identifier string) (*TUIZigbeeStatus, error) {
	var result *TUIZigbeeStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Get config for enable state
		configResult, err := conn.Call(ctx, "Zigbee.GetConfig", nil)
		if err != nil {
			return err
		}
		configData, err := json.Marshal(configResult)
		if err != nil {
			return err
		}
		var config struct {
			Enable bool `json:"enable"`
		}
		if err := json.Unmarshal(configData, &config); err != nil {
			return err
		}

		// Get status for network info
		statusResult, err := conn.Call(ctx, "Zigbee.GetStatus", nil)
		if err != nil {
			return err
		}
		statusData, err := json.Marshal(statusResult)
		if err != nil {
			return err
		}
		var status struct {
			NetworkState string `json:"network_state"`
			Channel      int    `json:"channel"`
			PANID        uint16 `json:"pan_id"`
		}
		if err := json.Unmarshal(statusData, &status); err != nil {
			return err
		}

		result = &TUIZigbeeStatus{
			Enabled:      config.Enable,
			NetworkState: status.NetworkState,
			Channel:      status.Channel,
			PANID:        status.PANID,
		}
		return nil
	})
	return result, err
}

// TUILoRaStatus holds LoRa status information for the TUI.
type TUILoRaStatus struct {
	Enabled   bool    `json:"enabled"`
	Frequency int64   `json:"frequency"`
	TxPower   int     `json:"tx_power"`
	RSSI      int     `json:"rssi,omitempty"`
	SNR       float64 `json:"snr,omitempty"`
}

// GetTUILoRaStatus returns the LoRa status for the TUI.
func (s *Service) GetTUILoRaStatus(ctx context.Context, identifier string) (*TUILoRaStatus, error) {
	var result *TUILoRaStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Get config - LoRa uses ID 100 by default
		configResult, err := conn.Call(ctx, "LoRa.GetConfig", map[string]any{"id": 100})
		if err != nil {
			return err
		}
		configData, err := json.Marshal(configResult)
		if err != nil {
			return err
		}
		var config struct {
			Enable bool  `json:"enable"`
			Freq   int64 `json:"freq"`
			TxP    int   `json:"txp"`
		}
		if err := json.Unmarshal(configData, &config); err != nil {
			return err
		}

		// Get status
		statusResult, err := conn.Call(ctx, "LoRa.GetStatus", map[string]any{"id": 100})
		if err != nil {
			return err
		}
		statusData, err := json.Marshal(statusResult)
		if err != nil {
			return err
		}
		var status struct {
			RSSI int     `json:"rssi"`
			SNR  float64 `json:"snr"`
		}
		if err := json.Unmarshal(statusData, &status); err != nil {
			return err
		}

		result = &TUILoRaStatus{
			Enabled:   config.Enable,
			Frequency: config.Freq,
			TxPower:   config.TxP,
			RSSI:      status.RSSI,
			SNR:       status.SNR,
		}
		return nil
	})
	return result, err
}

// TUISecurityStatus holds security status information for the TUI.
type TUISecurityStatus struct {
	AuthEnabled  bool   `json:"auth_enabled"`
	EcoMode      bool   `json:"eco_mode"`
	Discoverable bool   `json:"discoverable"`
	DebugMQTT    bool   `json:"debug_mqtt"`
	DebugWS      bool   `json:"debug_ws"`
	DebugUDP     bool   `json:"debug_udp"`
	DebugUDPAddr string `json:"debug_udp_addr,omitempty"`
}

// GetTUISecurityStatus returns the security status for the TUI.
func (s *Service) GetTUISecurityStatus(ctx context.Context, identifier string) (*TUISecurityStatus, error) {
	var result *TUISecurityStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Get client info for auth status
		info := conn.Info()

		// Get system config for debug and visibility settings
		sysResult, err := conn.Call(ctx, "Sys.GetConfig", nil)
		if err != nil {
			return err
		}
		sysData, err := json.Marshal(sysResult)
		if err != nil {
			return err
		}
		var sysConfig struct {
			Device *struct {
				EcoMode      *bool `json:"eco_mode"`
				Discoverable *bool `json:"discoverable"`
			} `json:"device"`
			Debug *struct {
				MQTT *struct {
					Enable *bool `json:"enable"`
				} `json:"mqtt"`
				Websocket *struct {
					Enable *bool `json:"enable"`
				} `json:"websocket"`
				UDP *struct {
					Addr *string `json:"addr"`
				} `json:"udp"`
			} `json:"debug"`
		}
		if err := json.Unmarshal(sysData, &sysConfig); err != nil {
			return err
		}

		result = &TUISecurityStatus{
			AuthEnabled: info.AuthEn,
		}

		if sysConfig.Device != nil {
			if sysConfig.Device.EcoMode != nil {
				result.EcoMode = *sysConfig.Device.EcoMode
			}
			if sysConfig.Device.Discoverable != nil {
				result.Discoverable = *sysConfig.Device.Discoverable
			}
		}

		if sysConfig.Debug != nil {
			if sysConfig.Debug.MQTT != nil && sysConfig.Debug.MQTT.Enable != nil {
				result.DebugMQTT = *sysConfig.Debug.MQTT.Enable
			}
			if sysConfig.Debug.Websocket != nil && sysConfig.Debug.Websocket.Enable != nil {
				result.DebugWS = *sysConfig.Debug.Websocket.Enable
			}
			if sysConfig.Debug.UDP != nil && sysConfig.Debug.UDP.Addr != nil && *sysConfig.Debug.UDP.Addr != "" {
				result.DebugUDP = true
				result.DebugUDPAddr = *sysConfig.Debug.UDP.Addr
			}
		}

		return nil
	})
	return result, err
}

