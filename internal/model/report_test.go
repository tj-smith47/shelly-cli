package model

import (
	"testing"
	"time"
)

const (
	testReportIP  = "192.168.1.100"
	testReportMAC = "AA:BB:CC:DD:EE:FF"
)

func TestNewDeviceReport(t *testing.T) {
	t.Parallel()

	before := time.Now()
	report := NewDeviceReport("status")
	after := time.Now()

	if report.ReportType != "status" {
		t.Errorf("ReportType = %q, want %q", report.ReportType, "status")
	}
	if report.Timestamp.Before(before) || report.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", report.Timestamp, before, after)
	}
	if report.Devices == nil {
		t.Error("Devices should not be nil")
	}
	if len(report.Devices) != 0 {
		t.Errorf("Devices len = %d, want 0", len(report.Devices))
	}
	if report.Summary == nil {
		t.Error("Summary should not be nil")
	}
	if len(report.Summary) != 0 {
		t.Errorf("Summary len = %d, want 0", len(report.Summary))
	}
}

func TestNewDeviceReport_DifferentTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		reportType string
	}{
		{"status report", "status"},
		{"inventory report", "inventory"},
		{"firmware report", "firmware"},
		{"empty type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			report := NewDeviceReport(tt.reportType)
			if report.ReportType != tt.reportType {
				t.Errorf("ReportType = %q, want %q", report.ReportType, tt.reportType)
			}
		})
	}
}

func TestDeviceReport_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	report := DeviceReport{
		Timestamp:  now,
		ReportType: "inventory",
		Devices: []DeviceReportInfo{
			{Name: "Device1", IP: "192.168.1.1", Online: true},
			{Name: "Device2", IP: "192.168.1.2", Online: false},
		},
		Summary: map[string]interface{}{
			"total":  2,
			"online": 1,
		},
	}

	if report.Timestamp != now {
		t.Errorf("Timestamp = %v, want %v", report.Timestamp, now)
	}
	if report.ReportType != "inventory" {
		t.Errorf("ReportType = %q, want %q", report.ReportType, "inventory")
	}
	if len(report.Devices) != 2 {
		t.Errorf("Devices len = %d, want 2", len(report.Devices))
	}
	if total, ok := report.Summary["total"].(int); !ok || total != 2 {
		t.Errorf("Summary[total] = %v, want 2", report.Summary["total"])
	}
}

func TestDeviceReportInfo(t *testing.T) {
	t.Parallel()

	info := DeviceReportInfo{
		Name:     "Living Room Switch",
		IP:       testReportIP,
		Model:    "Shelly Plus 1PM",
		Firmware: "1.0.0-stable",
		Online:   true,
		MAC:      testReportMAC,
	}

	if info.Name != "Living Room Switch" {
		t.Errorf("Name = %q, want %q", info.Name, "Living Room Switch")
	}
	if info.IP != testReportIP {
		t.Errorf("IP = %q, want %q", info.IP, testReportIP)
	}
	if info.Model != "Shelly Plus 1PM" {
		t.Errorf("Model = %q, want %q", info.Model, "Shelly Plus 1PM")
	}
	if info.Firmware != "1.0.0-stable" {
		t.Errorf("Firmware = %q, want %q", info.Firmware, "1.0.0-stable")
	}
	if !info.Online {
		t.Error("Online = false, want true")
	}
	if info.MAC != testReportMAC {
		t.Errorf("MAC = %q, want %q", info.MAC, testReportMAC)
	}
}

func TestDeviceReportInfo_Empty(t *testing.T) {
	t.Parallel()

	info := DeviceReportInfo{}

	if info.Name != "" {
		t.Errorf("Name = %q, want empty", info.Name)
	}
	if info.IP != "" {
		t.Errorf("IP = %q, want empty", info.IP)
	}
	if info.Online {
		t.Error("Online = true, want false")
	}
}
