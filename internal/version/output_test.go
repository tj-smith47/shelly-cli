package version

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

func TestNewOutput(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   "1.0.0",
		Commit:    "abc123",
		Date:      "2024-01-01",
		BuiltBy:   "test",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	output := NewOutput(info)

	if output.Version != info.Version {
		t.Errorf("Version = %q, want %q", output.Version, info.Version)
	}
	if output.Commit != info.Commit {
		t.Errorf("Commit = %q, want %q", output.Commit, info.Commit)
	}
	if output.Date != info.Date {
		t.Errorf("Date = %q, want %q", output.Date, info.Date)
	}
	if output.BuiltBy != info.BuiltBy {
		t.Errorf("BuiltBy = %q, want %q", output.BuiltBy, info.BuiltBy)
	}
	if output.GoVersion != info.GoVersion {
		t.Errorf("GoVersion = %q, want %q", output.GoVersion, info.GoVersion)
	}
	if output.OS != info.OS {
		t.Errorf("OS = %q, want %q", output.OS, info.OS)
	}
	if output.Arch != info.Arch {
		t.Errorf("Arch = %q, want %q", output.Arch, info.Arch)
	}
	if output.UpdateAvail != nil {
		t.Errorf("UpdateAvail should be nil initially, got %v", output.UpdateAvail)
	}
	if output.LatestVersion != nil {
		t.Errorf("LatestVersion should be nil initially, got %v", output.LatestVersion)
	}
}

func TestSetUpdateInfo_UpdateAvailable(t *testing.T) {
	t.Parallel()

	output := &Output{Version: testVersion1}
	output.SetUpdateInfo(testVersion2, true)

	if output.LatestVersion == nil || *output.LatestVersion != testVersion2 {
		t.Errorf("LatestVersion = %v, want %q", output.LatestVersion, testVersion2)
	}
	if output.UpdateAvail == nil || *output.UpdateAvail != availabilityYes {
		t.Errorf("UpdateAvail = %v, want %q", output.UpdateAvail, availabilityYes)
	}
}

func TestSetUpdateInfo_NoUpdate(t *testing.T) {
	t.Parallel()

	output := &Output{Version: testVersion2}
	output.SetUpdateInfo(testVersion2, false)

	if output.LatestVersion == nil || *output.LatestVersion != testVersion2 {
		t.Errorf("LatestVersion = %v, want %q", output.LatestVersion, testVersion2)
	}
	if output.UpdateAvail == nil || *output.UpdateAvail != "no" {
		t.Errorf("UpdateAvail = %v, want 'no'", output.UpdateAvail)
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	output := &Output{
		Version:   "1.2.3",
		Commit:    "def456",
		Date:      "2024-06-01",
		BuiltBy:   "goreleaser",
		GoVersion: "go1.22.0",
		OS:        "darwin",
		Arch:      "arm64",
	}

	var buf bytes.Buffer
	err := output.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Verify the output is valid JSON
	var decoded Output
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded.Version != output.Version {
		t.Errorf("decoded.Version = %q, want %q", decoded.Version, output.Version)
	}
}

func TestWriteJSON_WithUpdateInfo(t *testing.T) {
	t.Parallel()

	output := &Output{
		Version: "1.0.0",
	}
	output.SetUpdateInfo(testVersion2, true)

	var buf bytes.Buffer
	err := output.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Verify the output is valid JSON with update info
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded["update_available"] != availabilityYes {
		t.Errorf("update_available = %v, want %q", decoded["update_available"], availabilityYes)
	}
	if decoded["latest_version"] != testVersion2 {
		t.Errorf("latest_version = %v, want %q", decoded["latest_version"], testVersion2)
	}
}

func TestWriteJSONOutput_NoUpdateCheck(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   "1.0.0",
		Commit:    "abc",
		Date:      "2024-01-01",
		BuiltBy:   "test",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	var buf bytes.Buffer
	err := WriteJSONOutput(context.Background(), &buf, info, false, nil, nil)
	if err != nil {
		t.Fatalf("WriteJSONOutput() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded["version"] != "1.0.0" {
		t.Errorf("version = %v, want '1.0.0'", decoded["version"])
	}
	// Should not have update info
	if _, exists := decoded["update_available"]; exists {
		t.Error("update_available should not be present when checkUpdate is false")
	}
}

func TestWriteJSONOutput_WithUpdateCheck(t *testing.T) {
	// Use memory filesystem to prevent writes to real cache
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	info := Info{
		Version:   "1.0.0",
		Commit:    "abc",
		Date:      "2024-01-01",
		BuiltBy:   "test",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	// Mock fetcher that returns a newer version
	fetcher := func(_ context.Context) (string, error) {
		return testVersion2, nil
	}

	// Mock isNewer function
	isNewer := func(current, latest string) bool {
		return latest > current
	}

	var buf bytes.Buffer
	err := WriteJSONOutput(context.Background(), &buf, info, true, fetcher, isNewer)
	if err != nil {
		t.Fatalf("WriteJSONOutput() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded["update_available"] != availabilityYes {
		t.Errorf("update_available = %v, want %q", decoded["update_available"], availabilityYes)
	}
}

func TestWriteJSONOutput_FetcherError(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   "1.0.0",
		Commit:    "abc",
		Date:      "2024-01-01",
		BuiltBy:   "test",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	// Mock fetcher that returns an error
	fetcher := func(_ context.Context) (string, error) {
		return "", errors.New("network error")
	}

	var buf bytes.Buffer
	// Should still succeed, just without update info
	err := WriteJSONOutput(context.Background(), &buf, info, true, fetcher, nil)
	if err != nil {
		t.Fatalf("WriteJSONOutput() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Should not have update info when fetcher fails
	if _, exists := decoded["update_available"]; exists {
		t.Error("update_available should not be present when fetcher fails")
	}
}

func TestWriteJSONOutput_DevBuild(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   "dev",
		Commit:    "abc",
		Date:      "2024-01-01",
		BuiltBy:   "test",
		GoVersion: "go1.21.0",
		OS:        "linux",
		Arch:      "amd64",
	}

	// Mock fetcher that returns a newer version
	fetcher := func(_ context.Context) (string, error) {
		return testVersion2, nil
	}

	var buf bytes.Buffer
	err := WriteJSONOutput(context.Background(), &buf, info, true, fetcher, nil)
	if err != nil {
		t.Fatalf("WriteJSONOutput() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Should not have update info for dev builds
	if _, exists := decoded["update_available"]; exists {
		t.Error("update_available should not be present for dev builds")
	}
}
