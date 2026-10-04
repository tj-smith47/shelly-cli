package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTestPlugin creates an executable plugin script in dir.
func writeTestPlugin(t *testing.T, dir, name string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		t.Fatalf("create plugin: %v", err)
	}
	if _, err := f.WriteString("#!/bin/sh\necho ok\n"); err != nil {
		t.Fatalf("write plugin: %v", err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("sync plugin: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close plugin: %v", err)
	}
	// Linux can report "text file busy" when a just-written script is executed
	// before the kernel releases the write handle.
	time.Sleep(10 * time.Millisecond)
}

// The plugin guide promises that `shelly myext` runs the shelly-myext plugin.
func TestIsPluginInvocation(t *testing.T) {
	dir := t.TempDir()
	writeTestPlugin(t, dir, "shelly-zzdispatch")
	// A plugin named like a built-in command, and one named like a built-in alias.
	writeTestPlugin(t, dir, "shelly-status")
	writeTestPlugin(t, dir, "shelly-rpc")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"installed plugin", []string{"zzdispatch", "hello", "--flag"}, true},
		{"built-in command wins", []string{"status", "kitchen"}, false},
		{"built-in alias wins", []string{"rpc", "kitchen", "Sys.GetStatus"}, false},
		{"unknown name with no plugin", []string{"zznotinstalled"}, false},
		{"flag first", []string{"--help"}, false},
		{"no arguments", nil, false},
	}
	for _, tt := range tests {
		rootCmdMu.Lock()
		got := isPluginInvocation(rootCmd, tt.args)
		rootCmdMu.Unlock()
		if got != tt.want {
			t.Errorf("%s: isPluginInvocation(%v) = %v, want %v", tt.name, tt.args, got, tt.want)
		}
	}
}

// The example notify plugin is documented as `shelly notify ...`, so no
// built-in command or alias may claim that name.
func TestExamplePluginNamesAreNotBuiltins(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(filepath.Join(repoRoot, "examples", "plugins"))
	if err != nil {
		t.Fatalf("read examples/plugins: %v", err)
	}
	for _, e := range entries {
		name, ok := cutPluginPrefix(e.Name())
		if !ok {
			continue
		}
		rootCmdMu.Lock()
		target, _, findErr := rootCmd.Find([]string{name})
		rootCmdMu.Unlock()
		if findErr == nil && target != rootCmd {
			t.Errorf("example plugin %s is hidden by the built-in command %q; `shelly %s` would never reach it",
				e.Name(), target.CommandPath(), name)
		}
	}
}

func cutPluginPrefix(dirName string) (string, bool) {
	const prefix = "shelly-"
	if len(dirName) <= len(prefix) || dirName[:len(prefix)] != prefix {
		return "", false
	}
	return dirName[len(prefix):], true
}
