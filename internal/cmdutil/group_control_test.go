package cmdutil_test

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestResolveGroupDevices_Members(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{
			"living-room": {Devices: []string{"light-1", "light-2", "switch-1"}},
		},
	}
	f := cmdutil.NewFactory().SetConfigManager(config.NewTestManager(cfg))

	got, err := cmdutil.ResolveGroupDevices(f, "living-room")
	if err != nil {
		t.Fatalf("ResolveGroupDevices: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d devices, want 3", len(got))
	}
	for i, want := range []string{"light-1", "light-2", "switch-1"} {
		if got[i] != want {
			t.Errorf("device[%d] = %q, want %q", i, got[i], want)
		}
	}
}

func TestResolveGroupDevices_NotFound(t *testing.T) {
	t.Parallel()

	f := cmdutil.NewFactory().SetConfigManager(config.NewTestManager(&config.Config{}))

	_, err := cmdutil.ResolveGroupDevices(f, "missing")
	if err == nil {
		t.Fatal("expected error for missing group")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestResolveGroupDevices_Empty(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Groups: map[string]config.Group{"empty": {Devices: []string{}}},
	}
	f := cmdutil.NewFactory().SetConfigManager(config.NewTestManager(cfg))

	_, err := cmdutil.ResolveGroupDevices(f, "empty")
	if err == nil {
		t.Fatal("expected error for empty group")
	}
	if !strings.Contains(err.Error(), "no devices") {
		t.Errorf("error = %v, want 'no devices'", err)
	}
}
