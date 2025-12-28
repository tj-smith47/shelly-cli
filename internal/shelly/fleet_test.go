package shelly

import (
	"bytes"
	"testing"

	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

func TestIntegratorCredentials_Fields(t *testing.T) {
	t.Parallel()

	creds := IntegratorCredentials{
		Tag:   "test-tag",
		Token: "test-token",
	}

	if creds.Tag != "test-tag" {
		t.Errorf("expected Tag 'test-tag', got %q", creds.Tag)
	}
	if creds.Token != "test-token" {
		t.Errorf("expected Token 'test-token', got %q", creds.Token)
	}
}

func TestRelayAction_Constants(t *testing.T) {
	t.Parallel()

	if RelayOn != "on" {
		t.Errorf("expected RelayOn 'on', got %q", RelayOn)
	}
	if RelayOff != "off" {
		t.Errorf("expected RelayOff 'off', got %q", RelayOff)
	}
	if RelayToggle != "toggle" {
		t.Errorf("expected RelayToggle 'toggle', got %q", RelayToggle)
	}
}

func TestCapitalizeFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"a", "A"},
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"HELLO", "HELLO"},
		{"123abc", "123abc"},
		{"turn on", "Turn on"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			got := capitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReportBatchResults(t *testing.T) {
	t.Parallel()

	t.Run("all success", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		ios := iostreams.Test(nil, out, errOut)

		results := []integrator.BatchResult{
			{DeviceID: "dev1", Success: true},
			{DeviceID: "dev2", Success: true},
		}

		err := ReportBatchResults(ios, results, "turned on")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := out.String() + errOut.String()
		if output == "" {
			t.Error("expected some output")
		}
	})

	t.Run("some failures", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		ios := iostreams.Test(nil, out, errOut)

		results := []integrator.BatchResult{
			{DeviceID: "dev1", Success: true},
			{DeviceID: "dev2", Success: false, Error: "connection failed"},
		}

		err := ReportBatchResults(ios, results, "turned on")

		if err == nil {
			t.Error("expected error for partial failure")
		}
	})

	t.Run("all failures", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		ios := iostreams.Test(nil, out, errOut)

		results := []integrator.BatchResult{
			{DeviceID: "dev1", Success: false, Error: "error 1"},
			{DeviceID: "dev2", Success: false, Error: "error 2"},
		}

		err := ReportBatchResults(ios, results, "turned on")

		if err == nil {
			t.Error("expected error for all failures")
		}
	})

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}
		ios := iostreams.Test(nil, out, errOut)

		results := []integrator.BatchResult{}

		err := ReportBatchResults(ios, results, "turned on")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		output := out.String() + errOut.String()
		if output == "" {
			t.Error("expected warning about no devices")
		}
	})
}

func TestFleetConnection_Fields(t *testing.T) {
	t.Parallel()

	// Test that FleetConnection can hold references
	fc := &FleetConnection{
		Client:  nil,
		Manager: nil,
		ios:     nil,
	}

	if fc.Client != nil {
		t.Error("expected Client to be nil")
	}
	if fc.Manager != nil {
		t.Error("expected Manager to be nil")
	}
}
