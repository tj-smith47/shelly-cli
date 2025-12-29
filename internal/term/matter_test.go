package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayMatterStatus_Enabled(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.MatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   0,
	}
	DisplayMatterStatus(ios, status, "kitchen-light")

	output := out.String()
	if !strings.Contains(output, "Matter Status") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "Enabled") {
		t.Error("expected enabled status")
	}
	if !strings.Contains(output, "Commissionable") {
		t.Error("expected commissionable status")
	}
	if !strings.Contains(output, "shelly matter code kitchen-light") {
		t.Error("expected code command hint")
	}
}

func TestDisplayMatterStatus_Disabled(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.MatterStatus{
		Enabled: false,
	}
	DisplayMatterStatus(ios, status, "device")

	output := out.String()
	if !strings.Contains(output, "shelly matter enable device") {
		t.Error("expected enable command hint")
	}
}

func TestDisplayMatterStatus_Paired(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.MatterStatus{
		Enabled:        true,
		Commissionable: false,
		FabricsCount:   2,
	}
	DisplayMatterStatus(ios, status, "device")

	output := out.String()
	if !strings.Contains(output, "Paired Fabrics: 2") {
		t.Error("expected fabric count")
	}
}

func TestOutputMatterStatusJSON(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	status := model.MatterStatus{
		Enabled:        true,
		Commissionable: true,
		FabricsCount:   1,
	}
	err := OutputMatterStatusJSON(ios, status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "enabled") {
		t.Error("expected enabled field in JSON")
	}
}

func TestDisplayCommissioningInfo_Available(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := model.CommissioningInfo{
		Available:     true,
		ManualCode:    "12345-67890",
		QRCode:        "MT:ABC123",
		Discriminator: 3840,
		SetupPINCode:  12345678,
	}
	DisplayCommissioningInfo(ios, info, "192.168.1.100")

	output := out.String()
	if !strings.Contains(output, "Matter Pairing Code") {
		t.Error("expected pairing code header")
	}
	if !strings.Contains(output, "12345-67890") {
		t.Error("expected manual code")
	}
	if !strings.Contains(output, "MT:ABC123") {
		t.Error("expected QR data")
	}
}

func TestDisplayCommissioningInfo_NotAvailable(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := model.CommissioningInfo{
		Available:  false,
		ManualCode: "",
	}
	DisplayCommissioningInfo(ios, info, "192.168.1.100")

	output := out.String()
	if !strings.Contains(output, "not available via API") {
		t.Error("expected not available message")
	}
	if !strings.Contains(output, "192.168.1.100/matter") {
		t.Error("expected web UI hint")
	}
}

func TestDisplayAvailableCode_Full(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := model.CommissioningInfo{
		ManualCode:    "11111-22222",
		QRCode:        "MT:XYZ",
		Discriminator: 1000,
		SetupPINCode:  99999999,
	}
	DisplayAvailableCode(ios, info)

	output := out.String()
	if !strings.Contains(output, "11111-22222") {
		t.Error("expected manual code")
	}
	if !strings.Contains(output, "Discriminator: 1000") {
		t.Error("expected discriminator")
	}
	if !strings.Contains(output, "Setup PIN: 99999999") {
		t.Error("expected setup PIN")
	}
	if !strings.Contains(output, "Matter controller app") {
		t.Error("expected usage hint")
	}
}

func TestDisplayAvailableCode_MinimalInfo(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	info := model.CommissioningInfo{
		ManualCode: "33333-44444",
	}
	DisplayAvailableCode(ios, info)

	output := out.String()
	if !strings.Contains(output, "33333-44444") {
		t.Error("expected manual code")
	}
	// Should not show discriminator when 0
	if strings.Contains(output, "Discriminator: 0") {
		t.Error("should not show zero discriminator")
	}
}

func TestDisplayNotAvailable_WithIP(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayNotAvailable(ios, "192.168.1.50")

	output := out.String()
	if !strings.Contains(output, "Pairing Information") {
		t.Error("expected header")
	}
	if !strings.Contains(output, "not available via API") {
		t.Error("expected not available message")
	}
	if !strings.Contains(output, "http://192.168.1.50/matter") {
		t.Error("expected web UI URL")
	}
	if !strings.Contains(output, "QR code") {
		t.Error("expected QR code hint")
	}
}

func TestDisplayNotAvailable_NoIP(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayNotAvailable(ios, "")

	output := out.String()
	if !strings.Contains(output, "web UI at /matter") {
		t.Error("expected generic web UI hint")
	}
}
