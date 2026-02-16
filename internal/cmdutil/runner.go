// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/jq"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Output format constants.
const (
	outputJSON = "json"
	outputYAML = "yaml"
)

// logVerbose logs a message to stderr only if verbose mode is enabled.
func logVerbose(format string, args ...any) {
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "debug: "+format+"\n", args...)
	}
}

// CapConcurrency caps the requested concurrency to the global rate limit.
// If the requested value exceeds the global limit, it prints a warning and returns the capped value.
// This ensures user-requested parallelism doesn't exceed the rate limiter's global constraint.
func CapConcurrency(ios *iostreams.IOStreams, requested int) int {
	globalMax := config.GetGlobalMaxConcurrent()

	if requested > globalMax {
		ios.Warning("Requested concurrency %d exceeds global rate limit %d; capping to %d.\n"+
			"  Adjust ratelimit.global.max_concurrent in config to increase.",
			requested, globalMax, globalMax)
		return globalMax
	}
	return requested
}

// ComponentAction is a function that operates on a device component.
// It takes the context, shelly service, device identifier, and component ID.
type ComponentAction func(ctx context.Context, svc *shelly.Service, device string, id int) error

// DeviceAction is a function that operates on a device (no component ID).
type DeviceAction func(ctx context.Context, svc *shelly.Service, device string) error

// RunWithSpinner executes an action with a progress spinner.
// The spinner is automatically started before the action and stopped after.
// If ios is nil or not a TTY, progress messages are printed instead.
func RunWithSpinner(ctx context.Context, ios *iostreams.IOStreams, msg string, action func(context.Context) error) error {
	ios.StartProgress(msg)
	err := action(ctx)
	ios.StopProgress()
	return err
}

// RunWithSpinnerResult executes an action with a progress spinner and returns the result.
func RunWithSpinnerResult[T any](ctx context.Context, ios *iostreams.IOStreams, msg string, action func(context.Context) (T, error)) (T, error) {
	ios.StartProgress(msg)
	result, err := action(ctx)
	ios.StopProgress()
	return result, err
}

// RunBatch executes an action on multiple devices concurrently.
// It uses errgroup for concurrency, capped to the global rate limit.
// Errors from individual operations are logged but don't stop the batch.
// Returns an error only if context is cancelled.
func RunBatch(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, targets []string, concurrent int, action DeviceAction) error {
	if len(targets) == 0 {
		return nil
	}

	// Cap concurrency to global rate limit
	capped := CapConcurrency(ios, concurrent)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(capped)

	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

	// Add all lines upfront
	for _, target := range targets {
		mw.AddLine(target, "pending")
	}

	for _, target := range targets {
		t := target
		g.Go(func() error {
			mw.UpdateLine(t, iostreams.StatusRunning, "working...")

			if err := action(ctx, svc, t); err != nil {
				mw.UpdateLine(t, iostreams.StatusError, err.Error())
				return nil // Don't fail the whole batch
			}

			mw.UpdateLine(t, iostreams.StatusSuccess, "done")
			return nil
		})
	}

	err := g.Wait()
	mw.Finalize()
	mw.PrintSummary()

	return err
}

// RunBatchComponent executes a component action on multiple devices concurrently.
// Similar to RunBatch but passes a component ID to each action.
func RunBatchComponent(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, targets []string, componentID, concurrent int, action ComponentAction) error {
	return RunBatch(ctx, ios, svc, targets, concurrent, func(ctx context.Context, svc *shelly.Service, device string) error {
		return action(ctx, svc, device, componentID)
	})
}

// BatchResult holds the result of a batch operation.
type BatchResult struct {
	Device  string
	Success bool
	Message string
	Error   error
}

// RunBatchWithResults executes an action on multiple devices and collects results.
// Unlike RunBatch, this returns all results for further processing.
// Concurrency is capped to the global rate limit (silently, since no ios available).
func RunBatchWithResults(ctx context.Context, svc *shelly.Service, targets []string, concurrent int, action DeviceAction) []BatchResult {
	if len(targets) == 0 {
		return nil
	}

	// Cap concurrency to global rate limit (silently)
	globalMax := config.GetGlobalMaxConcurrent()
	capped := concurrent
	if concurrent > globalMax {
		capped = globalMax
	}

	results := make([]BatchResult, len(targets))
	resultChan := make(chan struct {
		idx    int
		result BatchResult
	}, len(targets))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(capped)

	for i, target := range targets {
		idx := i
		t := target
		g.Go(func() error {
			result := BatchResult{Device: t}

			if err := action(ctx, svc, t); err != nil {
				result.Error = err
				result.Message = err.Error()
			} else {
				result.Success = true
				result.Message = "success"
			}

			resultChan <- struct {
				idx    int
				result BatchResult
			}{idx, result}
			return nil
		})
	}

	// Wait for all goroutines and collect results
	go func() {
		if err := g.Wait(); err != nil {
			logVerbose("batch wait error: %v", err)
		}
		close(resultChan)
	}()

	for r := range resultChan {
		results[r.idx] = r.result
	}

	return results
}

// RunSimple executes a simple device operation with standard error handling.
// It wraps common command execution patterns:
//   - Creates a spinner with the given message
//   - Calls the action
//   - Prints success or error message.
func RunSimple(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, device string, componentID int, spinnerMsg, successMsg string, action ComponentAction) error {
	return RunWithSpinner(ctx, ios, spinnerMsg, func(ctx context.Context) error {
		if err := action(ctx, svc, device, componentID); err != nil {
			return err
		}
		ios.Success("%s", successMsg)
		return nil
	})
}

// PrintBatchSummary prints a summary of batch results.
func PrintBatchSummary(ios *iostreams.IOStreams, results []BatchResult) {
	var success, failed int
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}

	switch {
	case failed == 0:
		ios.Success("All %d operations succeeded", len(results))
	case success == 0:
		ios.Error("All %d operations failed", len(results))
	default:
		ios.Printf("%d succeeded, %d failed\n", success, failed)
	}
}

// PrintDryRun prints a dry-run preview showing the action and target devices.
// The description should be a human-readable action like "Would turn on switch (id:0)".
func PrintDryRun(ios *iostreams.IOStreams, description string, targets []string) {
	ios.Info("Dry run — showing what would be applied:")
	ios.Printf("\n")
	ios.Info("%s on %d device(s):", description, len(targets))
	for _, t := range targets {
		ios.Printf("  - %s\n", t)
	}
	ios.Printf("\nDry run complete. No changes were made.\n")
}

// StatusFetcher is a function that fetches component status.
type StatusFetcher[T any] func(ctx context.Context, svc *shelly.Service, device string, id int) (T, error)

// StatusDisplay is a function that displays status in human-readable format.
type StatusDisplay[T any] func(ios *iostreams.IOStreams, status T)

// RunStatus executes a status fetch with spinner and handles output formatting.
// It uses RunWithSpinnerResult for the fetch operation and supports JSON/YAML/default output.
func RunStatus[T any](
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc *shelly.Service,
	device string,
	componentID int,
	spinnerMsg string,
	fetcher StatusFetcher[T],
	display StatusDisplay[T],
) error {
	status, err := RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) (T, error) {
		return fetcher(ctx, svc, device, componentID)
	})
	if err != nil {
		return err
	}

	return PrintResult(ios, status, display)
}

// PrintResult outputs result data in the configured format (JSON, YAML, or human-readable).
// If --fields is set, prints available field names instead of data.
// If --jq is set, the jq filter is applied to the data regardless of output format.
func PrintResult[T any](ios *iostreams.IOStreams, data T, display StatusDisplay[T]) error {
	if jq.HasFields() {
		return jq.PrintFields(ios.Out, data)
	}
	if jq.HasFilter() {
		return jq.Apply(ios.Out, data, jq.GetFilter())
	}

	switch viper.GetString("output") {
	case outputJSON:
		enc := json.NewEncoder(ios.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case outputYAML:
		enc := yaml.NewEncoder(ios.Out)
		return enc.Encode(data)
	default:
		display(ios, data)
		return nil
	}
}

// DeviceStatusFetcher is a function that fetches device-level status (no component ID).
type DeviceStatusFetcher[T any] func(ctx context.Context, svc *shelly.Service, device string) (T, error)

// RunDeviceStatus executes a device status fetch with spinner and handles output formatting.
// Similar to RunStatus but doesn't take a component ID.
func RunDeviceStatus[T any](
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc *shelly.Service,
	device string,
	spinnerMsg string,
	fetcher DeviceStatusFetcher[T],
	display StatusDisplay[T],
) error {
	status, err := RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) (T, error) {
		return fetcher(ctx, svc, device)
	})
	if err != nil {
		return err
	}

	return PrintResult(ios, status, display)
}

// ListFetcher is a function that fetches a list of items from a device.
type ListFetcher[T any] func(ctx context.Context, svc *shelly.Service, device string) ([]T, error)

// ListDisplay is a function that displays a list in human-readable format.
type ListDisplay[T any] func(ios *iostreams.IOStreams, items []T)

// RunList executes a list fetch with spinner and handles output formatting.
// Returns early with a message if the list is empty.
func RunList[T any](
	ctx context.Context,
	ios *iostreams.IOStreams,
	svc *shelly.Service,
	device string,
	spinnerMsg string,
	emptyMsg string,
	fetcher ListFetcher[T],
	display ListDisplay[T],
) error {
	items, err := RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) ([]T, error) {
		return fetcher(ctx, svc, device)
	})
	if err != nil {
		return err
	}

	if len(items) == 0 {
		ios.NoResults(emptyMsg)
		return nil
	}

	return PrintListResult(ios, items, display)
}

// PrintListResult outputs list data in the configured format (JSON, YAML, or human-readable).
// If --fields is set, prints available field names instead of data.
// If --jq is set, the jq filter is applied to the data regardless of output format.
func PrintListResult[T any](ios *iostreams.IOStreams, items []T, display ListDisplay[T]) error {
	if jq.HasFields() {
		return jq.PrintFields(ios.Out, items)
	}
	if jq.HasFilter() {
		return jq.Apply(ios.Out, items, jq.GetFilter())
	}

	switch viper.GetString("output") {
	case outputJSON:
		enc := json.NewEncoder(ios.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(items)
	case outputYAML:
		enc := yaml.NewEncoder(ios.Out)
		return enc.Encode(items)
	default:
		display(ios, items)
		return nil
	}
}

// RunCachedDeviceStatus executes a device status fetch with cache integration.
// It checks cache first (unless --refresh), fetches if miss, and caches the result.
// Respects --offline flag (cache only mode).
func RunCachedDeviceStatus[T any](
	ctx context.Context,
	f *Factory,
	device string,
	cacheType string,
	cacheTTL time.Duration,
	spinnerMsg string,
	fetcher DeviceStatusFetcher[T],
	display StatusDisplay[T],
) error {
	// Check for flag conflict early
	if err := CheckCacheFlags(); err != nil {
		return err
	}

	ios := f.IOStreams()
	svc := f.ShellyService()

	result, err := CachedFetch(ctx, f, device, cacheType, cacheTTL, func(ctx context.Context) (T, error) {
		// Only show spinner for fresh fetches
		return RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) (T, error) {
			return fetcher(ctx, svc, device)
		})
	})
	if err != nil {
		return err
	}

	return PrintResult(ios, result.Data, display)
}

// RunCachedList executes a list fetch with cache integration.
// It checks cache first (unless --refresh), fetches if miss, and caches the result.
// Respects --offline flag (cache only mode).
func RunCachedList[T any](
	ctx context.Context,
	f *Factory,
	device string,
	cacheType string,
	cacheTTL time.Duration,
	spinnerMsg string,
	emptyMsg string,
	fetcher ListFetcher[T],
	display ListDisplay[T],
) error {
	// Check for flag conflict early
	if err := CheckCacheFlags(); err != nil {
		return err
	}

	ios := f.IOStreams()
	svc := f.ShellyService()

	result, err := CachedFetchList(ctx, f, device, cacheType, cacheTTL, func(ctx context.Context) ([]T, error) {
		// Only show spinner for fresh fetches
		return RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) ([]T, error) {
			return fetcher(ctx, svc, device)
		})
	})
	if err != nil {
		return err
	}

	if len(result.Data) == 0 {
		ios.NoResults(emptyMsg)
		return nil
	}

	return PrintListResult(ios, result.Data, display)
}

// RunCachedStatus executes a component status fetch with cache integration.
// Similar to RunCachedDeviceStatus but includes component ID.
func RunCachedStatus[T any](
	ctx context.Context,
	f *Factory,
	device string,
	componentID int,
	cacheType string,
	cacheTTL time.Duration,
	spinnerMsg string,
	fetcher StatusFetcher[T],
	display StatusDisplay[T],
) error {
	// Check for flag conflict early
	if err := CheckCacheFlags(); err != nil {
		return err
	}

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Include component ID in cache key
	cacheKey := fmt.Sprintf("%s:%d", cacheType, componentID)

	result, err := CachedFetch(ctx, f, device, cacheKey, cacheTTL, func(ctx context.Context) (T, error) {
		return RunWithSpinnerResult(ctx, ios, spinnerMsg, func(ctx context.Context) (T, error) {
			return fetcher(ctx, svc, device, componentID)
		})
	})
	if err != nil {
		return err
	}

	return PrintResult(ios, result.Data, display)
}

// InvalidateCacheAfterMutation invalidates the appropriate cache after a mutation.
// Call this after successful create/update/delete operations.
func InvalidateCacheAfterMutation(f *Factory, device, cacheType string) {
	fc := f.FileCache()
	if fc == nil {
		return
	}

	if err := fc.Invalidate(device, cacheType); err != nil {
		f.IOStreams().DebugErr("invalidate cache "+cacheType, err)
	}
}

// InvalidateCacheTypes invalidates multiple cache types for a device.
// Useful when a mutation affects multiple cached data types.
func InvalidateCacheTypes(f *Factory, device string, cacheTypes ...string) {
	fc := f.FileCache()
	if fc == nil {
		return
	}

	for _, cacheType := range cacheTypes {
		if err := fc.Invalidate(device, cacheType); err != nil {
			f.IOStreams().DebugErr("invalidate cache "+cacheType, err)
		}
	}
}

// CacheTypes provides access to cache type constants for convenience.
var CacheTypes = struct {
	Firmware   string
	System     string
	WiFi       string
	Security   string
	Cloud      string
	BLE        string
	MQTT       string
	Schedules  string
	Webhooks   string
	Virtuals   string
	Inputs     string
	KVS        string
	Scripts    string
	DeviceInfo string
	Components string
}{
	Firmware:   cache.TypeFirmware,
	System:     cache.TypeSystem,
	WiFi:       cache.TypeWiFi,
	Security:   cache.TypeSecurity,
	Cloud:      cache.TypeCloud,
	BLE:        cache.TypeBLE,
	MQTT:       cache.TypeMQTT,
	Schedules:  cache.TypeSchedules,
	Webhooks:   cache.TypeWebhooks,
	Virtuals:   cache.TypeVirtuals,
	Inputs:     cache.TypeInputs,
	KVS:        cache.TypeKVS,
	Scripts:    cache.TypeScripts,
	DeviceInfo: cache.TypeDeviceInfo,
	Components: cache.TypeComponents,
}
