// Package plugins provides plugin discovery, loading, and manifest management.
package plugins

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ManifestSchemaVersion is the current manifest schema version.
const ManifestSchemaVersion = "1"

// ManifestFileName is the name of the manifest file in plugin directories.
const ManifestFileName = "manifest.json"

// MigrationMarkerFile marks that migration has been completed.
const MigrationMarkerFile = ".migrated"

// Source types for plugins.
const (
	SourceTypeGitHub  = "github"
	SourceTypeURL     = "url"
	SourceTypeLocal   = "local"
	SourceTypeUnknown = "unknown"
)

// Manifest holds metadata for an installed plugin.
type Manifest struct {
	SchemaVersion        string        `json:"schema_version"`
	Name                 string        `json:"name"`
	Version              string        `json:"version,omitempty"`
	Description          string        `json:"description,omitempty"`
	InstalledAt          string        `json:"installed_at"`
	UpdatedAt            string        `json:"updated_at,omitempty"`
	Source               Source        `json:"source"`
	Binary               Binary        `json:"binary"`
	MinimumShellyVersion string        `json:"minimum_shelly_version,omitempty"`
	Capabilities         *Capabilities `json:"capabilities,omitempty"`
	Hooks                *Hooks        `json:"hooks,omitempty"`
}

// Source describes where a plugin was installed from.
type Source struct {
	Type  string `json:"type"`            // "github", "url", "local", "unknown"
	URL   string `json:"url,omitempty"`   // Full URL for github/url sources
	Ref   string `json:"ref,omitempty"`   // Git tag/commit for GitHub sources
	Asset string `json:"asset,omitempty"` // Release asset filename
	Path  string `json:"path,omitempty"`  // Original path for local sources
}

// Binary describes the plugin executable.
type Binary struct {
	Name     string `json:"name"`
	Checksum string `json:"checksum"`
	Platform string `json:"platform,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

// Capabilities defines what a plugin can do.
// All fields are optional for backward compatibility with existing plugins.
type Capabilities struct {
	// DeviceDetection indicates plugin can detect devices during discovery.
	DeviceDetection bool `json:"device_detection,omitempty"`

	// Platform is the device platform this plugin manages (e.g., "tasmota").
	Platform string `json:"platform,omitempty"`

	// Components lists controllable component types.
	// Values: "switch", "light", "cover", "sensor", "energy"
	Components []string `json:"components,omitempty"`

	// FirmwareUpdates indicates plugin supports firmware update operations.
	FirmwareUpdates bool `json:"firmware_updates,omitempty"`

	// Hints provides helpful messages for unsupported commands.
	// Keys are command names (e.g., "scene", "script", "schedule").
	// Values are user-friendly hints explaining alternatives.
	Hints map[string]string `json:"hints,omitempty"`
}

// Hooks defines executable entry points for integration.
// All fields are optional - plugins only implement hooks they support.
type Hooks struct {
	// Detect is called during discovery to probe if address is this platform.
	// Input: --address=<ip> [--auth-user=<user> --auth-pass=<pass>]
	// Output: JSON DeviceDetectionResult or exit code 1 if not this platform.
	Detect string `json:"detect,omitempty"`

	// Status returns device status.
	// Input: --address=<ip> [--auth-user=<user> --auth-pass=<pass>]
	// Output: JSON with power state, sensor readings, etc.
	Status string `json:"status,omitempty"`

	// Control executes device control commands.
	// Input: --address=<ip> --action=<on|off|toggle> --component=<switch|light> --id=<n>
	// Output: JSON with result.
	Control string `json:"control,omitempty"`

	// CheckUpdates checks for firmware updates.
	// Input: --address=<ip>
	// Output: JSON FirmwareUpdateInfo.
	CheckUpdates string `json:"check_updates,omitempty"`

	// ApplyUpdate applies firmware update.
	// Input: --address=<ip> [--url=<ota_url>] [--stage=<stable|beta>]
	// Output: JSON with success/error.
	ApplyUpdate string `json:"apply_update,omitempty"`
}

// NewManifest creates a new manifest with default values.
func NewManifest(name string, source Source) *Manifest {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Manifest{
		SchemaVersion: ManifestSchemaVersion,
		Name:          name,
		InstalledAt:   now,
		UpdatedAt:     now,
		Source:        source,
		Binary: Binary{
			Name:     PluginPrefix + name,
			Platform: runtime.GOOS + "-" + runtime.GOARCH,
		},
	}
}

// LoadManifest reads a manifest from a plugin directory.
func LoadManifest(pluginDir string) (*Manifest, error) {
	path := filepath.Join(pluginDir, ManifestFileName)
	data, err := os.ReadFile(path) //nolint:gosec // G304: pluginDir is from known plugins directory
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &m, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save(pluginDir string) error {
	path := filepath.Join(pluginDir, ManifestFileName)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec // G306: manifest is not sensitive
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// SetBinaryInfo computes and sets the binary checksum and size.
func (m *Manifest) SetBinaryInfo(binaryPath string) error {
	checksum, err := ComputeChecksum(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to compute checksum: %w", err)
	}
	m.Binary.Checksum = checksum

	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %w", err)
	}
	m.Binary.Size = info.Size()

	return nil
}

// MarkUpdated updates the UpdatedAt timestamp.
func (m *Manifest) MarkUpdated() {
	m.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

// ComputeChecksum calculates SHA256 checksum for a file.
func ComputeChecksum(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // G304: path is from known plugins directory
	if err != nil {
		return "", err
	}

	h := sha256.New()
	_, copyErr := io.Copy(h, f)
	closeErr := f.Close()

	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyChecksum verifies a file matches the manifest checksum.
func (m *Manifest) VerifyChecksum(binaryPath string) (bool, error) {
	if m.Binary.Checksum == "" {
		return false, fmt.Errorf("manifest has no checksum")
	}

	actual, err := ComputeChecksum(binaryPath)
	if err != nil {
		return false, err
	}

	return actual == m.Binary.Checksum, nil
}

// BinaryPath returns the full path to the plugin binary.
func (m *Manifest) BinaryPath(pluginDir string) string {
	return filepath.Join(pluginDir, m.Binary.Name)
}

// CanUpgrade returns whether this plugin can be automatically upgraded.
func (m *Manifest) CanUpgrade() bool {
	return m.Source.Type == SourceTypeGitHub || m.Source.Type == SourceTypeURL
}

// UpgradeMessage returns a user-friendly message about upgrade capability.
func (m *Manifest) UpgradeMessage() string {
	switch m.Source.Type {
	case SourceTypeGitHub:
		return fmt.Sprintf("Installed from GitHub: %s", m.Source.URL)
	case SourceTypeURL:
		return fmt.Sprintf("Installed from URL: %s", m.Source.URL)
	case SourceTypeLocal:
		return fmt.Sprintf("Installed from local file: %s (reinstall to upgrade)", m.Source.Path)
	case SourceTypeUnknown:
		return "Unknown source (migrated from old format, reinstall to enable auto-upgrade)"
	default:
		return "Unknown source type"
	}
}

// ParseGitHubSource creates a Source from a GitHub repo string.
func ParseGitHubSource(repoStr, tagName, assetName string) Source {
	// Remove gh: or github: prefix
	repoStr = strings.TrimPrefix(repoStr, "gh:")
	repoStr = strings.TrimPrefix(repoStr, "github:")

	return Source{
		Type:  SourceTypeGitHub,
		URL:   "https://github.com/" + repoStr,
		Ref:   tagName,
		Asset: assetName,
	}
}

// ParseURLSource creates a Source from a URL.
func ParseURLSource(url string) Source {
	return Source{
		Type: SourceTypeURL,
		URL:  url,
	}
}

// ParseLocalSource creates a Source from a local path.
func ParseLocalSource(path string) Source {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	return Source{
		Type: SourceTypeLocal,
		Path: absPath,
	}
}

// UnknownSource creates a Source for migrated plugins.
func UnknownSource() Source {
	return Source{
		Type: SourceTypeUnknown,
	}
}
