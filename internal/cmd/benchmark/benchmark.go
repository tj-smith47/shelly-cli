// Package benchmark provides the benchmark command for device performance testing.
package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	Iterations int
	Warmup     int
	JSONOutput bool
}

// Result holds benchmark results.
type Result struct {
	Device      string       `json:"device"`
	Iterations  int          `json:"iterations"`
	PingLatency LatencyStats `json:"ping_latency"`
	RPCLatency  LatencyStats `json:"rpc_latency"`
	Summary     string       `json:"summary"`
	Timestamp   time.Time    `json:"timestamp"`
}

// LatencyStats holds latency statistics.
type LatencyStats struct {
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Avg    time.Duration `json:"avg"`
	P50    time.Duration `json:"p50"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
	Errors int           `json:"errors"`
}

// NewCommand creates the benchmark command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Iterations: 10,
		Warmup:     2,
	}

	cmd := &cobra.Command{
		Use:     "benchmark <device>",
		Aliases: []string{"bench", "perf"},
		Short:   "Test device performance",
		Long: `Measure device performance including API latency and response times.

The benchmark runs multiple iterations to collect statistics on:
  - Ping latency (basic connectivity)
  - RPC latency (API call response time)

Results include min, max, average, and percentile statistics (P50, P95, P99).`,
		Example: `  # Basic benchmark (10 iterations)
  shelly benchmark kitchen-light

  # Extended benchmark
  shelly benchmark kitchen-light --iterations 50

  # JSON output for logging
  shelly benchmark kitchen-light -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Iterations, "iterations", "n", 10, "Number of iterations")
	cmd.Flags().IntVar(&opts.Warmup, "warmup", 2, "Number of warmup iterations (not counted)")
	cmd.Flags().BoolVarP(&opts.JSONOutput, "json", "o", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ios.Info("Benchmarking %s (%d iterations + %d warmup)...",
		device, opts.Iterations, opts.Warmup)
	ios.Println("")

	// Get connection for RPC tests
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	defer iostreams.CloseWithDebug("closing benchmark connection", conn)

	// Warmup
	if opts.Warmup > 0 {
		ios.Info("Warming up...")
		for range opts.Warmup {
			if _, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil); err != nil {
				ios.DebugErr("warmup call", err)
			}
		}
	}

	// Benchmark RPC calls
	ios.Info("Running RPC benchmark...")
	rpcLatencies := make([]time.Duration, 0, opts.Iterations)
	rpcErrors := 0

	for i := range opts.Iterations {
		start := time.Now()
		_, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		elapsed := time.Since(start)

		if err != nil {
			rpcErrors++
			ios.DebugErr("RPC benchmark call", err)
		} else {
			rpcLatencies = append(rpcLatencies, elapsed)
		}

		// Show progress
		if !opts.JSONOutput && (i+1)%5 == 0 {
			ios.Printf("  Progress: %d/%d\n", i+1, opts.Iterations)
		}
	}

	// Benchmark ping-style calls (lighter weight)
	ios.Info("Running ping benchmark...")
	pingLatencies := make([]time.Duration, 0, opts.Iterations)
	pingErrors := 0

	for range opts.Iterations {
		start := time.Now()
		_, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
		elapsed := time.Since(start)

		if err != nil {
			pingErrors++
			ios.DebugErr("ping benchmark call", err)
		} else {
			pingLatencies = append(pingLatencies, elapsed)
		}
	}

	// Calculate statistics
	rpcStats := calculateStats(rpcLatencies, rpcErrors)
	pingStats := calculateStats(pingLatencies, pingErrors)

	// Build result
	result := Result{
		Device:      device,
		Iterations:  opts.Iterations,
		PingLatency: pingStats,
		RPCLatency:  rpcStats,
		Summary:     getSummary(rpcStats),
		Timestamp:   time.Now(),
	}

	if opts.JSONOutput {
		return outputJSON(ios, result)
	}

	// Display results
	ios.Println("")
	ios.Success("Benchmark Complete")
	ios.Println("")

	ios.Printf("Device: %s\n", device)
	ios.Printf("Iterations: %d\n", opts.Iterations)
	ios.Println("")

	ios.Printf("RPC Latency:\n")
	displayStats(ios, rpcStats)

	ios.Printf("\nPing Latency:\n")
	displayStats(ios, pingStats)

	ios.Printf("\nSummary: %s\n", result.Summary)

	return nil
}

func calculateStats(latencies []time.Duration, errors int) LatencyStats {
	if len(latencies) == 0 {
		return LatencyStats{Errors: errors}
	}

	// Sort for percentiles
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate statistics
	var sum time.Duration
	for _, d := range sorted {
		sum += d
	}

	return LatencyStats{
		Min:    sorted[0],
		Max:    sorted[len(sorted)-1],
		Avg:    sum / time.Duration(len(sorted)),
		P50:    percentile(sorted, 50),
		P95:    percentile(sorted, 95),
		P99:    percentile(sorted, 99),
		Errors: errors,
	}
}

func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func getSummary(stats LatencyStats) string {
	avgMs := float64(stats.Avg.Microseconds()) / 1000

	switch {
	case avgMs < 50:
		return "Excellent - Very fast response times"
	case avgMs < 100:
		return "Good - Normal response times"
	case avgMs < 200:
		return "Fair - Slightly elevated latency"
	case avgMs < 500:
		return "Poor - High latency, check network"
	default:
		return "Critical - Very high latency, network issues likely"
	}
}

func displayStats(ios *iostreams.IOStreams, stats LatencyStats) {
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

func outputJSON(ios *iostreams.IOStreams, result Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	ios.Println(string(data))
	return nil
}
