package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestNewManifest tests NewManifest function.
func TestNewManifest(t *testing.T) {
	t.Parallel()

	// Truncate to seconds since RFC3339 format used by manifest drops nanoseconds
	before := time.Now().UTC().Truncate(time.Second)
	source := Source{Type: SourceTypeGitHub, URL: "https://github.com/user/repo"}
	manifest := NewManifest("test-plugin", source)
	after := time.Now().UTC().Add(time.Second).Truncate(time.Second) // Add 1s buffer

	if manifest.SchemaVersion != ManifestSchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", manifest.SchemaVersion, ManifestSchemaVersion)
	}
	if manifest.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", manifest.Name, "test-plugin")
	}
	if manifest.Source.Type != SourceTypeGitHub {
		t.Errorf("Source.Type = %q, want %q", manifest.Source.Type, SourceTypeGitHub)
	}
	if manifest.Binary.Name != PluginPrefix+"test-plugin" {
		t.Errorf("Binary.Name = %q, want %q", manifest.Binary.Name, PluginPrefix+"test-plugin")
	}
	expectedPlatform := runtime.GOOS + "-" + runtime.GOARCH
	if manifest.Binary.Platform != expectedPlatform {
		t.Errorf("Binary.Platform = %q, want %q", manifest.Binary.Platform, expectedPlatform)
	}

	// Check timestamps are set and valid
	installedAt, err := time.Parse(time.RFC3339, manifest.InstalledAt)
	if err != nil {
		t.Fatalf("failed to parse InstalledAt: %v", err)
	}
	if installedAt.Before(before) || installedAt.After(after) {
		t.Errorf("InstalledAt = %v, should be between %v and %v", installedAt, before, after)
	}
}

// TestManifest_Save tests Manifest Save method.
func TestManifest_Save(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	manifest := NewManifest("test", ParseLocalSource("/path/to/plugin"))
	manifest.Version = "1.0.0"
	manifest.Description = "Test plugin"

	if err := manifest.Save(tmpDir); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file was created
	manifestPath := filepath.Join(tmpDir, ManifestFileName)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("manifest file not created: %v", err)
	}

	// Verify we can load it back
	loaded, err := LoadManifest(tmpDir)
	if err != nil {
		t.Fatalf("LoadManifest() error: %v", err)
	}
	if loaded.Name != "test" {
		t.Errorf("loaded.Name = %q, want %q", loaded.Name, "test")
	}
	if loaded.Version != "1.0.0" {
		t.Errorf("loaded.Version = %q, want %q", loaded.Version, "1.0.0")
	}
}

// TestManifest_Save_InvalidDir tests Save with invalid directory.
func TestManifest_Save_InvalidDir(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("test", ParseLocalSource("/path"))
	err := manifest.Save("/nonexistent/directory/path")
	if err == nil {
		t.Error("Save() should fail with invalid directory")
	}
}

// TestLoadManifest_FileNotFound tests LoadManifest with missing file.
func TestLoadManifest_FileNotFound(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	_, err = LoadManifest(tmpDir)
	if err == nil {
		t.Error("LoadManifest() should fail when file doesn't exist")
	}
}

// TestLoadManifest_InvalidJSON tests LoadManifest with invalid JSON.
func TestLoadManifest_InvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Write invalid JSON
	manifestPath := filepath.Join(tmpDir, ManifestFileName)
	//nolint:gosec // G306: test file
	if err := os.WriteFile(manifestPath, []byte("not valid json"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = LoadManifest(tmpDir)
	if err == nil {
		t.Error("LoadManifest() should fail with invalid JSON")
	}
}

// TestManifest_SetBinaryInfo tests SetBinaryInfo method.
func TestManifest_SetBinaryInfo(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test binary
	binaryPath := filepath.Join(tmpDir, "shelly-test")
	//nolint:gosec // Test file
	if err := os.WriteFile(binaryPath, []byte("test binary content"), 0o755); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	manifest := NewManifest("test", ParseLocalSource(binaryPath))
	if err := manifest.SetBinaryInfo(binaryPath); err != nil {
		t.Fatalf("SetBinaryInfo() error: %v", err)
	}

	if manifest.Binary.Checksum == "" {
		t.Error("Checksum should not be empty after SetBinaryInfo")
	}
	if !strings.HasPrefix(manifest.Binary.Checksum, "sha256:") {
		t.Errorf("Checksum = %q, should start with 'sha256:'", manifest.Binary.Checksum)
	}
	if manifest.Binary.Size == 0 {
		t.Error("Size should not be 0 after SetBinaryInfo")
	}
}

// TestManifest_SetBinaryInfo_FileNotFound tests SetBinaryInfo with missing file.
func TestManifest_SetBinaryInfo_FileNotFound(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("test", ParseLocalSource("/path"))
	err := manifest.SetBinaryInfo("/nonexistent/path")
	if err == nil {
		t.Error("SetBinaryInfo() should fail with missing file")
	}
}

// TestManifest_MarkUpdated tests MarkUpdated method.
func TestManifest_MarkUpdated(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("test", ParseLocalSource("/path"))

	// Sleep to ensure time moves forward (at least 1 second since RFC3339 truncates to seconds)
	time.Sleep(1100 * time.Millisecond)

	// Truncate to seconds since RFC3339 format drops nanoseconds
	before := time.Now().UTC().Truncate(time.Second)
	manifest.MarkUpdated()
	after := time.Now().UTC().Add(time.Second).Truncate(time.Second) // Add 1s buffer

	// Verify UpdatedAt was updated
	updatedAt, err := time.Parse(time.RFC3339, manifest.UpdatedAt)
	if err != nil {
		t.Fatalf("failed to parse UpdatedAt: %v", err)
	}
	if updatedAt.Before(before) || updatedAt.After(after) {
		t.Errorf("UpdatedAt = %v, should be between %v and %v", updatedAt, before, after)
	}
}

// TestComputeChecksum tests ComputeChecksum function.
func TestComputeChecksum(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.bin")
	//nolint:gosec // Test file
	if err := os.WriteFile(testFile, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	checksum, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum() error: %v", err)
	}

	if !strings.HasPrefix(checksum, "sha256:") {
		t.Errorf("checksum = %q, should start with 'sha256:'", checksum)
	}

	// Verify it's deterministic
	checksum2, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum() second call error: %v", err)
	}
	if checksum != checksum2 {
		t.Errorf("checksum should be deterministic: %q != %q", checksum, checksum2)
	}
}

// TestComputeChecksum_FileNotFound tests ComputeChecksum with missing file.
func TestComputeChecksum_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := ComputeChecksum("/nonexistent/path")
	if err == nil {
		t.Error("ComputeChecksum() should fail with missing file")
	}
}

// TestManifest_VerifyChecksum tests VerifyChecksum method.
func TestManifest_VerifyChecksum(t *testing.T) {
	t.Parallel()

	tmpDir, err := os.MkdirTemp("", "shelly-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("warning: failed to remove temp dir: %v", err)
		}
	})

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.bin")
	//nolint:gosec // Test file
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Get the correct checksum
	checksum, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum() error: %v", err)
	}

	manifest := NewManifest("test", ParseLocalSource(testFile))
	manifest.Binary.Checksum = checksum

	// Verify matching checksum
	valid, err := manifest.VerifyChecksum(testFile)
	if err != nil {
		t.Fatalf("VerifyChecksum() error: %v", err)
	}
	if !valid {
		t.Error("VerifyChecksum() should return true for matching checksum")
	}

	// Test with wrong checksum
	manifest.Binary.Checksum = "sha256:wrongchecksum"
	valid, err = manifest.VerifyChecksum(testFile)
	if err != nil {
		t.Fatalf("VerifyChecksum() error: %v", err)
	}
	if valid {
		t.Error("VerifyChecksum() should return false for mismatched checksum")
	}
}

// TestManifest_VerifyChecksum_NoChecksum tests VerifyChecksum with empty checksum.
func TestManifest_VerifyChecksum_NoChecksum(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("test", ParseLocalSource("/path"))
	manifest.Binary.Checksum = ""

	_, err := manifest.VerifyChecksum("/some/path")
	if err == nil {
		t.Error("VerifyChecksum() should fail when manifest has no checksum")
	}
}

// TestManifest_BinaryPath tests BinaryPath method.
func TestManifest_BinaryPath(t *testing.T) {
	t.Parallel()

	manifest := NewManifest("test", ParseLocalSource("/path"))
	manifest.Binary.Name = "shelly-test"

	path := manifest.BinaryPath("/home/user/plugins/shelly-test")
	expected := filepath.Join("/home/user/plugins/shelly-test", "shelly-test") //nolint:gocritic // test path
	if path != expected {
		t.Errorf("BinaryPath() = %q, want %q", path, expected)
	}
}

// TestManifest_CanUpgrade tests CanUpgrade method.
func TestManifest_CanUpgrade(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sourceType string
		want       bool
	}{
		{"github source can upgrade", SourceTypeGitHub, true},
		{"url source can upgrade", SourceTypeURL, true},
		{"local source cannot upgrade", SourceTypeLocal, false},
		{"unknown source cannot upgrade", SourceTypeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manifest := NewManifest("test", Source{Type: tt.sourceType})
			if got := manifest.CanUpgrade(); got != tt.want {
				t.Errorf("CanUpgrade() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestManifest_UpgradeMessage tests UpgradeMessage method.
func TestManifest_UpgradeMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		source     Source
		wantSubstr string
	}{
		{
			name:       "github source",
			source:     Source{Type: SourceTypeGitHub, URL: "https://github.com/user/repo"},
			wantSubstr: "GitHub",
		},
		{
			name:       "url source",
			source:     Source{Type: SourceTypeURL, URL: "https://example.com/plugin.zip"},
			wantSubstr: "URL",
		},
		{
			name:       "local source",
			source:     Source{Type: SourceTypeLocal, Path: "/path/to/plugin"},
			wantSubstr: "local file",
		},
		{
			name:       "unknown source",
			source:     Source{Type: SourceTypeUnknown},
			wantSubstr: "Unknown source",
		},
		{
			name:       "empty type",
			source:     Source{Type: ""},
			wantSubstr: "Unknown source type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manifest := NewManifest("test", tt.source)
			msg := manifest.UpgradeMessage()
			if !strings.Contains(msg, tt.wantSubstr) {
				t.Errorf("UpgradeMessage() = %q, want to contain %q", msg, tt.wantSubstr)
			}
		})
	}
}

// TestParseGitHubSource tests ParseGitHubSource function.
func TestParseGitHubSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		repoStr   string
		tagName   string
		assetName string
		wantURL   string
		wantRef   string
		wantAsset string
	}{
		{
			name:      "basic repo",
			repoStr:   "user/repo",
			tagName:   "v1.0.0",
			assetName: "plugin-linux-amd64.tar.gz",
			wantURL:   "https://github.com/user/repo",
			wantRef:   "v1.0.0",
			wantAsset: "plugin-linux-amd64.tar.gz",
		},
		{
			name:      "with gh: prefix",
			repoStr:   "gh:user/repo",
			tagName:   "v2.0.0",
			assetName: "plugin.zip",
			wantURL:   "https://github.com/user/repo",
			wantRef:   "v2.0.0",
			wantAsset: "plugin.zip",
		},
		{
			name:      "with github: prefix",
			repoStr:   "github:user/repo",
			tagName:   "v1.5.0",
			assetName: "plugin.tar.gz",
			wantURL:   "https://github.com/user/repo",
			wantRef:   "v1.5.0",
			wantAsset: "plugin.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := ParseGitHubSource(tt.repoStr, tt.tagName, tt.assetName)

			if source.Type != SourceTypeGitHub {
				t.Errorf("Type = %q, want %q", source.Type, SourceTypeGitHub)
			}
			if source.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", source.URL, tt.wantURL)
			}
			if source.Ref != tt.wantRef {
				t.Errorf("Ref = %q, want %q", source.Ref, tt.wantRef)
			}
			if source.Asset != tt.wantAsset {
				t.Errorf("Asset = %q, want %q", source.Asset, tt.wantAsset)
			}
		})
	}
}

// TestParseURLSource tests ParseURLSource function.
func TestParseURLSource(t *testing.T) {
	t.Parallel()

	url := "https://example.com/plugin/v1.0.0/plugin-linux-amd64.tar.gz"
	source := ParseURLSource(url)

	if source.Type != SourceTypeURL {
		t.Errorf("Type = %q, want %q", source.Type, SourceTypeURL)
	}
	if source.URL != url {
		t.Errorf("URL = %q, want %q", source.URL, url)
	}
}

// TestParseLocalSource tests ParseLocalSource function.
func TestParseLocalSource(t *testing.T) {
	t.Parallel()

	path := "/home/user/downloads/shelly-plugin"
	source := ParseLocalSource(path)

	if source.Type != SourceTypeLocal {
		t.Errorf("Type = %q, want %q", source.Type, SourceTypeLocal)
	}
	// Path should be absolute
	if !filepath.IsAbs(source.Path) {
		t.Errorf("Path = %q, should be absolute", source.Path)
	}
}

// TestParseLocalSource_RelativePath tests ParseLocalSource with relative path.
func TestParseLocalSource_RelativePath(t *testing.T) {
	t.Parallel()

	path := "./plugin"
	source := ParseLocalSource(path)

	if source.Type != SourceTypeLocal {
		t.Errorf("Type = %q, want %q", source.Type, SourceTypeLocal)
	}
	// Should be converted to absolute
	if !filepath.IsAbs(source.Path) {
		t.Errorf("Path = %q, should be absolute", source.Path)
	}
}

// TestUnknownSource tests UnknownSource function.
func TestUnknownSource(t *testing.T) {
	t.Parallel()

	source := UnknownSource()

	if source.Type != SourceTypeUnknown {
		t.Errorf("Type = %q, want %q", source.Type, SourceTypeUnknown)
	}
}

// TestManifestSchemaVersion tests ManifestSchemaVersion constant.
func TestManifestSchemaVersion(t *testing.T) {
	t.Parallel()

	if ManifestSchemaVersion != "1" {
		t.Errorf("ManifestSchemaVersion = %q, want %q", ManifestSchemaVersion, "1")
	}
}

// TestManifestFileName tests ManifestFileName constant.
func TestManifestFileName(t *testing.T) {
	t.Parallel()

	if ManifestFileName != "manifest.json" {
		t.Errorf("ManifestFileName = %q, want %q", ManifestFileName, "manifest.json")
	}
}

// TestMigrationMarkerFile tests MigrationMarkerFile constant.
func TestMigrationMarkerFile(t *testing.T) {
	t.Parallel()

	if MigrationMarkerFile != ".migrated" {
		t.Errorf("MigrationMarkerFile = %q, want %q", MigrationMarkerFile, ".migrated")
	}
}

// TestSourceTypeConstants tests source type constants.
func TestSourceTypeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		want     string
	}{
		{"github", SourceTypeGitHub, "github"},
		{"url", SourceTypeURL, "url"},
		{"local", SourceTypeLocal, "local"},
		{"unknown", SourceTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.constant != tt.want {
				t.Errorf("constant = %q, want %q", tt.constant, tt.want)
			}
		})
	}
}

// TestCapabilities tests Capabilities struct.
func TestCapabilities(t *testing.T) {
	t.Parallel()

	caps := Capabilities{
		DeviceDetection: true,
		Platform:        "tasmota",
		Components:      []string{"switch", "light", "energy"},
		FirmwareUpdates: true,
		Hints: map[string]string{
			"scene":    "Tasmota uses Rules instead of scenes",
			"schedule": "Use Timers in Tasmota web UI",
		},
	}

	if !caps.DeviceDetection {
		t.Error("DeviceDetection = false, want true")
	}
	if caps.Platform != "tasmota" {
		t.Errorf("Platform = %q, want %q", caps.Platform, "tasmota")
	}
	if len(caps.Components) != 3 {
		t.Errorf("Components len = %d, want 3", len(caps.Components))
	}
	if !caps.FirmwareUpdates {
		t.Error("FirmwareUpdates = false, want true")
	}
	if len(caps.Hints) != 2 {
		t.Errorf("Hints len = %d, want 2", len(caps.Hints))
	}
}

// TestHooks tests Hooks struct.
func TestHooks(t *testing.T) {
	t.Parallel()

	hooks := Hooks{
		Detect:       "./shelly-tasmota detect",
		Status:       "./shelly-tasmota status",
		Control:      "./shelly-tasmota control",
		CheckUpdates: "./shelly-tasmota check-updates",
		ApplyUpdate:  "./shelly-tasmota apply-update",
	}

	if hooks.Detect != "./shelly-tasmota detect" {
		t.Errorf("Detect = %q, want %q", hooks.Detect, "./shelly-tasmota detect")
	}
	if hooks.Status != "./shelly-tasmota status" {
		t.Errorf("Status = %q, want %q", hooks.Status, "./shelly-tasmota status")
	}
	if hooks.Control != "./shelly-tasmota control" {
		t.Errorf("Control = %q, want %q", hooks.Control, "./shelly-tasmota control")
	}
	if hooks.CheckUpdates != "./shelly-tasmota check-updates" {
		t.Errorf("CheckUpdates = %q, want %q", hooks.CheckUpdates, "./shelly-tasmota check-updates")
	}
	if hooks.ApplyUpdate != "./shelly-tasmota apply-update" {
		t.Errorf("ApplyUpdate = %q, want %q", hooks.ApplyUpdate, "./shelly-tasmota apply-update")
	}
}

// TestSource tests Source struct.
func TestSource(t *testing.T) {
	t.Parallel()

	source := Source{
		Type:  SourceTypeGitHub,
		URL:   "https://github.com/user/repo",
		Ref:   "v1.0.0",
		Asset: "plugin-linux-amd64.tar.gz",
	}

	if source.Type != SourceTypeGitHub {
		t.Errorf("Type = %q, want %q", source.Type, SourceTypeGitHub)
	}
	if source.URL != "https://github.com/user/repo" {
		t.Errorf("URL = %q", source.URL)
	}
	if source.Ref != "v1.0.0" {
		t.Errorf("Ref = %q, want %q", source.Ref, "v1.0.0")
	}
	if source.Asset != "plugin-linux-amd64.tar.gz" {
		t.Errorf("Asset = %q", source.Asset)
	}
}

// TestBinary tests Binary struct.
func TestBinary(t *testing.T) {
	t.Parallel()

	binary := Binary{
		Name:     "shelly-tasmota",
		Checksum: "sha256:abc123",
		Platform: "linux-amd64",
		Size:     1024000,
	}

	if binary.Name != "shelly-tasmota" {
		t.Errorf("Name = %q, want %q", binary.Name, "shelly-tasmota")
	}
	if binary.Checksum != "sha256:abc123" {
		t.Errorf("Checksum = %q", binary.Checksum)
	}
	if binary.Platform != "linux-amd64" {
		t.Errorf("Platform = %q, want %q", binary.Platform, "linux-amd64")
	}
	if binary.Size != 1024000 {
		t.Errorf("Size = %d, want 1024000", binary.Size)
	}
}
