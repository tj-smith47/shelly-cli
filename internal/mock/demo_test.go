package mock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestIsDemoMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"enabled with 1", "1", true},
		{"enabled with true", "true", true},
		{"disabled with 0", "0", false},
		{"disabled with false", "false", false},
		{"disabled with empty", "", false},
		{"disabled with random value", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELLY_DEMO", tt.envValue)
			assert.Equal(t, tt.want, IsDemoMode())
		})
	}
}

//nolint:paralleltest // Tests use t.Setenv and os.Chdir, cannot run in parallel
func TestStart(t *testing.T) {
	t.Run("fails with default path when file missing", func(t *testing.T) {
		if err := os.Unsetenv("SHELLY_DEMO_FIXTURES"); err != nil {
			t.Logf("warning: %v", err)
		}

		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Logf("warning: %v", err)
			}
		}()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		_, err = Start()
		assert.Error(t, err)
	})

	t.Run("succeeds with custom fixture path", func(t *testing.T) {
		tmpDir := t.TempDir()
		fixturePath := filepath.Join(tmpDir, "fixtures.yaml")
		err := os.WriteFile(fixturePath, []byte(`version: "1"`), 0o600)
		require.NoError(t, err)

		t.Setenv("SHELLY_DEMO_FIXTURES", fixturePath)
		demo, err := Start()
		require.NoError(t, err)
		require.NotNil(t, demo)
		assert.Equal(t, "1", demo.Fixtures.Version)
	})
}

func TestStartWithPath(t *testing.T) {
	t.Parallel()

	t.Run("loads fixtures from path", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		fixturePath := filepath.Join(tmpDir, "fixtures.yaml")
		yaml := []byte(`
version: "1"
config:
  devices:
    - name: "Test Device"
      address: "192.168.1.1"
`)
		err := os.WriteFile(fixturePath, yaml, 0o600)
		require.NoError(t, err)

		demo, err := StartWithPath(fixturePath)
		require.NoError(t, err)
		require.NotNil(t, demo)
		assert.Equal(t, "1", demo.Fixtures.Version)
		assert.Len(t, demo.Fixtures.Config.Devices, 1)
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		t.Parallel()
		_, err := StartWithPath("/non/existent/path.yaml")
		assert.Error(t, err)
	})
}

func TestStartWithFixtures(t *testing.T) {
	t.Parallel()

	fixtures := &Fixtures{
		Version: "1",
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{
					Name:       "Device 1",
					Address:    "192.168.1.1",
					MAC:        "AA:BB:CC:DD:EE:FF",
					Generation: 2,
				},
			},
		},
	}

	demo, err := StartWithFixtures(fixtures)
	require.NoError(t, err)
	require.NotNil(t, demo)

	assert.Equal(t, fixtures, demo.Fixtures)
	assert.NotNil(t, demo.ConfigMgr)

	devices := demo.ConfigMgr.ListDevices()
	assert.Len(t, devices, 1)
}

func TestDemo_InjectIntoFactory(t *testing.T) {
	t.Parallel()

	fixtures := &Fixtures{
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{
					Name:       "Test Device",
					Address:    "192.168.1.1",
					Generation: 2,
				},
			},
		},
	}

	demo, err := StartWithFixtures(fixtures)
	require.NoError(t, err)

	factory := cmdutil.NewFactory()
	demo.InjectIntoFactory(factory)

	mgr, err := factory.ConfigManager()
	require.NoError(t, err)

	devices := mgr.ListDevices()
	assert.Len(t, devices, 1)
	_, ok := devices["test-device"]
	assert.True(t, ok)
}

func TestDemo_Cleanup(t *testing.T) {
	t.Parallel()

	cleanupCalled := false
	demo := &Demo{
		cleanup: []func(){
			func() { cleanupCalled = true },
		},
	}

	demo.Cleanup()
	assert.True(t, cleanupCalled)
}

func TestDemo_GetDeviceAddress(t *testing.T) {
	t.Parallel()

	fixtures := &Fixtures{
		Config: ConfigFixture{
			Devices: []DeviceFixture{
				{
					Name:    "Test Device",
					Address: "192.168.1.100",
				},
			},
		},
	}

	demo, err := StartWithFixtures(fixtures)
	require.NoError(t, err)

	t.Run("returns address for existing device", func(t *testing.T) {
		t.Parallel()
		addr := demo.GetDeviceAddress("Test Device")
		assert.Equal(t, "192.168.1.100", addr)
	})

	t.Run("returns address using normalized name", func(t *testing.T) {
		t.Parallel()
		addr := demo.GetDeviceAddress("test-device")
		assert.Equal(t, "192.168.1.100", addr)
	})

	t.Run("returns empty for non-existent device", func(t *testing.T) {
		t.Parallel()
		addr := demo.GetDeviceAddress("Non Existent")
		assert.Empty(t, addr)
	})
}
