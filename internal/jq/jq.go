// Package jq provides built-in jq filtering for CLI output.
// It uses gojq (pure Go jq implementation) to apply jq expressions
// to structured command output, eliminating the need for external jq.
package jq

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/output/synfmt"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// HasFields returns true if the --fields flag is set.
func HasFields() bool {
	return viper.GetBool("fields")
}

// HasFilter returns true if a jq filter expression is configured.
// Supports both single string and repeated --jq flags (StringArray).
func HasFilter() bool {
	filters := viper.GetStringSlice("jq")
	return len(filters) > 0
}

// GetFilter returns the configured jq filter expression.
// Multiple --jq flags are joined with " | " to form a pipeline.
func GetFilter() string {
	filters := viper.GetStringSlice("jq")
	return strings.Join(filters, " | ")
}

// Apply compiles and runs a jq expression against data, writing results to w.
// Data is first marshaled to JSON then unmarshaled to ensure gojq-compatible types.
// Each result is pretty-printed as JSON with syntax highlighting on TTY.
func Apply(w io.Writer, data any, expr string) error {
	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid jq expression: %w", err)
	}

	// Convert data to gojq-compatible types via JSON round-trip
	input, err := toJQInput(data)
	if err != nil {
		return fmt.Errorf("failed to convert data for jq: %w", err)
	}

	colorize := synfmt.ShouldHighlight()

	iter := query.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return fmt.Errorf("jq error: %w", err)
		}

		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode jq result: %w", err)
		}

		output := string(b)
		if colorize {
			output = theme.HighlightJSON(output)
		}

		if _, err := fmt.Fprintln(w, output); err != nil {
			return fmt.Errorf("failed to write jq result: %w", err)
		}
	}

	return nil
}

// toJQInput converts arbitrary Go data to gojq-compatible types (map[string]any, []any, etc.)
// by round-tripping through JSON.
func toJQInput(data any) (any, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PrintFields introspects data and prints available field names for use with --jq and --template.
func PrintFields(w io.Writer, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to inspect fields: %w", err)
	}

	var generic any
	if err := json.Unmarshal(b, &generic); err != nil {
		return fmt.Errorf("failed to inspect fields: %w", err)
	}

	// If it's an array, use the first element as the representative
	if arr, ok := generic.([]any); ok && len(arr) > 0 {
		generic = arr[0]
		if _, err := fmt.Fprintln(w, "(array element fields)"); err != nil {
			return err
		}
	}

	obj, ok := generic.(map[string]any)
	if !ok {
		_, err := fmt.Fprintln(w, "(scalar value, no fields available)")
		return err
	}

	var fields []string
	collectFields(obj, "", &fields)
	sort.Strings(fields)
	for _, f := range fields {
		if _, err := fmt.Fprintln(w, f); err != nil {
			return err
		}
	}
	return nil
}

// collectFields recursively collects dot-separated field paths from a map.
func collectFields(obj map[string]any, prefix string, fields *[]string) {
	for k, v := range obj {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}
		*fields = append(*fields, path)
		switch val := v.(type) {
		case map[string]any:
			collectFields(val, path, fields)
		case []any:
			if len(val) > 0 {
				if elem, ok := val[0].(map[string]any); ok {
					collectFields(elem, path+"[]", fields)
				}
			}
		}
	}
}
