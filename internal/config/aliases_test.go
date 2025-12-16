package config

import (
	"path/filepath"
	"strings"
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
			name:  "no placeholders no args",
			alias: Alias{Command: "switch status"},
			args:  []string{},
			want:  "switch status",
		},
		{
			name:  "no placeholders with args auto-appends",
			alias: Alias{Command: "device info"},
			args:  []string{"kitchen"},
			want:  "device info kitchen",
		},
		{
			name:  "no placeholders with multiple args auto-appends",
			alias: Alias{Command: "batch on"},
			args:  []string{"light1", "light2", "light3"},
			want:  "batch on light1 light2 light3",
		},
		{
			name:  "single placeholder consumes arg",
			alias: Alias{Command: "switch toggle $1"},
			args:  []string{"living-room"},
			want:  "switch toggle living-room",
		},
		{
			name:  "placeholder with extra args auto-appends remaining",
			alias: Alias{Command: "config export $1"},
			args:  []string{"kitchen", "-o", "yaml"},
			want:  "config export kitchen -o yaml",
		},
		{
			name:  "multiple placeholders",
			alias: Alias{Command: "switch $1 $2"},
			args:  []string{"on", "kitchen"},
			want:  "switch on kitchen",
		},
		{
			name:  "all args placeholder consumes all",
			alias: Alias{Command: "batch $@"},
			args:  []string{"device1", "device2", "device3"},
			want:  "batch device1 device2 device3",
		},
		{
			name:  "explicit $@ prevents auto-append",
			alias: Alias{Command: "echo $@"},
			args:  []string{"a", "b", "c"},
			want:  "echo a b c",
		},
		{
			name:  "unused placeholder kept as-is",
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

func TestManager_AliasOperations(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Test AddAlias
	err := m.AddAlias("ss", "switch status", false)
	if err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Test GetAlias
	alias, ok := m.GetAlias("ss")
	if !ok {
		t.Fatal("GetAlias() returned false")
	}
	if alias.Command != "switch status" {
		t.Errorf("expected command 'switch status', got %q", alias.Command)
	}

	// Test ListAliases
	aliases := m.ListAliases()
	if len(aliases) != 1 {
		t.Errorf("expected 1 alias, got %d", len(aliases))
	}

	// Test IsAlias
	if !m.IsAlias("ss") {
		t.Error("IsAlias() returned false for existing alias")
	}
	if m.IsAlias("nonexistent") {
		t.Error("IsAlias() returned true for nonexistent alias")
	}

	// Test RemoveAlias
	if err := m.RemoveAlias("ss"); err != nil {
		t.Fatalf("RemoveAlias() error: %v", err)
	}
	if _, ok := m.GetAlias("ss"); ok {
		t.Error("alias still exists after RemoveAlias()")
	}
}

func TestManager_ExportAliases(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if err := m.AddAlias("ss", "switch status", false); err != nil {
		t.Fatalf("AddAlias(ss) error: %v", err)
	}
	if err := m.AddAlias("st", "switch toggle $1", false); err != nil {
		t.Fatalf("AddAlias(st) error: %v", err)
	}

	exportPath := filepath.Join(tmpDir, "aliases.yaml")
	_, err := m.ExportAliases(exportPath)
	if err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}
}

func TestManager_ImportAliases(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Add and export aliases first
	if err := m.AddAlias("ss", "switch status", false); err != nil {
		t.Fatalf("AddAlias(ss) error: %v", err)
	}
	if err := m.AddAlias("st", "switch toggle $1", false); err != nil {
		t.Fatalf("AddAlias(st) error: %v", err)
	}

	exportPath := filepath.Join(tmpDir, "aliases.yaml")
	if _, err := m.ExportAliases(exportPath); err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}

	// Clear and import
	if err := m.RemoveAlias("ss"); err != nil {
		t.Fatalf("RemoveAlias(ss) error: %v", err)
	}
	if err := m.RemoveAlias("st"); err != nil {
		t.Fatalf("RemoveAlias(st) error: %v", err)
	}

	imported, skipped, err := m.ImportAliases(exportPath, false)
	if err != nil {
		t.Fatalf("ImportAliases() error: %v", err)
	}
	if imported != 2 {
		t.Errorf("expected 2 imported, got %d", imported)
	}
	if skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", skipped)
	}
	if _, ok := m.GetAlias("ss"); !ok {
		t.Error("alias 'ss' not imported")
	}
	if _, ok := m.GetAlias("st"); !ok {
		t.Error("alias 'st' not imported")
	}
}

func TestManager_ImportAliases_MergeMode(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Add and export aliases
	if err := m.AddAlias("ss", "switch status", false); err != nil {
		t.Fatalf("AddAlias(ss) error: %v", err)
	}
	if err := m.AddAlias("st", "switch toggle $1", false); err != nil {
		t.Fatalf("AddAlias(st) error: %v", err)
	}

	exportPath := filepath.Join(tmpDir, "aliases.yaml")
	if _, err := m.ExportAliases(exportPath); err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}

	// Import with merge (should skip existing)
	imported, skipped, err := m.ImportAliases(exportPath, true)
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

func TestExportAliases_Stdout(t *testing.T) {
	t.Parallel()

	// Create temp dir for isolated config
	tmpDir := t.TempDir()
	m := NewManager(filepath.Join(tmpDir, "config.yaml"))
	if err := m.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if err := m.AddAlias("test", "echo hello", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Export to stdout (empty filename)
	output, err := m.ExportAliases("")
	if err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}

	if output == "" {
		t.Error("expected non-empty output for stdout export")
	}

	// Verify YAML content
	if !strings.Contains(output, "test:") {
		t.Errorf("output missing alias name, got: %s", output)
	}
}
