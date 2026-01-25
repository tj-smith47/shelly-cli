// Package term provides terminal display functions for the CLI.
package term

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
)

// configRow holds a row with variable-depth keys and a value.
type configRow struct {
	keys  []string // hierarchical keys (e.g., ["defaults", "timeout"] or ["theme", "colors", "primary"])
	value string
}

// DisplayConfigTable prints a configuration map as a single consolidated table.
// Each nesting level gets its own column. Parent keys are vertically centered in their spans.
// Row lines appear between groups at any nesting level.
func DisplayConfigTable(ios *iostreams.IOStreams, configData any) error {
	configMap, ok := configData.(map[string]any)
	if !ok {
		return output.PrintJSON(configData)
	}

	// Collect all rows with their full key paths
	var rows []configRow
	collectConfigRows(configMap, nil, &rows)

	if len(rows) == 0 {
		ios.Info("No configuration settings found")
		return nil
	}

	maxDepth := configRowsMaxDepth(rows)
	sortConfigRows(rows)

	// Calculate which rows should display each key (vertically centered in spans)
	showKeyAt := calcCenteredKeyPositions(rows, maxDepth)

	// Calculate where separators should go (when any parent key changes)
	separatorRows := calcSeparatorPositions(rows)

	// Build the table
	builder := table.NewBuilder(configTableHeaders(maxDepth)...)

	for rowIdx, r := range rows {
		if startCol, hasSep := separatorRows[rowIdx]; hasSep {
			builder.AddSeparatorAt(startCol)
		}

		cells := make([]string, maxDepth+1)
		for col := range maxDepth {
			if showKeyAt[rowIdx][col] {
				cells[col] = r.keys[col]
			}
		}
		cells[maxDepth] = r.value

		builder.AddRow(cells...)
	}

	tbl := builder.WithModeStyle(ios).MergeEmptyHeaders().Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print config table", err)
	}
	ios.Println()

	return nil
}

// calcCenteredKeyPositions determines which row should display each key at each depth.
// Keys are vertically centered within their span of consecutive rows.
func calcCenteredKeyPositions(rows []configRow, maxDepth int) [][]bool {
	result := make([][]bool, len(rows))
	for i := range result {
		result[i] = make([]bool, maxDepth)
	}

	// For each column/depth, find spans and mark the middle row
	for col := range maxDepth {
		spanStart := 0
		var currentKey string

		for i, r := range rows {
			key := ""
			if col < len(r.keys) {
				key = r.keys[col]
			}

			// Check if parent keys match (a key only spans if all parent keys match)
			parentMatch := true
			for pc := range col {
				prevKey := ""
				if pc < len(rows[spanStart].keys) {
					prevKey = rows[spanStart].keys[pc]
				}
				currKey := ""
				if pc < len(r.keys) {
					currKey = r.keys[pc]
				}
				if prevKey != currKey {
					parentMatch = false
					break
				}
			}

			// If key changed or parent changed, close previous span and start new one
			if key != currentKey || !parentMatch {
				if currentKey != "" {
					// Mark the middle row of the previous span
					middleIdx := spanStart + (i-spanStart-1)/2
					result[middleIdx][col] = true
				}
				spanStart = i
				currentKey = key
			}
		}

		// Close final span
		if currentKey != "" {
			middleIdx := spanStart + (len(rows)-spanStart-1)/2
			result[middleIdx][col] = true
		}
	}

	return result
}

// calcSeparatorPositions determines where separator lines should appear.
// Returns a map from row index to starting column for the separator.
// A separator appears when any key at any depth changes, starting at that column.
func calcSeparatorPositions(rows []configRow) map[int]int {
	separators := make(map[int]int)

	for i := 1; i < len(rows); i++ {
		prev := rows[i-1]
		curr := rows[i]

		// Find the first column where keys differ
		maxLen := max(len(prev.keys), len(curr.keys))
		for col := range maxLen {
			prevKey := ""
			if col < len(prev.keys) {
				prevKey = prev.keys[col]
			}
			currKey := ""
			if col < len(curr.keys) {
				currKey = curr.keys[col]
			}

			if prevKey != currKey {
				// Keys differ at this column, separator starts here
				separators[i] = col
				break
			}
		}
	}

	return separators
}

// configRowsMaxDepth finds the maximum key depth across all rows.
func configRowsMaxDepth(rows []configRow) int {
	maxDepth := 0
	for _, r := range rows {
		if len(r.keys) > maxDepth {
			maxDepth = len(r.keys)
		}
	}
	return maxDepth
}

// sortConfigRows sorts rows by key path for proper grouping.
func sortConfigRows(rows []configRow) {
	slices.SortFunc(rows, func(a, b configRow) int {
		minLen := min(len(a.keys), len(b.keys))
		for i := range minLen {
			if cmp := strings.Compare(a.keys[i], b.keys[i]); cmp != 0 {
				return cmp
			}
		}
		return len(a.keys) - len(b.keys)
	})
}

// configTableHeaders builds header names for the config table.
// Uses "Setting" for all key columns (first only displays text) and "Value" for the last.
func configTableHeaders(maxDepth int) []string {
	headers := make([]string, maxDepth+1)
	headers[0] = "Setting"
	// Leave remaining key columns empty (visual spanning effect)
	for i := 1; i < maxDepth; i++ {
		headers[i] = ""
	}
	headers[maxDepth] = "Value"
	return headers
}

// collectConfigRows recursively collects rows from a config map.
func collectConfigRows(m map[string]any, parentKeys []string, rows *[]configRow) {
	for key, value := range m {
		currentKeys := append(slices.Clone(parentKeys), key)

		switch v := value.(type) {
		case map[string]any:
			// Recurse into nested map
			collectConfigRows(v, currentKeys, rows)
		case []any:
			*rows = append(*rows, configRow{
				keys:  currentKeys,
				value: formatSliceValue(v),
			})
		default:
			*rows = append(*rows, configRow{
				keys:  currentKeys,
				value: formatScalarValue(v),
			})
		}
	}
}

// formatScalarValue formats a scalar config value for display.
func formatScalarValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "<not set>"
	case bool:
		if val {
			return output.LabelTrue
		}
		return output.LabelFalse
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		if val == "" {
			return "<empty>"
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatSliceValue formats a slice as comma-separated values.
func formatSliceValue(slice []any) string {
	if len(slice) == 0 {
		return "<empty>"
	}
	parts := make([]string, len(slice))
	for i, v := range slice {
		parts[i] = formatScalarValue(v)
	}
	return strings.Join(parts, ", ")
}

// DisplaySceneList prints a table of scenes.
func DisplaySceneList(ios *iostreams.IOStreams, scenes []config.Scene) {
	builder := table.NewBuilder("Name", "Actions", "Description")
	for _, scene := range scenes {
		actions := output.FormatActionCount(len(scene.Actions))
		description := scene.Description
		if description == "" {
			description = "-"
		}
		builder.AddRow(scene.Name, actions, description)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print scenes table", err)
	}
	ios.Println()
	ios.Count("scene", len(scenes))
}

// DisplayAliasList prints a table of aliases.
func DisplayAliasList(ios *iostreams.IOStreams, aliases []config.NamedAlias) {
	builder := table.NewBuilder("Name", "Command", "Type")

	for _, alias := range aliases {
		aliasType := "command"
		if alias.Shell {
			aliasType = "shell"
		}
		builder.AddRow(alias.Name, alias.Command, aliasType)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print aliases table", err)
	}
	ios.Println()
	ios.Count("alias", len(aliases))
}

// DisplayResetableComponents lists available components that can be reset.
func DisplayResetableComponents(ios *iostreams.IOStreams, device string, configKeys []string) {
	ios.Title("Available components")
	ios.Printf("Specify a component to reset its configuration:\n")
	ios.Printf("\n")

	for _, key := range configKeys {
		ios.Printf("  shelly config reset %s %s\n", device, key)
	}
}

// DisplayTemplateDiffs prints a table of template comparison diffs.
func DisplayTemplateDiffs(ios *iostreams.IOStreams, templateName, deviceName string, diffs []model.ConfigDiff) {
	if len(diffs) == 0 {
		ios.Info("No differences - device matches template")
		return
	}

	ios.Title("Configuration Differences")
	ios.Printf("Template: %s  Device: %s\n\n", templateName, deviceName)

	builder := table.NewBuilder("Path", "Type", "Device Value", "Template Value")
	for _, d := range diffs {
		builder.AddRow(d.Path, d.DiffType, output.FormatDisplayValue(d.OldValue), output.FormatDisplayValue(d.NewValue))
	}
	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print template diffs table", err)
	}
	ios.Printf("\n%d difference(s) found\n", len(diffs))
}

// DisplayDeviceTemplateList prints a table of device configuration templates.
func DisplayDeviceTemplateList(ios *iostreams.IOStreams, templates []config.DeviceTemplate) {
	ios.Title("Configuration Templates")
	ios.Println()

	builder := table.NewBuilder("Name", "Model", "Gen", "Source", "Created")
	for _, t := range templates {
		source := t.SourceDevice
		if source == "" {
			source = "-"
		}
		created := t.CreatedAt
		if len(created) > 10 {
			created = created[:10] // Just the date part
		}
		builder.AddRow(t.Name, t.Model, fmt.Sprintf("Gen%d", t.Generation), source, created)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print templates table", err)
	}
	ios.Printf("\n%d template(s)\n", len(templates))
}
