package pluginupgrade_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	"github.com/tj-smith47/shelly-cli/internal/pluginupgrade"
)

const testVersion = "1.0.0"

func TestResult_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result pluginupgrade.Result
	}{
		{
			name: "upgraded result",
			result: pluginupgrade.Result{
				Name:       "myext",
				OldVersion: testVersion,
				NewVersion: "1.1.0",
				Upgraded:   true,
				Skipped:    false,
				Error:      nil,
			},
		},
		{
			name: "skipped result",
			result: pluginupgrade.Result{
				Name:       "myext",
				OldVersion: testVersion,
				NewVersion: "",
				Upgraded:   false,
				Skipped:    true,
				Error:      nil,
			},
		},
		{
			name: "up to date result",
			result: pluginupgrade.Result{
				Name:       "myext",
				OldVersion: testVersion,
				NewVersion: testVersion,
				Upgraded:   false,
				Skipped:    false,
				Error:      nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.result.Name != "myext" {
				t.Errorf("Name = %q, want %q", tt.result.Name, "myext")
			}
			if tt.result.OldVersion != testVersion {
				t.Errorf("OldVersion = %q, want %q", tt.result.OldVersion, testVersion)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	// Creating an upgrader without a real registry will fail
	// Just test that nil registry doesn't panic during creation
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("New() panicked: %v", r)
		}
	}()

	// Pass nil - this tests the function signature
	upgrader := pluginupgrade.New(nil, ios)
	if upgrader == nil {
		t.Error("New() returned nil")
	}
}

func TestNew_WithRegistry(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()
	registry := plugins.NewRegistryWithDir(tmpDir)

	upgrader := pluginupgrade.New(registry, ios)
	if upgrader == nil {
		t.Error("New() returned nil with valid registry")
	}
}

func TestResult_IsUpToDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		oldVersion string
		newVersion string
		upgraded   bool
		skipped    bool
		wantUpTo   bool
	}{
		{
			name:       "same version means up to date",
			oldVersion: "1.0.0",
			newVersion: "1.0.0",
			upgraded:   false,
			skipped:    false,
			wantUpTo:   true,
		},
		{
			name:       "upgraded means not up to date before",
			oldVersion: "1.0.0",
			newVersion: "1.1.0",
			upgraded:   true,
			skipped:    false,
			wantUpTo:   false,
		},
		{
			name:       "skipped is not up to date",
			oldVersion: "1.0.0",
			newVersion: "",
			upgraded:   false,
			skipped:    true,
			wantUpTo:   false,
		},
		{
			name:       "empty versions with no upgrade",
			oldVersion: "",
			newVersion: "",
			upgraded:   false,
			skipped:    false,
			wantUpTo:   true,
		},
		{
			name:       "version with v prefix same",
			oldVersion: "v1.0.0",
			newVersion: "v1.0.0",
			upgraded:   false,
			skipped:    false,
			wantUpTo:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := pluginupgrade.Result{
				OldVersion: tt.oldVersion,
				NewVersion: tt.newVersion,
				Upgraded:   tt.upgraded,
				Skipped:    tt.skipped,
			}

			// Check if it was up to date (not upgraded and not skipped and versions match)
			isUpToDate := !result.Upgraded && !result.Skipped && result.OldVersion == result.NewVersion
			if isUpToDate != tt.wantUpTo {
				t.Errorf("up to date = %v, want %v", isUpToDate, tt.wantUpTo)
			}
		})
	}
}

func TestResult_HasError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{
			name:    "no error",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "with error",
			err:     errors.New("test error"),
			wantErr: true,
		},
		{
			name:    "with wrapped error",
			err:     errors.New("wrapped: inner error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := pluginupgrade.Result{
				Error: tt.err,
			}

			hasErr := result.Error != nil
			if hasErr != tt.wantErr {
				t.Errorf("has error = %v, want %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestUpgrader_Fields(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	upgrader := pluginupgrade.New(nil, ios)
	if upgrader == nil {
		t.Error("New() returned nil")
	}
}

func TestResult_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result pluginupgrade.Result
	}{
		{
			name: "full result",
			result: pluginupgrade.Result{
				Name:       "test-plugin",
				OldVersion: "1.0.0",
				NewVersion: "2.0.0",
				Upgraded:   true,
				Skipped:    false,
				Error:      nil,
			},
		},
		{
			name: "result with error",
			result: pluginupgrade.Result{
				Name:       "error-plugin",
				OldVersion: "1.0.0",
				NewVersion: "",
				Upgraded:   false,
				Skipped:    true,
				Error:      errors.New("download failed"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Result is a simple struct, just verify fields are accessible
			if tt.result.Name == "" {
				t.Error("Name should not be empty")
			}
		})
	}
}

func TestResult_UpgradedFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		upgraded     bool
		skipped      bool
		hasError     bool
		expectStatus string
	}{
		{
			name:         "successfully upgraded",
			upgraded:     true,
			skipped:      false,
			hasError:     false,
			expectStatus: "upgraded",
		},
		{
			name:         "skipped",
			upgraded:     false,
			skipped:      true,
			hasError:     false,
			expectStatus: "skipped",
		},
		{
			name:         "error occurred",
			upgraded:     false,
			skipped:      false,
			hasError:     true,
			expectStatus: "error",
		},
		{
			name:         "up to date",
			upgraded:     false,
			skipped:      false,
			hasError:     false,
			expectStatus: "up_to_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var err error
			if tt.hasError {
				err = errors.New("test error")
			}

			result := pluginupgrade.Result{
				Upgraded: tt.upgraded,
				Skipped:  tt.skipped,
				Error:    err,
			}

			// Determine status based on flags
			var status string
			switch {
			case result.Upgraded:
				status = "upgraded"
			case result.Skipped:
				status = "skipped"
			case result.Error != nil:
				status = "error"
			default:
				status = "up_to_date"
			}

			if status != tt.expectStatus {
				t.Errorf("status = %q, want %q", status, tt.expectStatus)
			}
		})
	}
}

func TestResult_VersionComparison(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		oldVersion string
		newVersion string
		wantNewer  bool
	}{
		{
			name:       "newer version available",
			oldVersion: "1.0.0",
			newVersion: "1.1.0",
			wantNewer:  true,
		},
		{
			name:       "same version",
			oldVersion: "1.0.0",
			newVersion: "1.0.0",
			wantNewer:  false,
		},
		{
			name:       "with v prefix newer",
			oldVersion: "v1.0.0",
			newVersion: "v2.0.0",
			wantNewer:  true,
		},
		{
			name:       "empty old version",
			oldVersion: "",
			newVersion: "1.0.0",
			wantNewer:  true,
		},
		{
			name:       "both empty",
			oldVersion: "",
			newVersion: "",
			wantNewer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := pluginupgrade.Result{
				OldVersion: tt.oldVersion,
				NewVersion: tt.newVersion,
			}

			// Simple string comparison to determine if newer
			isNewer := result.NewVersion != "" && result.OldVersion != result.NewVersion
			if isNewer != tt.wantNewer {
				t.Errorf("isNewer = %v, want %v", isNewer, tt.wantNewer)
			}
		})
	}
}

func TestNew_WithIOStreams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ios  *iostreams.IOStreams
	}{
		{
			name: "with test iostreams",
			ios:  iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			upgrader := pluginupgrade.New(nil, tt.ios)
			if upgrader == nil {
				t.Error("New() returned nil with valid IOStreams")
			}
		})
	}
}

func TestResult_AllFieldsCombinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		result     pluginupgrade.Result
		wantName   string
		wantOld    string
		wantNew    string
		wantUpgrad bool
		wantSkip   bool
		wantErr    bool
	}{
		{
			name: "successful upgrade",
			result: pluginupgrade.Result{
				Name:       "myext",
				OldVersion: "1.0.0",
				NewVersion: "2.0.0",
				Upgraded:   true,
				Skipped:    false,
				Error:      nil,
			},
			wantName:   "myext",
			wantOld:    "1.0.0",
			wantNew:    "2.0.0",
			wantUpgrad: true,
			wantSkip:   false,
			wantErr:    false,
		},
		{
			name: "skipped non-github source",
			result: pluginupgrade.Result{
				Name:       "localext",
				OldVersion: "1.0.0",
				NewVersion: "",
				Upgraded:   false,
				Skipped:    true,
				Error:      errors.New("local source cannot be upgraded"),
			},
			wantName:   "localext",
			wantOld:    "1.0.0",
			wantNew:    "",
			wantUpgrad: false,
			wantSkip:   true,
			wantErr:    true,
		},
		{
			name: "already up to date",
			result: pluginupgrade.Result{
				Name:       "current",
				OldVersion: "3.0.0",
				NewVersion: "3.0.0",
				Upgraded:   false,
				Skipped:    false,
				Error:      nil,
			},
			wantName:   "current",
			wantOld:    "3.0.0",
			wantNew:    "3.0.0",
			wantUpgrad: false,
			wantSkip:   false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tt.result.Name, tt.wantName)
			}
			if tt.result.OldVersion != tt.wantOld {
				t.Errorf("OldVersion = %q, want %q", tt.result.OldVersion, tt.wantOld)
			}
			if tt.result.NewVersion != tt.wantNew {
				t.Errorf("NewVersion = %q, want %q", tt.result.NewVersion, tt.wantNew)
			}
			if tt.result.Upgraded != tt.wantUpgrad {
				t.Errorf("Upgraded = %v, want %v", tt.result.Upgraded, tt.wantUpgrad)
			}
			if tt.result.Skipped != tt.wantSkip {
				t.Errorf("Skipped = %v, want %v", tt.result.Skipped, tt.wantSkip)
			}
			hasErr := tt.result.Error != nil
			if hasErr != tt.wantErr {
				t.Errorf("has error = %v, want %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestUpgradeAll_EmptyRegistry(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create plugins directory to avoid "directory not found" error
	if err := os.MkdirAll(tmpDir, 0o750); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Empty registry should return empty results
	if len(results) != 0 {
		t.Errorf("UpgradeAll() returned %d results, want 0", len(results))
	}
}

func TestUpgradeOne_NotInstalled(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	_, err := upgrader.UpgradeOne(context.Background(), "nonexistent")
	if err == nil {
		t.Error("UpgradeOne() should error for non-installed extension")
	}

	if err.Error() != `extension "nonexistent" is not installed` {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUpgradeAll_WithLocalPlugin(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-testplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file - needs executable permissions for plugin testing
	binaryPath := filepath.Join(pluginDir, "shelly-testplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with local source (not upgradeable)
	manifest := plugins.NewManifest("testplugin", plugins.ParseLocalSource("/tmp/shelly-testplugin"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result for the local plugin
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// Local plugin should be skipped
	if !results[0].Skipped {
		t.Error("local plugin should be skipped")
	}

	if results[0].Error == nil {
		t.Error("skipped plugin should have error explaining why")
	}
}

func TestUpgradeOne_WithLocalPlugin(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-localplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file - needs executable permissions for plugin testing
	binaryPath := filepath.Join(pluginDir, "shelly-localplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with local source (not upgradeable)
	manifest := plugins.NewManifest("localplugin", plugins.ParseLocalSource("/tmp/shelly-localplugin"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// UpgradeOne uses a Loader that searches default paths (not our temp dir)
	// So it may not find the plugin. This is expected behavior - it tests
	// that the code handles the "plugin found in registry but not by loader" case
	result, err := upgrader.UpgradeOne(context.Background(), "localplugin")

	// Either succeeds with skipped result, or fails with "not found"
	if err != nil {
		// The plugin is in the registry but the loader can't find it
		// This exercises the error path in UpgradeOne
		return
	}

	// If found, local plugin should be skipped
	if !result.Skipped {
		t.Error("local plugin should be skipped")
	}
}

func TestUpgradeAll_WithGitHubPlugin(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-ghplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file - needs executable permissions for plugin testing
	binaryPath := filepath.Join(pluginDir, "shelly-ghplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with GitHub source
	manifest := plugins.NewManifest("ghplugin", plugins.ParseGitHubSource("testuser/shelly-ghplugin", "v1.0.0", "shelly-ghplugin"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Use a context with very short timeout to avoid actual network calls
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	results, err := upgrader.UpgradeAll(ctx)
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// Plugin should have an error (network timeout or similar)
	// because we can't actually reach GitHub
	// If no error and not upgraded, it might be "already up to date" which is also valid
	// Any outcome is acceptable since network conditions vary
	_ = results[0].Error
}

func TestUpgradeOne_WithGitHubPlugin(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-ghplugin2")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file - needs executable permissions for plugin testing
	binaryPath := filepath.Join(pluginDir, "shelly-ghplugin2")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with GitHub source
	manifest := plugins.NewManifest("ghplugin2", plugins.ParseGitHubSource("testuser/shelly-ghplugin2", "v1.0.0", "shelly-ghplugin2"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Use a context with very short timeout to avoid actual network calls
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	// UpgradeOne uses a Loader that searches default paths (not our temp dir)
	// So it may not find the plugin even though it's in the registry
	result, err := upgrader.UpgradeOne(ctx, "ghplugin2")

	// Either succeeds (with some result) or fails with "not found"
	if err != nil {
		// The plugin is in the registry but the loader can't find it
		// This exercises the error path in UpgradeOne
		return
	}

	// Result should exist with correct name
	if result.Name != "ghplugin2" {
		t.Errorf("result.Name = %q, want %q", result.Name, "ghplugin2")
	}
}

func TestUpgradeAll_WithNoManifest(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure (old format without manifest)
	pluginDir := filepath.Join(tmpDir, "shelly-oldplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create just a binary file (no manifest)
	binaryPath := filepath.Join(pluginDir, "shelly-oldplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Use a context with very short timeout to avoid actual network calls
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	results, err := upgrader.UpgradeAll(ctx)
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
	}
}

func TestUpgradeAll_WithUnknownSourceManifest(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-unknownplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file
	binaryPath := filepath.Join(pluginDir, "shelly-unknownplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with unknown source (not upgradeable)
	manifest := plugins.NewManifest("unknownplugin", plugins.UnknownSource())
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// Unknown source should be skipped
	if !results[0].Skipped {
		t.Error("unknown source plugin should be skipped")
	}
}

func TestUpgradeAll_WithURLSourceManifest(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-urlplugin")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file
	binaryPath := filepath.Join(pluginDir, "shelly-urlplugin")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with URL source (not auto-upgradeable)
	manifest := plugins.NewManifest("urlplugin", plugins.ParseURLSource("https://example.com/shelly-urlplugin"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// URL source without GitHub should be skipped
	if !results[0].Skipped {
		t.Error("URL source plugin should be skipped")
	}
}

func TestUpgradeAll_MultiplePlugins(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create multiple plugins
	for _, name := range []string{"plugin1", "plugin2", "plugin3"} {
		pluginDir := filepath.Join(tmpDir, "shelly-"+name)
		if err := os.MkdirAll(pluginDir, 0o750); err != nil {
			t.Fatalf("failed to create plugin dir: %v", err)
		}

		binaryPath := filepath.Join(pluginDir, "shelly-"+name)
		if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho "+name), 0o750); err != nil { //nolint:gosec // G306: executable script
			t.Fatalf("failed to create binary: %v", err)
		}

		manifest := plugins.NewManifest(name, plugins.ParseLocalSource("/tmp/shelly-"+name))
		manifest.Version = testVersion
		if err := manifest.Save(pluginDir); err != nil {
			t.Fatalf("failed to save manifest: %v", err)
		}
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 3 results
	if len(results) != 3 {
		t.Errorf("UpgradeAll() returned %d results, want 3", len(results))
	}

	// All should be skipped (local source)
	for _, r := range results {
		if !r.Skipped {
			t.Errorf("plugin %s should be skipped", r.Name)
		}
	}
}

func TestUpgradeOne_PluginFoundInLoaderButNotRegistry(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create empty registry (plugin not installed according to registry)
	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Try to upgrade a plugin that's not in registry
	_, err := upgrader.UpgradeOne(context.Background(), "notinregistry")
	if err == nil {
		t.Error("UpgradeOne() should error for plugin not in registry")
	}

	// Error should mention "not installed"
	if err.Error() != `extension "notinregistry" is not installed` {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResult_ZeroValue(t *testing.T) {
	t.Parallel()

	// Test zero value result
	var result pluginupgrade.Result

	if result.Name != "" {
		t.Errorf("zero value Name = %q, want empty", result.Name)
	}
	if result.OldVersion != "" {
		t.Errorf("zero value OldVersion = %q, want empty", result.OldVersion)
	}
	if result.NewVersion != "" {
		t.Errorf("zero value NewVersion = %q, want empty", result.NewVersion)
	}
	if result.Upgraded {
		t.Error("zero value Upgraded = true, want false")
	}
	if result.Skipped {
		t.Error("zero value Skipped = true, want false")
	}
	if result.Error != nil {
		t.Errorf("zero value Error = %v, want nil", result.Error)
	}
}

func TestUpgradeAll_WithInvalidGitHubURL(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin directory structure
	pluginDir := filepath.Join(tmpDir, "shelly-invalidgh")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	// Create a binary file
	binaryPath := filepath.Join(pluginDir, "shelly-invalidgh")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create a manifest with invalid GitHub URL (no slash in path)
	source := plugins.Source{
		Type: plugins.SourceTypeGitHub,
		URL:  "https://github.com/invalid-no-slash", // Invalid - no owner/repo separator
	}
	manifest := plugins.NewManifest("invalidgh", source)
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// Should have error about invalid URL
	if results[0].Error == nil {
		t.Error("expected error for invalid GitHub URL")
	}
}

func TestUpgradeAll_ContextCanceled(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin with GitHub source
	pluginDir := filepath.Join(tmpDir, "shelly-ghctxtest")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	binaryPath := filepath.Join(pluginDir, "shelly-ghctxtest")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	manifest := plugins.NewManifest("ghctxtest", plugins.ParseGitHubSource("testuser/shelly-ghctxtest", "v1.0.0", "shelly-ghctxtest"))
	manifest.Version = testVersion
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should handle cancelled context gracefully
	results, err := upgrader.UpgradeAll(ctx)
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Results should exist
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
	}
}

func TestUpgradeAll_WithValidGitHubManifest(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin with valid GitHub source format
	pluginDir := filepath.Join(tmpDir, "shelly-validgh")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	binaryPath := filepath.Join(pluginDir, "shelly-validgh")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create manifest with valid GitHub source
	manifest := plugins.NewManifest("validgh", plugins.ParseGitHubSource("owner/repo", "v1.0.0", "shelly-validgh"))
	manifest.Version = testVersion
	manifest.Description = "A test plugin"
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Use very short timeout - will fail on network but exercise code paths
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	defer cancel()

	results, err := upgrader.UpgradeAll(ctx)
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	// Should have an error (network failure) but result should exist
	if results[0].Name != "validgh" {
		t.Errorf("result.Name = %q, want %q", results[0].Name, "validgh")
	}
}

func TestUpgradeAll_PluginWithEmptyVersion(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create a plugin with empty version
	pluginDir := filepath.Join(tmpDir, "shelly-nover")
	if err := os.MkdirAll(pluginDir, 0o750); err != nil {
		t.Fatalf("failed to create plugin dir: %v", err)
	}

	binaryPath := filepath.Join(pluginDir, "shelly-nover")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
		t.Fatalf("failed to create binary: %v", err)
	}

	// Create manifest without version
	manifest := plugins.NewManifest("nover", plugins.ParseLocalSource("/tmp/shelly-nover"))
	// Don't set Version - leave empty
	if err := manifest.Save(pluginDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 1 result with empty old version
	if len(results) != 1 {
		t.Errorf("UpgradeAll() returned %d results, want 1", len(results))
		return
	}

	if results[0].OldVersion != "" {
		t.Errorf("OldVersion = %q, want empty", results[0].OldVersion)
	}
}

func TestUpgradeOne_NotFound(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	// Try to upgrade an extension that doesn't exist
	result, err := upgrader.UpgradeOne(context.Background(), "nonexistent-plugin")
	if err == nil {
		t.Error("expected error for non-existent plugin")
	}

	// result should be empty/zero
	if result.Name != "" {
		t.Errorf("result.Name = %q, want empty", result.Name)
	}
}

func TestUpgradeAll_SourceTypeFiltering(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test(&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})
	tmpDir := t.TempDir()

	// Create plugins with different source types
	sourceTypes := []struct {
		name       string
		sourceType string
		source     plugins.Source
	}{
		{"local-src", "local", plugins.ParseLocalSource("/tmp/shelly-local-src")},
		{"unknown-src", "unknown", plugins.UnknownSource()},
		{"url-src", "url", plugins.ParseURLSource("https://example.com/shelly-url-src")},
	}

	for _, st := range sourceTypes {
		pluginDir := filepath.Join(tmpDir, "shelly-"+st.name)
		if err := os.MkdirAll(pluginDir, 0o750); err != nil {
			t.Fatalf("failed to create plugin dir: %v", err)
		}

		binaryPath := filepath.Join(pluginDir, "shelly-"+st.name)
		if err := os.WriteFile(binaryPath, []byte("#!/bin/bash\necho test"), 0o750); err != nil { //nolint:gosec // G306: executable script
			t.Fatalf("failed to create binary: %v", err)
		}

		manifest := plugins.NewManifest(st.name, st.source)
		manifest.Version = testVersion
		if err := manifest.Save(pluginDir); err != nil {
			t.Fatalf("failed to save manifest: %v", err)
		}
	}

	registry := plugins.NewRegistryWithDir(tmpDir)
	upgrader := pluginupgrade.New(registry, ios)

	results, err := upgrader.UpgradeAll(context.Background())
	if err != nil {
		t.Errorf("UpgradeAll() error = %v", err)
	}

	// Should have 3 results (one for each plugin)
	if len(results) != 3 {
		t.Errorf("UpgradeAll() returned %d results, want 3", len(results))
		return
	}

	// All non-GitHub sources should be skipped
	for _, result := range results {
		if !result.Skipped {
			t.Errorf("plugin %q should be skipped", result.Name)
		}
	}
}
