// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
// It follows the gh CLI and kubectl patterns for dependency injection and command helpers.
package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
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

	// Browser provides cross-platform URL opening capabilities.
	Browser func() browser.Browser

	// Cached instances - set after first call to avoid re-initialization.
	ioStreams     *iostreams.IOStreams
	cfg           *config.Config
	shellyService *shelly.Service
	browserInst   browser.Browser
}

// NewFactory creates a Factory with production dependencies.
// Dependencies are lazily initialized on first access.
func NewFactory() *Factory {
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

	f.Browser = func() browser.Browser {
		if f.browserInst == nil {
			f.browserInst = browser.New()
		}
		return f.browserInst
	}

	return f
}

// NewWithIOStreams creates a Factory with a custom IOStreams instance.
// Useful for testing or when you need to override standard I/O.
func NewWithIOStreams(ios *iostreams.IOStreams) *Factory {
	f := NewFactory()
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

// SetBrowser sets a custom browser instance on an existing factory.
// This modifies the factory in-place and returns it for chaining.
func (f *Factory) SetBrowser(b browser.Browser) *Factory {
	f.browserInst = b
	origBrowser := f.Browser
	f.Browser = func() browser.Browser {
		if f.browserInst != nil {
			return f.browserInst
		}
		return origBrowser()
	}
	return f
}

// =============================================================================
// Context Helpers
// =============================================================================

// WithTimeout creates a child context with the specified timeout.
// This is a convenience wrapper around context.WithTimeout.
func (f *Factory) WithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

// WithDefaultTimeout creates a child context with DefaultTimeout (10s).
// Useful for standard device operations.
func (f *Factory) WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeout)
}

// =============================================================================
// Config Accessors
// =============================================================================

// GetDevice retrieves a device from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetDevice(name string) *model.Device {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if dev, ok := cfg.Devices[name]; ok {
		return &dev
	}
	return nil
}

// GetGroup retrieves a device group from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetGroup(name string) *config.Group {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if grp, ok := cfg.Groups[name]; ok {
		return &grp
	}
	return nil
}

// GetAlias retrieves an alias from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetAlias(name string) *config.Alias {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if alias, ok := cfg.Aliases[name]; ok {
		return &alias
	}
	return nil
}

// GetScene retrieves a scene from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetScene(name string) *config.Scene {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if scene, ok := cfg.Scenes[name]; ok {
		return &scene
	}
	return nil
}

// GetTemplate retrieves a template from config by name.
// Returns nil if not found or config fails to load.
func (f *Factory) GetTemplate(name string) *config.Template {
	cfg, err := f.Config()
	if err != nil {
		return nil
	}
	if tmpl, ok := cfg.Templates[name]; ok {
		return &tmpl
	}
	return nil
}

// =============================================================================
// Device Resolution
// =============================================================================

// ResolveAddress resolves a device identifier to its address.
// Checks config registry first, falls back to using identifier as address.
func (f *Factory) ResolveAddress(identifier string) string {
	if dev := f.GetDevice(identifier); dev != nil && dev.Address != "" {
		return dev.Address
	}
	return identifier
}

// ResolveDevice resolves identifier to full device info.
// Returns (device, true) if found in config, (nil, false) otherwise.
func (f *Factory) ResolveDevice(identifier string) (*model.Device, bool) {
	cfg, err := f.Config()
	if err != nil {
		return nil, false
	}
	if dev, ok := cfg.Devices[identifier]; ok {
		return &dev, true
	}
	return nil, false
}

// =============================================================================
// Target Expansion
// =============================================================================

// ExpandTargets expands a list of device identifiers.
// Handles --group flag, --all flag, and individual device args.
// Returns resolved addresses ready for operations.
func (f *Factory) ExpandTargets(args []string, groupName string, all bool) ([]string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	var targets []string

	switch {
	case all:
		// Get all registered devices
		for _, dev := range cfg.Devices {
			if dev.Address != "" {
				targets = append(targets, dev.Address)
			}
		}
	case groupName != "":
		// Get devices from group
		group, ok := cfg.Groups[groupName]
		if !ok {
			return nil, fmt.Errorf("group %q not found", groupName)
		}
		for _, name := range group.Devices {
			targets = append(targets, f.ResolveAddress(name))
		}
	default:
		// Use provided args
		for _, arg := range args {
			targets = append(targets, f.ResolveAddress(arg))
		}
	}

	if len(targets) == 0 {
		return nil, errors.New("no target devices specified")
	}

	return targets, nil
}

// =============================================================================
// Confirmation Helpers
// =============================================================================

// ConfirmAction prompts for confirmation, respecting the yes flag.
// Returns true if confirmed (or yes is true), false otherwise.
func (f *Factory) ConfirmAction(message string, yes bool) (bool, error) {
	if yes {
		return true, nil
	}
	return f.IOStreams().Confirm(message, false)
}

// ConfirmDangerousAction prompts for dangerous operation confirmation.
// Requires typing "yes" to confirm. Respects yes flag.
func (f *Factory) ConfirmDangerousAction(message string, yes bool) (bool, error) {
	if yes {
		return true, nil
	}
	return f.IOStreams().ConfirmDanger(message)
}

// =============================================================================
// Output Format Helpers
// =============================================================================

// OutputFormat returns the current output format (json, yaml, table, or "").
func (f *Factory) OutputFormat() string {
	return viper.GetString("output")
}

// IsJSONOutput returns true if JSON output is requested.
func (f *Factory) IsJSONOutput() bool {
	return f.OutputFormat() == "json"
}

// IsYAMLOutput returns true if YAML output is requested.
func (f *Factory) IsYAMLOutput() bool {
	return f.OutputFormat() == "yaml"
}

// IsStructuredOutput returns true if JSON or YAML output is requested.
func (f *Factory) IsStructuredOutput() bool {
	format := f.OutputFormat()
	return format == "json" || format == "yaml"
}

// =============================================================================
// Logger Convenience
// =============================================================================

// Logger returns the structured logger from IOStreams.
func (f *Factory) Logger() *iostreams.Logger {
	return f.IOStreams().Logger()
}
