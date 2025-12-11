// Package command provides the batch command subcommand for raw RPC calls.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/helpers"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the batch command command.
func NewCommand() *cobra.Command {
	var (
		groupName    string
		all          bool
		timeout      time.Duration
		concurrent   int
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:     "command <method> [params-json] [device...]",
		Aliases: []string{"cmd", "rpc"},
		Short:   "Send RPC command to devices",
		Long: `Send a raw RPC command to multiple devices simultaneously.

The method is the RPC method name (e.g., "Switch.Set", "Shelly.GetStatus").
Params should be a JSON object (e.g., '{"id":0,"on":true}').`,
		Example: `  # Get status from all devices in a group
  shelly batch command "Shelly.GetStatus" --group living-room

  # Turn on switch 0 on specific devices
  shelly batch command "Switch.Set" '{"id":0,"on":true}' light-1 light-2

  # Set brightness on all devices
  shelly batch command "Light.Set" '{"id":0,"brightness":50}' --all

  # Using alias
  shelly batch rpc "Switch.Toggle" '{"id":0}' --group bedroom`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			method := args[0]

			// Parse params if provided
			var params map[string]any
			deviceArgs := args[1:]
			if len(args) > 1 && isJSON(args[1]) {
				if err := json.Unmarshal([]byte(args[1]), &params); err != nil {
					return fmt.Errorf("invalid JSON params: %w", err)
				}
				deviceArgs = args[2:]
			}

			targets, err := helpers.ResolveBatchTargets(groupName, all, deviceArgs)
			if err != nil {
				return err
			}
			return run(targets, method, params, timeout, concurrent, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&groupName, "group", "g", "", "Target device group")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Target all registered devices")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json, yaml")

	return cmd
}

// isJSON returns true if the string looks like a JSON object.
func isJSON(s string) bool {
	return s != "" && s[0] == '{'
}

// Result holds the result of a batch RPC operation.
type Result struct {
	Device   string `json:"device" yaml:"device"`
	Response any    `json:"response,omitempty" yaml:"response,omitempty"`
	Error    string `json:"error,omitempty" yaml:"error,omitempty"`
}

func run(targets []string, method string, params map[string]any, timeout time.Duration, concurrent int, outputFormat string) error {
	svc := shelly.NewService()
	results := make(chan Result, len(targets))

	// Semaphore for concurrency control
	sem := make(chan struct{}, concurrent)

	var wg sync.WaitGroup
	for _, target := range targets {
		wg.Add(1)
		go func(device string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resp, err := svc.RawRPC(ctx, device, method, params)
			result := Result{Device: device}
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Response = resp
			}
			results <- result
		}(target)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	collected := make([]Result, 0, len(targets))
	var failed int
	for result := range results {
		collected = append(collected, result)
		if result.Error != "" {
			failed++
		}
	}

	// Output results
	switch outputFormat {
	case "yaml":
		if err := output.PrintYAML(collected); err != nil {
			return err
		}
	default:
		if err := output.PrintJSON(collected); err != nil {
			return err
		}
	}

	// Print summary
	succeeded := len(collected) - failed
	if failed > 0 {
		iostreams.Warning("%d/%d devices failed", failed, len(targets))
		return fmt.Errorf("%d/%d devices failed", failed, len(targets))
	}
	iostreams.Info("Command sent to %d device(s)", succeeded)
	return nil
}
