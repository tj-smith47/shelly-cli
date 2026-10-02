package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// TestSaveWithoutLock_RefusesLiveConfigOnOsFs proves the guard that protects the
// user's real config: even when code reaches a save against the live config path
// on a real OS filesystem (the exact path getDefaultManager() resolves to), the
// write is refused under `go test`. HOME/XDG are redirected to a temp dir so a
// regression cannot touch the real config — it refuses (or, on regression,
// writes) inside the temp dir, never ~/.config/shelly.
func TestSaveWithoutLock_RefusesLiveConfigOnOsFs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	live := filepath.Join(tmp, "shelly", "config.yaml")
	m := &Manager{
		path:   live,
		fs:     afero.NewOsFs(),
		config: &Config{Devices: nil},
		loaded: true,
	}

	if err := m.saveWithoutLock(); err == nil {
		t.Fatal("saveWithoutLock wrote the live config via OS FS under test; guard failed")
	}

	exists, err := afero.Exists(afero.NewOsFs(), live)
	if err != nil {
		t.Fatalf("stat %q: %v", live, err)
	}
	if exists {
		t.Fatalf("guard returned an error but still created %q", live)
	}
}

// TestSaveWithoutLock_AllowsTempPathOnOsFs confirms the guard is surgical: an OS
// filesystem write to a non-live path (a temp file) is still permitted, so
// legitimate OS-FS temp-dir tests keep working unchanged.
func TestSaveWithoutLock_AllowsTempPathOnOsFs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	other := filepath.Join(tmp, "elsewhere", "scratch.yaml")
	m := &Manager{
		path:   other,
		fs:     afero.NewOsFs(),
		config: &Config{Devices: map[string]model.Device{}},
		loaded: true,
	}

	if err := m.saveWithoutLock(); err != nil {
		t.Fatalf("guard wrongly blocked a non-live OS-FS write: %v", err)
	}
	exists, err := afero.Exists(afero.NewOsFs(), other)
	if err != nil {
		t.Fatalf("stat %q: %v", other, err)
	}
	if !exists {
		t.Fatalf("expected %q to be written", other)
	}
}

// TestSetSetting_RefusesLiveConfigOnOsFs is the regression for the second clobber:
// `config set` persists through viper, which bypasses the Manager. With an OS
// filesystem and viper pointed at the live config path, a `config set` under test
// must be refused and must not create the file. HOME/XDG are redirected to a temp
// dir so a regression cannot touch the user's real config.
func TestSetSetting_RefusesLiveConfigOnOsFs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	SetFs(afero.NewOsFs())
	t.Cleanup(func() { SetFs(nil) })

	live := filepath.Join(tmp, "shelly", "config.yaml")
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(live)

	if err := SetSetting("editor", "code --wait=true"); err == nil {
		t.Fatal("SetSetting wrote the live config via viper on OS FS under test; guard failed")
	}

	exists, err := afero.Exists(afero.NewOsFs(), live)
	if err != nil {
		t.Fatalf("stat %q: %v", live, err)
	}
	if exists {
		t.Fatalf("guard returned an error but viper still created %q", live)
	}
}

// TestLoad_IgnoresLiveConfigOnOsFs proves a test never reads the user's live
// registry: a live config holding a device loads as empty under `go test`,
// while the same file at another path loads normally.
func TestLoad_IgnoresLiveConfigOnOsFs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	body := []byte("devices:\n  kitchen:\n    name: kitchen\n    address: 10.0.0.9\n")
	live := filepath.Join(tmp, "shelly", "config.yaml")
	other := filepath.Join(tmp, "elsewhere", "config.yaml")
	osfs := afero.NewOsFs()
	for _, p := range []string{live, other} {
		if err := osfs.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := afero.WriteFile(osfs, p, body, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	for path, want := range map[string]int{live: 0, other: 1} {
		m := &Manager{path: path, fs: osfs}
		if err := m.Load(); err != nil {
			t.Fatalf("Load(%s) error = %v", path, err)
		}
		if got := len(m.config.Devices); got != want {
			t.Errorf("Load(%s) devices = %d, want %d", path, got, want)
		}
	}
}
