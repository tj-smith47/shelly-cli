// Package audit provides the audit command for security auditing devices.
package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	Factory *cmdutil.Factory
	All     bool
	Devices []string
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
  - Authentication status (password protection)
  - Cloud connection exposure
  - Firmware version (security patches)

The audit flags potential security concerns such as:
  - Devices without authentication enabled
  - Devices connected to cloud with auth disabled
  - Outdated firmware that may have vulnerabilities

Use --all to audit all registered devices.`,
		Example: `  # Audit a single device
  shelly audit kitchen-light

  # Audit multiple devices
  shelly audit light-1 switch-2

  # Audit all registered devices
  shelly audit --all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices = args
			if opts.All {
				registered := config.ListDevices()
				if len(registered) == 0 {
					opts.Factory.IOStreams().Warning("No devices registered. Run 'shelly discover mdns --register' first.")
					return nil
				}
				opts.Devices = make([]string, 0, len(registered))
				for name := range registered {
					opts.Devices = append(opts.Devices, name)
				}
			} else if len(args) == 0 {
				return fmt.Errorf("specify device(s) or use --all")
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Audit all registered devices")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	ios.Println("")
	ios.Println(theme.Title().Render("Shelly Security Audit"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println("")

	totalIssues := 0
	totalWarnings := 0

	for _, device := range opts.Devices {
		result := svc.AuditDevice(ctx, device)
		term.DisplayAuditResult(ios, result)
		totalIssues += len(result.Issues)
		totalWarnings += len(result.Warnings)
	}

	// Summary
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
