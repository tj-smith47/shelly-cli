package shelly

import (
	"testing"
)

func TestSyncResult_Fields(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()

		result := SyncResult{
			Config: map[string]any{
				"switch:0": map[string]any{
					"name": "Test",
				},
			},
			Err: nil,
		}

		if result.Config == nil {
			t.Error("expected Config to be set")
		}
		if result.Err != nil {
			t.Error("expected Err to be nil")
		}
	})

	t.Run("error result", func(t *testing.T) {
		t.Parallel()

		result := SyncResult{
			Config: nil,
			Err:    &testError{msg: "connection failed"},
		}

		if result.Config != nil {
			t.Error("expected Config to be nil for error result")
		}
		if result.Err == nil {
			t.Error("expected Err to be set")
		}
	})
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
