// Package utils provides common functionality shared across CLI commands.
package utils

import (
	"testing"
)

func TestDeepEqualJSON_NestedStructures(t *testing.T) {
	t.Parallel()

	a := map[string]any{
		"outer": map[string]any{
			"inner": []any{1, 2, 3},
		},
	}
	b := map[string]any{
		"outer": map[string]any{
			"inner": []any{1, 2, 3},
		},
	}

	if !DeepEqualJSON(a, b) {
		t.Error("DeepEqualJSON() should return true for equal nested structures")
	}
}

func TestDeepEqualJSON_DifferentTypes(t *testing.T) {
	t.Parallel()

	a := map[string]int{"a": 1}
	b := []int{1}

	if DeepEqualJSON(a, b) {
		t.Error("DeepEqualJSON() should return false for different types")
	}
}

func TestDeepEqualJSON_Numbers(t *testing.T) {
	t.Parallel()

	// JSON numbers are all float64 when unmarshaled
	a := 42
	b := 42.0

	// After JSON round-trip, these should be equal
	if !DeepEqualJSON(a, b) {
		t.Error("DeepEqualJSON() should return true for equivalent numbers")
	}
}

func TestDeepEqualJSON_Booleans(t *testing.T) {
	t.Parallel()

	if !DeepEqualJSON(true, true) {
		t.Error("DeepEqualJSON(true, true) should be true")
	}
	if DeepEqualJSON(true, false) {
		t.Error("DeepEqualJSON(true, false) should be false")
	}
}

func TestDeepEqualJSON_UnmarshalableTypes(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshaled to JSON
	a := make(chan int)
	b := make(chan int)

	// Should return false when marshaling fails
	if DeepEqualJSON(a, b) {
		t.Error("DeepEqualJSON() should return false for unmarshalable types")
	}
}

func TestDeepEqualJSON_MixedNilAndEmpty(t *testing.T) {
	t.Parallel()

	// nil and empty slice both marshal to "null" and "[]" respectively
	var nilSlice []int
	emptySlice := []int{}

	// These are actually different in JSON (null vs [])
	if DeepEqualJSON(nilSlice, emptySlice) {
		t.Error("DeepEqualJSON() should return false for nil vs empty slice")
	}
}
