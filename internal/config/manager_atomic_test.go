package config

import (
	"errors"
	"testing"

	"github.com/spf13/afero"
)

// readErrFs fails Open for a single path with a non-not-exist error, simulating
// EIO/EACCES on the config file while leaving every other path working.
type readErrFs struct {
	afero.Fs
	failPath string
	err      error
}

func (f readErrFs) Open(name string) (afero.File, error) {
	if name == f.failPath {
		return nil, f.err
	}
	return f.Fs.Open(name)
}

// renameErrFs fails every Rename, simulating a crash/ENOSPC at the atomic-swap step.
type renameErrFs struct {
	afero.Fs
	err error
}

func (f renameErrFs) Rename(oldname, newname string) error {
	return f.err
}

//nolint:paralleltest // modifies global state via SetFs
func TestManager_Load_ReadErrorAborts(t *testing.T) {
	base := afero.NewMemMapFs()
	// A real config exists; the read itself fails (not a missing file).
	if err := afero.WriteFile(base, testConfigPath, []byte("devices:\n  d1:\n    name: d1\n"), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	SetFs(readErrFs{Fs: base, failPath: testConfigPath, err: errors.New("simulated I/O error")})
	t.Cleanup(func() { SetFs(nil) })

	m := NewManager(testConfigPath)
	if err := m.Load(); err == nil {
		t.Fatal("Load() must return an error on a non-not-exist read failure, " +
			"never silently fall back to an empty config (which a later Save would persist over the real registry)")
	}
}

//nolint:paralleltest // modifies global state via SetFs
func TestManager_Load_MissingFileIsClean(t *testing.T) {
	SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { SetFs(nil) })

	m := NewManager("/does/not/exist/config.yaml")
	if err := m.Load(); err != nil {
		t.Fatalf("a missing config file is a normal first run, want nil error, got: %v", err)
	}
}

//nolint:paralleltest // modifies global state via SetFs
func TestManager_atomicWrite_RenameFailurePreservesOriginal(t *testing.T) {
	base := afero.NewMemMapFs()
	if err := afero.WriteFile(base, testConfigPath, []byte("original"), 0o600); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	fs := renameErrFs{Fs: base, err: errors.New("simulated rename failure")}

	m := NewManager(testConfigPath)
	if err := m.atomicWrite(fs, []byte("new content")); err == nil {
		t.Fatal("atomicWrite must return an error when the rename fails")
	}

	// The live config must be untouched — the failure happened before the swap.
	got, err := afero.ReadFile(base, testConfigPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(got) != "original" {
		t.Errorf("original config clobbered by a failed write, got %q", got)
	}

	// The temp file must not be left behind.
	tmpExists, err := afero.Exists(base, testConfigPath+".tmp")
	if err != nil {
		t.Fatalf("stat temp: %v", err)
	}
	if tmpExists {
		t.Error("temp file leaked after a failed rename")
	}
}

//nolint:paralleltest // modifies global state via SetFs
func TestManager_Save_LeavesNoTempFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	SetFs(fs)
	t.Cleanup(func() { SetFs(nil) })

	m := NewManager(testConfigPath)
	if err := m.Load(); err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := m.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	tmpExists, err := afero.Exists(fs, testConfigPath+".tmp")
	if err != nil {
		t.Fatalf("stat temp: %v", err)
	}
	if tmpExists {
		t.Error("Save left a .tmp file behind; the atomic swap should remove it")
	}
	cfgExists, err := afero.Exists(fs, testConfigPath)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if !cfgExists {
		t.Error("Save did not write the config file")
	}
}
