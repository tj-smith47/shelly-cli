package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/provisioning"
)

func TestDisplayBLEProvisionResult(t *testing.T) {
	t.Parallel()

	t.Run("success result", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		start := time.Now().Add(-2 * time.Second)
		result := &provisioning.BLEProvisionResult{
			Success:     true,
			StartedAt:   start,
			CompletedAt: time.Now(),
			Device: &provisioning.BLEDevice{
				Name:       "ShellyPlus1PM",
				Address:    "AA:BB:CC:DD:EE:FF",
				Model:      "SNSW-001P16EU",
				Generation: 2,
			},
		}

		DisplayBLEProvisionResult(ios, result, "MyNetwork")

		output := out.String()
		if !strings.Contains(output, "successfully") {
			t.Error("output should contain 'successfully'")
		}
		if !strings.Contains(output, "ShellyPlus1PM") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "AA:BB:CC:DD:EE:FF") {
			t.Error("output should contain device address")
		}
		if !strings.Contains(output, "MyNetwork") {
			t.Error("output should contain WiFi SSID")
		}
	})

	t.Run("failure result", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()
		start := time.Now().Add(-1 * time.Second)
		testErr := &testError{msg: "connection timeout"}
		result := &provisioning.BLEProvisionResult{
			Success:     false,
			StartedAt:   start,
			CompletedAt: time.Now(),
			Error:       testErr,
		}

		DisplayBLEProvisionResult(ios, result, "")

		errOutput := errOut.String()
		if !strings.Contains(errOutput, "failed") {
			t.Errorf("output should contain 'failed', got %q", errOutput)
		}
	})

	t.Run("nil result", func(t *testing.T) {
		t.Parallel()

		ios, _, errOut := testIOStreams()

		DisplayBLEProvisionResult(ios, nil, "")

		errOutput := errOut.String()
		if !strings.Contains(errOutput, "No provisioning result") {
			t.Errorf("output should contain error message, got %q", errOutput)
		}
	})

	t.Run("result with device model", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		start := time.Now().Add(-1 * time.Second)
		result := &provisioning.BLEProvisionResult{
			Success:     true,
			StartedAt:   start,
			CompletedAt: time.Now(),
			Device: &provisioning.BLEDevice{
				Name:    "ShellyPlus2PM",
				Address: "11:22:33:44:55:66",
				Model:   "SNSW-002P16EU",
			},
		}

		DisplayBLEProvisionResult(ios, result, "TestNetwork")

		output := out.String()
		if !strings.Contains(output, "SNSW-002P16EU") {
			t.Error("output should contain device model")
		}
	})

	t.Run("success without device", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		start := time.Now().Add(-500 * time.Millisecond)
		result := &provisioning.BLEProvisionResult{
			Success:     true,
			StartedAt:   start,
			CompletedAt: time.Now(),
			Device:      nil,
		}

		DisplayBLEProvisionResult(ios, result, "WiFiNetwork")

		output := out.String()
		if !strings.Contains(output, "successfully") {
			t.Error("output should contain success message")
		}
	})
}

func TestDisplayBLEDevice(t *testing.T) {
	t.Parallel()

	t.Run("full device info", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		device := &provisioning.BLEDevice{
			Name:       "ShellyPlus1",
			Address:    "AA:BB:CC:DD:EE:FF",
			Model:      "SNSW-001P8EU",
			RSSI:       -65,
			Generation: 2,
		}

		DisplayBLEDevice(ios, device)

		output := out.String()
		if !strings.Contains(output, "ShellyPlus1") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "AA:BB:CC:DD:EE:FF") {
			t.Error("output should contain device address")
		}
		if !strings.Contains(output, "SNSW-001P8EU") {
			t.Error("output should contain model")
		}
		if !strings.Contains(output, "-65 dBm") {
			t.Error("output should contain signal strength")
		}
		if !strings.Contains(output, "Gen2") {
			t.Error("output should contain generation")
		}
	})

	t.Run("minimal device info", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		device := &provisioning.BLEDevice{
			Name:    "Unknown",
			Address: "11:22:33:44:55:66",
		}

		DisplayBLEDevice(ios, device)

		output := out.String()
		if !strings.Contains(output, "Unknown") {
			t.Error("output should contain device name")
		}
		if !strings.Contains(output, "11:22:33:44:55:66") {
			t.Error("output should contain device address")
		}
		// Should not contain optional fields
		if strings.Contains(output, "Model:") {
			t.Error("output should not contain Model when empty")
		}
	})

	t.Run("nil device", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()

		DisplayBLEDevice(ios, nil)

		output := out.String()
		if output != "" {
			t.Errorf("output should be empty for nil device, got %q", output)
		}
	})

	t.Run("device with zero RSSI", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		device := &provisioning.BLEDevice{
			Name:    "Device",
			Address: "AA:BB:CC:DD:EE:FF",
			RSSI:    0,
		}

		DisplayBLEDevice(ios, device)

		output := out.String()
		if strings.Contains(output, "Signal:") {
			t.Error("output should not contain Signal when RSSI is 0")
		}
	})

	t.Run("device with zero generation", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		device := &provisioning.BLEDevice{
			Name:       "Device",
			Address:    "AA:BB:CC:DD:EE:FF",
			Generation: 0,
		}

		DisplayBLEDevice(ios, device)

		output := out.String()
		if strings.Contains(output, "Generation:") {
			t.Error("output should not contain Generation when 0")
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
