// Package updates provides the firmware updates subcommand.
// This command provides an interactive workflow for discovering and applying
// firmware updates across all devices, with support for both Shelly and plugin-managed platforms.
package updates

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Release stage constants.
const (
	stageStable = "stable"
	stageBeta   = "beta"
)

// NewCommand creates the firmware updates command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		allFlag      bool
		stableFlag   bool
		betaFlag     bool
		devicesFlag  string
		platformFlag string
		yesFlag      bool
		parallelFlag int
	)

	cmd := &cobra.Command{
		Use:     "updates",
		Aliases: []string{"upd"},
		Short:   "Interactive firmware update workflow",
		Long: `Check for and apply firmware updates with an interactive workflow.

By default, runs in interactive mode - displays devices with available updates
and prompts for selection. Use --all with --yes for non-interactive batch updates.

Supports both native Shelly devices and plugin-managed devices (Tasmota, etc.).`,
		Example: `  # Interactive mode - check and select devices
  shelly firmware updates

  # Update all devices to stable (non-interactive)
  shelly firmware updates --all --yes

  # Update all devices to beta (non-interactive)
  shelly firmware updates --all --beta --yes

  # Update specific devices to stable
  shelly firmware updates --devices=kitchen,bedroom --stable

  # Update only Tasmota devices
  shelly firmware updates --all --platform=tasmota --yes

  # Update specific devices interactively
  shelly firmware updates --devices=kitchen,bedroom`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := runOptions{
				all:      allFlag,
				stable:   stableFlag,
				beta:     betaFlag,
				devices:  devicesFlag,
				platform: platformFlag,
				yes:      yesFlag,
				parallel: parallelFlag,
			}
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().BoolVar(&allFlag, "all", false, "Update all devices with available updates")
	cmd.Flags().BoolVar(&stableFlag, "stable", false, "Use stable release channel (default)")
	cmd.Flags().BoolVar(&betaFlag, "beta", false, "Use beta/development release channel")
	cmd.Flags().StringVar(&devicesFlag, "devices", "", "Comma-separated list of specific devices to update")
	cmd.Flags().StringVar(&platformFlag, "platform", "", "Only update devices of this platform (e.g., tasmota)")
	cmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().IntVar(&parallelFlag, "parallel", 3, "Number of devices to update in parallel")

	return cmd
}

type runOptions struct {
	all      bool
	stable   bool
	beta     bool
	devices  string
	platform string
	yes      bool
	parallel int
}

// deviceUpdateInfo holds information about a device's firmware update status.
type deviceUpdateInfo struct {
	Name        string
	Device      model.Device
	FwInfo      *shelly.FirmwareInfo
	HasUpdate   bool
	HasBeta     bool
	UpdateError error
}

func run(ctx context.Context, f *cmdutil.Factory, opts runOptions) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Load and filter devices
	deviceConfigs, err := loadAndFilterDevices(ios, opts)
	if err != nil {
		return err
	}
	if len(deviceConfigs) == 0 {
		return nil // Warning already shown
	}

	// Check all devices for updates using platform-aware checking
	results := svc.CheckFirmwareAllPlatforms(ctx, ios, deviceConfigs)

	// Build list of devices with updates
	devicesWithUpdates := buildUpdateList(results, deviceConfigs)
	if len(devicesWithUpdates) == 0 {
		ios.Info("All devices are up to date")
		return nil
	}

	// Display devices with available updates
	displayUpdatesTable(ios, devicesWithUpdates)

	// Determine which devices to update and what stage
	toUpdate, stage := selectDevicesForUpdate(ios, devicesWithUpdates, opts)
	if len(toUpdate) == 0 {
		ios.Info("No devices selected for update")
		return nil
	}

	// Confirm and execute updates
	return confirmAndExecuteUpdates(ctx, ios, svc, toUpdate, stage, opts)
}

// loadAndFilterDevices loads config and filters devices based on options.
// Returns empty map (not nil) with no error if no devices match; caller should check len().
func loadAndFilterDevices(ios *iostreams.IOStreams, opts runOptions) (map[string]model.Device, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Devices) == 0 {
		ios.Warning("No devices registered. Use 'shelly device add' to add devices.")
		return make(map[string]model.Device), nil
	}

	deviceConfigs := filterDevices(cfg.Devices, opts.devices, opts.platform)
	if len(deviceConfigs) == 0 {
		ios.Warning("No devices matched the specified filters")
	}

	return deviceConfigs, nil
}

// buildUpdateList builds a sorted list of devices that have updates available.
func buildUpdateList(results []shelly.FirmwareCheckResult, deviceConfigs map[string]model.Device) []deviceUpdateInfo {
	var devicesWithUpdates []deviceUpdateInfo
	for _, r := range results {
		device := deviceConfigs[r.Name]
		info := deviceUpdateInfo{
			Name:   r.Name,
			Device: device,
			FwInfo: r.Info,
		}
		if r.Err != nil {
			info.UpdateError = r.Err
		} else if r.Info != nil {
			info.HasUpdate = r.Info.HasUpdate
			info.HasBeta = r.Info.Beta != "" && r.Info.Beta != r.Info.Current
		}
		if info.HasUpdate || info.HasBeta {
			devicesWithUpdates = append(devicesWithUpdates, info)
		}
	}

	// Sort by device name
	sort.Slice(devicesWithUpdates, func(i, j int) bool {
		return devicesWithUpdates[i].Name < devicesWithUpdates[j].Name
	})

	return devicesWithUpdates
}

// confirmAndExecuteUpdates confirms with user and executes the updates.
func confirmAndExecuteUpdates(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, toUpdate []deviceUpdateInfo, stage string, opts runOptions) error {
	// Confirm if not --yes
	if !opts.yes {
		msg := fmt.Sprintf("Update %d device(s) to %s?", len(toUpdate), stage)
		confirmed, confirmErr := ios.Confirm(msg, false)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			ios.Warning("Update cancelled")
			return nil
		}
	}

	// Convert to DeviceUpdateStatus for the service
	updateStatuses := make([]shelly.DeviceUpdateStatus, len(toUpdate))
	for i, d := range toUpdate {
		updateStatuses[i] = shelly.DeviceUpdateStatus{
			Name:      d.Name,
			Info:      d.FwInfo,
			HasUpdate: d.HasUpdate,
		}
	}

	// Perform updates
	updateResults := svc.UpdateDevices(ctx, ios, updateStatuses, shelly.UpdateOpts{
		Beta:        stage == stageBeta,
		Parallelism: opts.parallel,
	})

	// Display results
	termResults := make([]term.UpdateResult, len(updateResults))
	for i, r := range updateResults {
		termResults[i] = term.UpdateResult{
			Name:    r.Name,
			Success: r.Success,
			Err:     r.Err,
		}
	}
	term.DisplayUpdateResults(ios, termResults)

	// Exit code based on results
	for _, r := range updateResults {
		if !r.Success {
			return fmt.Errorf("some updates failed")
		}
	}
	return nil
}

// filterDevices filters devices based on --devices and --platform flags.
func filterDevices(devices map[string]model.Device, devicesList, platform string) map[string]model.Device {
	result := make(map[string]model.Device)

	// If --devices is specified, filter by names
	var selectedNames map[string]bool
	if devicesList != "" {
		selectedNames = make(map[string]bool)
		for _, name := range strings.Split(devicesList, ",") {
			selectedNames[strings.TrimSpace(name)] = true
		}
	}

	for name, device := range devices {
		// Filter by name if --devices specified
		if selectedNames != nil && !selectedNames[name] {
			continue
		}

		// Filter by platform if --platform specified
		if platform != "" {
			devicePlatform := device.Platform
			if devicePlatform == "" {
				devicePlatform = "shelly"
			}
			if devicePlatform != platform {
				continue
			}
		}

		result[name] = device
	}

	return result
}

// displayUpdatesTable shows a table of devices with available updates.
func displayUpdatesTable(ios *iostreams.IOStreams, devices []deviceUpdateInfo) {
	ios.Println("")
	ios.Printf("%s\n\n", theme.Bold().Render("Devices with available updates:"))

	table := output.NewTable("#", "Device", "Platform", "Current", "Stable", "Beta")
	for i, d := range devices {
		platform := d.FwInfo.Platform
		if platform == "" {
			platform = "shelly"
		}

		stable := d.FwInfo.Available
		if stable == "" {
			stable = output.LabelPlaceholder
		}

		beta := d.FwInfo.Beta
		if beta == "" {
			beta = output.LabelPlaceholder
		}

		table.AddRow(
			fmt.Sprintf("%d", i+1),
			d.Name,
			platform,
			d.FwInfo.Current,
			stable,
			beta,
		)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
	ios.Println("")
}

// selectDevicesForUpdate determines which devices to update based on options.
// Returns the devices to update and the release stage (stageStable or stageBeta).
func selectDevicesForUpdate(ios *iostreams.IOStreams, devices []deviceUpdateInfo, opts runOptions) (selected []deviceUpdateInfo, stage string) {
	// Determine release stage
	stage = stageStable
	if opts.beta {
		stage = stageBeta
	}

	// Non-interactive mode: --all or --devices specified
	if opts.all || opts.devices != "" {
		return filterDevicesByStage(devices, stage), stage
	}

	// Interactive mode
	return interactiveDeviceSelect(ios, devices, stage)
}

// filterDevicesByStage filters devices based on the requested stage.
func filterDevicesByStage(devices []deviceUpdateInfo, stage string) []deviceUpdateInfo {
	var result []deviceUpdateInfo
	for _, d := range devices {
		if stage == stageBeta && d.HasBeta {
			result = append(result, d)
		} else if d.HasUpdate {
			result = append(result, d)
		}
	}
	return result
}

// interactiveDeviceSelect prompts user to select devices interactively.
func interactiveDeviceSelect(ios *iostreams.IOStreams, devices []deviceUpdateInfo, stage string) (selected []deviceUpdateInfo, selectedStage string) {
	ios.Println("Options:")
	ios.Println("  [a] Update all to stable")
	if anyHasBeta(devices) {
		ios.Println("  [b] Update all to beta")
	}
	ios.Println("  [s] Select specific devices")
	ios.Println("  [q] Quit")
	ios.Println("")

	choice, err := ios.Input("Choose option", "a")
	if err != nil {
		return nil, stage
	}

	return handleInteractiveChoice(ios, devices, choice, stage)
}

// handleInteractiveChoice processes the user's interactive menu choice.
func handleInteractiveChoice(ios *iostreams.IOStreams, devices []deviceUpdateInfo, choice, stage string) (selected []deviceUpdateInfo, selectedStage string) {
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "a":
		return filterDevicesByStage(devices, stageStable), stageStable
	case "b":
		return filterDevicesByStage(devices, stageBeta), stageBeta
	case "s":
		return selectSpecificDevices(ios, devices)
	case "q", "":
		return nil, stage
	default:
		return nil, stage
	}
}

// selectSpecificDevices prompts user to select individual devices.
func selectSpecificDevices(ios *iostreams.IOStreams, devices []deviceUpdateInfo) (selected []deviceUpdateInfo, stage string) {
	ios.Println("")
	ios.Println("Enter device numbers separated by commas (e.g., 1,3,5)")
	ios.Println("Or enter 'all' for all devices")
	ios.Println("")

	input, err := ios.Input("Devices", "all")
	if err != nil {
		return nil, stageStable
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return nil, stageStable
	}

	if strings.EqualFold(input, "all") {
		for _, d := range devices {
			if d.HasUpdate {
				selected = append(selected, d)
			}
		}
		return selected, stageStable
	}

	// Parse device numbers
	for _, part := range strings.Split(input, ",") {
		var num int
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(part), "%d", &num); scanErr == nil {
			if num >= 1 && num <= len(devices) {
				selected = append(selected, devices[num-1])
			}
		}
	}

	if len(selected) == 0 {
		ios.Warning("No valid devices selected")
		return nil, stageStable
	}

	// Ask for stage
	ios.Println("")
	stageChoice, stageErr := ios.Input("Release channel (stable/beta)", stageStable)
	if stageErr != nil {
		return selected, stageStable
	}

	stage = stageStable
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(stageChoice)), "b") {
		stage = stageBeta
	}

	return selected, stage
}

// anyHasBeta returns true if any device has a beta update available.
func anyHasBeta(devices []deviceUpdateInfo) bool {
	for _, d := range devices {
		if d.HasBeta {
			return true
		}
	}
	return false
}
