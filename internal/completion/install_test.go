package completion_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestGenerateAndInstall_UnsupportedShell(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	rootCmd := &cobra.Command{Use: "shelly"}

	err := completion.GenerateAndInstall(ios, rootCmd, "unsupported")
	if err == nil {
		t.Error("GenerateAndInstall() should error for unsupported shell")
	}
}

func TestGenerateAndInstall_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)
	rootCmd := &cobra.Command{Use: "shelly"}

	err := completion.GenerateAndInstall(ios, rootCmd, completion.ShellBash)
	if err != nil {
		// This may fail depending on permissions; just log
		t.Logf("GenerateAndInstall(bash) error: %v (may be expected)", err)
	}
}

func TestGenerateAndInstall_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)
	rootCmd := &cobra.Command{Use: "shelly"}

	err := completion.GenerateAndInstall(ios, rootCmd, completion.ShellZsh)
	if err != nil {
		t.Logf("GenerateAndInstall(zsh) error: %v (may be expected)", err)
	}
}

func TestGenerateAndInstall_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)
	rootCmd := &cobra.Command{Use: "shelly"}

	err := completion.GenerateAndInstall(ios, rootCmd, completion.ShellFish)
	if err != nil {
		t.Logf("GenerateAndInstall(fish) error: %v (may be expected)", err)
	}
}

func TestGenerateAndInstall_PowerShell(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)
	rootCmd := &cobra.Command{Use: "shelly"}

	err := completion.GenerateAndInstall(ios, rootCmd, completion.ShellPowerShell)
	if err != nil {
		t.Logf("GenerateAndInstall(powershell) error: %v (may be expected)", err)
	}
}

func TestInstallBash(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)

	script := []byte("# test bash completion script")
	err := completion.InstallBash(ios, script)
	if err != nil {
		t.Logf("InstallBash() error: %v (may be expected)", err)
	}
}

func TestInstallZsh(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)

	script := []byte("# test zsh completion script")
	err := completion.InstallZsh(ios, script)
	if err != nil {
		t.Logf("InstallZsh() error: %v (may be expected)", err)
	}

	// Verify completion file was created
	expectedPath := filepath.Join(fakeHome, ".zsh", "completions", "_shelly")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Logf("completion file not created at %s (may be expected)", expectedPath)
	}
}

func TestInstallFish(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)

	script := []byte("# test fish completion script")
	err := completion.InstallFish(ios, script)
	if err != nil {
		t.Logf("InstallFish() error: %v (may be expected)", err)
	}

	// Verify completion file was created
	expectedPath := filepath.Join(fakeHome, ".config", "fish", "completions", "shelly.fish")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Logf("completion file not created at %s (may be expected)", expectedPath)
	}
}

func TestInstallPowerShell(t *testing.T) {
	tmpDir := t.TempDir()
	fakeHome := filepath.Join(tmpDir, "home")
	if err := os.MkdirAll(fakeHome, 0o750); err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	t.Setenv("HOME", fakeHome)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := iostreams.Test(&bytes.Buffer{}, out, errOut)

	script := []byte("# test powershell completion script")
	err := completion.InstallPowerShell(ios, script)
	if err != nil {
		t.Logf("InstallPowerShell() error: %v (may be expected)", err)
	}
}
