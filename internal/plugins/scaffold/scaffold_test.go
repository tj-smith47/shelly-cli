package scaffold_test

import (
	"bytes"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins/scaffold"
)

const testScaffoldDir = "/test/scaffold"

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestBash(t *testing.T) {
	tests := []struct {
		name    string
		extName string
		cmdName string
	}{
		{
			name:    "simple extension",
			extName: "shelly-myext",
			cmdName: "myext",
		},
		{
			name:    "hyphenated name",
			extName: "shelly-my-extension",
			cmdName: "my-extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			config.SetFs(fs)
			t.Cleanup(func() { config.SetFs(nil) })

			// Create test directory
			if err := fs.MkdirAll(testScaffoldDir, 0o755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			err := scaffold.Bash(testScaffoldDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Bash() error = %v", err)
			}

			// Check script file exists
			scriptPath := testScaffoldDir + "/" + tt.extName
			info, err := fs.Stat(scriptPath)
			if err != nil {
				t.Errorf("failed to stat script: %v", err)
			}

			// Verify script is executable
			if info != nil && info.Mode()&0o100 == 0 {
				t.Error("script should be executable")
			}

			// Read and verify content
			content, err := afero.ReadFile(fs, scriptPath)
			if err != nil {
				t.Fatalf("failed to read script: %v", err)
			}

			// Check shebang
			if len(content) < 20 || string(content[:2]) != "#!" {
				t.Error("script should start with shebang")
			}

			// Check README exists
			readmePath := testScaffoldDir + "/README.md"
			if _, err := fs.Stat(readmePath); err != nil {
				t.Error("README.md not created")
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestGo(t *testing.T) {
	tests := []struct {
		name    string
		extName string
		cmdName string
	}{
		{
			name:    "simple extension",
			extName: "shelly-myext",
			cmdName: "myext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			config.SetFs(fs)
			t.Cleanup(func() { config.SetFs(nil) })

			// Create test directory
			if err := fs.MkdirAll(testScaffoldDir, 0o755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			err := scaffold.Go(testScaffoldDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Go() error = %v", err)
			}

			// Check main.go exists
			mainPath := testScaffoldDir + "/main.go"
			if _, err := fs.Stat(mainPath); err != nil {
				t.Error("main.go not created")
			}

			// Read and verify main.go content
			content, err := afero.ReadFile(fs, mainPath)
			if err != nil {
				t.Fatalf("failed to read main.go: %v", err)
			}

			if len(content) == 0 {
				t.Error("main.go is empty")
			}

			// Check package declaration
			if len(content) < 12 || string(content[:12]) != "package main" {
				t.Error("main.go should have package main")
			}

			// Check go.mod exists
			modPath := testScaffoldDir + "/go.mod"
			if _, err := fs.Stat(modPath); err != nil {
				t.Error("go.mod not created")
			}

			// Check Makefile exists
			makefilePath := testScaffoldDir + "/Makefile"
			if _, err := fs.Stat(makefilePath); err != nil {
				t.Error("Makefile not created")
			}

			// Check README exists
			readmePath := testScaffoldDir + "/README.md"
			if _, err := fs.Stat(readmePath); err != nil {
				t.Error("README.md not created")
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestPython(t *testing.T) {
	tests := []struct {
		name    string
		extName string
		cmdName string
	}{
		{
			name:    "simple extension",
			extName: "shelly-myext",
			cmdName: "myext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			config.SetFs(fs)
			t.Cleanup(func() { config.SetFs(nil) })

			// Create test directory
			if err := fs.MkdirAll(testScaffoldDir, 0o755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			err := scaffold.Python(testScaffoldDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Python() error = %v", err)
			}

			// Check script file exists
			scriptPath := testScaffoldDir + "/" + tt.extName
			info, err := fs.Stat(scriptPath)
			if err != nil {
				t.Errorf("failed to stat script: %v", err)
			}

			// Verify script is executable
			if info != nil && info.Mode()&0o100 == 0 {
				t.Error("script should be executable")
			}

			// Read and verify content
			content, err := afero.ReadFile(fs, scriptPath)
			if err != nil {
				t.Fatalf("failed to read script: %v", err)
			}

			// Check shebang for Python
			if len(content) < 20 || string(content[:2]) != "#!" {
				t.Error("script should start with shebang")
			}

			// Check README exists
			readmePath := testScaffoldDir + "/README.md"
			if _, err := fs.Stat(readmePath); err != nil {
				t.Error("README.md not created")
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestReadme(t *testing.T) {
	tests := []struct {
		name    string
		extName string
		cmdName string
	}{
		{
			name:    "simple extension",
			extName: "shelly-myext",
			cmdName: "myext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			config.SetFs(fs)
			t.Cleanup(func() { config.SetFs(nil) })

			// Create test directory
			if err := fs.MkdirAll(testScaffoldDir, 0o755); err != nil {
				t.Fatalf("failed to create test dir: %v", err)
			}

			err := scaffold.Readme(testScaffoldDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Readme() error = %v", err)
			}

			readmePath := testScaffoldDir + "/README.md"
			content, err := afero.ReadFile(fs, readmePath)
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}

			if len(content) == 0 {
				t.Error("README.md is empty")
			}

			// Check it contains the extension name
			if !containsBytes(content, []byte(tt.extName)) {
				t.Error("README.md should contain extension name")
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already normalized",
			input: "myext",
			want:  "myext",
		},
		{
			name:  "with prefix",
			input: "shelly-myext",
			want:  "myext",
		},
		{
			name:  "uppercase",
			input: "MyExt",
			want:  "myext",
		},
		{
			name:  "uppercase with prefix",
			input: "shelly-MyExt",
			want:  "myext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := scaffold.NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFullName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "myext",
			want:  "shelly-myext",
		},
		{
			name:  "already has prefix",
			input: "shelly-myext",
			want:  "shelly-myext",
		},
		{
			name:  "uppercase",
			input: "MyExt",
			want:  "shelly-myext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := scaffold.FullName(tt.input)
			if got != tt.want {
				t.Errorf("FullName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestBash_ReadOnlyDirectory(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	config.SetFs(roFs)
	t.Cleanup(func() { config.SetFs(nil) })

	// Try to create in a read-only filesystem
	err := scaffold.Bash("/test/scaffold", "shelly-test", "test")
	if err == nil {
		t.Error("Bash() should error for read-only filesystem")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestGo_ReadOnlyDirectory(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	config.SetFs(roFs)
	t.Cleanup(func() { config.SetFs(nil) })

	err := scaffold.Go("/test/scaffold", "shelly-test", "test")
	if err == nil {
		t.Error("Go() should error for read-only filesystem")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestPython_ReadOnlyDirectory(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	config.SetFs(roFs)
	t.Cleanup(func() { config.SetFs(nil) })

	err := scaffold.Python("/test/scaffold", "shelly-test", "test")
	if err == nil {
		t.Error("Python() should error for read-only filesystem")
	}
}

//nolint:paralleltest // Test modifies global state via config.SetFs
func TestReadme_ReadOnlyDirectory(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	roFs := afero.NewReadOnlyFs(baseFs)
	config.SetFs(roFs)
	t.Cleanup(func() { config.SetFs(nil) })

	err := scaffold.Readme("/test/scaffold", "shelly-test", "test")
	if err == nil {
		t.Error("Readme() should error for read-only filesystem")
	}
}

// containsBytes checks if content contains substr.
func containsBytes(content, substr []byte) bool {
	return bytes.Contains(content, substr)
}
