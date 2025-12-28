// Package device provides device-level operations for Shelly devices.
package device

import (
	"context"

	"github.com/tj-smith47/shelly-go/gen2/components"

	"github.com/tj-smith47/shelly-cli/internal/client"
)

// SysStatus represents system status information.
type SysStatus struct {
	MAC             string
	Uptime          int
	Time            string
	Unixtime        int64
	RAMFree         int
	RAMSize         int
	FSFree          int
	FSSize          int
	RestartRequired bool
	CfgRev          int
	UpdateAvailable string
}

// SysConfig represents system configuration.
type SysConfig struct {
	Name         string
	Timezone     string
	Lat          float64
	Lng          float64
	EcoMode      bool
	Discoverable bool
	Profile      string
	SNTPServer   string
}

// GetSysStatus returns system status information.
func (s *Service) GetSysStatus(ctx context.Context, identifier string) (*SysStatus, error) {
	var result *SysStatus
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		status, err := sys.GetStatus(ctx)
		if err != nil {
			return err
		}

		result = &SysStatus{
			MAC:             status.MAC,
			Uptime:          status.Uptime,
			RAMFree:         status.RAMFree,
			RAMSize:         status.RAMSize,
			FSFree:          status.FSFree,
			FSSize:          status.FSSize,
			RestartRequired: status.RestartRequired,
			CfgRev:          status.CfgRev,
		}
		if status.Time != nil {
			result.Time = *status.Time
		}
		if status.Unixtime != nil {
			result.Unixtime = *status.Unixtime
		}
		if status.AvailableUpdates != nil && status.AvailableUpdates.Stable != nil {
			result.UpdateAvailable = status.AvailableUpdates.Stable.Version
		}
		return nil
	})
	return result, err
}

// GetSysConfig returns system configuration.
func (s *Service) GetSysConfig(ctx context.Context, identifier string) (*SysConfig, error) {
	var result *SysConfig
	err := s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		config, err := sys.GetConfig(ctx)
		if err != nil {
			return err
		}

		result = &SysConfig{}
		extractDeviceConfig(config.Device, result)
		extractLocationConfig(config.Location, result)
		if config.SNTP != nil && config.SNTP.Server != nil {
			result.SNTPServer = *config.SNTP.Server
		}
		return nil
	})
	return result, err
}

// SetSysName sets the device name.
func (s *Service) SetSysName(ctx context.Context, identifier, name string) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		return sys.SetConfig(ctx, &components.SysConfig{
			Device: &components.SysDeviceConfig{
				Name: &name,
			},
		})
	})
}

// SetSysTimezone sets the device timezone.
func (s *Service) SetSysTimezone(ctx context.Context, identifier, tz string) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		return sys.SetConfig(ctx, &components.SysConfig{
			Location: &components.SysLocationConfig{
				TZ: &tz,
			},
		})
	})
}

// SetSysLocation sets the device location (latitude and longitude).
func (s *Service) SetSysLocation(ctx context.Context, identifier string, lat, lng float64) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		return sys.SetConfig(ctx, &components.SysConfig{
			Location: &components.SysLocationConfig{
				Lat: &lat,
				Lng: &lng,
			},
		})
	})
}

// SetSysEcoMode enables or disables eco mode.
func (s *Service) SetSysEcoMode(ctx context.Context, identifier string, enable bool) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		return sys.SetConfig(ctx, &components.SysConfig{
			Device: &components.SysDeviceConfig{
				EcoMode: &enable,
			},
		})
	})
}

// SetSysDiscoverable enables or disables device discoverability.
func (s *Service) SetSysDiscoverable(ctx context.Context, identifier string, discoverable bool) error {
	return s.parent.WithConnection(ctx, identifier, func(conn *client.Client) error {
		sys := components.NewSys(conn.RPCClient())
		return sys.SetConfig(ctx, &components.SysConfig{
			Device: &components.SysDeviceConfig{
				Discoverable: &discoverable,
			},
		})
	})
}

func extractDeviceConfig(device *components.SysDeviceConfig, result *SysConfig) {
	if device == nil {
		return
	}
	if device.Name != nil {
		result.Name = *device.Name
	}
	if device.EcoMode != nil {
		result.EcoMode = *device.EcoMode
	}
	if device.Discoverable != nil {
		result.Discoverable = *device.Discoverable
	}
	if device.Profile != nil {
		result.Profile = *device.Profile
	}
}

func extractLocationConfig(location *components.SysLocationConfig, result *SysConfig) {
	if location == nil {
		return
	}
	if location.TZ != nil {
		result.Timezone = *location.TZ
	}
	if location.Lat != nil {
		result.Lat = *location.Lat
	}
	if location.Lng != nil {
		result.Lng = *location.Lng
	}
}
