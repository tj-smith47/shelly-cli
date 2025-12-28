// Package report provides the report command for generating reports.
package report

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds the command options.
type Options struct {
	flags.OutputFlags
	Type   string
	Output string
}

// NewCommand creates the report command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Type: "devices",
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
	cmd.Flags().StringVar(&opts.Output, "output-file", "", "Output file path")
	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "json", "json", "text")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	var report model.DeviceReport
	var spinnerMsg string

	switch opts.Type {
	case "devices":
		spinnerMsg = "Generating device report..."
	case "energy":
		spinnerMsg = "Generating energy report..."
	case "audit":
		spinnerMsg = "Generating security audit report..."
	default:
		return fmt.Errorf("unknown report type: %s", opts.Type)
	}

	err = cmdutil.RunWithSpinner(ctx, ios, spinnerMsg, func(ctx context.Context) error {
		switch opts.Type {
		case "devices":
			report = svc.GenerateDevicesReport(ctx, cfg.Devices)
		case "energy":
			report = svc.GenerateEnergyReport(ctx, cfg.Devices)
		case "audit":
			report = svc.GenerateAuditReport(ctx, cfg.Devices)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return term.OutputReport(ios, report, opts.Format, opts.Output)
}
