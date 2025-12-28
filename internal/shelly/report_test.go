package shelly

import (
	"testing"
)

const (
	testApp = "Plus1PM"
)

func TestDeviceInfoResult_Fields(t *testing.T) {
	t.Parallel()

	info := deviceInfoResult{
		ID:  "shellyplus1pm-aabbcc",
		MAC: testMAC,
		App: testApp,
		Ver: "1.2.0",
	}

	if info.ID != "shellyplus1pm-aabbcc" {
		t.Errorf("expected ID 'shellyplus1pm-aabbcc', got %q", info.ID)
	}
	if info.MAC != testMAC {
		t.Errorf("expected MAC %q, got %q", testMAC, info.MAC)
	}
	if info.App != testApp {
		t.Errorf("expected App %q, got %q", testApp, info.App)
	}
	if info.Ver != "1.2.0" {
		t.Errorf("expected Ver '1.2.0', got %q", info.Ver)
	}
}

func TestParseDeviceInfo(t *testing.T) {
	t.Parallel()

	t.Run("valid map", func(t *testing.T) {
		t.Parallel()

		rawResult := map[string]any{
			"id":  "shellyplus1pm-aabbcc",
			"mac": testMAC,
			"app": testApp,
			"ver": "1.2.0",
		}

		info, ok := parseDeviceInfo(rawResult)

		if !ok {
			t.Error("expected parseDeviceInfo to succeed")
		}
		if info.ID != "shellyplus1pm-aabbcc" {
			t.Errorf("expected ID 'shellyplus1pm-aabbcc', got %q", info.ID)
		}
		if info.App != testApp {
			t.Errorf("expected App %q, got %q", testApp, info.App)
		}
	})

	t.Run("partial data", func(t *testing.T) {
		t.Parallel()

		rawResult := map[string]any{
			"id": "shellyplus1pm",
		}

		info, ok := parseDeviceInfo(rawResult)

		if !ok {
			t.Error("expected parseDeviceInfo to succeed")
		}
		if info.ID != "shellyplus1pm" {
			t.Errorf("expected ID 'shellyplus1pm', got %q", info.ID)
		}
		// Other fields should be empty
		if info.MAC != "" {
			t.Errorf("expected empty MAC, got %q", info.MAC)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		rawResult := map[string]any{}

		info, ok := parseDeviceInfo(rawResult)

		if !ok {
			t.Error("expected parseDeviceInfo to succeed for empty map")
		}
		if info.ID != "" {
			t.Errorf("expected empty ID, got %q", info.ID)
		}
	})
}

func TestExtractPower(t *testing.T) {
	t.Parallel()

	t.Run("switch with power", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"switch:0": map[string]any{
				"output": true,
				"apower": 150.5,
			},
		}

		power, ok := extractPower(rawStatus)

		if !ok {
			t.Error("expected extractPower to succeed")
		}
		if power != 150.5 {
			t.Errorf("expected power 150.5, got %f", power)
		}
	})

	t.Run("em with power", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"em:0": map[string]any{
				"apower": 2500.0,
			},
		}

		power, ok := extractPower(rawStatus)

		if !ok {
			t.Error("expected extractPower to succeed")
		}
		if power != 2500.0 {
			t.Errorf("expected power 2500.0, got %f", power)
		}
	})

	t.Run("no power data", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"temperature:0": map[string]any{
				"tC": 25.0,
			},
		}

		_, ok := extractPower(rawStatus)

		if ok {
			t.Error("expected extractPower to fail for status without power")
		}
	})

	t.Run("empty status", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{}

		_, ok := extractPower(rawStatus)

		if ok {
			t.Error("expected extractPower to fail for empty status")
		}
	})
}

func TestExtractAuthEnabled(t *testing.T) {
	t.Parallel()

	t.Run("auth enabled", func(t *testing.T) {
		t.Parallel()

		rawInfo := map[string]any{
			"auth_en": true,
		}

		enabled := extractAuthEnabled(rawInfo)

		if !enabled {
			t.Error("expected auth to be enabled")
		}
	})

	t.Run("auth disabled", func(t *testing.T) {
		t.Parallel()

		rawInfo := map[string]any{
			"auth_en": false,
		}

		enabled := extractAuthEnabled(rawInfo)

		if enabled {
			t.Error("expected auth to be disabled")
		}
	})

	t.Run("missing auth field", func(t *testing.T) {
		t.Parallel()

		rawInfo := map[string]any{
			"id": "device",
		}

		enabled := extractAuthEnabled(rawInfo)

		if enabled {
			t.Error("expected auth to default to false when missing")
		}
	})
}

func TestExtractCloudConnected(t *testing.T) {
	t.Parallel()

	t.Run("cloud connected", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"connected": true,
		}

		connected := extractCloudConnected(rawStatus)

		if !connected {
			t.Error("expected cloud to be connected")
		}
	})

	t.Run("cloud not connected", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"connected": false,
		}

		connected := extractCloudConnected(rawStatus)

		if connected {
			t.Error("expected cloud to not be connected")
		}
	})

	t.Run("missing connected field", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{}

		connected := extractCloudConnected(rawStatus)

		if connected {
			t.Error("expected cloud to default to false when missing")
		}
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		connected := extractCloudConnected(nil)

		if connected {
			t.Error("expected cloud to default to false for nil input")
		}
	})
}

func TestParseDeviceInfo_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		info, ok := parseDeviceInfo(nil)

		if !ok {
			t.Error("expected parseDeviceInfo to succeed for nil")
		}
		if info == nil {
			t.Fatal("expected non-nil result")
		}
		// All fields should be empty
		if info.ID != "" || info.MAC != "" || info.App != "" || info.Ver != "" {
			t.Error("expected all empty fields for nil input")
		}
	})

	t.Run("unmarshalable type", func(t *testing.T) {
		t.Parallel()

		// A channel cannot be marshaled to JSON
		ch := make(chan int)

		_, ok := parseDeviceInfo(ch)

		if ok {
			t.Error("expected parseDeviceInfo to fail for unmarshalable type")
		}
	})
}

func TestExtractPower_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		power, ok := extractPower(nil)

		if ok {
			t.Error("expected extractPower to fail for nil input")
		}
		if power != 0 {
			t.Errorf("expected power 0, got %f", power)
		}
	})

	t.Run("unmarshalable type", func(t *testing.T) {
		t.Parallel()

		ch := make(chan int)

		_, ok := extractPower(ch)

		if ok {
			t.Error("expected extractPower to fail for unmarshalable type")
		}
	})

	t.Run("nested non-map value", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"sys": "string value",
		}

		_, ok := extractPower(rawStatus)

		if ok {
			t.Error("expected extractPower to fail for non-map value")
		}
	})

	t.Run("pm with power", func(t *testing.T) {
		t.Parallel()

		rawStatus := map[string]any{
			"pm:0": map[string]any{
				"apower": 1250.0,
			},
		}

		power, ok := extractPower(rawStatus)

		if !ok {
			t.Error("expected extractPower to succeed")
		}
		if power != 1250.0 {
			t.Errorf("expected power 1250.0, got %f", power)
		}
	})
}

func TestExtractAuthEnabled_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		enabled := extractAuthEnabled(nil)

		if enabled {
			t.Error("expected auth to default to false for nil input")
		}
	})

	t.Run("unmarshalable type", func(t *testing.T) {
		t.Parallel()

		ch := make(chan int)

		enabled := extractAuthEnabled(ch)

		if enabled {
			t.Error("expected auth to default to false for unmarshalable type")
		}
	})

	t.Run("non-bool auth_en", func(t *testing.T) {
		t.Parallel()

		// auth_en as string should fail to unmarshal into bool
		rawInfo := map[string]any{
			"auth_en": "yes",
		}

		enabled := extractAuthEnabled(rawInfo)

		if enabled {
			t.Error("expected auth to default to false for non-bool auth_en")
		}
	})
}
