package shelly_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// componentMutationFixtures returns a single Gen2 device carrying switch and RGB
// components, served by the mock RPC server.
func componentMutationFixtures() *mock.Fixtures {
	return &mock.Fixtures{
		Version: "1",
		Config: mock.ConfigFixture{
			Devices: []mock.DeviceFixture{
				{
					Name:       "dev",
					Address:    "192.168.1.50",
					MAC:        "AA:BB:CC:DD:EE:10",
					Type:       "SNSW-001P16EU",
					Model:      "Shelly Plus 1PM",
					Generation: 2,
				},
			},
		},
		DeviceStates: map[string]mock.DeviceState{
			"dev": {
				"switch:0": map[string]any{"output": false},
				"rgb:0":    map[string]any{"output": false},
			},
		},
	}
}

// seededCache returns a FileCache (backed by an in-memory FS) holding a fresh
// component-status entry for device, so a later invalidation is observable.
func seededCache(t *testing.T, device string) *cache.FileCache {
	t.Helper()
	fc, err := cache.NewWithFs("/cache", afero.NewMemMapFs())
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	if err := fc.Set(device, cache.TypeComponents, map[string]any{"seed": true}, cache.TTLComponents); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
	entry, err := fc.Get(device, cache.TypeComponents)
	if err != nil || entry == nil {
		t.Fatalf("seed not present: entry=%v err=%v", entry, err)
	}
	return fc
}

func cacheHasComponents(t *testing.T, fc *cache.FileCache, device string) bool {
	t.Helper()
	entry, err := fc.Get(device, cache.TypeComponents)
	if err != nil {
		t.Fatalf("cache get: %v", err)
	}
	return entry != nil
}

// TestComponentMutations_InvalidateCache verifies that a successful component
// mutation drops the device's cached component status (so a follow-up status read
// reflects the change), and that a failed mutation leaves the cache untouched.
// It exercises the three mutation shapes the sweep touched: withComponentAction
// (SwitchOn), a WithDevice toggle (SwitchToggle), and a WithConnection set
// (RGBSet).
//
//nolint:paralleltest // installs a process-global default config manager
func TestComponentMutations_InvalidateCache(t *testing.T) {
	demo, err := mock.StartWithFixtures(componentMutationFixtures())
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	defer demo.Cleanup()
	config.SetDefaultManager(demo.ConfigMgr)
	// Re-arm the lazy default manager rather than leaving it nil, so later tests in
	// the package (which call config.Get via the global) don't hit a nil manager.
	t.Cleanup(config.ResetDefaultManagerForTesting)

	ctx := context.Background()

	t.Run("SwitchOn invalidates on success", func(t *testing.T) {
		fc := seededCache(t, "dev")
		svc := shelly.New(shelly.NewConfigResolver(), shelly.WithFileCache(fc))
		if err := svc.SwitchOn(ctx, "dev", 0); err != nil {
			t.Fatalf("SwitchOn: %v", err)
		}
		if cacheHasComponents(t, fc, "dev") {
			t.Error("component cache not invalidated after SwitchOn")
		}
	})

	t.Run("SwitchToggle invalidates on success", func(t *testing.T) {
		fc := seededCache(t, "dev")
		svc := shelly.New(shelly.NewConfigResolver(), shelly.WithFileCache(fc))
		if _, err := svc.SwitchToggle(ctx, "dev", 0); err != nil {
			t.Fatalf("SwitchToggle: %v", err)
		}
		if cacheHasComponents(t, fc, "dev") {
			t.Error("component cache not invalidated after SwitchToggle")
		}
	})

	t.Run("RGBSet invalidates on success", func(t *testing.T) {
		fc := seededCache(t, "dev")
		svc := shelly.New(shelly.NewConfigResolver(), shelly.WithFileCache(fc))
		if err := svc.RGBSet(ctx, "dev", 0, shelly.BuildRGBSetParams(255, 128, 64, 50, true)); err != nil {
			t.Fatalf("RGBSet: %v", err)
		}
		if cacheHasComponents(t, fc, "dev") {
			t.Error("component cache not invalidated after RGBSet")
		}
	})

	t.Run("failed mutation leaves cache intact", func(t *testing.T) {
		fc := seededCache(t, "ghost")
		svc := shelly.New(shelly.NewConfigResolver(), shelly.WithFileCache(fc))
		// Bound the unknown-device resolution, which otherwise retries generation
		// detection against the unreachable name for several seconds.
		fctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		if err := svc.SwitchOn(fctx, "ghost", 0); err == nil {
			t.Fatal("SwitchOn against unknown device unexpectedly succeeded")
		}
		if !cacheHasComponents(t, fc, "ghost") {
			t.Error("cache invalidated despite mutation failure")
		}
	})
}
