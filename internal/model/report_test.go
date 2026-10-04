package model

import (
	"encoding/json"
	"reflect"
	"testing"
)

const (
	testReportIP  = "192.168.1.100"
	testReportMAC = "AA:BB:CC:DD:EE:FF"
)

func TestNewAuditReportRow(t *testing.T) {
	t.Parallel()

	result := &AuditResult{
		Device:     "kitchen",
		Address:    testReportIP,
		Reachable:  true,
		Issues:     []string{"Authentication is DISABLED - device is unprotected"},
		Warnings:   []string{},
		AuthStatus: &AuthAudit{AuthEnabled: false},
		CloudAudit: &CloudAudit{Connected: true},
		FWAudit:    &FirmwareAudit{Current: "1.4.4", Available: "1.5.0", HasUpdate: true},
	}
	row := NewAuditReportRow(result)

	if row.Name != "kitchen" || row.IP != testReportIP || !row.Reachable {
		t.Errorf("row = %+v", row)
	}
	if row.AuthEnabled == nil || *row.AuthEnabled {
		t.Errorf("AuthEnabled = %v, want false", row.AuthEnabled)
	}
	if row.CloudConnected == nil || !*row.CloudConnected {
		t.Errorf("CloudConnected = %v, want true", row.CloudConnected)
	}
	if row.FirmwareCurrent != "1.4.4" || row.FirmwareAvailable != "1.5.0" || row.FirmwareOutdated == nil || !*row.FirmwareOutdated {
		t.Errorf("firmware = %+v", row)
	}
	if !reflect.DeepEqual(row.Issues, result.Issues) {
		t.Errorf("Issues = %v, want %v", row.Issues, result.Issues)
	}
}

func TestNewAuditReportRow_UnreachableLeavesChecksNull(t *testing.T) {
	t.Parallel()

	row := NewAuditReportRow(&AuditResult{Device: "attic", Issues: []string{"Device unreachable"}})
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"auth_enabled", "cloud_connected", "firmware_outdated"} {
		v, ok := doc[key]
		if !ok || v != nil {
			t.Errorf("%s = %v (present %v), want null", key, v, ok)
		}
	}
}

func TestReport_JSONKeys(t *testing.T) {
	t.Parallel()

	report := DevicesReport{
		ReportType: ReportTypeDevices,
		Devices:    []DeviceReportRow{{Name: "kitchen", IP: testReportIP, MAC: testReportMAC, Online: true}},
		Summary:    DevicesReportSummary{Total: 1, Online: 1},
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"timestamp", "report_type", "devices", "summary"} {
		if _, ok := doc[key]; !ok {
			t.Errorf("missing key %q in %s", key, data)
		}
	}
	row, ok := doc["devices"].([]any)[0].(map[string]any)
	if !ok {
		t.Fatalf("devices[0] is not an object: %s", data)
	}
	for _, key := range []string{"name", "ip", "model", "generation", "firmware", "mac", "online"} {
		if _, ok := row[key]; !ok {
			t.Errorf("missing row key %q in %s", key, data)
		}
	}
}
