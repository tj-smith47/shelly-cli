package term

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/version"
)

func TestDisplayUpdateAvailable(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	DisplayUpdateAvailable(ios, "1.0.0", "2.0.0")

	// Warning goes to stderr, success message goes to stdout
	output := out.String() + errOut.String()
	if output == "" {
		t.Error("DisplayUpdateAvailable should produce output")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("output should contain current version")
	}
	if !strings.Contains(output, "2.0.0") {
		t.Error("output should contain available version")
	}
}

func TestDisplayUpToDate(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayUpToDate(ios)

	output := out.String()
	if output == "" {
		t.Error("DisplayUpToDate should produce output")
	}
	if !strings.Contains(output, "latest") {
		t.Error("output should mention latest version")
	}
}

func TestDisplayUpdateResult_UpdateAvailable(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	DisplayUpdateResult(ios, "1.0.0", "2.0.0", true, nil)

	// Warning goes to stderr, other messages go to stdout
	output := out.String() + errOut.String()
	if output == "" {
		t.Error("DisplayUpdateResult should produce output")
	}
	if !strings.Contains(output, "1.0.0") || !strings.Contains(output, "2.0.0") {
		t.Error("output should contain version info")
	}
}

func TestDisplayUpdateResult_UpToDate(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayUpdateResult(ios, "1.0.0", "1.0.0", false, nil)

	output := out.String()
	if output == "" {
		t.Error("DisplayUpdateResult should produce output")
	}
	if !strings.Contains(output, "latest") {
		t.Error("output should mention up to date")
	}
}

func TestDisplayUpdateResult_WithCacheError(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	cacheErr := errors.New("cache write failed")
	DisplayUpdateResult(ios, "1.0.0", "2.0.0", true, cacheErr)

	output := out.String()
	if output == "" {
		t.Error("DisplayUpdateResult should produce output even with cache error")
	}
}

func TestDisplayVersionInfo_Full(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayVersionInfo(ios, "1.0.0", "abc123", "2024-01-01", "goreleaser", "go1.21", "linux", "amd64")

	output := out.String()
	if output == "" {
		t.Error("DisplayVersionInfo should produce output")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("output should contain version")
	}
	if !strings.Contains(output, "abc123") {
		t.Error("output should contain commit")
	}
	if !strings.Contains(output, "2024-01-01") {
		t.Error("output should contain date")
	}
	if !strings.Contains(output, "goreleaser") {
		t.Error("output should contain built by")
	}
	if !strings.Contains(output, "go1.21") {
		t.Error("output should contain go version")
	}
	if !strings.Contains(output, "linux") {
		t.Error("output should contain os")
	}
	if !strings.Contains(output, "amd64") {
		t.Error("output should contain arch")
	}
}

func TestDisplayVersionInfo_Unknown(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayVersionInfo(ios, "1.0.0", "unknown", "unknown", "unknown", "go1.21", "darwin", "arm64")

	output := out.String()
	if output == "" {
		t.Error("DisplayVersionInfo should produce output")
	}
	// Should not contain "unknown" fields
	if strings.Contains(output, "commit:") {
		t.Error("output should not contain commit line for unknown")
	}
}

func TestDisplayVersionInfo_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayVersionInfo(ios, "1.0.0", "", "", "", "go1.21", "windows", "386")

	output := out.String()
	if output == "" {
		t.Error("DisplayVersionInfo should produce output")
	}
	// Should not contain empty fields
	if strings.Contains(output, "built:") {
		t.Error("output should not contain built line for empty")
	}
}

func TestRunUpdateCheck_Success(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	checker := func(_ context.Context) (*version.UpdateResult, error) {
		return &version.UpdateResult{
			CurrentVersion:  "1.0.0",
			LatestVersion:   "2.0.0",
			UpdateAvailable: true,
		}, nil
	}

	ctx := context.Background()
	RunUpdateCheck(ctx, ios, checker)

	output := out.String()
	if output == "" {
		t.Error("RunUpdateCheck should produce output")
	}
}

func TestRunUpdateCheck_Error(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	checker := func(_ context.Context) (*version.UpdateResult, error) {
		return nil, errors.New("network error")
	}

	ctx := context.Background()
	RunUpdateCheck(ctx, ios, checker)

	// Should not panic on error
	_ = out.String()
}

func TestRunUpdateCheck_DevBuild(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	checker := func(_ context.Context) (*version.UpdateResult, error) {
		return &version.UpdateResult{
			SkippedDevBuild: true,
		}, nil
	}

	ctx := context.Background()
	RunUpdateCheck(ctx, ios, checker)

	// Dev builds produce no output
	output := out.String()
	if strings.Contains(output, "Update") || strings.Contains(output, "latest") {
		t.Error("RunUpdateCheck should not produce version output for dev builds")
	}
}

func TestDisplayUpdateStatus_HasUpdate(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayUpdateStatus(ios, "1.0.0", "2.0.0", true, "https://github.com/releases/v2.0.0")

	output := out.String()
	if output == "" {
		t.Error("DisplayUpdateStatus should produce output")
	}
	if !strings.Contains(output, "Update available") {
		t.Error("output should mention update available")
	}
	if !strings.Contains(output, "github.com") {
		t.Error("output should contain release URL")
	}
}

func TestDisplayUpdateStatus_NoUpdate(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayUpdateStatus(ios, "1.0.0", "1.0.0", false, "")

	output := out.String()
	if output == "" {
		t.Error("DisplayUpdateStatus should produce output")
	}
	if !strings.Contains(output, "latest") {
		t.Error("output should mention already at latest")
	}
}
