// Package check provides the firmware check subcommand.
package check

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

var allFlag bool

// NewCommand creates the firmware check command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "check [device]",
		Aliases: []string{"ck"},
		Short:   "Check for firmware updates",
		Long: `Check if firmware updates are available for a device.

Use --all to check all registered devices.`,
		Example: `  # Check a specific device
  shelly firmware check living-room

  # Check all registered devices
  shelly firmware check --all`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if allFlag {
				return runAll(f, cmd.Context())
			}
			if len(args) == 0 {
				return fmt.Errorf("device name required (or use --all)")
			}
			return run(f, cmd.Context(), args[0])
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "Check all registered devices")

	return cmd
}

func run(f *cmdutil.Factory, ctx context.Context, device string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, device,
		"Checking for updates...",
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.FirmwareInfo, error) {
			return svc.CheckFirmware(ctx, device)
		},
		displayFirmwareInfo)
}

func displayFirmwareInfo(ios *iostreams.IOStreams, info *shelly.FirmwareInfo) {
	ios.Println(theme.Bold().Render("Firmware Information"))
	ios.Println("")

	// Device info
	ios.Printf("  Device:     %s (%s)\n", info.DeviceID, info.DeviceModel)
	ios.Printf("  Generation: Gen%d\n", info.Generation)
	ios.Println("")

	// Version info
	ios.Printf("  Current:    %s\n", info.Current)

	switch {
	case info.HasUpdate:
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(true))
	case info.Available != "":
		ios.Printf("  Available:  %s %s\n",
			info.Available,
			output.RenderUpdateStatus(false))
	default:
		ios.Printf("  Available:  %s\n", output.RenderUpdateStatus(false))
	}

	if info.Beta != "" {
		ios.Printf("  Beta:       %s\n", info.Beta)
	}
}

func runAll(f *cmdutil.Factory, ctx context.Context) error {
	ios := f.IOStreams()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	devices := cfg.Devices
	if len(devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return nil
	}

	svc := f.ShellyService()

	type result struct {
		name string
		info *shelly.FirmwareInfo
		err  error
	}

	var (
		results []result
		mu      sync.Mutex
	)

	ios.StartProgress("Checking firmware on all devices...")

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5) // Limit concurrent checks

	for name := range devices {
		deviceName := name
		g.Go(func() error {
			info, checkErr := svc.CheckFirmware(gctx, deviceName)
			mu.Lock()
			results = append(results, result{name: deviceName, info: info, err: checkErr})
			mu.Unlock()
			return nil // Don't fail the whole group on individual errors
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("errgroup wait", err)
	}
	ios.StopProgress()

	// Build table
	table := output.NewTable("Device", "Current", "Available", "Status")
	updatesAvailable := 0

	for _, r := range results {
		var status, current, available string
		if r.err != nil {
			status = output.RenderErrorState()
			current = output.LabelPlaceholder
			available = r.err.Error()
		} else {
			current = r.info.Current
			if r.info.HasUpdate {
				status = output.RenderBoolState(true, "update available", "")
				available = r.info.Available
				updatesAvailable++
			} else {
				status = output.FormatPlaceholder("up to date")
				available = output.LabelPlaceholder
			}
		}
		table.AddRow(r.name, current, available, status)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	ios.Println("")
	if updatesAvailable > 0 {
		ios.Success("%d device(s) have updates available", updatesAvailable)
	} else {
		ios.Info("All devices are up to date")
	}

	return nil
}
