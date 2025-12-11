// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// ConfigResolver resolves device identifiers using the config package.
type ConfigResolver struct{}

// NewConfigResolver creates a new config-based device resolver.
func NewConfigResolver() *ConfigResolver {
	return &ConfigResolver{}
}

// Resolve resolves a device identifier to a model.Device.
func (r *ConfigResolver) Resolve(identifier string) (model.Device, error) {
	cfgDevice, err := config.ResolveDevice(identifier)
	if err != nil {
		return model.Device{}, err
	}

	return configDeviceToModel(cfgDevice), nil
}

// configDeviceToModel converts a config.Device to a model.Device.
func configDeviceToModel(d config.Device) model.Device {
	result := model.Device{
		Name:       d.Name,
		Address:    d.Address,
		Generation: d.Generation,
		Type:       d.Type,
		Model:      d.Model,
	}

	if d.Auth != nil {
		result.Auth = &model.Auth{
			Username: d.Auth.Username,
			Password: d.Auth.Password,
		}
	}

	return result
}

// NewService creates a new Shelly service with the default config resolver.
func NewService() *Service {
	return New(NewConfigResolver())
}
