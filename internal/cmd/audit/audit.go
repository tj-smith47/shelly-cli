// Package audit provides the audit command for security auditing devices.
package audit

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	All     bool
	Devices []string
	Checks  model.AuditChecks
}

// NewCommand creates the audit command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "audit [device...]",
		Aliases: []string{"security", "sec"},
		Short:   "Security audit for devices",
		Long: `Perform a security audit on Shelly devices.

Checks performed:
  - auth:     Authentication status (password protection)
  - cloud:    Cloud connection exposure (cloud connected without a password)
  - firmware: Firmware version (security patches)

Every check runs by default. Pass one or more --check-<name> flags to run
only those checks.

With no device named, every registered device is audited (the same as --all).

Use -o json or -o yaml for one structured result per device, suitable for
scripting.`,
		Example: `  # Audit a single device
  shelly audit kitchen-light

  # Audit multiple devices
  shelly audit light-1 switch-2

  # Audit all registered devices
  shelly audit

  # Only check firmware and authentication, as JSON
  shelly audit --check-firmware --check-auth -o json

  # Only check cloud exposure on one device
  shelly audit kitchen-light --check-cloud`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices = args
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Audit all registered devices (the default when no device is named)")
	cmd.Flags().BoolVar(&opts.Checks.Auth, "check-auth", false, "Check authentication (password protection)")
	cmd.Flags().BoolVar(&opts.Checks.Cloud, "check-cloud", false, "Check cloud connection exposure")
	cmd.Flags().BoolVar(&opts.Checks.Firmware, "check-firmware", false, "Check for firmware updates")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()
	structured := cmdutil.StructuredOutput()

	devices := opts.Devices
	if opts.All || len(devices) == 0 {
		cfg, err := opts.Factory.Config()
		if err != nil {
			return err
		}
		devices = slices.Sorted(maps.Keys(cfg.Devices))
		if len(devices) == 0 && !structured {
			ios.Warning("No devices registered. Run 'shelly discover mdns --register' first.")
			return nil
		}
	}

	checks := opts.Checks
	if !checks.Any() {
		checks = model.AllAuditChecks()
	}

	results, err := cmdutil.RunWithSpinnerResult(ctx, ios, fmt.Sprintf("Auditing %d devices...", len(devices)),
		func(ctx context.Context) ([]*model.AuditResult, error) {
			return svc.AuditDevices(ctx, devices, checks), nil
		})
	if err != nil {
		return err
	}
	if structured {
		return cmdutil.PrintListResult(ios, results, nil)
	}

	ios.Println("")
	ios.Println(theme.Title().Render("Shelly Security Audit"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println("")
	totalIssues := 0
	totalWarnings := 0
	for _, result := range results {
		totalIssues += len(result.Issues)
		totalWarnings += len(result.Warnings)
		term.DisplayAuditResult(ios, result)
	}

	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	if totalIssues == 0 && totalWarnings == 0 {
		ios.Success("No security issues found!")
	} else {
		if totalIssues > 0 {
			ios.Printf("%s %d security issue(s) found\n",
				theme.StatusWarn().Render("⚠"),
				totalIssues)
		}
		if totalWarnings > 0 {
			ios.Info("%d warning(s) - review recommended", totalWarnings)
		}
	}
	ios.Println("")

	return nil
}
