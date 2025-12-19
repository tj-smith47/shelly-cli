// Package benchmark provides the benchmark command for device performance testing.
package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Options holds the command options.
type Options struct {
	Iterations int
	Warmup     int
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
  shelly benchmark kitchen-light --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Iterations, "iterations", "n", 10, "Number of iterations")
	cmd.Flags().IntVar(&opts.Warmup, "warmup", 2, "Number of warmup iterations (not counted)")

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

		// Show progress (only in table mode)
		if !output.WantsStructured() && (i+1)%5 == 0 {
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
	rpcStats := utils.CalculateLatencyStats(rpcLatencies, rpcErrors)
	pingStats := utils.CalculateLatencyStats(pingLatencies, pingErrors)

	// Build result
	result := model.BenchmarkResult{
		Device:      device,
		Iterations:  opts.Iterations,
		PingLatency: pingStats,
		RPCLatency:  rpcStats,
		Summary:     output.FormatBenchmarkSummary(rpcStats),
		Timestamp:   time.Now(),
	}

	if output.WantsStructured() {
		return output.FormatOutput(ios.Out, result)
	}

	// Display results
	ios.Println("")
	ios.Success("Benchmark Complete")
	ios.Println("")

	ios.Printf("Device: %s\n", device)
	ios.Printf("Iterations: %d\n", opts.Iterations)
	ios.Println("")

	ios.Printf("RPC Latency:\n")
	term.DisplayLatencyStats(ios, rpcStats)

	ios.Printf("\nPing Latency:\n")
	term.DisplayLatencyStats(ios, pingStats)

	ios.Printf("\nSummary: %s\n", result.Summary)

	return nil
}
