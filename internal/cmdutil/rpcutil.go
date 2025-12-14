package cmdutil

import (
	"encoding/json"
	"fmt"
)

// UnmarshalRPCResult converts an RPC call result (any) to a typed struct.
// This is a helper to avoid the repetitive marshal/unmarshal pattern.
func UnmarshalRPCResult[T any](result any) (T, error) {
	var target T

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return target, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &target); err != nil {
		return target, fmt.Errorf("failed to parse result: %w", err)
	}

	return target, nil
}

// UnmarshalRPCResultToMap converts an RPC call result to a map of raw JSON messages.
// This is useful when you need to iterate over dynamic component keys.
func UnmarshalRPCResultToMap(result any) (map[string]json.RawMessage, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	return m, nil
}
