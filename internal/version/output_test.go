package version

import (
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
		Version:   testVersion1,
		Commit:    "abc123",
		Date:      testDate,
		BuiltBy:   testBuiltBy,
		GoVersion: testGoVersion,
		OS:        testOS,
		Arch:      testArch,
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

// keyVersion is the JSON key of the version number.
const keyVersion = "version"

func TestOutput_JSONKeys(t *testing.T) {
	t.Parallel()

	output := &Output{Version: testVersion1}
	output.SetUpdateInfo(testVersion2, true)
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	for _, key := range []string{keyVersion, "commit", "date", "built_by", "go_version", "os", "arch"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("JSON has no %q key: %s", key, data)
		}
	}
	if decoded["update_available"] != availabilityYes || decoded["latest_version"] != testVersion2 {
		t.Errorf("update keys = %v, %v; want %q, %q", decoded["update_available"], decoded["latest_version"], availabilityYes, testVersion2)
	}
}

func TestBuildOutput_NoUpdateCheck(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   testVersion1,
		Commit:    testCommitAbc,
		Date:      testDate,
		BuiltBy:   testBuiltBy,
		GoVersion: testGoVersion,
		OS:        testOS,
		Arch:      testArch,
	}

	out := BuildOutput(context.Background(), info, false, nil, nil)

	if out.Version != testVersion1 {
		t.Errorf("Version = %q, want %q", out.Version, testVersion1)
	}
	// Should not have update info
	if out.UpdateAvail != nil {
		t.Error("update_available should not be present when checkUpdate is false")
	}
}

func TestBuildOutput_WithUpdateCheck(t *testing.T) {
	// Use memory filesystem to prevent writes to real cache
	config.SetFs(afero.NewMemMapFs())
	t.Cleanup(func() { config.SetFs(nil) })
	t.Setenv("HOME", "/test/home")

	info := Info{
		Version:   testVersion1,
		Commit:    testCommitAbc,
		Date:      testDate,
		BuiltBy:   testBuiltBy,
		GoVersion: testGoVersion,
		OS:        testOS,
		Arch:      testArch,
	}

	// Mock fetcher that returns a newer version
	fetcher := func(_ context.Context) (string, error) {
		return testVersion2, nil
	}

	// Mock isNewer function
	isNewer := func(current, latest string) bool {
		return latest > current
	}

	out := BuildOutput(context.Background(), info, true, fetcher, isNewer)

	if out.UpdateAvail == nil || *out.UpdateAvail != availabilityYes {
		t.Errorf("UpdateAvail = %v, want %q", out.UpdateAvail, availabilityYes)
	}
}

func TestBuildOutput_FetcherError(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   testVersion1,
		Commit:    testCommitAbc,
		Date:      testDate,
		BuiltBy:   testBuiltBy,
		GoVersion: testGoVersion,
		OS:        testOS,
		Arch:      testArch,
	}

	// Mock fetcher that returns an error
	fetcher := func(_ context.Context) (string, error) {
		return "", errors.New("network error")
	}

	// Should still succeed, just without update info
	out := BuildOutput(context.Background(), info, true, fetcher, nil)

	// Should not have update info when fetcher fails
	if out.UpdateAvail != nil {
		t.Error("update_available should not be present when fetcher fails")
	}
}

func TestBuildOutput_DevBuild(t *testing.T) {
	t.Parallel()

	info := Info{
		Version:   DevVersion,
		Commit:    testCommitAbc,
		Date:      testDate,
		BuiltBy:   testBuiltBy,
		GoVersion: testGoVersion,
		OS:        testOS,
		Arch:      testArch,
	}

	// Mock fetcher that returns a newer version
	fetcher := func(_ context.Context) (string, error) {
		return testVersion2, nil
	}

	out := BuildOutput(context.Background(), info, true, fetcher, nil)

	// Should not have update info for dev builds
	if out.UpdateAvail != nil {
		t.Error("update_available should not be present for dev builds")
	}
}
