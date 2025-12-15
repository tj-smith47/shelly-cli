// Package report provides the report command for generating reports.
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Type   string
	Output string
	Format string
}

// DeviceReport holds device information for reporting.
type DeviceReport struct {
	Timestamp  time.Time              `json:"timestamp"`
	ReportType string                 `json:"report_type"`
	Devices    []DeviceInfo           `json:"devices"`
	Summary    map[string]interface{} `json:"summary"`
}

// DeviceInfo holds individual device information.
type DeviceInfo struct {
	Name     string `json:"name"`
	IP       string `json:"ip,omitempty"`
	Model    string `json:"model,omitempty"`
	Firmware string `json:"firmware,omitempty"`
	Online   bool   `json:"online"`
	MAC      string `json:"mac,omitempty"`
}

// deviceInfoResult holds parsed device info from API.
type deviceInfoResult struct {
	ID  string `json:"id"`
	MAC string `json:"mac"`
	App string `json:"app"`
	Ver string `json:"ver"`
}

// parseDeviceInfo parses raw API result into deviceInfoResult.
func parseDeviceInfo(rawResult any) (*deviceInfoResult, bool) {
	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return nil, false
	}
	var info deviceInfoResult
	if err := json.Unmarshal(jsonBytes, &info); err != nil {
		return nil, false
	}
	return &info, true
}

// extractPower extracts active power from device status.
func extractPower(rawStatus any) (float64, bool) {
	statusBytes, err := json.Marshal(rawStatus)
	if err != nil {
		return 0, false
	}
	var status map[string]interface{}
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return 0, false
	}
	// Look for switch:0 or em:0 with power info
	for _, val := range status {
		if valMap, ok := val.(map[string]interface{}); ok {
			if power, ok := valMap["apower"].(float64); ok {
				return power, true
			}
		}
	}
	return 0, false
}

// extractAuthEnabled extracts auth_en from device info.
func extractAuthEnabled(rawInfo any) bool {
	infoBytes, err := json.Marshal(rawInfo)
	if err != nil {
		return false
	}
	var info struct {
		Auth bool `json:"auth_en"`
	}
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return false
	}
	return info.Auth
}

// extractCloudConnected extracts connected status from cloud status.
func extractCloudConnected(rawStatus any) bool {
	statusBytes, err := json.Marshal(rawStatus)
	if err != nil {
		return false
	}
	var status struct {
		Connected bool `json:"connected"`
	}
	if err := json.Unmarshal(statusBytes, &status); err != nil {
		return false
	}
	return status.Connected
}

// NewCommand creates the report command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Type:   "devices",
		Format: "json",
	}

	cmd := &cobra.Command{
		Use:     "report",
		Aliases: []string{"generate", "export"},
		Short:   "Generate reports",
		Long: `Generate reports about devices, energy usage, or security audits.

Report types:
  devices  - Device inventory and status
  energy   - Energy consumption summary
  audit    - Security audit report

Output formats:
  json   - JSON format (default)
  text   - Human-readable text`,
		Example: `  # Generate device report
  shelly report --type devices

  # Save report to file
  shelly report --type devices -o report.json

  # Generate energy report
  shelly report --type energy

  # Text format report
  shelly report --type devices --format text`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "devices", "Report type: devices, energy, audit")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&opts.Format, "format", "json", "Output format: json, text")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	switch opts.Type {
	case "devices":
		return runDevicesReport(ctx, f, opts)
	case "energy":
		return runEnergyReport(ctx, f, opts)
	case "audit":
		return runAuditReport(ctx, f, opts)
	default:
		return fmt.Errorf("unknown report type: %s", opts.Type)
	}
}

func runDevicesReport(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ios.StartProgress("Generating device report...")

	report := DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "devices",
		Devices:    []DeviceInfo{},
		Summary:    make(map[string]interface{}),
	}

	var online, offline int

	for name, deviceCfg := range cfg.Devices {
		info := DeviceInfo{
			Name:   name,
			IP:     deviceCfg.Address,
			Online: false,
		}

		// Try to get device info
		conn, err := svc.Connect(ctx, name)
		if err != nil {
			report.Devices = append(report.Devices, info)
			offline++
			continue
		}

		rawResult, callErr := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		iostreams.CloseWithDebug("closing device report connection", conn)

		if callErr != nil {
			report.Devices = append(report.Devices, info)
			offline++
			continue
		}

		if deviceInfo, ok := parseDeviceInfo(rawResult); ok {
			info.Online = true
			info.Model = deviceInfo.App
			info.Firmware = deviceInfo.Ver
			info.MAC = deviceInfo.MAC
			online++
		}

		if !info.Online {
			offline++
		}

		report.Devices = append(report.Devices, info)
	}

	ios.StopProgress()

	report.Summary["total"] = len(report.Devices)
	report.Summary["online"] = online
	report.Summary["offline"] = offline

	return outputReport(ios, report, opts)
}

func runEnergyReport(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ios.StartProgress("Generating energy report...")

	report := DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "energy",
		Devices:    []DeviceInfo{},
		Summary:    make(map[string]interface{}),
	}

	var totalPower float64
	var devicesWithEnergy int

	for name := range cfg.Devices {
		conn, err := svc.Connect(ctx, name)
		if err != nil {
			continue
		}

		rawStatus, err := conn.Call(ctx, "Shelly.GetStatus", nil)
		iostreams.CloseWithDebug("closing energy report connection", conn)

		if err != nil {
			continue
		}

		if power, ok := extractPower(rawStatus); ok {
			totalPower += power
			devicesWithEnergy++
		}
	}

	ios.StopProgress()

	report.Summary["total_power_w"] = totalPower
	report.Summary["devices_reporting"] = devicesWithEnergy

	return outputReport(ios, report, opts)
}

func runAuditReport(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ios.StartProgress("Generating security audit report...")

	report := DeviceReport{
		Timestamp:  time.Now(),
		ReportType: "audit",
		Devices:    []DeviceInfo{},
		Summary:    make(map[string]interface{}),
	}

	var authEnabled, cloudEnabled, outdated int

	for name := range cfg.Devices {
		conn, err := svc.Connect(ctx, name)
		if err != nil {
			continue
		}

		rawDeviceInfo, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		if err == nil && extractAuthEnabled(rawDeviceInfo) {
			authEnabled++
		}

		rawCloudStatus, err := conn.Call(ctx, "Cloud.GetStatus", nil)
		if err == nil && extractCloudConnected(rawCloudStatus) {
			cloudEnabled++
		}

		iostreams.CloseWithDebug("closing audit report connection", conn)
	}

	ios.StopProgress()

	report.Summary["devices_scanned"] = len(cfg.Devices)
	report.Summary["auth_enabled"] = authEnabled
	report.Summary["auth_disabled"] = len(cfg.Devices) - authEnabled
	report.Summary["cloud_connected"] = cloudEnabled
	report.Summary["outdated_firmware"] = outdated

	return outputReport(ios, report, opts)
}

func outputReport(ios *iostreams.IOStreams, report DeviceReport, opts *Options) error {
	var output string

	switch opts.Format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
		output = string(data)

	case "text":
		output = formatTextReport(report)

	default:
		return fmt.Errorf("unknown format: %s", opts.Format)
	}

	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(output), 0o600); err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		ios.Success("Report saved to: %s", opts.Output)
		return nil
	}

	ios.Println(output)
	return nil
}

func formatTextReport(report DeviceReport) string {
	var result string

	result += fmt.Sprintf("Shelly %s Report\n", report.ReportType)
	result += fmt.Sprintf("Generated: %s\n\n", report.Timestamp.Format(time.RFC3339))

	if len(report.Devices) > 0 {
		result += "Devices:\n"
		for _, d := range report.Devices {
			status := "offline"
			if d.Online {
				status = "online"
			}
			result += fmt.Sprintf("  - %s (%s): %s\n", d.Name, d.IP, status)
			if d.Model != "" {
				result += fmt.Sprintf("    Model: %s, Firmware: %s\n", d.Model, d.Firmware)
			}
		}
		result += "\n"
	}

	result += "Summary:\n"
	for k, v := range report.Summary {
		result += fmt.Sprintf("  %s: %v\n", k, v)
	}

	return result
}
