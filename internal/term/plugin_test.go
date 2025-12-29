package term

import (
	"errors"
	"strings"
	"testing"
)

func TestDisplayPluginUpgradeResults_AllUpgraded(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	results := []PluginUpgradeResult{
		{Name: "plugin1", OldVersion: "1.0.0", NewVersion: "2.0.0", Upgraded: true},
		{Name: "plugin2", OldVersion: "1.5.0", NewVersion: "2.0.0", Upgraded: true},
	}
	DisplayPluginUpgradeResults(ios, results)

	output := out.String()
	if !strings.Contains(output, "plugin1") {
		t.Error("expected plugin1 name")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("expected old version")
	}
	if !strings.Contains(output, "2.0.0") {
		t.Error("expected new version")
	}
	if !strings.Contains(output, "2 upgraded") {
		t.Error("expected upgraded count")
	}
	if !strings.Contains(output, "0 skipped") {
		t.Error("expected skipped count")
	}
	if !strings.Contains(output, "0 failed") {
		t.Error("expected failed count")
	}
}

func TestDisplayPluginUpgradeResults_MixedResults(t *testing.T) {
	t.Parallel()

	ios, out, errOut := testIOStreams()
	results := []PluginUpgradeResult{
		{Name: "upgraded", OldVersion: "1.0", NewVersion: "2.0", Upgraded: true},
		{Name: "skipped", OldVersion: "1.0", Skipped: true, Error: errors.New("network unavailable")},
		{Name: "failed", OldVersion: "1.0", Error: errors.New("download failed")},
		{Name: "current", OldVersion: "2.0", Upgraded: false},
	}
	DisplayPluginUpgradeResults(ios, results)

	// Success messages go to stdout
	if !strings.Contains(out.String(), "Upgraded from") {
		t.Error("expected upgrade message")
	}
	// Warning (skipped) goes to stderr
	if !strings.Contains(errOut.String(), "Skipped") {
		t.Error("expected skipped message")
	}
	// Error (failed) goes to stderr
	if !strings.Contains(errOut.String(), "Failed") {
		t.Error("expected failed message")
	}
	// Info (already latest) goes to stdout
	if !strings.Contains(out.String(), "Already at latest") {
		t.Error("expected already latest message")
	}
	// Summary goes to stdout
	if !strings.Contains(out.String(), "1 upgraded") {
		t.Error("expected 1 upgraded")
	}
	if !strings.Contains(out.String(), "1 skipped") {
		t.Error("expected 1 skipped")
	}
	if !strings.Contains(out.String(), "1 failed") {
		t.Error("expected 1 failed")
	}
}

func TestDisplayPluginUpgradeResult_Upgraded(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := PluginUpgradeResult{
		Name:       "my-plugin",
		OldVersion: "1.0.0",
		NewVersion: "1.1.0",
		Upgraded:   true,
	}
	DisplayPluginUpgradeResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "Upgraded my-plugin") {
		t.Error("expected upgrade message")
	}
	if !strings.Contains(output, "from 1.0.0 to 1.1.0") {
		t.Error("expected version info")
	}
}

func TestDisplayPluginUpgradeResult_Failed(t *testing.T) {
	t.Parallel()

	ios, _, errOut := testIOStreams()
	result := PluginUpgradeResult{
		Name:  "broken-plugin",
		Error: errors.New("connection refused"),
	}
	DisplayPluginUpgradeResult(ios, result)

	output := errOut.String()
	if !strings.Contains(output, "Failed to upgrade broken-plugin") {
		t.Error("expected failure message")
	}
	if !strings.Contains(output, "connection refused") {
		t.Error("expected error details")
	}
}

func TestDisplayPluginUpgradeResult_AlreadyLatest(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := PluginUpgradeResult{
		Name:       "current-plugin",
		OldVersion: "2.0.0",
		Upgraded:   false,
	}
	DisplayPluginUpgradeResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "already at latest version") {
		t.Error("expected already latest message")
	}
	if !strings.Contains(output, "2.0.0") {
		t.Error("expected version")
	}
}
