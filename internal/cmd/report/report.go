// Package report provides the report command for generating reports.
package report

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	flags.OutputFlags
	Output string
	Type   string
}

// NewCommand creates the report command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory: f,
		Type:    model.ReportTypeDevices,
	}

	cmd := &cobra.Command{
		Use:     "report",
		Aliases: []string{"generate"},
		Short:   "Generate reports",
		Long: `Generate a report about every registered device: an inventory, the current
power draw, or a security audit. All devices are queried at the same time, Gen1
and Gen2+ alike, and each device has one row, sorted by name.

Report types:
  devices  - name, ip, model, generation, firmware, mac and online for each
             device; summary: total, online, offline
  energy   - online, reporting (the device has a power meter) and power_w for
             each device; summary: total, online, offline, devices_reporting,
             total_power_w
  audit    - the checks of 'shelly audit': reachable, auth_enabled,
             cloud_connected, firmware_current, firmware_available,
             firmware_outdated, issues and warnings for each device; summary:
             devices_scanned, reachable, unreachable, auth_enabled,
             auth_disabled, cloud_connected, outdated_firmware, issues, warnings

A value the device did not report (it is offline, or the check failed) is
null in the audit rows. The document has timestamp, report_type, devices and
summary.

Output formats (--format, or the global -o):
  json   - JSON (default)
  yaml   - YAML
  text   - human-readable table (-o table is the same)

Progress messages go to stderr, so stdout carries only the report.`,
		Example: `  # Device inventory as JSON
  shelly report --type devices -o json

  # Names of the devices that are offline
  shelly report --type devices -o json | jq -r '.devices[] | select(.online == false) | .name'

  # Current power draw of every device, as a table
  shelly report --type energy -o text

  # Total power in watts
  shelly report --type energy -o json | jq '.summary.total_power_w'

  # Security audit as YAML
  shelly report --type audit -o yaml

  # Devices with a firmware update available
  shelly report --type audit -o json | jq -r '.devices[] | select(.firmware_outdated == true) | .name'

  # Save a report to a file
  shelly report --type devices --output-file report.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", model.ReportTypeDevices, "Report type: devices, energy, audit")
	cmd.Flags().StringVar(&opts.Output, "output-file", "", "Output file path")
	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "json", "json", "yaml", "text", "table")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	switch opts.Type {
	case model.ReportTypeDevices, model.ReportTypeEnergy, model.ReportTypeAudit:
	default:
		return fmt.Errorf("unknown report type %q: use devices, energy or audit", opts.Type)
	}

	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	msg := fmt.Sprintf("Generating %s report for %d devices...", opts.Type, len(cfg.Devices))
	switch opts.Type {
	case model.ReportTypeEnergy:
		report, err := cmdutil.RunWithSpinnerResult(ctx, ios, msg, func(ctx context.Context) (model.EnergyReport, error) {
			return svc.GenerateEnergyReport(ctx, cfg.Devices), nil
		})
		if err != nil {
			return err
		}
		return term.OutputReport(ios, report, output.FormatEnergyReportText, opts.Format, opts.Output)
	case model.ReportTypeAudit:
		report, err := cmdutil.RunWithSpinnerResult(ctx, ios, msg, func(ctx context.Context) (model.AuditReport, error) {
			return svc.GenerateAuditReport(ctx, cfg.Devices), nil
		})
		if err != nil {
			return err
		}
		return term.OutputReport(ios, report, output.FormatAuditReportText, opts.Format, opts.Output)
	default:
		report, err := cmdutil.RunWithSpinnerResult(ctx, ios, msg, func(ctx context.Context) (model.DevicesReport, error) {
			return svc.GenerateDevicesReport(ctx, cfg.Devices), nil
		})
		if err != nil {
			return err
		}
		return term.OutputReport(ios, report, output.FormatDevicesReportText, opts.Format, opts.Output)
	}
}
