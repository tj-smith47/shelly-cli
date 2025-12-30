package config

import (
	"context"
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

//nolint:paralleltest // Test modifies global state via t.Setenv and ResetDefaultManagerForTesting.
func TestExpandAliasArgs(t *testing.T) {
	// Note: This test modifies global state, cannot be parallel
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	ResetDefaultManagerForTesting()

	// Add a regular alias
	if err := AddAlias("ss", "switch status", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Add a shell alias
	if err := AddAlias("lsalias", "ls -la", true); err != nil {
		t.Fatalf("AddAlias(shell) error: %v", err)
	}

	t.Run("empty args", func(t *testing.T) {
		expanded, isShell := ExpandAliasArgs([]string{})
		if isShell {
			t.Error("isShell should be false for empty args")
		}
		if len(expanded) != 0 {
			t.Errorf("expected empty expanded, got %v", expanded)
		}
	})

	t.Run("non-alias arg", func(t *testing.T) {
		expanded, isShell := ExpandAliasArgs([]string{"unknown", "arg1"})
		if isShell {
			t.Error("isShell should be false for non-alias")
		}
		if len(expanded) != 2 || expanded[0] != "unknown" {
			t.Errorf("expected [unknown arg1], got %v", expanded)
		}
	})

	t.Run("regular alias expansion", func(t *testing.T) {
		expanded, isShell := ExpandAliasArgs([]string{"ss", "kitchen"})
		if isShell {
			t.Error("isShell should be false for regular alias")
		}
		// "switch status" + "kitchen" -> ["switch", "status", "kitchen"]
		if len(expanded) != 3 {
			t.Errorf("expected 3 args, got %v", expanded)
		}
		if expanded[0] != "switch" || expanded[1] != "status" {
			t.Errorf("expected [switch status ...], got %v", expanded)
		}
	})

	t.Run("shell alias expansion", func(t *testing.T) {
		expanded, isShell := ExpandAliasArgs([]string{"lsalias"})
		if !isShell {
			t.Error("isShell should be true for shell alias")
		}
		if len(expanded) != 1 || expanded[0] != "ls -la" {
			t.Errorf("expected [ls -la], got %v", expanded)
		}
	})
}

func TestReservedCommands(t *testing.T) {
	t.Parallel()

	// Verify critical commands are reserved
	criticalCommands := []string{"help", "version", "config", "alias", "device", "switch"}
	for _, cmd := range criticalCommands {
		if !ReservedCommands[cmd] {
			t.Errorf("expected %q to be reserved", cmd)
		}
	}

	// Verify non-existent commands are not reserved
	if ReservedCommands["nonexistent"] {
		t.Error("expected 'nonexistent' to not be reserved")
	}
}

func TestExecuteShellAlias(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("empty args returns 0", func(t *testing.T) {
		t.Parallel()
		code := ExecuteShellAlias(ctx, []string{})
		if code != 0 {
			t.Errorf("ExecuteShellAlias([]) = %d, want 0", code)
		}
	})

	t.Run("successful command returns 0", func(t *testing.T) {
		t.Parallel()
		code := ExecuteShellAlias(ctx, []string{"true"})
		if code != 0 {
			t.Errorf("ExecuteShellAlias([true]) = %d, want 0", code)
		}
	})

	t.Run("failing command returns exit code", func(t *testing.T) {
		t.Parallel()
		code := ExecuteShellAlias(ctx, []string{"exit 42"})
		if code != 42 {
			t.Errorf("ExecuteShellAlias([exit 42]) = %d, want 42", code)
		}
	})

	t.Run("command not found returns 1", func(t *testing.T) {
		t.Parallel()
		code := ExecuteShellAlias(ctx, []string{"nonexistent_command_12345"})
		// Should return non-zero (typically 127 or 1)
		if code == 0 {
			t.Error("ExecuteShellAlias([nonexistent]) = 0, want non-zero")
		}
	})

	t.Run("canceled context returns 1", func(t *testing.T) {
		t.Parallel()
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		code := ExecuteShellAlias(canceledCtx, []string{"sleep 10"})
		// Should return 1 for context error (not an exit error)
		if code == 0 {
			t.Error("ExecuteShellAlias with canceled context = 0, want non-zero")
		}
	})
}

func TestExecuteShellAlias_NoShell(t *testing.T) {
	// Test the fallback to /bin/sh when SHELL is not set
	// Note: This test modifies environment, cannot be parallel with other shell tests
	t.Setenv("SHELL", "") // Clear SHELL to trigger fallback

	ctx := context.Background()
	code := ExecuteShellAlias(ctx, []string{"true"})
	if code != 0 {
		t.Errorf("ExecuteShellAlias with no SHELL = %d, want 0", code)
	}
}

//nolint:gocyclo // Integration test covering all package-level alias functions.
func TestPackageLevelAliasFunctions(t *testing.T) {
	// Note: This test modifies global state, cannot be parallel
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	ResetDefaultManagerForTesting()

	// Test AddAlias
	if err := AddAlias("test1", "device info", false); err != nil {
		t.Fatalf("AddAlias() error: %v", err)
	}

	// Test GetAlias
	alias, ok := GetAlias("test1")
	if !ok {
		t.Fatal("GetAlias() returned false for existing alias")
	}
	if alias.Command != "device info" {
		t.Errorf("GetAlias() command = %q, want %q", alias.Command, "device info")
	}

	// Test IsAlias
	if !IsAlias("test1") {
		t.Error("IsAlias() returned false for existing alias")
	}
	if IsAlias("nonexistent") {
		t.Error("IsAlias() returned true for nonexistent alias")
	}

	// Add another alias for list tests
	if err := AddAlias("aaa", "switch on", false); err != nil {
		t.Fatalf("AddAlias(aaa) error: %v", err)
	}

	// Test ListAliases
	aliases := ListAliases()
	if len(aliases) != 2 {
		t.Errorf("ListAliases() returned %d aliases, want 2", len(aliases))
	}
	if _, ok := aliases["test1"]; !ok {
		t.Error("ListAliases() missing test1")
	}

	// Test ListAliasesSorted
	sorted := ListAliasesSorted()
	if len(sorted) != 2 {
		t.Errorf("ListAliasesSorted() returned %d aliases, want 2", len(sorted))
	}
	// Should be sorted: aaa comes before test1
	if sorted[0].Name != "aaa" {
		t.Errorf("ListAliasesSorted() first = %q, want %q", sorted[0].Name, "aaa")
	}

	// Test ExportAliases (to stdout)
	output, err := ExportAliases("")
	if err != nil {
		t.Fatalf("ExportAliases() error: %v", err)
	}
	if !strings.Contains(output, "test1:") {
		t.Error("ExportAliases() output missing test1")
	}

	// Test ExportAliases (to file)
	exportPath := filepath.Join(tmpDir, "export.yaml")
	_, err = ExportAliases(exportPath)
	if err != nil {
		t.Fatalf("ExportAliases(file) error: %v", err)
	}

	// Test RemoveAlias
	if err := RemoveAlias("test1"); err != nil {
		t.Fatalf("RemoveAlias() error: %v", err)
	}
	if IsAlias("test1") {
		t.Error("alias test1 still exists after RemoveAlias()")
	}

	// Test ImportAliases with merge=false (overwrites all)
	imported, skipped, err := ImportAliases(exportPath, false)
	if err != nil {
		t.Fatalf("ImportAliases() error: %v", err)
	}
	// merge=false: both test1 and aaa are imported (overwrite mode)
	if imported != 2 {
		t.Errorf("ImportAliases() imported = %d, want 2", imported)
	}
	if skipped != 0 {
		t.Errorf("ImportAliases() skipped = %d, want 0", skipped)
	}

	// Test ImportAliases with merge=true (skips existing)
	imported, skipped, err = ImportAliases(exportPath, true)
	if err != nil {
		t.Fatalf("ImportAliases(merge) error: %v", err)
	}
	// merge=true: both already exist, so both skipped
	if imported != 0 {
		t.Errorf("ImportAliases(merge) imported = %d, want 0", imported)
	}
	if skipped != 2 {
		t.Errorf("ImportAliases(merge) skipped = %d, want 2", skipped)
	}
}
