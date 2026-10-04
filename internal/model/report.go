package model

import "time"

// Report types selectable with `shelly report --type`.
const (
	ReportTypeDevices = "devices"
	ReportTypeEnergy  = "energy"
	ReportTypeAudit   = "audit"
)

// Report is the document `shelly report` prints: one row per registered
// device, sorted by name, and a summary of the rows.
type Report[Row, Summary any] struct {
	Timestamp  time.Time `json:"timestamp" yaml:"timestamp"`
	ReportType string    `json:"report_type" yaml:"report_type"`
	Devices    []Row     `json:"devices" yaml:"devices"`
	Summary    Summary   `json:"summary" yaml:"summary"`
}

// DevicesReport is the device inventory report.
type DevicesReport = Report[DeviceReportRow, DevicesReportSummary]

// EnergyReport is the power consumption report.
type EnergyReport = Report[EnergyReportRow, EnergyReportSummary]

// AuditReport is the security audit report.
type AuditReport = Report[AuditReportRow, AuditReportSummary]

// DeviceReportRow is one device in the inventory report. For a device that
// did not answer, model, generation and mac are the registered values and
// firmware is empty.
type DeviceReportRow struct {
	Name       string `json:"name" yaml:"name"`
	IP         string `json:"ip" yaml:"ip"`
	Model      string `json:"model" yaml:"model"`
	Generation int    `json:"generation" yaml:"generation"`
	Firmware   string `json:"firmware" yaml:"firmware"`
	MAC        string `json:"mac" yaml:"mac"`
	Online     bool   `json:"online" yaml:"online"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DevicesReportSummary counts the devices in the inventory report.
type DevicesReportSummary struct {
	Total   int `json:"total" yaml:"total"`
	Online  int `json:"online" yaml:"online"`
	Offline int `json:"offline" yaml:"offline"`
}

// EnergyReportRow is one device in the energy report. Reporting is true when
// the device has at least one power meter (Gen1 meter or emeter, Gen2+ EM,
// EM1, PM, PM1 or a switch that measures power); PowerW is the sum of their
// active power.
type EnergyReportRow struct {
	Name      string  `json:"name" yaml:"name"`
	Online    bool    `json:"online" yaml:"online"`
	Reporting bool    `json:"reporting" yaml:"reporting"`
	PowerW    float64 `json:"power_w" yaml:"power_w"`
	Error     string  `json:"error,omitempty" yaml:"error,omitempty"`
}

// EnergyReportSummary totals the energy report.
type EnergyReportSummary struct {
	Total            int     `json:"total" yaml:"total"`
	Online           int     `json:"online" yaml:"online"`
	Offline          int     `json:"offline" yaml:"offline"`
	DevicesReporting int     `json:"devices_reporting" yaml:"devices_reporting"`
	TotalPowerW      float64 `json:"total_power_w" yaml:"total_power_w"`
}

// AuditReportRow is one device in the security audit report. A field the
// audit could not determine (the device did not answer, or the check failed)
// is null.
type AuditReportRow struct {
	Name              string   `json:"name" yaml:"name"`
	IP                string   `json:"ip" yaml:"ip"`
	Reachable         bool     `json:"reachable" yaml:"reachable"`
	AuthEnabled       *bool    `json:"auth_enabled" yaml:"auth_enabled"`
	CloudConnected    *bool    `json:"cloud_connected" yaml:"cloud_connected"`
	FirmwareCurrent   string   `json:"firmware_current" yaml:"firmware_current"`
	FirmwareAvailable string   `json:"firmware_available" yaml:"firmware_available"`
	FirmwareOutdated  *bool    `json:"firmware_outdated" yaml:"firmware_outdated"`
	Issues            []string `json:"issues" yaml:"issues"`
	Warnings          []string `json:"warnings" yaml:"warnings"`
}

// AuditReportSummary counts the audit findings. AuthDisabled counts reachable
// devices without a password; unreachable devices are counted in Unreachable
// only.
type AuditReportSummary struct {
	DevicesScanned   int `json:"devices_scanned" yaml:"devices_scanned"`
	Reachable        int `json:"reachable" yaml:"reachable"`
	Unreachable      int `json:"unreachable" yaml:"unreachable"`
	AuthEnabled      int `json:"auth_enabled" yaml:"auth_enabled"`
	AuthDisabled     int `json:"auth_disabled" yaml:"auth_disabled"`
	CloudConnected   int `json:"cloud_connected" yaml:"cloud_connected"`
	OutdatedFirmware int `json:"outdated_firmware" yaml:"outdated_firmware"`
	Issues           int `json:"issues" yaml:"issues"`
	Warnings         int `json:"warnings" yaml:"warnings"`
}

// NewAuditReportRow converts one device's audit result into a report row.
func NewAuditReportRow(r *AuditResult) AuditReportRow {
	row := AuditReportRow{
		Name:      r.Device,
		IP:        r.Address,
		Reachable: r.Reachable,
		Issues:    r.Issues,
		Warnings:  r.Warnings,
	}
	if r.AuthStatus != nil {
		row.AuthEnabled = &r.AuthStatus.AuthEnabled
	}
	if r.CloudAudit != nil {
		row.CloudConnected = &r.CloudAudit.Connected
	}
	if r.FWAudit != nil {
		row.FirmwareCurrent = r.FWAudit.Current
		row.FirmwareAvailable = r.FWAudit.Available
		row.FirmwareOutdated = &r.FWAudit.HasUpdate
	}
	return row
}
