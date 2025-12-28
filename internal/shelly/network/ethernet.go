// Package network provides network-related services for Shelly devices.
package network

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// EthernetStatus holds ethernet status information.
type EthernetStatus struct {
	IP string `json:"ip"`
}

// EthernetConfig holds ethernet configuration.
type EthernetConfig struct {
	Enable     bool   `json:"enable"`
	IPv4Mode   string `json:"ipv4mode,omitempty"`
	IP         string `json:"ip,omitempty"`
	Netmask    string `json:"netmask,omitempty"`
	GW         string `json:"gw,omitempty"`
	Nameserver string `json:"nameserver,omitempty"`
}

// EthernetService provides ethernet-related operations for Shelly devices.
type EthernetService struct {
	provider ConnectionProvider
}

// NewEthernetService creates a new ethernet service.
func NewEthernetService(provider ConnectionProvider) *EthernetService {
	return &EthernetService{provider: provider}
}

// GetStatus returns the ethernet status.
func (s *EthernetService) GetStatus(ctx context.Context, identifier string) (*EthernetStatus, error) {
	var result *EthernetStatus
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
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

// GetConfig returns the ethernet configuration.
func (s *EthernetService) GetConfig(ctx context.Context, identifier string) (*EthernetConfig, error) {
	var result *EthernetConfig
	err := s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		eth := components.NewEthernet(conn.RPCClient())
		config, err := eth.GetConfig(ctx)
		if err != nil {
			return err
		}
		result = &EthernetConfig{}
		if config.Enable != nil {
			result.Enable = *config.Enable
		}
		if config.IPv4Mode != nil {
			result.IPv4Mode = *config.IPv4Mode
		}
		if config.IP != nil {
			result.IP = *config.IP
		}
		if config.Netmask != nil {
			result.Netmask = *config.Netmask
		}
		if config.GW != nil {
			result.GW = *config.GW
		}
		if config.Nameserver != nil {
			result.Nameserver = *config.Nameserver
		}
		return nil
	})
	return result, err
}

// EthernetSetConfigParams holds parameters for SetConfig.
type EthernetSetConfigParams struct {
	Enable     *bool
	IPv4Mode   string
	IP         string
	Netmask    string
	GW         string
	Nameserver string
}

// SetConfig updates the ethernet configuration.
func (s *EthernetService) SetConfig(ctx context.Context, identifier string, params EthernetSetConfigParams) error {
	return s.provider.WithConnection(ctx, identifier, func(conn *client.Client) error {
		eth := components.NewEthernet(conn.RPCClient())

		ethCfg := &components.EthernetConfig{}
		if params.Enable != nil {
			ethCfg.Enable = params.Enable
		}
		if params.IPv4Mode != "" {
			ethCfg.IPv4Mode = &params.IPv4Mode
		}
		if params.IP != "" {
			ethCfg.IP = &params.IP
		}
		if params.Netmask != "" {
			ethCfg.Netmask = &params.Netmask
		}
		if params.GW != "" {
			ethCfg.GW = &params.GW
		}
		if params.Nameserver != "" {
			ethCfg.Nameserver = &params.Nameserver
		}

		return eth.SetConfig(ctx, ethCfg)
	})
}
