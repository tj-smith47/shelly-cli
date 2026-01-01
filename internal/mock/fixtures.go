// Package mock provides demo mode infrastructure for testing and demonstrations.
// It enables running the CLI with mock device data instead of requiring real devices.
package mock

import (
	"os"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Fixtures holds all demo mode data loaded from YAML.
type Fixtures struct {
	Version string `yaml:"version"`

	// Config represents registered devices, groups, scenes, aliases
	Config ConfigFixture `yaml:"config"`

	// DeviceStates maps device names to their component states (HTTP responses)
	DeviceStates map[string]DeviceState `yaml:"device_states"`

	// Fleet represents enterprise fleet devices (WebSocket responses)
	Fleet FleetFixture `yaml:"fleet"`

	// Discovery represents devices found via mDNS/scanning
	Discovery []DiscoveredDevice `yaml:"discovery"`
}

// ConfigFixture holds registered config data.
type ConfigFixture struct {
	Devices []DeviceFixture `yaml:"devices"`
	Groups  []GroupFixture  `yaml:"groups"`
	Scenes  []SceneFixture  `yaml:"scenes"`
	Aliases []AliasFixture  `yaml:"aliases"`
}

// DeviceFixture represents a registered device.
type DeviceFixture struct {
	Name        string `yaml:"name"`
	Address     string `yaml:"address"`
	MAC         string `yaml:"mac"`
	Model       string `yaml:"model"`
	Type        string `yaml:"type"`
	Generation  int    `yaml:"generation"`
	Platform    string `yaml:"platform,omitempty"`
	AuthUser    string `yaml:"auth_user,omitempty"`
	AuthPass    string `yaml:"auth_pass,omitempty"`
	AuthEnabled bool   `yaml:"auth_enabled,omitempty"`
}

// GroupFixture represents a device group.
type GroupFixture struct {
	Name    string   `yaml:"name"`
	Devices []string `yaml:"devices"`
}

// SceneFixture represents a scene with actions.
type SceneFixture struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description,omitempty"`
	Actions     []SceneActionFixture `yaml:"actions"`
}

// SceneActionFixture represents a single scene action.
type SceneActionFixture struct {
	Device string         `yaml:"device"`
	Method string         `yaml:"method"`
	Params map[string]any `yaml:"params,omitempty"`
}

// AliasFixture represents a command alias.
type AliasFixture struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Shell   bool   `yaml:"shell,omitempty"`
}

// DeviceState holds component states for HTTP mocking.
// Keys are component keys like "switch:0", "cover:0", "sys", "relay".
type DeviceState map[string]any

// FleetFixture holds enterprise fleet data.
type FleetFixture struct {
	Organization string               `yaml:"organization"`
	Devices      []FleetDeviceFixture `yaml:"devices"`
}

// FleetDeviceFixture represents a fleet device.
type FleetDeviceFixture struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Model    string `yaml:"model"`
	Online   bool   `yaml:"online"`
	Firmware string `yaml:"firmware,omitempty"`
}

// DiscoveredDevice represents a device found via discovery.
type DiscoveredDevice struct {
	Name       string `yaml:"name"`
	Address    string `yaml:"address"`
	MAC        string `yaml:"mac"`
	Model      string `yaml:"model"`
	Generation int    `yaml:"generation"`
}

// LoadFixtures loads fixtures from a YAML file.
func LoadFixtures(path string) (*Fixtures, error) {
	data, err := afero.ReadFile(config.Fs(), path)
	if err != nil {
		return nil, err
	}

	return LoadFixturesFromBytes(data)
}

// LoadFixturesFromBytes loads fixtures from YAML bytes.
func LoadFixturesFromBytes(data []byte) (*Fixtures, error) {
	var f Fixtures
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, err
	}

	if f.DeviceStates == nil {
		f.DeviceStates = make(map[string]DeviceState)
	}

	return &f, nil
}

// DefaultFixturePath returns the default fixture file path.
func DefaultFixturePath() string {
	if path := os.Getenv("SHELLY_DEMO_FIXTURES"); path != "" {
		return path
	}
	return "testdata/demo/fixtures.yaml"
}
