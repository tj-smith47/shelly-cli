package shelly

import (
	"testing"
)

func TestComponentTypePM_Constants(t *testing.T) {
	t.Parallel()

	if ComponentTypePM != "pm" {
		t.Errorf("expected ComponentTypePM 'pm', got %q", ComponentTypePM)
	}
	if ComponentTypePM1 != "pm1" {
		t.Errorf("expected ComponentTypePM1 'pm1', got %q", ComponentTypePM1)
	}
}
