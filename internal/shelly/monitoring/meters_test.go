package monitoring

import (
	"testing"
)

func TestMaxComponentID(t *testing.T) {
	t.Parallel()

	if maxComponentID != 10 {
		t.Errorf("maxComponentID = %d, want 10", maxComponentID)
	}
}
