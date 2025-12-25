// Package cmd implements the plugin commands for the Tasmota plugin.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/examples/plugins/shelly-tasmota/tasmota"
)

// FirmwareUpdateInfo is returned by the check_updates hook.
// Matches the structure expected by shelly-cli.
type FirmwareUpdateInfo struct {
	CurrentVersion  string `json:"current_version"`
	LatestStable    string `json:"latest_stable,omitempty"`
	LatestBeta      string `json:"latest_beta,omitempty"`
	HasUpdate       bool   `json:"has_update"`
	HasBetaUpdate   bool   `json:"has_beta_update,omitempty"`
	OTAURLStable    string `json:"ota_url_stable,omitempty"`
	OTAURLBeta      string `json:"ota_url_beta,omitempty"`
	ChipType        string `json:"chip_type,omitempty"`
	Variant         string `json:"variant,omitempty"`
	ReleaseNotesURL string `json:"release_notes_url,omitempty"`
}

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	HTMLURL    string `json:"html_url"`
}

// Tasmota OTA URL patterns and chip types.
const (
	otaBaseURL     = "http://ota.tasmota.com"
	tasmotaRepoAPI = "https://api.github.com/repos/arendst/Tasmota/releases"

	chipESP8266 = "ESP8266"
	chipESP32   = "ESP32"
	chipESP32S2 = "ESP32-S2"
	chipESP32S3 = "ESP32-S3"
	chipESP32C3 = "ESP32-C3"
)

var checkUpdatesFlags struct {
	address  string
	authUser string
	authPass string
	timeout  time.Duration
}

// NewCheckUpdatesCmd creates the check-updates command.
func NewCheckUpdatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-updates",
		Short: "Check for firmware updates on a Tasmota device",
		Long: `Query a Tasmota device for its current firmware version and check
GitHub for the latest available releases.

Returns FirmwareUpdateInfo JSON with current version, latest stable/beta
versions, and OTA URLs for updating.`,
		Example: `  shelly-tasmota check-updates --address=192.168.1.50
  shelly-tasmota check-updates --address=192.168.1.50 --auth-user=admin --auth-pass=secret`,
		RunE: runCheckUpdates,
	}

	cmd.Flags().StringVar(&checkUpdatesFlags.address, "address", "", "Device IP address (required)")
	cmd.Flags().StringVar(&checkUpdatesFlags.authUser, "auth-user", "", "HTTP auth username")
	cmd.Flags().StringVar(&checkUpdatesFlags.authPass, "auth-pass", "", "HTTP auth password")
	cmd.Flags().DurationVar(&checkUpdatesFlags.timeout, "timeout", 15*time.Second, "Request timeout")

	if err := cmd.MarkFlagRequired("address"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to mark flag required: %v\n", err)
	}

	return cmd
}

func runCheckUpdates(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), checkUpdatesFlags.timeout)
	defer cancel()

	client := tasmota.NewClient(checkUpdatesFlags.address, checkUpdatesFlags.authUser, checkUpdatesFlags.authPass)

	// Get current firmware info from device
	fwInfo, err := client.GetStatusFirmware(ctx)
	if err != nil {
		return fmt.Errorf("failed to get firmware info: %w", err)
	}

	// Parse the current version and determine chip type
	currentVersion := parseVersion(fwInfo.Version)
	chipType := determineChipType(fwInfo.Hardware, fwInfo.Version)
	variant := determineVariant(fwInfo.Version)

	// Fetch latest releases from GitHub
	latestStable, latestBeta, err := fetchLatestReleases(ctx)
	if err != nil {
		// Return partial info even if GitHub check fails
		result := FirmwareUpdateInfo{
			CurrentVersion: currentVersion,
			ChipType:       chipType,
			Variant:        variant,
		}
		return outputJSON(result)
	}

	// Build result
	result := FirmwareUpdateInfo{
		CurrentVersion:  currentVersion,
		ChipType:        chipType,
		Variant:         variant,
		ReleaseNotesURL: "https://github.com/arendst/Tasmota/releases",
	}

	// Set stable release info
	if latestStable != nil {
		stableVer := strings.TrimPrefix(latestStable.TagName, "v")
		result.LatestStable = stableVer
		result.HasUpdate = compareVersions(currentVersion, stableVer) < 0
		result.OTAURLStable = buildOTAURL(chipType, variant, true)
		result.ReleaseNotesURL = latestStable.HTMLURL
	}

	// Set beta release info
	if latestBeta != nil {
		betaVer := strings.TrimPrefix(latestBeta.TagName, "v")
		result.LatestBeta = betaVer
		result.HasBetaUpdate = compareVersions(currentVersion, betaVer) < 0
		result.OTAURLBeta = buildOTAURL(chipType, variant, false)
	}

	return outputJSON(result)
}

// parseVersion extracts the version number from Tasmota's version string.
// Example: "14.3.0(tasmota)" becomes "14.3.0".
func parseVersion(fullVersion string) string {
	// Remove parenthetical suffix like "(tasmota)" or "(tasmota-lite)"
	if idx := strings.Index(fullVersion, "("); idx > 0 {
		return strings.TrimSpace(fullVersion[:idx])
	}
	return fullVersion
}

// determineChipType identifies the chip type from hardware info.
// Returns one of the chip* constants.
func determineChipType(hardware, version string) string {
	hw := strings.ToUpper(hardware)
	ver := strings.ToLower(version)

	// Check version string for ESP32 variants
	switch {
	case strings.Contains(ver, "tasmota32s3"):
		return chipESP32S3
	case strings.Contains(ver, "tasmota32s2"):
		return chipESP32S2
	case strings.Contains(ver, "tasmota32c3"), strings.Contains(ver, "tasmota32c2"):
		return chipESP32C3
	case strings.Contains(ver, "tasmota32"):
		return chipESP32
	}

	// Check hardware string
	switch {
	case strings.Contains(hw, chipESP32S3):
		return chipESP32S3
	case strings.Contains(hw, chipESP32S2):
		return chipESP32S2
	case strings.Contains(hw, "ESP32-C3"), strings.Contains(hw, "ESP32-C2"):
		return chipESP32C3
	case strings.Contains(hw, chipESP32):
		return chipESP32
	case strings.Contains(hw, chipESP8266), strings.Contains(hw, "ESP8285"):
		return chipESP8266
	}

	// Default to ESP8266 (most common)
	return chipESP8266
}

// determineVariant extracts the Tasmota variant from the version string.
// Examples: "tasmota", "tasmota-lite", "tasmota-sensors", "tasmota32", etc.
func determineVariant(version string) string {
	lower := strings.ToLower(version)

	// Extract variant from parenthetical suffix
	re := regexp.MustCompile(`\(([^)]+)\)`)
	if matches := re.FindStringSubmatch(lower); len(matches) > 1 {
		return matches[1]
	}

	// Check for known variants in version string
	variants := []string{
		"tasmota32-bluetooth", "tasmota32-display", "tasmota32-ir",
		"tasmota32-lvgl", "tasmota32-webcam", "tasmota32",
		"tasmota-sensors", "tasmota-display", "tasmota-ir",
		"tasmota-lite", "tasmota-minimal", "tasmota-knx",
		"tasmota-zbbridge", "tasmota",
	}

	for _, v := range variants {
		if strings.Contains(lower, v) {
			return v
		}
	}

	return "tasmota"
}

// fetchLatestReleases fetches the latest stable and beta releases from GitHub.
func fetchLatestReleases(ctx context.Context) (stable, beta *GitHubRelease, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tasmotaRepoAPI, http.NoBody)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "shelly-tasmota-plugin/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // Best effort close

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, nil, err
	}

	var latestStable, latestBeta *GitHubRelease
	for i := range releases {
		r := &releases[i]
		if r.Prerelease {
			if latestBeta == nil {
				latestBeta = r
			}
		} else {
			if latestStable == nil {
				latestStable = r
			}
		}
		// Stop once we have both
		if latestStable != nil && latestBeta != nil {
			break
		}
	}

	return latestStable, latestBeta, nil
}

// buildOTAURL constructs the OTA URL for a given chip type and variant.
func buildOTAURL(chipType, variant string, stable bool) string {
	// Base path depends on chip type
	var basePath string
	switch chipType {
	case chipESP32S2:
		basePath = "/tasmota32s2"
	case chipESP32S3:
		basePath = "/tasmota32s3"
	case chipESP32C3:
		basePath = "/tasmota32c3"
	case chipESP32:
		basePath = "/tasmota32"
	default:
		basePath = "/tasmota"
	}

	// Add /release for stable builds
	if stable {
		basePath += "/release"
	}

	// Determine filename based on variant
	filename := variantToFilename(variant, chipType)

	return otaBaseURL + basePath + "/" + filename
}

// variantFilenames maps variant names to their OTA filenames (without extension).
var variantFilenames = map[string]string{
	"tasmota-lite":        "tasmota-lite",
	"tasmota-minimal":     "tasmota-minimal",
	"tasmota-sensors":     "tasmota-sensors",
	"tasmota-display":     "tasmota-display",
	"tasmota-ir":          "tasmota-ir",
	"tasmota-knx":         "tasmota-knx",
	"tasmota-zbbridge":    "tasmota-zbbridge",
	"tasmota32-bluetooth": "tasmota32-bluetooth",
	"tasmota32-display":   "tasmota32-display",
	"tasmota32-ir":        "tasmota32-ir",
	"tasmota32-lvgl":      "tasmota32-lvgl",
	"tasmota32-webcam":    "tasmota32-webcam",
	"tasmota32":           "tasmota32",
}

// variantToFilename maps a variant to its OTA filename.
func variantToFilename(variant, chipType string) string {
	isESP32 := strings.HasPrefix(chipType, "ESP32")

	// Look up the base filename
	baseName, ok := variantFilenames[variant]
	if !ok {
		// Default to base tasmota variant
		if isESP32 {
			baseName = "tasmota32"
		} else {
			baseName = "tasmota"
		}
	}

	// Add extension based on chip type
	if isESP32 {
		return baseName + ".bin"
	}
	return baseName + ".bin.gz"
}

// compareVersions compares two semantic version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	aParts := parseVersionParts(a)
	bParts := parseVersionParts(b)

	// Compare each part
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var aVal, bVal int
		if i < len(aParts) {
			aVal = aParts[i]
		}
		if i < len(bParts) {
			bVal = bParts[i]
		}

		if aVal < bVal {
			return -1
		}
		if aVal > bVal {
			return 1
		}
	}

	return 0
}

// parseVersionParts splits a version string into integer parts.
func parseVersionParts(version string) []int {
	// Remove any suffix after the version number
	version = strings.Split(version, "-")[0]
	version = strings.Split(version, "+")[0]

	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))

	for _, p := range parts {
		var val int
		if _, err := fmt.Sscanf(p, "%d", &val); err == nil {
			result = append(result, val)
		}
	}

	return result
}
