package mock

import (
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// NewConfigManager creates a config.Manager from fixtures.
// This returns a real *config.Manager populated with fixture data,
// allowing seamless integration with the factory pattern.
func NewConfigManager(fixtures *Fixtures) *config.Manager {
	cfg := FixturesToConfig(fixtures)
	return config.NewTestManager(cfg)
}

// FixturesToConfig converts fixture data to a config.Config struct.
func FixturesToConfig(fixtures *Fixtures) *config.Config {
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Groups:  make(map[string]config.Group),
		Scenes:  make(map[string]config.Scene),
		Aliases: make(map[string]config.Alias),
	}

	for _, d := range fixtures.Config.Devices {
		key := config.NormalizeDeviceName(d.Name)
		dev := model.Device{
			Name:       d.Name,
			Address:    d.Address,
			MAC:        d.MAC,
			Model:      d.Model,
			Type:       d.Type,
			Generation: d.Generation,
		}
		if d.AuthUser != "" || d.AuthPass != "" {
			dev.Auth = &model.Auth{
				Username: d.AuthUser,
				Password: d.AuthPass,
			}
		}
		cfg.Devices[key] = dev
	}

	for _, g := range fixtures.Config.Groups {
		cfg.Groups[g.Name] = config.Group{
			Devices: g.Devices,
		}
	}

	for _, s := range fixtures.Config.Scenes {
		actions := make([]config.SceneAction, len(s.Actions))
		for i, a := range s.Actions {
			actions[i] = config.SceneAction{
				Device: a.Device,
				Method: a.Method,
				Params: a.Params,
			}
		}
		cfg.Scenes[s.Name] = config.Scene{
			Name:        s.Name,
			Description: s.Description,
			Actions:     actions,
		}
	}

	for _, a := range fixtures.Config.Aliases {
		cfg.Aliases[a.Name] = config.Alias{
			Command: a.Command,
			Shell:   a.Shell,
		}
	}

	return cfg
}

// FixturesToConfigWithMockURLs converts fixtures to config with mock server URLs.
func FixturesToConfigWithMockURLs(fixtures *Fixtures, server *DeviceServer) *config.Config {
	cfg := &config.Config{
		Devices: make(map[string]model.Device),
		Groups:  make(map[string]config.Group),
		Scenes:  make(map[string]config.Scene),
		Aliases: make(map[string]config.Alias),
	}

	for _, d := range fixtures.Config.Devices {
		key := config.NormalizeDeviceName(d.Name)
		dev := model.Device{
			Name:       d.Name,
			Address:    server.DeviceURL(d.Name),
			MAC:        d.MAC,
			Model:      d.Model,
			Type:       d.Type,
			Generation: d.Generation,
			Platform:   d.Platform,
		}
		if d.AuthUser != "" || d.AuthPass != "" {
			dev.Auth = &model.Auth{
				Username: d.AuthUser,
				Password: d.AuthPass,
			}
		}
		cfg.Devices[key] = dev
	}

	for _, g := range fixtures.Config.Groups {
		cfg.Groups[g.Name] = config.Group{
			Devices: g.Devices,
		}
	}

	for _, s := range fixtures.Config.Scenes {
		actions := make([]config.SceneAction, len(s.Actions))
		for i, a := range s.Actions {
			actions[i] = config.SceneAction{
				Device: a.Device,
				Method: a.Method,
				Params: a.Params,
			}
		}
		cfg.Scenes[s.Name] = config.Scene{
			Name:        s.Name,
			Description: s.Description,
			Actions:     actions,
		}
	}

	for _, a := range fixtures.Config.Aliases {
		cfg.Aliases[a.Name] = config.Alias{
			Command: a.Command,
			Shell:   a.Shell,
		}
	}

	return cfg
}
