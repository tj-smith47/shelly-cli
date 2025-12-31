package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestDisplayDeviceAliases_NoAliases(t *testing.T) {
	t.Parallel()

	var stdin, stdout, stderr bytes.Buffer
	ios := iostreams.Test(&stdin, &stdout, &stderr)

	DisplayDeviceAliases(ios, "test-device", nil)

	output := stdout.String()
	if output == "" {
		t.Error("expected output for no aliases")
	}
	if !strings.Contains(output, "No aliases") {
		t.Error("expected 'No aliases' message")
	}
}

func TestDisplayDeviceAliases_WithAliases(t *testing.T) {
	t.Parallel()

	var stdin, stdout, stderr bytes.Buffer
	ios := iostreams.Test(&stdin, &stdout, &stderr)

	DisplayDeviceAliases(ios, "test-device", []string{"alias1", "alias2"})

	output := stdout.String()
	if output == "" {
		t.Error("expected output for aliases")
	}
	if !strings.Contains(output, "alias1") {
		t.Error("expected 'alias1' in output")
	}
	if !strings.Contains(output, "alias2") {
		t.Error("expected 'alias2' in output")
	}
}

func TestDisplayAliasAdded(t *testing.T) {
	t.Parallel()

	var stdin, stdout, stderr bytes.Buffer
	ios := iostreams.Test(&stdin, &stdout, &stderr)

	DisplayAliasAdded(ios, "test-device", "new-alias")

	output := stdout.String()
	if output == "" {
		t.Error("expected success message")
	}
	if !strings.Contains(output, "Added") {
		t.Error("expected 'Added' in output")
	}
	if !strings.Contains(output, "new-alias") {
		t.Error("expected alias name in output")
	}
}

func TestDisplayAliasRemoved(t *testing.T) {
	t.Parallel()

	var stdin, stdout, stderr bytes.Buffer
	ios := iostreams.Test(&stdin, &stdout, &stderr)

	DisplayAliasRemoved(ios, "test-device", "old-alias")

	output := stdout.String()
	if output == "" {
		t.Error("expected success message")
	}
	if !strings.Contains(output, "Removed") {
		t.Error("expected 'Removed' in output")
	}
	if !strings.Contains(output, "old-alias") {
		t.Error("expected alias name in output")
	}
}
