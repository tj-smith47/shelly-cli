// Package dashboard provides the energy dashboard command.
package dashboard

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Data represents aggregated energy dashboard data.
type Data struct {
	Timestamp     time.Time      `json:"timestamp"`
	TotalPower    float64        `json:"total_power_w"`
	TotalEnergy   float64        `json:"total_energy_wh,omitempty"`
	DeviceCount   int            `json:"device_count"`
	OnlineCount   int            `json:"online_count"`
	OfflineCount  int            `json:"offline_count"`
	Devices       []DeviceStatus `json:"devices"`
	EstimatedCost *float64       `json:"estimated_cost,omitempty"`
	CostCurrency  string         `json:"cost_currency,omitempty"`
	CostPerKwh    float64        `json:"cost_per_kwh,omitempty"`
}

// DeviceStatus represents energy status for a single device.
type DeviceStatus struct {
	Device      string           `json:"device"`
	Online      bool             `json:"online"`
	Error       string           `json:"error,omitempty"`
	TotalPower  float64          `json:"total_power_w"`
	TotalEnergy float64          `json:"total_energy_wh,omitempty"`
	Components  []ComponentPower `json:"components,omitempty"`
}

// ComponentPower represents power data for a single component.
type ComponentPower struct {
	Type    string  `json:"type"`
	ID      int     `json:"id"`
	Power   float64 `json:"power_w"`
	Voltage float64 `json:"voltage_v,omitempty"`
	Current float64 `json:"current_a,omitempty"`
	Energy  float64 `json:"energy_wh,omitempty"`
}

// NewCommand creates the energy dashboard command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		devices      []string
		costPerKwh   float64
		costCurrency string
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show energy dashboard for all devices",
		Long: `Display an aggregated energy dashboard showing power consumption across all devices.

Shows total power consumption, per-device breakdown, and optional cost estimation.
By default, queries all registered devices. Use --devices to specify a subset.

Examples:
  # Show dashboard for all registered devices
  shelly energy dashboard

  # Show dashboard for specific devices
  shelly energy dashboard --devices kitchen,living-room,garage

  # Include cost estimation at $0.12 per kWh
  shelly energy dashboard --cost 0.12 --currency USD`,
		Aliases: []string{"dash", "summary"},
		Example: `  # Show dashboard for all registered devices
  shelly energy dashboard

  # Show dashboard for specific devices
  shelly energy dashboard --devices kitchen,living-room

  # Include cost estimation
  shelly energy dashboard --cost 0.15 --currency EUR

  # Output as JSON
  shelly energy dashboard -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, devices, costPerKwh, costCurrency)
		},
	}

	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to include (default: all registered)")
	cmd.Flags().Float64Var(&costPerKwh, "cost", 0, "Cost per kWh for estimation")
	cmd.Flags().StringVar(&costCurrency, "currency", "USD", "Currency for cost display")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, costPerKwh float64, currency string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get device list
	if len(devices) == 0 {
		// Use all registered devices
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices found. Register devices using 'shelly device add' or specify --devices")
		return nil
	}

	// Sort for consistent output
	sort.Strings(devices)

	// Collect data from all devices concurrently
	dashboard := collectDashboardData(ctx, ios, svc, devices)

	// Add cost estimation if configured
	if costPerKwh > 0 {
		dashboard.CostPerKwh = costPerKwh
		dashboard.CostCurrency = currency
		cost := (dashboard.TotalEnergy / 1000) * costPerKwh
		dashboard.EstimatedCost = &cost
	}

	// Output results
	return cmdutil.PrintResult(ios, dashboard, displayDashboard)
}

func collectDashboardData(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, devices []string) Data {
	dashboard := Data{
		Timestamp:   time.Now(),
		DeviceCount: len(devices),
		Devices:     make([]DeviceStatus, len(devices)),
	}

	// Collect status concurrently
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrent requests

	for i, device := range devices {
		idx := i
		dev := device
		g.Go(func() error {
			status := collectDeviceStatus(ctx, svc, dev)
			dashboard.Devices[idx] = status
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("collecting dashboard data", err)
	}

	// Aggregate totals
	for _, dev := range dashboard.Devices {
		if dev.Online {
			dashboard.OnlineCount++
			dashboard.TotalPower += dev.TotalPower
			dashboard.TotalEnergy += dev.TotalEnergy
		} else {
			dashboard.OfflineCount++
		}
	}

	return dashboard
}

func collectDeviceStatus(ctx context.Context, svc *shelly.Service, device string) DeviceStatus {
	status := DeviceStatus{
		Device: device,
		Online: true,
	}

	// Collect each component type
	collectEMComponents(ctx, svc, device, &status)
	collectEM1Components(ctx, svc, device, &status)
	collectPMComponents(ctx, svc, device, &status)
	collectPM1Components(ctx, svc, device, &status)

	// Mark offline if no components found
	if len(status.Components) == 0 {
		if _, pingErr := svc.ListPMComponents(ctx, device); pingErr != nil {
			status.Online = false
			status.Error = "device unreachable"
		}
	}

	return status
}

func collectEMComponents(ctx context.Context, svc *shelly.Service, device string, status *DeviceStatus) {
	emIDs, err := svc.ListEMComponents(ctx, device)
	if err != nil {
		return
	}
	for _, id := range emIDs {
		emStatus, err := svc.GetEMStatus(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentPower{
			Type:    "EM",
			ID:      id,
			Power:   emStatus.TotalActivePower,
			Voltage: emStatus.AVoltage,
			Current: emStatus.TotalCurrent,
		}
		status.Components = append(status.Components, comp)
		status.TotalPower += emStatus.TotalActivePower
	}
}

func collectEM1Components(ctx context.Context, svc *shelly.Service, device string, status *DeviceStatus) {
	em1IDs, err := svc.ListEM1Components(ctx, device)
	if err != nil {
		return
	}
	for _, id := range em1IDs {
		em1Status, err := svc.GetEM1Status(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentPower{
			Type:    "EM1",
			ID:      id,
			Power:   em1Status.ActPower,
			Voltage: em1Status.Voltage,
			Current: em1Status.Current,
		}
		status.Components = append(status.Components, comp)
		status.TotalPower += em1Status.ActPower
	}
}

func collectPMComponents(ctx context.Context, svc *shelly.Service, device string, status *DeviceStatus) {
	pmIDs, err := svc.ListPMComponents(ctx, device)
	if err != nil {
		return
	}
	for _, id := range pmIDs {
		pmStatus, err := svc.GetPMStatus(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentPower{
			Type:    "PM",
			ID:      id,
			Power:   pmStatus.APower,
			Voltage: pmStatus.Voltage,
			Current: pmStatus.Current,
		}
		if pmStatus.AEnergy != nil {
			comp.Energy = pmStatus.AEnergy.Total
			status.TotalEnergy += pmStatus.AEnergy.Total
		}
		status.Components = append(status.Components, comp)
		status.TotalPower += pmStatus.APower
	}
}

func collectPM1Components(ctx context.Context, svc *shelly.Service, device string, status *DeviceStatus) {
	pm1IDs, err := svc.ListPM1Components(ctx, device)
	if err != nil {
		return
	}
	for _, id := range pm1IDs {
		pm1Status, err := svc.GetPM1Status(ctx, device, id)
		if err != nil {
			continue
		}
		comp := ComponentPower{
			Type:    "PM1",
			ID:      id,
			Power:   pm1Status.APower,
			Voltage: pm1Status.Voltage,
			Current: pm1Status.Current,
		}
		if pm1Status.AEnergy != nil {
			comp.Energy = pm1Status.AEnergy.Total
			status.TotalEnergy += pm1Status.AEnergy.Total
		}
		status.Components = append(status.Components, comp)
		status.TotalPower += pm1Status.APower
	}
}

func displayDashboard(ios *iostreams.IOStreams, data Data) {
	// Header
	ios.Printf("%s\n", theme.Bold().Render("Energy Dashboard"))
	ios.Printf("Timestamp: %s\n\n", data.Timestamp.Format(time.RFC3339))

	// Summary section
	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Devices:     %d total (%d online, %d offline)\n",
		data.DeviceCount, data.OnlineCount, data.OfflineCount)
	ios.Printf("  Total Power: %s\n", theme.StyledPower(data.TotalPower))

	if data.TotalEnergy > 0 {
		ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy))
	}

	if data.EstimatedCost != nil {
		ios.Printf("  Est. Cost:   %s %.2f/kWh = %s %.4f\n",
			data.CostCurrency, data.CostPerKwh,
			data.CostCurrency, *data.EstimatedCost)
	}

	ios.Printf("\n")

	// Device breakdown table
	ios.Printf("%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Device", "Status", "Power", "Components")

	for _, dev := range data.Devices {
		statusStr := theme.StatusOK().Render("online")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		}

		powerStr := output.FormatPower(dev.TotalPower)
		if !dev.Online {
			powerStr = "-"
		}

		compStr := formatComponentSummary(dev.Components)

		table.AddRow(dev.Device, statusStr, powerStr, compStr)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}

func formatComponentSummary(components []ComponentPower) string {
	if len(components) == 0 {
		return "-"
	}

	counts := make(map[string]int)
	for _, c := range components {
		counts[c.Type]++
	}

	parts := make([]string, 0, len(counts))
	for typ, count := range counts {
		parts = append(parts, fmt.Sprintf("%d %s", count, typ))
	}

	return fmt.Sprintf("%d (%s)", len(components), strings.Join(parts, ", "))
}
