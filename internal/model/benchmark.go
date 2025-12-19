// Package model provides domain types for the shelly-cli.
package model

import "time"

// BenchmarkResult holds the results of a device performance benchmark.
type BenchmarkResult struct {
	Device      string       `json:"device"`
	Iterations  int          `json:"iterations"`
	PingLatency LatencyStats `json:"ping_latency"`
	RPCLatency  LatencyStats `json:"rpc_latency"`
	Summary     string       `json:"summary"`
	Timestamp   time.Time    `json:"timestamp"`
}

// LatencyStats holds latency statistics for benchmark measurements.
type LatencyStats struct {
	Min    time.Duration `json:"min"`
	Max    time.Duration `json:"max"`
	Avg    time.Duration `json:"avg"`
	P50    time.Duration `json:"p50"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
	Errors int           `json:"errors"`
}
