package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayLoRaStatus_Full(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	full := model.LoRaFullStatus{
		Config: &model.LoRaConfig{
			ID:   0,
			Freq: 868000000,
			BW:   125,
			DR:   7,
			TxP:  14,
		},
		Status: &model.LoRaStatus{
			RSSI: -80,
			SNR:  8.5,
		},
	}
	DisplayLoRaStatus(ios, full)

	output := out.String()
	if !strings.Contains(output, "LoRa Add-on Status") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Configuration") {
		t.Error("expected config section")
	}
	if !strings.Contains(output, "868000000") {
		t.Error("expected frequency")
	}
	if !strings.Contains(output, "MHz") {
		t.Error("expected MHz conversion")
	}
	if !strings.Contains(output, "125") {
		t.Error("expected bandwidth")
	}
	if !strings.Contains(output, "14 dBm") {
		t.Error("expected TX power")
	}
	if !strings.Contains(output, "Last Packet") {
		t.Error("expected last packet section")
	}
	if !strings.Contains(output, "-80 dBm") {
		t.Error("expected RSSI")
	}
	if !strings.Contains(output, "8.5 dB") {
		t.Error("expected SNR")
	}
}

func TestDisplayLoRaStatus_ConfigOnly(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	full := model.LoRaFullStatus{
		Config: &model.LoRaConfig{
			ID:   0,
			Freq: 915000000,
			BW:   250,
			DR:   10,
			TxP:  20,
		},
		Status: nil,
	}
	DisplayLoRaStatus(ios, full)

	output := out.String()
	if !strings.Contains(output, "Configuration") {
		t.Error("expected config section")
	}
	if strings.Contains(output, "Last Packet") {
		t.Error("should not show last packet without status")
	}
}

func TestDisplayLoRaStatus_StatusOnly(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	full := model.LoRaFullStatus{
		Config: nil,
		Status: &model.LoRaStatus{
			RSSI: -90,
			SNR:  5.0,
		},
	}
	DisplayLoRaStatus(ios, full)

	output := out.String()
	if strings.Contains(output, "Configuration") {
		t.Error("should not show config without config data")
	}
	if !strings.Contains(output, "Last Packet") {
		t.Error("expected last packet section")
	}
}

func TestDisplayLoRaStatus_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	full := model.LoRaFullStatus{}
	DisplayLoRaStatus(ios, full)

	output := out.String()
	if !strings.Contains(output, "LoRa Add-on Status") {
		t.Error("expected header even with empty status")
	}
}

func TestOutputLoRaStatusJSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	full := model.LoRaFullStatus{
		Config: &model.LoRaConfig{
			ID:   0,
			Freq: 868000000,
			BW:   125,
			DR:   7,
			TxP:  14,
		},
		Status: &model.LoRaStatus{
			RSSI: -75,
			SNR:  10.0,
		},
	}
	err := OutputLoRaStatusJSON(ios, full)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "868000000") {
		t.Error("expected frequency in JSON")
	}
	if !strings.Contains(output, "-75") {
		t.Error("expected RSSI in JSON")
	}
}
