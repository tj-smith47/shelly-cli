package term

import (
	"strings"
	"testing"
)

func TestDisplayDoctorHeader(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDoctorHeader(ios)

	output := out.String()
	if !strings.Contains(output, "Shelly CLI Doctor") {
		t.Error("expected header title")
	}
}

func TestDisplayDoctorSummary_NoIssues(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDoctorSummary(ios, 0, 0)

	output := out.String()
	if !strings.Contains(output, "No issues found") {
		t.Error("expected no issues message")
	}
	if !strings.Contains(output, "healthy") {
		t.Error("expected healthy message")
	}
}

func TestDisplayDoctorSummary_OnlyWarnings(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDoctorSummary(ios, 0, 3)

	output := out.String()
	if !strings.Contains(output, "No critical issues found") {
		t.Error("expected no critical issues message")
	}
	if !strings.Contains(output, "3 warning") {
		t.Error("expected warning count")
	}
}

func TestDisplayDoctorSummary_WithIssues(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDoctorSummary(ios, 2, 1)

	output := out.String()
	if !strings.Contains(output, "2 issue") {
		t.Error("expected issue count")
	}
	if !strings.Contains(output, "1 additional warning") {
		t.Error("expected warning count")
	}
}

func TestDisplayDoctorSummary_OnlyIssues(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayDoctorSummary(ios, 5, 0)

	output := out.String()
	if !strings.Contains(output, "5 issue") {
		t.Error("expected issue count")
	}
	if !strings.Contains(output, "shelly doctor --full") {
		t.Error("expected full command hint")
	}
}

func TestWarnStdout(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	WarnStdout(ios, "Test warning: %s", "details")

	output := out.String()
	if !strings.Contains(output, "Test warning: details") {
		t.Error("expected warning message")
	}
	// Should contain warning symbol
	if !strings.Contains(output, "⚠") {
		t.Error("expected warning symbol")
	}
}

func TestCheckCLIVersion_DevBuild(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	// This will detect dev build since version package won't have real version set
	issues, warnings := CheckCLIVersion(ios)

	output := out.String()
	if !strings.Contains(output, "CLI Version") {
		t.Error("expected header")
	}
	// Should handle gracefully regardless of version state
	_ = issues
	_ = warnings
	_ = output
}
