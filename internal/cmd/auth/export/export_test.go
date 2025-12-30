package export

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

func TestNewCommand(t *testing.T) {
	t.Parallel()
	cmd := NewCommand(cmdutil.NewFactory())

	if cmd == nil {
		t.Fatal("NewCommand returned nil")
	}

	if cmd.Use == "" {
		t.Error("Use is empty")
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestNewCommand_Structure(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	// Test Use
	if cmd.Use != "export [device...]" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export [device...]")
	}

	// Test Aliases
	wantAliases := []string{"exp", "backup"}
	if len(cmd.Aliases) != len(wantAliases) {
		t.Errorf("Aliases = %v, want %v", cmd.Aliases, wantAliases)
	} else {
		for i, alias := range wantAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Aliases[%d] = %q, want %q", i, cmd.Aliases[i], alias)
			}
		}
	}

	// Test Long
	if cmd.Long == "" {
		t.Error("Long description is empty")
	}

	// Test Example
	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}

func TestNewCommand_Flags(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	tests := []struct {
		name      string
		shorthand string
		defValue  string
	}{
		{"output", "o", "credentials.json"},
		{"all", "a", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			flag := cmd.Flags().Lookup(tt.name)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.name)
			}
			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
			}
			if flag.DefValue != tt.defValue {
				t.Errorf("flag %q default = %q, want %q", tt.name, flag.DefValue, tt.defValue)
			}
		})
	}
}

func TestNewCommand_Help(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	cmd := NewCommand(tf.Factory)

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("--help should not error: %v", err)
	}
}

func TestNewCommand_ExampleContent(t *testing.T) {
	t.Parallel()

	cmd := NewCommand(cmdutil.NewFactory())

	wantPatterns := []string{
		"shelly auth export",
		"--all",
		"-o",
	}

	for _, pattern := range wantPatterns {
		if !strings.Contains(cmd.Example, pattern) {
			t.Errorf("expected Example to contain %q", pattern)
		}
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Output:   "test.json",
		Password: "secret",
	}
	opts.All = true

	if opts.Output != "test.json" {
		t.Errorf("Output = %q, want %q", opts.Output, "test.json")
	}
	if opts.Password != "secret" {
		t.Errorf("Password = %q, want %q", opts.Password, "secret")
	}
	if !opts.All {
		t.Error("All should be true")
	}
}

func TestExecute_NoDevicesNoAll(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{}) // No devices, no --all
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// The command may succeed (if no credentials) or fail (if credentials exist)
	// Either behavior is acceptable for this test
	err := cmd.Execute()
	if err != nil {
		t.Logf("Execute() error = %v (may be expected)", err)
	}
}

func TestRun_AllWithNoCredentials(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)

	opts := &Options{
		Output: "/tmp/test-export.json",
	}
	opts.All = true

	err := run(context.Background(), tf.Factory, []string{}, opts)
	if err != nil {
		t.Logf("run() error = %v (may be expected)", err)
	}

	out := tf.OutString()
	if strings.Contains(out, "No credentials") {
		t.Logf("Output shows no credentials")
	}
}

func TestRun_NoDevicesNoAll_ReturnsError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// Add a device with credentials so creds is non-empty
	tf.Config.Devices["test-device"] = model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret",
		},
	}

	opts := &Options{
		Output: filepath.Join(t.TempDir(), "test-export.json"),
	}
	// All is false (default), no devices provided

	err := run(t.Context(), tf.Factory, []string{}, opts)
	if err == nil {
		t.Fatal("expected error when no devices specified and --all not set")
	}
	if !strings.Contains(err.Error(), "specify devices or use --all") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRun_AllWithCredentials_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// Add devices with credentials
	tf.Config.Devices["kitchen"] = model.Device{
		Name:    "kitchen",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "kitchen-secret",
		},
	}
	tf.Config.Devices["bedroom"] = model.Device{
		Name:    "bedroom",
		Address: "192.168.1.101",
		Auth: &model.Auth{
			Username: "user",
			Password: "bedroom-secret",
		},
	}

	outputPath := filepath.Join(t.TempDir(), "exported-creds.json")
	opts := &Options{
		Output: outputPath,
	}
	opts.All = true

	err := run(t.Context(), tf.Factory, []string{}, opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify success message
	out := tf.OutString()
	if !strings.Contains(out, "Exported 2 credential(s)") {
		t.Errorf("expected success message with 2 credentials, got: %s", out)
	}

	// Verify file was created
	data, err := os.ReadFile(outputPath) //nolint:gosec // Test file with known path
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	// Verify JSON structure
	var export map[string]any
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if export["version"] != "1.0" {
		t.Errorf("expected version 1.0, got %v", export["version"])
	}
	if _, ok := export["exported_at"]; !ok {
		t.Error("expected exported_at field")
	}
	creds, ok := export["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("expected credentials map, got %T", export["credentials"])
	}
	if len(creds) != 2 {
		t.Errorf("expected 2 credentials, got %d", len(creds))
	}
}

func TestRun_FilterSpecificDevices_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// Add multiple devices with credentials
	tf.Config.Devices["kitchen"] = model.Device{
		Name:    "kitchen",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "kitchen-secret",
		},
	}
	tf.Config.Devices["bedroom"] = model.Device{
		Name:    "bedroom",
		Address: "192.168.1.101",
		Auth: &model.Auth{
			Username: "user",
			Password: "bedroom-secret",
		},
	}
	tf.Config.Devices["living-room"] = model.Device{
		Name:    "living-room",
		Address: "192.168.1.102",
		Auth: &model.Auth{
			Username: "admin",
			Password: "living-secret",
		},
	}

	outputPath := filepath.Join(t.TempDir(), "exported-creds.json")
	opts := &Options{
		Output: outputPath,
	}
	opts.All = false // Not --all, filter by device names

	// Export only kitchen and bedroom
	err := run(t.Context(), tf.Factory, []string{"kitchen", "bedroom"}, opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify success message
	out := tf.OutString()
	if !strings.Contains(out, "Exported 2 credential(s)") {
		t.Errorf("expected success message with 2 credentials, got: %s", out)
	}

	// Verify file was created and contains only filtered devices
	data, err := os.ReadFile(outputPath) //nolint:gosec // Test file with known path
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var export map[string]any
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	creds, ok := export["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("expected credentials map, got %T", export["credentials"])
	}
	if len(creds) != 2 {
		t.Errorf("expected 2 credentials, got %d", len(creds))
	}
	if _, ok := creds["kitchen"]; !ok {
		t.Error("expected kitchen in credentials")
	}
	if _, ok := creds["bedroom"]; !ok {
		t.Error("expected bedroom in credentials")
	}
	if _, ok := creds["living-room"]; ok {
		t.Error("unexpected living-room in credentials")
	}
}

func TestRun_FilterNoMatchingDevices_Warning(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// Add device with credentials
	tf.Config.Devices["kitchen"] = model.Device{
		Name:    "kitchen",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "kitchen-secret",
		},
	}

	outputPath := filepath.Join(t.TempDir(), "exported-creds.json")
	opts := &Options{
		Output: outputPath,
	}
	opts.All = false

	// Try to export a device that doesn't exist
	err := run(t.Context(), tf.Factory, []string{"nonexistent"}, opts)
	if err != nil {
		t.Fatalf("run() should not error for non-matching filter: %v", err)
	}

	// Warning messages go to stderr
	errOut := tf.ErrString()
	if !strings.Contains(errOut, "No matching credentials found") {
		t.Errorf("expected warning about no matching credentials, got: %s", errOut)
	}
}

func TestRun_CreateNestedDirectory_Success(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tf.Config.Devices["test-device"] = model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret",
		},
	}

	// Use a nested path that needs to be created
	outputPath := filepath.Join(t.TempDir(), "nested", "dir", "creds.json")
	opts := &Options{
		Output: outputPath,
	}
	opts.All = true

	err := run(t.Context(), tf.Factory, []string{}, opts)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestRun_WriteFileError(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tf.Config.Devices["test-device"] = model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret",
		},
	}

	// Use an invalid path (write to a directory that exists as a file)
	tmpFile := filepath.Join(t.TempDir(), "file-not-dir")
	if err := os.WriteFile(tmpFile, []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	outputPath := filepath.Join(tmpFile, "creds.json")
	opts := &Options{
		Output: outputPath,
	}
	opts.All = true

	err := run(t.Context(), tf.Factory, []string{}, opts)
	if err == nil {
		t.Fatal("expected error when writing to invalid path")
	}
	// Error could be "create directory" or "write file" depending on OS
	if !strings.Contains(err.Error(), "create directory") && !strings.Contains(err.Error(), "write file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExecute_WithAllFlag(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tf.Config.Devices["test-device"] = model.Device{
		Name:    "test-device",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret",
		},
	}

	outputPath := filepath.Join(t.TempDir(), "creds.json")

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--all", "-o", outputPath})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestExecute_WithSpecificDevices(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	tf.Config.Devices["kitchen"] = model.Device{
		Name:    "kitchen",
		Address: "192.168.1.100",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret",
		},
	}
	tf.Config.Devices["bedroom"] = model.Device{
		Name:    "bedroom",
		Address: "192.168.1.101",
		Auth: &model.Auth{
			Username: "admin",
			Password: "secret2",
		},
	}

	outputPath := filepath.Join(t.TempDir(), "creds.json")

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"kitchen", "-o", outputPath})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify file was created with only kitchen
	data, err := os.ReadFile(outputPath) //nolint:gosec // Test file with known path
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var export map[string]any
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	creds, ok := export["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("expected credentials map, got %T", export["credentials"])
	}
	if len(creds) != 1 {
		t.Errorf("expected 1 credential, got %d", len(creds))
	}
	if _, ok := creds["kitchen"]; !ok {
		t.Error("expected kitchen in credentials")
	}
}

func TestExecute_NoCredentialsWarning(t *testing.T) {
	t.Parallel()

	tf := factory.NewTestFactory(t)
	// No devices with credentials

	outputPath := filepath.Join(t.TempDir(), "creds.json")

	var buf bytes.Buffer
	cmd := NewCommand(tf.Factory)
	cmd.SetContext(t.Context())
	cmd.SetArgs([]string{"--all", "-o", outputPath})
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() should not error: %v", err)
	}

	// Warning messages go to stderr
	errOut := tf.ErrString()
	if !strings.Contains(errOut, "No credentials found") {
		t.Errorf("expected warning about no credentials, got: %s", errOut)
	}
}
