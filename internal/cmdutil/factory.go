// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
// It follows the gh CLI and kubectl patterns for dependency injection and command helpers.
package cmdutil

import (
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Factory provides dependencies to commands through lazy initialization.
// This pattern allows for easy testing by replacing the factory functions
// with mocks, while also avoiding initialization until dependencies are needed.
type Factory struct {
	// IOStreams provides access to stdin/stdout/stderr and terminal capabilities.
	IOStreams func() *iostreams.IOStreams

	// Config provides access to the CLI configuration.
	Config func() (*config.Config, error)

	// ShellyService provides the business logic service for device operations.
	ShellyService func() *shelly.Service

	// Cached instances - set after first call to avoid re-initialization.
	ioStreams     *iostreams.IOStreams
	cfg           *config.Config
	shellyService *shelly.Service
}

// New creates a Factory with production dependencies.
// Dependencies are lazily initialized on first access.
func New() *Factory {
	f := &Factory{}

	f.IOStreams = func() *iostreams.IOStreams {
		if f.ioStreams == nil {
			f.ioStreams = iostreams.System()
		}
		return f.ioStreams
	}

	f.Config = func() (*config.Config, error) {
		if f.cfg == nil {
			cfg, err := config.Load()
			if err != nil {
				return nil, err
			}
			f.cfg = cfg
		}
		return f.cfg, nil
	}

	f.ShellyService = func() *shelly.Service {
		if f.shellyService == nil {
			f.shellyService = shelly.NewService()
		}
		return f.shellyService
	}

	return f
}

// NewWithIOStreams creates a Factory with a custom IOStreams instance.
// Useful for testing or when you need to override standard I/O.
func NewWithIOStreams(ios *iostreams.IOStreams) *Factory {
	f := New()
	f.ioStreams = ios
	f.IOStreams = func() *iostreams.IOStreams {
		return ios
	}
	return f
}

// SetIOStreams sets a custom IOStreams instance on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetIOStreams(ios *iostreams.IOStreams) *Factory {
	f.ioStreams = ios
	origIOStreams := f.IOStreams
	f.IOStreams = func() *iostreams.IOStreams {
		if f.ioStreams != nil {
			return f.ioStreams
		}
		return origIOStreams()
	}
	return f
}

// SetConfig sets a custom config instance on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetConfig(cfg *config.Config) *Factory {
	f.cfg = cfg
	origConfig := f.Config
	f.Config = func() (*config.Config, error) {
		if f.cfg != nil {
			return f.cfg, nil
		}
		return origConfig()
	}
	return f
}

// SetShellyService sets a custom shelly service on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetShellyService(svc *shelly.Service) *Factory {
	f.shellyService = svc
	origService := f.ShellyService
	f.ShellyService = func() *shelly.Service {
		if f.shellyService != nil {
			return f.shellyService
		}
		return origService()
	}
	return f
}

// MustConfig returns the configuration, panicking on error.
// Use this only when config errors should be fatal (e.g., during initialization).
func (f *Factory) MustConfig() *config.Config {
	cfg, err := f.Config()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	return cfg
}
