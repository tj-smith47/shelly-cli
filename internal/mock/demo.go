package mock

import (
	"os"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Demo coordinates all mock components for demo mode.
type Demo struct {
	Fixtures  *Fixtures
	ConfigMgr *config.Manager
	cleanup   []func()
}

// IsDemoMode returns true if demo mode is enabled via environment variable.
func IsDemoMode() bool {
	val := os.Getenv("SHELLY_DEMO")
	return val == "1" || val == "true"
}

// Start initializes demo mode from the default fixture path.
func Start() (*Demo, error) {
	return StartWithPath(DefaultFixturePath())
}

// StartWithPath initializes demo mode from a specific fixture file.
func StartWithPath(path string) (*Demo, error) {
	fixtures, err := LoadFixtures(path)
	if err != nil {
		return nil, err
	}

	return StartWithFixtures(fixtures)
}

// StartWithFixtures initializes demo mode from pre-loaded fixtures.
func StartWithFixtures(fixtures *Fixtures) (*Demo, error) {
	d := &Demo{
		Fixtures:  fixtures,
		ConfigMgr: NewConfigManager(fixtures),
	}

	return d, nil
}

// InjectIntoFactory configures a cmdutil.Factory to use mock components.
func (d *Demo) InjectIntoFactory(f *cmdutil.Factory) {
	f.SetConfigManager(d.ConfigMgr)
}

// Cleanup shuts down all mock servers and resources.
func (d *Demo) Cleanup() {
	for _, fn := range d.cleanup {
		fn()
	}
}

// GetDeviceAddress returns the address for a device.
func (d *Demo) GetDeviceAddress(deviceName string) string {
	if dev, ok := d.ConfigMgr.GetDevice(deviceName); ok {
		return dev.Address
	}
	return ""
}
