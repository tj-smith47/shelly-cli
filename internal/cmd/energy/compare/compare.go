// Package compare provides the energy compare command.
package compare

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tj-smith47/shelly-go/gen2/components"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// ComparisonData represents energy comparison results.
type ComparisonData struct {
	Period      string         `json:"period"`
	From        time.Time      `json:"from"`
	To          time.Time      `json:"to"`
	Devices     []DeviceEnergy `json:"devices"`
	TotalEnergy float64        `json:"total_energy_kwh"`
	MaxEnergy   float64        `json:"max_energy_kwh"`
	MinEnergy   float64        `json:"min_energy_kwh"`
}

// DeviceEnergy represents energy data for a single device.
type DeviceEnergy struct {
	Device     string  `json:"device"`
	Energy     float64 `json:"energy_kwh"`
	AvgPower   float64 `json:"avg_power_w"`
	PeakPower  float64 `json:"peak_power_w"`
	DataPoints int     `json:"data_points"`
	Online     bool    `json:"online"`
	Error      string  `json:"error,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

// NewCommand creates the energy compare command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		devices []string
		period  string
		from    string
		to      string
	)

	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare energy usage between devices",
		Long: `Compare energy consumption across multiple devices for a specified time period.

Shows each device's total energy consumption, average power, and percentage
of the total consumption. Useful for identifying high-energy consumers.

By default, compares all registered devices. Use --devices to specify a subset.`,
		Example: `  # Compare all devices for the last day
  shelly energy compare

  # Compare specific devices for the last week
  shelly energy compare --devices kitchen,living-room,garage --period week

  # Compare for a specific date range
  shelly energy compare --from "2025-01-01" --to "2025-01-07"

  # Output as JSON
  shelly energy compare -o json`,
		Aliases: []string{"cmp", "diff"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, devices, period, from, to)
		},
	}

	cmd.Flags().StringSliceVar(&devices, "devices", nil, "Devices to compare (default: all registered)")
	cmd.Flags().StringVarP(&period, "period", "p", "day", "Time period (hour, day, week, month)")
	cmd.Flags().StringVar(&from, "from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	cmd.Flags().StringVar(&to, "to", "", "End time (RFC3339 or YYYY-MM-DD)")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, devices []string, period, from, to string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()
	cfg, err := f.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get device list
	if len(devices) == 0 {
		for name := range cfg.Devices {
			devices = append(devices, name)
		}
	}

	if len(devices) == 0 {
		ios.Warning("No devices found. Register devices using 'shelly device add' or specify --devices")
		return nil
	}

	if len(devices) < 2 {
		ios.Warning("At least 2 devices are required for comparison. Found: %d", len(devices))
		return nil
	}

	sort.Strings(devices)

	// Calculate time range
	startTS, endTS, err := shelly.CalculateTimeRange(period, from, to)
	if err != nil {
		return fmt.Errorf("invalid time range: %w", err)
	}

	// Collect comparison data
	comparison := collectComparisonData(ctx, ios, svc, devices, period, startTS, endTS)

	// Calculate percentages
	if comparison.TotalEnergy > 0 {
		for i := range comparison.Devices {
			comparison.Devices[i].Percentage = (comparison.Devices[i].Energy / comparison.TotalEnergy) * 100
		}
	}

	// Output results
	return cmdutil.PrintResult(ios, comparison, displayComparison)
}

func collectComparisonData(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, devices []string, period string, startTS, endTS *int64) ComparisonData {
	comparison := ComparisonData{
		Period:  period,
		Devices: make([]DeviceEnergy, len(devices)),
	}

	if startTS != nil {
		comparison.From = time.Unix(*startTS, 0)
	}
	if endTS != nil {
		comparison.To = time.Unix(*endTS, 0)
	}

	// Collect data concurrently
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for i, device := range devices {
		idx := i
		dev := device
		g.Go(func() error {
			energy := collectDeviceEnergy(ctx, svc, dev, startTS, endTS)
			comparison.Devices[idx] = energy
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		ios.DebugErr("collecting comparison data", err)
	}

	// Calculate totals and find min/max
	comparison.MinEnergy = -1
	for _, dev := range comparison.Devices {
		if dev.Online {
			comparison.TotalEnergy += dev.Energy
			if dev.Energy > comparison.MaxEnergy {
				comparison.MaxEnergy = dev.Energy
			}
			if comparison.MinEnergy < 0 || dev.Energy < comparison.MinEnergy {
				comparison.MinEnergy = dev.Energy
			}
		}
	}
	if comparison.MinEnergy < 0 {
		comparison.MinEnergy = 0
	}

	return comparison
}

func collectDeviceEnergy(ctx context.Context, svc *shelly.Service, device string, startTS, endTS *int64) DeviceEnergy {
	result := DeviceEnergy{
		Device: device,
		Online: true,
	}

	// Try EM data first
	emData, err := svc.GetEMDataHistory(ctx, device, 0, startTS, endTS)
	if err == nil && emData != nil && len(emData.Data) > 0 {
		result.Energy, result.AvgPower, result.PeakPower, result.DataPoints = calculateEMMetrics(emData)
		return result
	}

	// Try EM1 data
	em1Data, err := svc.GetEM1DataHistory(ctx, device, 0, startTS, endTS)
	if err == nil && em1Data != nil && len(em1Data.Data) > 0 {
		result.Energy, result.AvgPower, result.PeakPower, result.DataPoints = calculateEM1Metrics(em1Data)
		return result
	}

	// Try to get current power from PM/PM1 for devices without history
	currentPower := collectCurrentPower(ctx, svc, device)
	if currentPower > 0 {
		// No historical data, but device is online with current reading
		result.AvgPower = currentPower
		result.PeakPower = currentPower
		result.DataPoints = 1
		result.Error = "no historical data"
		return result
	}

	// Mark offline if no data available
	result.Online = false
	result.Error = "no data available"
	return result
}

func collectCurrentPower(ctx context.Context, svc *shelly.Service, device string) float64 {
	var totalPower float64

	// Try PM
	pmIDs, err := svc.ListPMComponents(ctx, device)
	if err == nil {
		for _, id := range pmIDs {
			if status, err := svc.GetPMStatus(ctx, device, id); err == nil {
				totalPower += status.APower
			}
		}
	}

	// Try PM1
	pm1IDs, err := svc.ListPM1Components(ctx, device)
	if err == nil {
		for _, id := range pm1IDs {
			if status, err := svc.GetPM1Status(ctx, device, id); err == nil {
				totalPower += status.APower
			}
		}
	}

	return totalPower
}

func calculateEMMetrics(data *components.EMDataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	var totalPower float64

	for _, block := range data.Data {
		for _, values := range block.Values {
			totalPower += values.TotalActivePower
			if values.TotalActivePower > peakPower {
				peakPower = values.TotalActivePower
			}
			// Energy = Power (W) * Time (s) / 3600 = Wh
			energy += values.TotalActivePower * float64(block.Period) / 3600.0
			dataPoints++
		}
	}

	if dataPoints > 0 {
		avgPower = totalPower / float64(dataPoints)
	}

	// Convert Wh to kWh
	energy /= 1000.0
	return
}

func calculateEM1Metrics(data *components.EM1DataGetDataResult) (energy, avgPower, peakPower float64, dataPoints int) {
	var totalPower float64

	for _, block := range data.Data {
		for _, values := range block.Values {
			totalPower += values.ActivePower
			if values.ActivePower > peakPower {
				peakPower = values.ActivePower
			}
			// Energy = Power (W) * Time (s) / 3600 = Wh
			energy += values.ActivePower * float64(block.Period) / 3600.0
			dataPoints++
		}
	}

	if dataPoints > 0 {
		avgPower = totalPower / float64(dataPoints)
	}

	// Convert Wh to kWh
	energy /= 1000.0
	return
}

func displayComparison(ios *iostreams.IOStreams, data ComparisonData) {
	// Header
	ios.Printf("%s\n", theme.Bold().Render("Energy Comparison"))
	ios.Printf("Period: %s\n", data.Period)
	if !data.From.IsZero() {
		ios.Printf("From:   %s\n", data.From.Format("2006-01-02 15:04:05"))
	}
	if !data.To.IsZero() {
		ios.Printf("To:     %s\n", data.To.Format("2006-01-02 15:04:05"))
	}
	ios.Printf("\n")

	// Summary
	ios.Printf("%s\n", theme.Bold().Render("Summary"))
	ios.Printf("  Total Energy: %s\n", theme.StyledEnergy(data.TotalEnergy*1000))
	ios.Printf("  Max Device:   %s\n", theme.StyledEnergy(data.MaxEnergy*1000))
	ios.Printf("  Min Device:   %s\n", theme.StyledEnergy(data.MinEnergy*1000))
	ios.Printf("\n")

	// Sort by energy consumption (descending)
	sorted := make([]DeviceEnergy, len(data.Devices))
	copy(sorted, data.Devices)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Energy > sorted[j].Energy
	})

	// Device comparison table
	ios.Printf("%s\n", theme.Bold().Render("Device Breakdown"))

	table := output.NewTable("Rank", "Device", "Energy", "Avg Power", "Peak Power", "Share", "Status")

	for i, dev := range sorted {
		rank := fmt.Sprintf("#%d", i+1)

		statusStr := theme.StatusOK().Render("ok")
		if !dev.Online {
			statusStr = theme.StatusError().Render("offline")
		} else if dev.Error != "" {
			statusStr = theme.StatusWarn().Render(truncate(dev.Error, 15))
		}

		energyStr := "-"
		avgStr := "-"
		peakStr := "-"
		shareStr := "-"

		if dev.Online {
			energyStr = output.FormatEnergy(dev.Energy * 1000) // dev.Energy is kWh, FormatEnergy takes Wh
			avgStr = output.FormatPower(dev.AvgPower)
			peakStr = output.FormatPower(dev.PeakPower)
			if dev.Percentage > 0 {
				shareStr = fmt.Sprintf("%.1f%%", dev.Percentage)
			}
		}

		table.AddRow(rank, dev.Device, energyStr, avgStr, peakStr, shareStr, statusStr)
	}

	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}

	// Show visual bar chart
	ios.Printf("\n%s\n", theme.Bold().Render("Energy Distribution"))
	displayBarChart(ios, sorted, data.MaxEnergy)
}

func displayBarChart(ios *iostreams.IOStreams, devices []DeviceEnergy, maxEnergy float64) {
	if maxEnergy <= 0 {
		return
	}

	maxNameLen := 0
	for _, dev := range devices {
		if len(dev.Device) > maxNameLen {
			maxNameLen = len(dev.Device)
		}
	}

	barWidth := 40

	for _, dev := range devices {
		if !dev.Online || dev.Energy <= 0 {
			continue
		}

		name := padRight(dev.Device, maxNameLen)
		barLen := int((dev.Energy / maxEnergy) * float64(barWidth))
		if barLen < 1 {
			barLen = 1
		}

		bar := strings.Repeat("█", barLen)
		ios.Printf("  %s │ %s %.2f kWh\n", name, theme.Highlight().Render(bar), dev.Energy)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
