package term

import (
	"strings"
	"testing"
	"time"

	"github.com/tj-smith47/shelly-go/integrator"
)

func TestDisplayFleetStatus(t *testing.T) {
	t.Parallel()

	t.Run("with devices", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		statuses := []*integrator.DeviceStatus{
			{
				DeviceID: "device1",
				Host:     "192.168.1.100",
				Online:   true,
				LastSeen: time.Now().Add(-5 * time.Minute),
			},
			{
				DeviceID: "device2",
				Host:     "192.168.1.101",
				Online:   false,
				LastSeen: time.Now().Add(-1 * time.Hour),
			},
		}

		DisplayFleetStatus(ios, statuses)

		output := out.String()
		if !strings.Contains(output, "device1") {
			t.Error("output should contain device1")
		}
		if !strings.Contains(output, "device2") {
			t.Error("output should contain device2")
		}
		if !strings.Contains(output, "192.168.1.100") {
			t.Error("output should contain host")
		}
	})

	t.Run("empty devices", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()

		DisplayFleetStatus(ios, []*integrator.DeviceStatus{})

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No devices found") {
			t.Errorf("output should contain 'No devices found', got %q", allOutput)
		}
	})
}

func TestDisplayFleetHealth(t *testing.T) {
	t.Parallel()

	t.Run("with health data", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		health := []*integrator.DeviceHealth{
			{
				DeviceID:      "device1",
				Online:        true,
				LastSeen:      time.Now().Add(-5 * time.Minute),
				OnlineCount:   100,
				OfflineCount:  2,
				ActivityCount: 50,
			},
			{
				DeviceID:      "device2",
				Online:        false,
				LastSeen:      time.Now().Add(-1 * time.Hour),
				OnlineCount:   10,
				OfflineCount:  20,
				ActivityCount: 5,
			},
		}

		DisplayFleetHealth(ios, health, 30*time.Minute)

		output := out.String()
		if !strings.Contains(output, "device1") {
			t.Error("output should contain device1")
		}
		if !strings.Contains(output, "Summary:") {
			t.Error("output should contain 'Summary:'")
		}
	})

	t.Run("empty health data", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()

		DisplayFleetHealth(ios, []*integrator.DeviceHealth{}, 30*time.Minute)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No health data") {
			t.Errorf("output should contain 'No health data', got %q", allOutput)
		}
	})

	t.Run("warning status", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		health := []*integrator.DeviceHealth{
			{
				DeviceID:     "device1",
				Online:       true,
				LastSeen:     time.Now().Add(-5 * time.Minute),
				OnlineCount:  10,
				OfflineCount: 6, // More than half of online
			},
		}

		DisplayFleetHealth(ios, health, 30*time.Minute)

		output := out.String()
		if !strings.Contains(output, "warning") {
			t.Error("output should contain warning count in summary")
		}
	})
}

func TestDisplayFleetStats(t *testing.T) {
	t.Parallel()

	t.Run("with stats", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		stats := &integrator.FleetStats{
			TotalDevices:     10,
			OnlineDevices:    8,
			OfflineDevices:   2,
			TotalConnections: 15,
			TotalGroups:      3,
		}

		DisplayFleetStats(ios, stats)

		output := out.String()
		if !strings.Contains(output, "Total Devices:") {
			t.Error("output should contain 'Total Devices:'")
		}
		if !strings.Contains(output, "10") {
			t.Error("output should contain device count")
		}
		if !strings.Contains(output, "Online:") {
			t.Error("output should contain 'Online:'")
		}
		if !strings.Contains(output, "8") {
			t.Error("output should contain online count")
		}
	})

	t.Run("nil stats", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()

		DisplayFleetStats(ios, nil)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No statistics available") {
			t.Errorf("output should contain 'No statistics available', got %q", allOutput)
		}
	})
}

func TestDetermineHealthStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		health    *integrator.DeviceHealth
		threshold time.Duration
		want      string
	}{
		{
			name: "healthy device",
			health: &integrator.DeviceHealth{
				Online:       true,
				LastSeen:     time.Now().Add(-5 * time.Minute),
				OnlineCount:  100,
				OfflineCount: 5,
			},
			threshold: 30 * time.Minute,
			want:      healthStatusHealthy,
		},
		{
			name: "offline device",
			health: &integrator.DeviceHealth{
				Online:   false,
				LastSeen: time.Now().Add(-1 * time.Hour),
			},
			threshold: 30 * time.Minute,
			want:      healthStatusUnhealthy,
		},
		{
			name: "high offline count",
			health: &integrator.DeviceHealth{
				Online:       true,
				LastSeen:     time.Now().Add(-5 * time.Minute),
				OnlineCount:  10,
				OfflineCount: 6,
			},
			threshold: 30 * time.Minute,
			want:      healthStatusWarning,
		},
		{
			name: "not seen within threshold",
			health: &integrator.DeviceHealth{
				Online:       true,
				LastSeen:     time.Now().Add(-1 * time.Hour),
				OnlineCount:  100,
				OfflineCount: 5,
			},
			threshold: 30 * time.Minute,
			want:      healthStatusWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := determineHealthStatus(tt.health, tt.threshold)
			if got != tt.want {
				t.Errorf("determineHealthStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTimeSince(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "zero time",
			time: time.Time{},
			want: "never",
		},
		{
			name: "just now",
			time: time.Now().Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "minutes ago",
			time: time.Now().Add(-5 * time.Minute),
			want: "minutes ago",
		},
		{
			name: "1 minute ago",
			time: time.Now().Add(-1 * time.Minute),
			want: "minute ago",
		},
		{
			name: "hours ago",
			time: time.Now().Add(-3 * time.Hour),
			want: "hours ago",
		},
		{
			name: "1 hour ago",
			time: time.Now().Add(-1 * time.Hour),
			want: "hour ago",
		},
		{
			name: "days ago",
			time: time.Now().Add(-48 * time.Hour),
			want: "days ago",
		},
		{
			name: "1 day ago",
			time: time.Now().Add(-24 * time.Hour),
			want: "day ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatTimeSince(tt.time)
			if !strings.Contains(got, tt.want) {
				t.Errorf("formatTimeSince() = %v, want to contain %v", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "less than minute",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "1 minute",
			duration: 1 * time.Minute,
			want:     "1 minute ago",
		},
		{
			name:     "5 minutes",
			duration: 5 * time.Minute,
			want:     "5 minutes ago",
		},
		{
			name:     "1 hour",
			duration: 1 * time.Hour,
			want:     "1 hour ago",
		},
		{
			name:     "5 hours",
			duration: 5 * time.Hour,
			want:     "5 hours ago",
		},
		{
			name:     "1 day",
			duration: 24 * time.Hour,
			want:     "1 day ago",
		},
		{
			name:     "3 days",
			duration: 72 * time.Hour,
			want:     "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"positive", 42, "42"},
		{"large", 1000, "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatInt(tt.n)
			if got != tt.want {
				t.Errorf("formatInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
