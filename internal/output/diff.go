// Package output provides output formatting utilities for the CLI.
package output

import (
	"reflect"
	"sort"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

// CompareConfigs recursively compares two configuration maps and returns structured diffs.
func CompareConfigs(source, target map[string]any) []model.ConfigDiff {
	return compareConfigsRecursive("", source, target)
}

func compareConfigsRecursive(prefix string, source, target map[string]any) []model.ConfigDiff {
	var diffs []model.ConfigDiff

	// Get all keys from both
	allKeys := make(map[string]bool)
	for k := range source {
		allKeys[k] = true
	}
	for k := range target {
		allKeys[k] = true
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		sourceVal, sourceOK := source[key]
		targetVal, targetOK := target[key]

		switch {
		case !sourceOK:
			// Key only in target
			diffs = append(diffs, model.ConfigDiff{
				Path:     path,
				DiffType: model.DiffAdded,
				NewValue: targetVal,
			})
		case !targetOK:
			// Key only in source
			diffs = append(diffs, model.ConfigDiff{
				Path:     path,
				DiffType: model.DiffRemoved,
				OldValue: sourceVal,
			})
		default:
			// Key in both - compare values
			sourceMap, sourceIsMap := sourceVal.(map[string]any)
			targetMap, targetIsMap := targetVal.(map[string]any)

			if sourceIsMap && targetIsMap {
				// Recursively compare nested objects
				diffs = append(diffs, compareConfigsRecursive(path, sourceMap, targetMap)...)
			} else if !reflect.DeepEqual(sourceVal, targetVal) {
				diffs = append(diffs, model.ConfigDiff{
					Path:     path,
					DiffType: model.DiffChanged,
					OldValue: sourceVal,
					NewValue: targetVal,
				})
			}
		}
	}

	return diffs
}
