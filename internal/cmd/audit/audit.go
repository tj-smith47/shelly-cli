// Package audit provides the audit command for security auditing devices.
package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds the command options.
type Options struct {
	All bool
}

// Result holds the results of a device security audit.
type Result struct {
	Device     string
	Address    string
	Issues     []string
	Warnings   []string
	InfoItems  []string
	Reachable  bool
	AuthStatus *AuthAudit
	CloudAudit *CloudAudit
	FWAudit    *FirmwareAudit
}

// AuthAudit holds authentication audit results.
type AuthAudit struct {
	AuthEnabled bool
}

// CloudAudit holds cloud audit results.
type CloudAudit struct {
	Connected bool
}

// FirmwareAudit holds firmware audit results.
type FirmwareAudit struct {
	Current   string
	Available string
	HasUpdate bool
}

// NewCommand creates the audit command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			if opts.All {
				return runAll(cmd.Context(), f)
			}
			if len(args) == 0 {
				return fmt.Errorf("specify device(s) or use --all")
			}
			return run(cmd.Context(), f, args)
		},
	}

	cmd.Flags().BoolVar(&opts.All, "all", false, "Audit all registered devices")

	return cmd
}

func runAll(ctx context.Context, f *cmdutil.Factory) error {
	devices := config.ListDevices()
	if len(devices) == 0 {
		ios := f.IOStreams()
		ios.Warning("No devices registered. Run 'shelly discover mdns --register' first.")
		return nil
	}

	deviceList := make([]string, 0, len(devices))
	for name := range devices {
		deviceList = append(deviceList, name)
	}

	return run(ctx, f, deviceList)
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Println("")
	ios.Println(theme.Title().Render("Shelly Security Audit"))
	ios.Println(theme.Dim().Render(strings.Repeat("━", 50)))
	ios.Println("")

	totalIssues := 0
	totalWarnings := 0

	for _, device := range devices {
		result := auditDevice(ctx, svc, device)
		printAuditResult(ios, result)
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

func auditDevice(ctx context.Context, svc *shelly.Service, device string) *Result {
	result := &Result{
		Device:    device,
		Issues:    []string{},
		Warnings:  []string{},
		InfoItems: []string{},
	}

	// Resolve device address
	if d, ok := config.GetDevice(device); ok {
		result.Address = d.Address
	} else {
		result.Address = device
	}

	// Try to ping device first
	info, err := svc.DevicePing(ctx, device)
	if err != nil {
		result.Reachable = false
		result.Issues = append(result.Issues, "Device unreachable")
		return result
	}
	result.Reachable = true

	// Check authentication status
	result.AuthStatus = &AuthAudit{
		AuthEnabled: info.AuthEn,
	}
	if !info.AuthEn {
		result.Issues = append(result.Issues, "Authentication is DISABLED - device is unprotected")
	} else {
		result.InfoItems = append(result.InfoItems, "Authentication enabled")
	}

	// Check cloud status
	cloudStatus, err := svc.GetCloudStatus(ctx, device)
	if err == nil {
		result.CloudAudit = &CloudAudit{
			Connected: cloudStatus.Connected,
		}
		switch {
		case cloudStatus.Connected && !info.AuthEn:
			result.Issues = append(result.Issues, "Cloud connected but NO AUTH - exposed to internet!")
		case cloudStatus.Connected:
			result.InfoItems = append(result.InfoItems, "Cloud connected (with auth)")
		default:
			result.InfoItems = append(result.InfoItems, "Cloud not connected (local only)")
		}
	}

	// Check firmware
	fwInfo, err := svc.CheckFirmware(ctx, device)
	if err == nil {
		result.FWAudit = &FirmwareAudit{
			Current:   fwInfo.Current,
			Available: fwInfo.Available,
			HasUpdate: fwInfo.HasUpdate,
		}
		if fwInfo.HasUpdate {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Firmware update available: %s -> %s", fwInfo.Current, fwInfo.Available))
		} else {
			result.InfoItems = append(result.InfoItems,
				fmt.Sprintf("Firmware up to date (%s)", fwInfo.Current))
		}
	}

	return result
}

func printAuditResult(ios *iostreams.IOStreams, result *Result) {
	ios.Printf("%s %s\n",
		theme.Bold().Render(result.Device),
		theme.Dim().Render(fmt.Sprintf("(%s)", result.Address)))

	if !result.Reachable {
		ios.Printf("  %s Device unreachable - cannot audit\n", theme.StatusWarn().Render("⚠"))
		ios.Println("")
		return
	}

	// Print issues (security concerns)
	for _, issue := range result.Issues {
		ios.Printf("  %s %s\n", theme.StatusWarn().Render("⚠"), issue)
	}

	// Print warnings
	for _, warn := range result.Warnings {
		ios.Printf("  %s %s\n", theme.StatusWarn().Render("!"), warn)
	}

	// Print info items
	for _, info := range result.InfoItems {
		ios.Printf("  %s %s\n", theme.StatusOK().Render("✓"), info)
	}

	ios.Println("")
}
