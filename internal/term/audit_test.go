package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayAuditResult_Reachable(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &model.AuditResult{
		Device:    "kitchen-light",
		Address:   "192.168.1.100",
		Reachable: true,
		Issues:    []string{"No authentication enabled", "Firmware outdated"},
		Warnings:  []string{"Cloud not configured"},
		InfoItems: []string{"HTTPS enabled", "Local access available"},
	}
	DisplayAuditResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "kitchen-light") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "192.168.1.100") {
		t.Error("expected device address")
	}
	if !strings.Contains(output, "No authentication enabled") {
		t.Error("expected issue")
	}
	if !strings.Contains(output, "Firmware outdated") {
		t.Error("expected second issue")
	}
	if !strings.Contains(output, "Cloud not configured") {
		t.Error("expected warning")
	}
	if !strings.Contains(output, "HTTPS enabled") {
		t.Error("expected info item")
	}
	if !strings.Contains(output, "Local access available") {
		t.Error("expected second info item")
	}
}

func TestDisplayAuditResult_Unreachable(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &model.AuditResult{
		Device:    "offline-device",
		Address:   "192.168.1.50",
		Reachable: false,
	}
	DisplayAuditResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "offline-device") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "unreachable") {
		t.Error("expected unreachable message")
	}
	if !strings.Contains(output, "cannot audit") {
		t.Error("expected cannot audit message")
	}
}

func TestDisplayAuditResult_NoIssues(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &model.AuditResult{
		Device:    "secure-device",
		Address:   "192.168.1.200",
		Reachable: true,
		Issues:    []string{},
		Warnings:  []string{},
		InfoItems: []string{"All checks passed"},
	}
	DisplayAuditResult(ios, result)

	output := out.String()
	if !strings.Contains(output, "secure-device") {
		t.Error("expected device name")
	}
	if !strings.Contains(output, "All checks passed") {
		t.Error("expected info item")
	}
}
