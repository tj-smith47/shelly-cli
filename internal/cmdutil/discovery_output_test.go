package cmdutil

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

type discoveredItem struct {
	ID string `json:"id"`
}

func setOutputFormat(t *testing.T, format, tmpl string) {
	t.Helper()
	viper.Set("output", format)
	viper.Set("template", tmpl)
	t.Cleanup(func() {
		viper.Set("output", "")
		viper.Set("template", "")
	})
}

//nolint:paralleltest // sets the process-global output format
func TestPrintDiscovered_Formats(t *testing.T) {
	items := []discoveredItem{{ID: "shelly1-A1B2C3"}}
	display := func(ios *iostreams.IOStreams, items []discoveredItem) {
		ios.Printf("TABLE %d\n", len(items))
	}

	tests := []struct {
		name, format, tmpl string
		items              []discoveredItem
		want               string
	}{
		{"table", "", "", items, "TABLE 1\n"},
		{"table empty", "", "", nil, "No devices found\n"},
		{"json", "json", "", items, "[\n  {\n    \"id\": \"shelly1-A1B2C3\"\n  }\n]\n"},
		{"json empty", "json", "", nil, "[]\n"},
		{"yaml", "yaml", "", items, "- id: shelly1-A1B2C3\n"},
		{"yml alias", "yml", "", items, "- id: shelly1-A1B2C3\n"},
		{"yaml empty", "yaml", "", nil, "[]\n"},
		{"template", "template", "{{range .}}{{.ID}};{{end}}", items, "shelly1-A1B2C3;"},
		{"template empty", "template", "{{len .}}", nil, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setOutputFormat(t, tt.format, tt.tmpl)
			var out, errOut bytes.Buffer
			ios := iostreams.Test(nil, &out, &errOut)

			if err := PrintDiscovered(ios, tt.items, display, "devices"); err != nil {
				t.Fatalf("PrintDiscovered: %v", err)
			}
			if got := strings.TrimRight(out.String(), "\n"); got != strings.TrimRight(tt.want, "\n") {
				t.Errorf("stdout = %q, want %q", out.String(), tt.want)
			}
			if errOut.Len() != 0 {
				t.Errorf("stderr = %q, want empty", errOut.String())
			}
		})
	}
}

//nolint:paralleltest // sets the process-global output format
func TestStatusStreams(t *testing.T) {
	for _, format := range []string{"", "table", "text"} {
		t.Run("human "+format, func(t *testing.T) {
			setOutputFormat(t, format, "")
			ios := iostreams.Test(nil, &bytes.Buffer{}, &bytes.Buffer{})
			if got := StatusStreams(ios); got != ios {
				t.Error("StatusStreams should return the command streams for human-readable output")
			}
		})
	}
	for _, format := range []string{"json", "yaml", "template"} {
		t.Run("structured "+format, func(t *testing.T) {
			setOutputFormat(t, format, "")
			var out, errOut bytes.Buffer
			status := StatusStreams(iostreams.Test(nil, &out, &errOut))

			status.Info("Scanning %d addresses", 4)
			status.NoResults("devices", "a hint")
			status.Added("device", 2)
			status.Count("device", 3)
			status.Println("line")

			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing: status lines must not mix into structured output", out.String())
			}
			for _, want := range []string{"Scanning 4 addresses", "No devices found", "a hint", "Added 2 devices", "Found 3 devices", "line"} {
				if !strings.Contains(errOut.String(), want) {
					t.Errorf("stderr is missing %q; got %q", want, errOut.String())
				}
			}
		})
	}
}

// TestRunHTTPDiscovery_StructuredStdoutIsClean checks that the scan's progress
// lines go to stderr when the caller passes the status streams.
//
//nolint:paralleltest // sets the process-global output format
func TestRunHTTPDiscovery_StructuredStdoutIsClean(t *testing.T) {
	setOutputFormat(t, "json", "")
	var out, errOut bytes.Buffer
	ios := iostreams.Test(nil, &out, &errOut)

	// The test binary refuses non-loopback connections, so the scan of this
	// documentation range finds nothing and returns at once.
	devices, err := RunHTTPDiscovery(t.Context(), StatusStreams(ios), 0, []string{"192.0.2.0/30"})
	if err != nil {
		t.Fatalf("RunHTTPDiscovery: %v", err)
	}
	if err := PrintDiscovered(ios, devices, nil, "devices"); err != nil {
		t.Fatalf("PrintDiscovered: %v", err)
	}

	var got []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not a JSON list: %v\n%s", err, out.String())
	}
	if !strings.Contains(errOut.String(), "Scanning") {
		t.Errorf("scan progress should be on stderr; got %q", errOut.String())
	}
}
