// Package network provides network-related services for Shelly devices.
package network

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// MQTTStatus holds MQTT status information.
type MQTTStatus struct {
	Connected bool `json:"connected"`
}

// MQTTConfig holds MQTT configuration.
type MQTTConfig struct {
	Enable      bool   `json:"enable"`
	Server      string `json:"server,omitempty"`
	User        string `json:"user,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	TopicPrefix string `json:"topic_prefix,omitempty"`
	SSLCA       string `json:"ssl_ca,omitempty"`
	RPCNTF      bool   `json:"rpc_ntf,omitempty"`
	StatusNTF   bool   `json:"status_ntf,omitempty"`
}

// MQTTService provides MQTT-related operations for Shelly devices.
type MQTTService struct {
	provider ConnectionProvider
}

// NewMQTTService creates a new MQTT service.
func NewMQTTService(provider ConnectionProvider) *MQTTService {
	return &MQTTService{provider: provider}
}

// GetStatus returns the MQTT status.
func (s *MQTTService) GetStatus(ctx context.Context, identifier string) (*MQTTStatus, error) {
	var result *MQTTStatus
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// GetConfig returns the MQTT configuration.
func (s *MQTTService) GetConfig(ctx context.Context, identifier string) (*MQTTConfig, error) {
	var result *MQTTConfig
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mqtt := components.NewMQTT(conn.RPCClient())
		config, err := mqtt.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = &MQTTConfig{}
		if config.Enable != nil {
			result.Enable = *config.Enable
		}
		if config.Server != nil {
			result.Server = *config.Server
		}
		if config.User != nil {
			result.User = *config.User
		}
		if config.ClientID != nil {
			result.ClientID = *config.ClientID
		}
		if config.TopicPrefix != nil {
			result.TopicPrefix = *config.TopicPrefix
		}
		if config.RPCNTF != nil {
			result.RPCNTF = *config.RPCNTF
		}
		if config.StatusNTF != nil {
			result.StatusNTF = *config.StatusNTF
		}
		if config.SSLCA != nil {
			result.SSLCA = *config.SSLCA
		}
		return nil
	})
	return result, err
}

// SetConfigParams holds parameters for SetConfig.
type SetConfigParams struct {
	Enable      *bool
	Server      string
	User        string
	Password    string
	ClientID    string
	TopicPrefix string
	SSLCA       string // TLS settings: "", "*", "ca.pem", or "user_ca.pem"
}

// SetConfig updates the MQTT configuration.
func (s *MQTTService) SetConfig(ctx context.Context, identifier string, params SetConfigParams) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		mqtt := components.NewMQTT(conn.RPCClient())

		mqttCfg := &components.MQTTConfig{}
		if params.Enable != nil {
			mqttCfg.Enable = params.Enable
		}
		if params.Server != "" {
			mqttCfg.Server = &params.Server
		}
		if params.User != "" {
			mqttCfg.User = &params.User
		}
		if params.Password != "" {
			mqttCfg.Pass = &params.Password
		}
		if params.TopicPrefix != "" {
			mqttCfg.TopicPrefix = &params.TopicPrefix
		}
		if params.ClientID != "" {
			mqttCfg.ClientID = &params.ClientID
		}
		if params.SSLCA != "" {
			mqttCfg.SSLCA = &params.SSLCA
		}

		return mqtt.SetConfig(ctx, mqttCfg)
	})
}
