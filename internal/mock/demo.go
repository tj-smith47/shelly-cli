package mock

import (
	"context"
	"os"
	"sync"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Demo coordinates all mock components for demo mode.
type Demo struct {
	Fixtures     *Fixtures
	ConfigMgr    *config.Manager
	DeviceServer *DeviceServer
	cleanup      []func()
}

var (
	currentDemo   *Demo
	currentDemoMu sync.RWMutex
)

// GetCurrentDemo returns the current demo instance if demo mode is active.
func GetCurrentDemo() *Demo {
	currentDemoMu.RLock()
	defer currentDemoMu.RUnlock()
	return currentDemo
}

// setCurrentDemo sets the current demo instance.
func setCurrentDemo(d *Demo) {
	currentDemoMu.Lock()
	defer currentDemoMu.Unlock()
	currentDemo = d
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
	deviceServer := NewDeviceServer(fixtures)

	cfg := FixturesToConfigWithMockURLs(fixtures, deviceServer)
	configMgr := config.NewTestManager(cfg)

	d := &Demo{
		Fixtures:     fixtures,
		ConfigMgr:    configMgr,
		DeviceServer: deviceServer,
		cleanup:      []func(){deviceServer.Close},
	}

	// Set as current demo for global access
	setCurrentDemo(d)

	return d, nil
}

// mockResolver resolves devices using the mock's config manager directly.
// This avoids reliance on global state for test isolation.
type mockResolver struct {
	mgr *config.Manager
}

func (r *mockResolver) Resolve(identifier string) (model.Device, error) {
	return r.mgr.ResolveDevice(identifier)
}

func (r *mockResolver) ResolveWithGeneration(_ context.Context, identifier string) (model.Device, error) {
	return r.mgr.ResolveDevice(identifier)
}

// InjectIntoFactory configures a cmdutil.Factory to use mock components.
// It injects both the config manager AND a shelly service with a resolver
// bound to the mock's config manager. This ensures test isolation when
// tests run in parallel across packages.
func (d *Demo) InjectIntoFactory(f *cmdutil.Factory) {
	f.SetConfigManager(d.ConfigMgr)
	config.SetDefaultManager(d.ConfigMgr)

	// Create a shelly service with a resolver bound to this mock's config manager.
	// This is critical for test isolation - without it, parallel tests would
	// stomp on each other's global config manager.
	resolver := &mockResolver{mgr: d.ConfigMgr}
	svc := shelly.New(resolver)
	f.SetShellyService(svc)
}

// Cleanup shuts down all mock servers and resources.
func (d *Demo) Cleanup() {
	for _, fn := range d.cleanup {
		fn()
	}
	// Clear global reference
	setCurrentDemo(nil)
}

// GetDeviceAddress returns the mock server URL for a device.
func (d *Demo) GetDeviceAddress(deviceName string) string {
	if dev, ok := d.ConfigMgr.GetDevice(deviceName); ok {
		return dev.Address
	}
	return ""
}
