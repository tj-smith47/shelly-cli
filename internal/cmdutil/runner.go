// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// logVerbose logs a message to stderr only if verbose mode is enabled.
func logVerbose(format string, args ...any) {
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "debug: "+format+"\n", args...)
	}
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
// It uses errgroup for concurrency with a configurable limit.
// Errors from individual operations are logged but don't stop the batch.
// Returns an error only if context is cancelled.
func RunBatch(ctx context.Context, ios *iostreams.IOStreams, svc *shelly.Service, targets []string, concurrent int, action DeviceAction) error {
	if len(targets) == 0 {
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrent)

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
func RunBatchWithResults(ctx context.Context, svc *shelly.Service, targets []string, concurrent int, action DeviceAction) []BatchResult {
	if len(targets) == 0 {
		return nil
	}

	results := make([]BatchResult, len(targets))
	resultChan := make(chan struct {
		idx    int
		result BatchResult
	}, len(targets))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrent)

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
