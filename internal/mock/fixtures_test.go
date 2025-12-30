package mock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFixtures(t *testing.T) {
	t.Parallel()

	t.Run("loads valid fixtures", func(t *testing.T) {
		t.Parallel()
		fixtures, err := LoadFixtures("../../testdata/demo/fixtures.yaml")
		require.NoError(t, err)
		require.NotNil(t, fixtures)

		assert.Equal(t, "1", fixtures.Version)
		assert.GreaterOrEqual(t, len(fixtures.Config.Devices), 1)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		t.Parallel()
		_, err := LoadFixtures("/non/existent/file.yaml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		badFile := filepath.Join(tmpDir, "bad.yaml")
		err := os.WriteFile(badFile, []byte("not: valid: yaml: ["), 0o600)
		require.NoError(t, err)

		_, err = LoadFixtures(badFile)
		assert.Error(t, err)
	})
}

func TestLoadFixturesFromBytes(t *testing.T) {
	t.Parallel()

	t.Run("loads valid YAML", func(t *testing.T) {
		t.Parallel()
		yaml := []byte(`
version: "1"
config:
  devices:
    - name: "Test Device"
      address: "192.168.1.1"
      mac: "AA:BB:CC:DD:EE:FF"
      model: "Test Model"
      type: "SHSW-1"
      generation: 2
`)
		fixtures, err := LoadFixturesFromBytes(yaml)
		require.NoError(t, err)
		require.NotNil(t, fixtures)

		assert.Equal(t, "1", fixtures.Version)
		assert.Len(t, fixtures.Config.Devices, 1)
		assert.Equal(t, "Test Device", fixtures.Config.Devices[0].Name)
	})

	t.Run("initializes nil maps", func(t *testing.T) {
		t.Parallel()
		yaml := []byte(`version: "1"`)
		fixtures, err := LoadFixturesFromBytes(yaml)
		require.NoError(t, err)
		assert.NotNil(t, fixtures.DeviceStates)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		t.Parallel()
		_, err := LoadFixturesFromBytes([]byte("not: valid: yaml: ["))
		assert.Error(t, err)
	})
}

//nolint:paralleltest // Tests use t.Setenv/os.Unsetenv, cannot run in parallel
func TestDefaultFixturePath(t *testing.T) {
	t.Run("uses environment variable when set", func(t *testing.T) {
		t.Setenv("SHELLY_DEMO_FIXTURES", "/custom/path.yaml")
		assert.Equal(t, "/custom/path.yaml", DefaultFixturePath())
	})

	t.Run("returns default path when env not set", func(t *testing.T) {
		if err := os.Unsetenv("SHELLY_DEMO_FIXTURES"); err != nil {
			t.Logf("warning: %v", err)
		}
		assert.Equal(t, "testdata/demo/fixtures.yaml", DefaultFixturePath())
	})
}

//nolint:paralleltest // Subtests share fixtures object
func TestFixturesStructure(t *testing.T) {
	yaml := []byte(`
version: "1"
config:
  devices:
    - name: "Device 1"
      address: "192.168.1.1"
      mac: "AA:BB:CC:DD:EE:01"
      model: "Model 1"
      type: "SHSW-1"
      generation: 1
      auth_user: "admin"
      auth_pass: "secret"
  groups:
    - name: "test-group"
      devices: ["Device 1"]
  scenes:
    - name: "test-scene"
      description: "Test scene"
      actions:
        - device: "Device 1"
          method: "Switch.Set"
          params:
            on: true
  aliases:
    - name: "testalias"
      command: "switch on device"
      shell: true
device_states:
  "Device 1":
    "switch:0":
      output: true
fleet:
  organization: "Test Corp"
  devices:
    - id: "fleet-1"
      name: "Fleet Device"
      model: "SHSW-1"
      online: true
      firmware: "1.0.0"
discovery:
  - name: "Discovered"
    address: "192.168.1.100"
    mac: "FF:FF:FF:FF:FF:FF"
    model: "SHPLG-S"
    generation: 1
`)

	fixtures, err := LoadFixturesFromBytes(yaml)
	require.NoError(t, err)

	t.Run("parses devices", func(t *testing.T) {
		require.Len(t, fixtures.Config.Devices, 1)
		d := fixtures.Config.Devices[0]
		assert.Equal(t, "Device 1", d.Name)
		assert.Equal(t, "192.168.1.1", d.Address)
		assert.Equal(t, "AA:BB:CC:DD:EE:01", d.MAC)
		assert.Equal(t, "Model 1", d.Model)
		assert.Equal(t, "SHSW-1", d.Type)
		assert.Equal(t, 1, d.Generation)
		assert.Equal(t, "admin", d.AuthUser)
		assert.Equal(t, "secret", d.AuthPass)
	})

	t.Run("parses groups", func(t *testing.T) {
		require.Len(t, fixtures.Config.Groups, 1)
		g := fixtures.Config.Groups[0]
		assert.Equal(t, "test-group", g.Name)
		assert.Equal(t, []string{"Device 1"}, g.Devices)
	})

	t.Run("parses scenes", func(t *testing.T) {
		require.Len(t, fixtures.Config.Scenes, 1)
		s := fixtures.Config.Scenes[0]
		assert.Equal(t, "test-scene", s.Name)
		assert.Equal(t, "Test scene", s.Description)
		require.Len(t, s.Actions, 1)
		assert.Equal(t, "Device 1", s.Actions[0].Device)
		assert.Equal(t, "Switch.Set", s.Actions[0].Method)
		assert.Equal(t, true, s.Actions[0].Params["on"])
	})

	t.Run("parses aliases", func(t *testing.T) {
		require.Len(t, fixtures.Config.Aliases, 1)
		a := fixtures.Config.Aliases[0]
		assert.Equal(t, "testalias", a.Name)
		assert.Equal(t, "switch on device", a.Command)
		assert.True(t, a.Shell)
	})

	t.Run("parses device states", func(t *testing.T) {
		require.NotNil(t, fixtures.DeviceStates)
		state, ok := fixtures.DeviceStates["Device 1"]
		require.True(t, ok)
		switchData := state["switch:0"]
		require.NotNil(t, switchData, "switch:0 should exist")
	})

	t.Run("parses fleet", func(t *testing.T) {
		assert.Equal(t, "Test Corp", fixtures.Fleet.Organization)
		require.Len(t, fixtures.Fleet.Devices, 1)
		fd := fixtures.Fleet.Devices[0]
		assert.Equal(t, "fleet-1", fd.ID)
		assert.Equal(t, "Fleet Device", fd.Name)
		assert.Equal(t, "SHSW-1", fd.Model)
		assert.True(t, fd.Online)
		assert.Equal(t, "1.0.0", fd.Firmware)
	})

	t.Run("parses discovery", func(t *testing.T) {
		require.Len(t, fixtures.Discovery, 1)
		dd := fixtures.Discovery[0]
		assert.Equal(t, "Discovered", dd.Name)
		assert.Equal(t, "192.168.1.100", dd.Address)
		assert.Equal(t, "FF:FF:FF:FF:FF:FF", dd.MAC)
		assert.Equal(t, "SHPLG-S", dd.Model)
		assert.Equal(t, 1, dd.Generation)
	})
}
