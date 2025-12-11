package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestValidateAliasName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "myalias", false},
		{"valid with dash", "my-alias", false},
		{"valid with underscore", "my_alias", false},
		{"empty", "", true},
		{"whitespace", "my alias", true},
		{"tab", "my\talias", true},
		{"newline", "my\nalias", true},
		{"reserved help", "help", true},
		{"reserved version", "version", true},
		{"reserved alias", "alias", true},
		{"reserved switch", "switch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateAliasName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAliasName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestExpandAlias(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alias Alias
		args  []string
		want  string
	}{
		{
			name:  "no placeholders",
			alias: Alias{Command: "switch status"},
			args:  []string{},
			want:  "switch status",
		},
		{
			name:  "single placeholder",
			alias: Alias{Command: "switch toggle $1"},
			args:  []string{"living-room"},
			want:  "switch toggle living-room",
		},
		{
			name:  "multiple placeholders",
			alias: Alias{Command: "switch $1 $2"},
			args:  []string{"on", "kitchen"},
			want:  "switch on kitchen",
		},
		{
			name:  "all args placeholder",
			alias: Alias{Command: "batch $@"},
			args:  []string{"device1", "device2", "device3"},
			want:  "batch device1 device2 device3",
		},
		{
			name:  "mixed placeholders",
			alias: Alias{Command: "script $1 run $@"},
			args:  []string{"mydevice", "arg1", "arg2"},
			want:  "script mydevice run mydevice arg1 arg2",
		},
		{
			name:  "unused placeholder",
			alias: Alias{Command: "switch $1 $2 $3"},
			args:  []string{"on"},
			want:  "switch on $2 $3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ExpandAlias(tt.alias, tt.args)
			if got != tt.want {
				t.Errorf("ExpandAlias() = %q, want %q", got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // Test modifies global config state
func TestConfig_AliasOperations(t *testing.T) {
	// Reset config state
	cfgMu.Lock()
	cfg = nil
	cfgOnce = sync.Once{}
	cfgMu.Unlock()

	c := Get()

	// Test AddAlias
	err := c.AddAlias("ss", "switch status", false)
	if err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Test GetAlias
	alias := c.GetAlias("ss")
	if alias == nil {
		t.Fatal("GetAlias() returned nil")
	}
	if alias.Command != "switch status" {
		t.Errorf("expected command 'switch status', got %q", alias.Command)
	}

	// Test ListAliases
	aliases := c.ListAliases()
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}

	// Test IsAlias
	if !c.IsAlias("ss") {
		t.Error("IsAlias() returned false for existing alias")
	}
	if c.IsAlias("nonexistent") {
		t.Error("IsAlias() returned true for nonexistent alias")
	}

	// Test RemoveAlias
	c.RemoveAlias("ss")
	if c.GetAlias("ss") != nil {
		t.Error("alias still exists after RemoveAlias()")
	}
}

//nolint:paralleltest // Test modifies global config state
func TestConfig_ImportExportAliases(t *testing.T) {
	// Reset config state
	cfgMu.Lock()
	cfg = nil
	cfgOnce = sync.Once{}
	cfgMu.Unlock()

	c := Get()

	// Add some aliases.
	if err := c.AddAlias("ss", "switch status", false); err != nil {
		t.Fatalf("AddAlias(ss) error: %v", err)
	}
	if err := c.AddAlias("st", "switch toggle $1", false); err != nil {
		t.Fatalf("AddAlias(st) error: %v", err)
	}

	// Create temp dir.
	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Export
	exportPath := filepath.Join(tmpDir, "aliases.yaml")
	err = c.ExportAliases(exportPath)
	if err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}

	// Clear aliases
	c.RemoveAlias("ss")
	c.RemoveAlias("st")
	if len(c.ListAliases()) != 0 {
		t.Fatal("aliases not cleared")
	}

	// Import
	imported, skipped, err := c.ImportAliases(exportPath, false)
	if err != nil {
		t.Fatalf("ImportAliases() error: %v", err)
	}
	if imported != 2 {
		t.Errorf("expected 2 imported, got %d", imported)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}

	// Verify imported
	if c.GetAlias("ss") == nil {
		t.Error("alias 'ss' not imported")
	}
	if c.GetAlias("st") == nil {
		t.Error("alias 'st' not imported")
	}

	// Test merge mode (should skip existing)
	imported, skipped, err = c.ImportAliases(exportPath, true)
	if err != nil {
		t.Fatalf("ImportAliases(merge) error: %v", err)
	}
	if imported != 0 {
		t.Errorf("expected 0 imported in merge mode, got %d", imported)
	}
	if skipped != 2 {
		t.Errorf("expected 2 skipped in merge mode, got %d", skipped)
	}
}
