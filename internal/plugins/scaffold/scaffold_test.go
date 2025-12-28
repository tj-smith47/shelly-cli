package scaffold_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/plugins/scaffold"
)

func TestBash(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			tmpDir := t.TempDir()

			err := scaffold.Bash(tmpDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Bash() error = %v", err)
			}

			// Check script file exists
			scriptPath := filepath.Join(tmpDir, tt.extName)
			info, err := os.Stat(scriptPath)
			if os.IsNotExist(err) {
				t.Errorf("script file not created: %s", scriptPath)
			} else if err != nil {
				t.Errorf("failed to stat script: %v", err)
			}

			// Verify script is executable
			if info != nil && info.Mode()&0o100 == 0 {
				t.Error("script should be executable")
			}

			// Read and verify content
			content, err := os.ReadFile(scriptPath) //nolint:gosec // G304: scriptPath from t.TempDir()
			if err != nil {
				t.Fatalf("failed to read script: %v", err)
			}

			// Check shebang
			if len(content) < 20 || string(content[:2]) != "#!" {
				t.Error("script should start with shebang")
			}

			// Check README exists
			readmePath := filepath.Join(tmpDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) {
				t.Error("README.md not created")
			}
		})
	}
}

func TestGo(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			tmpDir := t.TempDir()

			err := scaffold.Go(tmpDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Go() error = %v", err)
			}

			// Check main.go exists
			mainPath := filepath.Join(tmpDir, "main.go")
			if _, err := os.Stat(mainPath); os.IsNotExist(err) {
				t.Error("main.go not created")
			}

			// Read and verify main.go content
			content, err := os.ReadFile(mainPath) //nolint:gosec // G304: mainPath from t.TempDir()
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
			modPath := filepath.Join(tmpDir, "go.mod")
			if _, err := os.Stat(modPath); os.IsNotExist(err) {
				t.Error("go.mod not created")
			}

			// Check Makefile exists
			makefilePath := filepath.Join(tmpDir, "Makefile")
			if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
				t.Error("Makefile not created")
			}

			// Check README exists
			readmePath := filepath.Join(tmpDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) {
				t.Error("README.md not created")
			}
		})
	}
}

func TestPython(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			tmpDir := t.TempDir()

			err := scaffold.Python(tmpDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Python() error = %v", err)
			}

			// Check script file exists
			scriptPath := filepath.Join(tmpDir, tt.extName)
			info, err := os.Stat(scriptPath)
			if os.IsNotExist(err) {
				t.Errorf("script file not created: %s", scriptPath)
			} else if err != nil {
				t.Errorf("failed to stat script: %v", err)
			}

			// Verify script is executable
			if info != nil && info.Mode()&0o100 == 0 {
				t.Error("script should be executable")
			}

			// Read and verify content
			content, err := os.ReadFile(scriptPath) //nolint:gosec // G304: scriptPath from t.TempDir()
			if err != nil {
				t.Fatalf("failed to read script: %v", err)
			}

			// Check shebang for Python
			if len(content) < 20 || string(content[:2]) != "#!" {
				t.Error("script should start with shebang")
			}

			// Check README exists
			readmePath := filepath.Join(tmpDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) {
				t.Error("README.md not created")
			}
		})
	}
}

func TestReadme(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			tmpDir := t.TempDir()

			err := scaffold.Readme(tmpDir, tt.extName, tt.cmdName)
			if err != nil {
				t.Fatalf("Readme() error = %v", err)
			}

			readmePath := filepath.Join(tmpDir, "README.md")
			content, err := os.ReadFile(readmePath) //nolint:gosec // G304: readmePath from t.TempDir()
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

func TestBash_InvalidDirectory(t *testing.T) {
	t.Parallel()

	// Try to create in a directory that doesn't exist
	err := scaffold.Bash("/nonexistent/path/that/should/fail", "shelly-test", "test")
	if err == nil {
		t.Error("Bash() should error for invalid directory")
	}
}

func TestGo_InvalidDirectory(t *testing.T) {
	t.Parallel()

	err := scaffold.Go("/nonexistent/path/that/should/fail", "shelly-test", "test")
	if err == nil {
		t.Error("Go() should error for invalid directory")
	}
}

func TestPython_InvalidDirectory(t *testing.T) {
	t.Parallel()

	err := scaffold.Python("/nonexistent/path/that/should/fail", "shelly-test", "test")
	if err == nil {
		t.Error("Python() should error for invalid directory")
	}
}

func TestReadme_InvalidDirectory(t *testing.T) {
	t.Parallel()

	err := scaffold.Readme("/nonexistent/path/that/should/fail", "shelly-test", "test")
	if err == nil {
		t.Error("Readme() should error for invalid directory")
	}
}

// containsBytes checks if content contains substr.
func containsBytes(content, substr []byte) bool {
	return bytes.Contains(content, substr)
}
