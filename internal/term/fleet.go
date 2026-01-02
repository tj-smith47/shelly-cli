// Package term provides terminal display functions for the CLI.
package term

import (
	"fmt"
	"time"

	"github.com/tj-smith47/shelly-go/integrator"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output/table"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Health status constants.
const (
	healthStatusHealthy   = "healthy"
	healthStatusWarning   = "warning"
	healthStatusUnhealthy = "unhealthy"
)

// DisplayFleetStatus displays the status of devices in the fleet.
func DisplayFleetStatus(ios *iostreams.IOStreams, statuses []*integrator.DeviceStatus) {
	if len(statuses) == 0 {
		ios.Info("No devices found")
		return
	}

	builder := table.NewBuilder("STATUS", "DEVICE", "HOST", "LAST SEEN")
	for _, s := range statuses {
		status := theme.StatusError().Render("○")
		if s.Online {
			status = theme.StatusOK().Render("●")
		}

		lastSeen := formatTimeSince(s.LastSeen)
		builder.AddRow(status, s.DeviceID, s.Host, lastSeen)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print fleet status table", err)
	}
}

// DisplayFleetHealth displays health information for fleet devices.
// The threshold parameter specifies how long since last seen before a device is considered unhealthy.
func DisplayFleetHealth(ios *iostreams.IOStreams, health []*integrator.DeviceHealth, threshold time.Duration) {
	if len(health) == 0 {
		ios.Info("No health data available")
		return
	}

	var healthy, warning, unhealthy int
	for _, h := range health {
		status := determineHealthStatus(h, threshold)
		switch status {
		case healthStatusHealthy:
			healthy++
		case healthStatusWarning:
			warning++
		case healthStatusUnhealthy:
			unhealthy++
		}
	}

	ios.Printf("Summary: %d healthy, %d warning, %d unhealthy\n\n",
		healthy, warning, unhealthy)

	builder := table.NewBuilder("STATUS", "DEVICE", "ONLINE", "LAST SEEN", "ACTIVITY")
	for _, h := range health {
		status := determineHealthStatus(h, threshold)
		var statusIcon string
		switch status {
		case healthStatusHealthy:
			statusIcon = theme.StatusOK().Render("✓")
		case healthStatusWarning:
			statusIcon = theme.StatusWarn().Render("!")
		case healthStatusUnhealthy:
			statusIcon = theme.StatusError().Render("✗")
		}

		online := "no"
		if h.Online {
			online = "yes"
		}

		builder.AddRow(
			statusIcon,
			h.DeviceID,
			online,
			formatTimeSince(h.LastSeen),
			fmt.Sprintf("%d", h.ActivityCount),
		)
	}

	tbl := builder.WithModeStyle(ios).Build()
	if err := tbl.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print fleet health table", err)
	}
}

// DisplayFleetStats displays aggregate fleet statistics.
func DisplayFleetStats(ios *iostreams.IOStreams, stats *integrator.FleetStats) {
	if stats == nil {
		ios.Info("No statistics available")
		return
	}

	ios.Printf("Total Devices:  %d\n", stats.TotalDevices)
	ios.Printf("  Online:       %d\n", stats.OnlineDevices)
	ios.Printf("  Offline:      %d\n", stats.OfflineDevices)
	ios.Printf("Connections:    %d\n", stats.TotalConnections)
	ios.Printf("Groups:         %d\n", stats.TotalGroups)
}

func determineHealthStatus(h *integrator.DeviceHealth, threshold time.Duration) string {
	if !h.Online {
		return healthStatusUnhealthy
	}
	// Consider device unhealthy if offline count exceeds online count
	if h.OfflineCount > h.OnlineCount/2 {
		return healthStatusWarning
	}
	// Consider device unhealthy if not seen within threshold
	if time.Since(h.LastSeen) > threshold {
		return healthStatusWarning
	}
	return healthStatusHealthy
}

func formatTimeSince(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return formatDuration(d.Round(time.Minute))
	case d < 24*time.Hour:
		return formatDuration(d.Round(time.Hour))
	default:
		return formatDuration(d.Round(24 * time.Hour))
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return formatInt(m) + " minutes ago"
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return formatInt(h) + " hours ago"
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return formatInt(days) + " days ago"
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}
