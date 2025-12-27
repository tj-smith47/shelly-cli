package model

import "time"

// DeviceReport holds device information for reporting.
type DeviceReport struct {
	Timestamp  time.Time              `json:"timestamp"`
	ReportType string                 `json:"report_type"`
	Devices    []DeviceReportInfo     `json:"devices"`
	Summary    map[string]interface{} `json:"summary"`
}

// DeviceReportInfo holds individual device information for reports.
type DeviceReportInfo struct {
	Name     string `json:"name"`
	IP       string `json:"ip,omitempty"`
	Model    string `json:"model,omitempty"`
	Firmware string `json:"firmware,omitempty"`
	Online   bool   `json:"online"`
	MAC      string `json:"mac,omitempty"`
}

// NewDeviceReport creates a new device report with the given type.
func NewDeviceReport(reportType string) DeviceReport {
	return DeviceReport{
		Timestamp:  time.Now(),
		ReportType: reportType,
		Devices:    []DeviceReportInfo{},
		Summary:    make(map[string]interface{}),
	}
}
