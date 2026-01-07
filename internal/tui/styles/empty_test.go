package styles

import (
	"strings"
	"testing"
)

func TestEmptyState(t *testing.T) {
	t.Parallel()

	result := EmptyState("Test message", 40, 10)

	if !strings.Contains(result, "Test message") {
		t.Error("expected result to contain the message")
	}
}

func TestEmptyStateWithBorder(t *testing.T) {
	t.Parallel()

	result := EmptyStateWithBorder("Test message", 44, 12)

	if !strings.Contains(result, "Test message") {
		t.Error("expected result to contain the message")
	}
}

func TestNoDevicesOnline(t *testing.T) {
	t.Parallel()

	result := NoDevicesOnline(44, 12)

	if !strings.Contains(result, "No devices online") {
		t.Error("expected result to contain 'No devices online'")
	}
}

func TestNoDataAvailable(t *testing.T) {
	t.Parallel()

	result := NoDataAvailable(44, 12)

	if !strings.Contains(result, "No data available") {
		t.Error("expected result to contain 'No data available'")
	}
}

func TestNoItemsFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		itemType string
		expected string
	}{
		{"devices", "No devices found"},
		{"scripts", "No scripts found"},
		{"webhooks", "No webhooks found"},
	}

	for _, tt := range tests {
		t.Run(tt.itemType, func(t *testing.T) {
			t.Parallel()

			result := NoItemsFound(tt.itemType, 44, 12)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected result to contain %q", tt.expected)
			}
		})
	}
}

func TestNoDeviceSelected(t *testing.T) {
	t.Parallel()

	result := NoDeviceSelected(44, 12)

	if !strings.Contains(result, "No device selected") {
		t.Error("expected result to contain 'No device selected'")
	}
}

func TestNoItemsConfigured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		itemType string
		expected string
	}{
		{"webhooks", "No webhooks configured"},
		{"schedules", "No schedules configured"},
	}

	for _, tt := range tests {
		t.Run(tt.itemType, func(t *testing.T) {
			t.Parallel()

			result := NoItemsConfigured(tt.itemType, 44, 12)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected result to contain %q", tt.expected)
			}
		})
	}
}
