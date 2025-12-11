// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"crypto/md5" //nolint:gosec // Required for Shelly digest auth (HA1 hash)
	"encoding/hex"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
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

// MQTTStatus holds MQTT status information.
type MQTTStatus struct {
	Connected bool `json:"connected"`
}

// GetMQTTStatus returns the MQTT status.
func (s *Service) GetMQTTStatus(ctx context.Context, identifier string) (*MQTTStatus, error) {
	var result *MQTTStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mqtt := components.NewMQTT(conn.RPCClient())
		status, err := mqtt.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &MQTTStatus{
			Connected: status.Connected,
		}
		return nil
	})
	return result, err
}

// GetMQTTConfig returns the MQTT configuration.
func (s *Service) GetMQTTConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var result map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mqtt := components.NewMQTT(conn.RPCClient())
		config, err := mqtt.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = map[string]any{
			"enable":       config.Enable,
			"server":       config.Server,
			"user":         config.User,
			"client_id":    config.ClientID,
			"topic_prefix": config.TopicPrefix,
			"rpc_ntf":      config.RPCNTF,
			"status_ntf":   config.StatusNTF,
		}
		return nil
	})
	return result, err
}

// SetMQTTConfig updates the MQTT configuration.
func (s *Service) SetMQTTConfig(ctx context.Context, identifier string, enable *bool, server, user, password, topicPrefix string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mqtt := components.NewMQTT(conn.RPCClient())

		mqttCfg := &components.MQTTConfig{}
		if enable != nil {
			mqttCfg.Enable = enable
		}
		if server != "" {
			mqttCfg.Server = &server
		}
		if user != "" {
			mqttCfg.User = &user
		}
		if password != "" {
			mqttCfg.Pass = &password
		}
		if topicPrefix != "" {
			mqttCfg.TopicPrefix = &topicPrefix
		}

		return mqtt.SetConfig(ctx, mqttCfg)
	})
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

// EthernetStatus holds ethernet status information.
type EthernetStatus struct {
	IP string `json:"ip"`
}

// GetEthernetStatus returns the ethernet status.
func (s *Service) GetEthernetStatus(ctx context.Context, identifier string) (*EthernetStatus, error) {
	var result *EthernetStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		eth := components.NewEthernet(conn.RPCClient())
		status, err := eth.GetStatus(ctx)
		if err != nil {
			return err
		}
		result = &EthernetStatus{}
		if status.IP != nil {
			result.IP = *status.IP
		}
		return nil
	})
	return result, err
}

// GetEthernetConfig returns the ethernet configuration.
func (s *Service) GetEthernetConfig(ctx context.Context, identifier string) (map[string]any, error) {
	var result map[string]any
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		eth := components.NewEthernet(conn.RPCClient())
		config, err := eth.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = map[string]any{
			"enable":     config.Enable,
			"ipv4mode":   config.IPv4Mode,
			"ip":         config.IP,
			"netmask":    config.Netmask,
			"gw":         config.GW,
			"nameserver": config.Nameserver,
		}
		return nil
	})
	return result, err
}

// SetEthernetConfig updates the ethernet configuration.
func (s *Service) SetEthernetConfig(ctx context.Context, identifier string, enable *bool, ipv4Mode, ip, netmask, gw, nameserver string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		eth := components.NewEthernet(conn.RPCClient())

		ethCfg := &components.EthernetConfig{}
		if enable != nil {
			ethCfg.Enable = enable
		}
		if ipv4Mode != "" {
			ethCfg.IPv4Mode = &ipv4Mode
		}
		if ip != "" {
			ethCfg.IP = &ip
		}
		if netmask != "" {
			ethCfg.Netmask = &netmask
		}
		if gw != "" {
			ethCfg.GW = &gw
		}
		if nameserver != "" {
			ethCfg.Nameserver = &nameserver
		}

		return eth.SetConfig(ctx, ethCfg)
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

// AuthStatus holds authentication status information.
type AuthStatus struct {
	Enabled bool   `json:"enabled"`
	User    string `json:"user,omitempty"`
	Realm   string `json:"realm,omitempty"`
}

// GetAuthStatus returns the authentication status for a device.
func (s *Service) GetAuthStatus(ctx context.Context, identifier string) (*AuthStatus, error) {
	info, err := s.DeviceInfo(ctx, identifier)
	if err != nil {
		return nil, err
	}
	return &AuthStatus{
		Enabled: info.AuthEn,
	}, nil
}

// SetAuth configures device authentication.
// If password is empty, authentication is disabled.
func (s *Service) SetAuth(ctx context.Context, identifier, user, realm, password string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		params := map[string]any{
			"user":  user,
			"realm": realm,
		}
		if password != "" {
			// Calculate HA1 = MD5(user:realm:password)
			ha1 := calculateHA1(user, realm, password)
			params["ha1"] = ha1
		}
		_, err := conn.Call(ctx, "Shelly.SetAuth", params)
		return err
	})
}

// DisableAuth disables device authentication.
func (s *Service) DisableAuth(ctx context.Context, identifier string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		// Setting ha1 to null disables authentication
		params := map[string]any{
			"user":  "admin",
			"realm": "",
			"ha1":   nil,
		}
		_, err := conn.Call(ctx, "Shelly.SetAuth", params)
		return err
	})
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

// calculateHA1 calculates the HA1 hash for digest authentication.
// MD5 is required by the Shelly device protocol - not a security concern since
// this is a password hash transmitted over a local network to the device.
func calculateHA1(user, realm, password string) string {
	data := user + ":" + realm + ":" + password
	hash := md5.Sum([]byte(data)) //nolint:gosec // Required by Shelly digest auth protocol
	return hex.EncodeToString(hash[:])
}
