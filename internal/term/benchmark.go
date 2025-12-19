// Package term provides terminal display functions.
package term

import (
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// DisplayLatencyStats prints latency statistics to the terminal.
func DisplayLatencyStats(ios *iostreams.IOStreams, stats model.LatencyStats) {
	ios.Printf("  Min: %v\n", stats.Min)
	ios.Printf("  Max: %v\n", stats.Max)
	ios.Printf("  Avg: %v\n", stats.Avg)
	ios.Printf("  P50: %v\n", stats.P50)
	ios.Printf("  P95: %v\n", stats.P95)
	ios.Printf("  P99: %v\n", stats.P99)
	if stats.Errors > 0 {
		ios.Printf("  Errors: %d\n", stats.Errors)
	}
}
