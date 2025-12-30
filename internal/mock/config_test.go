package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	fixtures := &Fixtures{
		Version: "1",
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{
					Name:       "Test Device",
					Address:    "192.168.1.1",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Model:      "Test Model",
					Type:       "SHSW-1",
					Generation: 2,
				},
			},
		},
	}

	mgr := NewConfigManager(fixtures)
	require.NotNil(t, mgr)

	devices := mgr.ListDevices()
	assert.Len(t, devices, 1)
}

func TestFixturesToConfig(t *testing.T) {
	t.Parallel()

	t.Run("converts devices", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Devices: []DeviceFixture{
					{
						Name:       "Device 1",
						Address:    "192.168.1.1",
						MAC:        "AA:BB:CC:DD:EE:01",
						Model:      "Model 1",
						Type:       "SHSW-1",
						Generation: 1,
					},
					{
						Name:       "Device 2",
						Address:    "192.168.1.2",
						MAC:        "AA:BB:CC:DD:EE:02",
						Model:      "Model 2",
						Type:       "SNSW-001P16EU",
						Generation: 2,
						AuthUser:   "admin",
						AuthPass:   "secret",
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		require.NotNil(t, cfg)
		assert.Len(t, cfg.Devices, 2)

		d1, ok := cfg.Devices["device-1"]
		require.True(t, ok)
		assert.Equal(t, "Device 1", d1.Name)
		assert.Equal(t, "192.168.1.1", d1.Address)
		assert.Equal(t, "AA:BB:CC:DD:EE:01", d1.MAC)
		assert.Equal(t, "Model 1", d1.Model)
		assert.Equal(t, "SHSW-1", d1.Type)
		assert.Equal(t, 1, d1.Generation)
		assert.Nil(t, d1.Auth)

		d2, ok := cfg.Devices["device-2"]
		require.True(t, ok)
		assert.Equal(t, "Device 2", d2.Name)
		require.NotNil(t, d2.Auth)
		assert.Equal(t, "admin", d2.Auth.Username)
		assert.Equal(t, "secret", d2.Auth.Password)
	})

	t.Run("converts groups", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Groups: []GroupFixture{
					{
						Name:    "test-group",
						Devices: []string{"device1", "device2"},
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		assert.Len(t, cfg.Groups, 1)
		g, ok := cfg.Groups["test-group"]
		require.True(t, ok)
		assert.Equal(t, "test-group", g.Name)
		assert.Equal(t, []string{"device1", "device2"}, g.Devices)
	})

	t.Run("converts scenes", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Scenes: []SceneFixture{
					{
						Name:        "test-scene",
						Description: "Test description",
						Actions: []SceneActionFixture{
							{
								Device: "device1",
								Method: "Switch.Set",
								Params: map[string]any{"on": true},
							},
						},
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		assert.Len(t, cfg.Scenes, 1)
		s, ok := cfg.Scenes["test-scene"]
		require.True(t, ok)
		assert.Equal(t, "test-scene", s.Name)
		assert.Equal(t, "Test description", s.Description)
		require.Len(t, s.Actions, 1)
		assert.Equal(t, "device1", s.Actions[0].Device)
		assert.Equal(t, "Switch.Set", s.Actions[0].Method)
		assert.Equal(t, true, s.Actions[0].Params["on"])
	})

	t.Run("converts aliases", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Aliases: []AliasFixture{
					{
						Name:    "test-alias",
						Command: "switch on device",
						Shell:   true,
					},
					{
						Name:    "simple-alias",
						Command: "status",
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		assert.Len(t, cfg.Aliases, 2)

		a1, ok := cfg.Aliases["test-alias"]
		require.True(t, ok)
		assert.Equal(t, "test-alias", a1.Name)
		assert.Equal(t, "switch on device", a1.Command)
		assert.True(t, a1.Shell)

		a2, ok := cfg.Aliases["simple-alias"]
		require.True(t, ok)
		assert.False(t, a2.Shell)
	})

	t.Run("handles empty fixtures", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{}

		cfg := FixturesToConfig(fixtures)
		require.NotNil(t, cfg)
		assert.Empty(t, cfg.Devices)
		assert.Empty(t, cfg.Groups)
		assert.Empty(t, cfg.Scenes)
		assert.Empty(t, cfg.Aliases)
	})

	t.Run("handles auth with only password", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Devices: []DeviceFixture{
					{
						Name:     "Device",
						Address:  "192.168.1.1",
						AuthPass: "password-only",
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		d, ok := cfg.Devices["device"]
		require.True(t, ok)
		require.NotNil(t, d.Auth)
		assert.Empty(t, d.Auth.Username)
		assert.Equal(t, "password-only", d.Auth.Password)
	})

	t.Run("handles auth with only username", func(t *testing.T) {
		t.Parallel()
		fixtures := &Fixtures{
			Config: ConfigFixture{
				Devices: []DeviceFixture{
					{
						Name:     "Device",
						Address:  "192.168.1.1",
						AuthUser: "username-only",
					},
				},
			},
		}

		cfg := FixturesToConfig(fixtures)
		d, ok := cfg.Devices["device"]
		require.True(t, ok)
		require.NotNil(t, d.Auth)
		assert.Equal(t, "username-only", d.Auth.Username)
		assert.Empty(t, d.Auth.Password)
	})
}
