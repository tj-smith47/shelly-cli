package utils

import (
	"bytes"
	"encoding/json"
)

// DeepEqualJSON compares two values for equality by comparing their JSON representations.
func DeepEqualJSON(a, b any) bool {
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)
	if aErr != nil || bErr != nil {
		return false
	}
	return bytes.Equal(aJSON, bJSON)
}
