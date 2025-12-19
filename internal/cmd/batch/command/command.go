// Package command provides the batch command subcommand for raw RPC calls.
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// NewCommand creates the batch command command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
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
Params should be a JSON object (e.g., '{"id":0,"on":true}').

Target devices can be specified multiple ways:
  - As arguments: device names or addresses after the method/params
  - Via stdin: pipe device names (one per line or space-separated)
  - Via group: --group flag targets all devices in a group
  - Via all: --all flag targets all registered devices

Priority: explicit args > stdin > group > all

Results are output as JSON or YAML (use -o yaml). Each result includes
the device name and either the response or error message.`,
		Example: `  # Get status from all devices in a group
  shelly batch command "Shelly.GetStatus" --group living-room

  # Turn on switch 0 on specific devices
  shelly batch command "Switch.Set" '{"id":0,"on":true}' light-1 light-2

  # Set brightness on all devices
  shelly batch command "Light.Set" '{"id":0,"brightness":50}' --all

  # Using alias
  shelly batch rpc "Switch.Toggle" '{"id":0}' --group bedroom

  # Output as YAML
  shelly batch command "Shelly.GetDeviceInfo" --all -o yaml

  # Pipe device names from a file
  cat devices.txt | shelly batch command "Shelly.GetStatus"

  # Pipe from device list command
  shelly device list -o json | jq -r '.[].name' | shelly batch command "Shelly.Reboot"

  # Get status of Gen2+ devices and extract uptime
  shelly device list -o json | jq -r '.[] | select(.generation >= 2) | .name' | \
    shelly batch command "Shelly.GetStatus" | jq '.[] | {device, uptime: .response.sys.uptime}'

  # Check firmware versions across all devices
  shelly batch command "Shelly.GetDeviceInfo" --all | jq '.[] | {device, fw: .response.fw_id}'`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := args[0]

			// Parse params if provided
			var params map[string]any
			deviceArgs := args[1:]
			if len(args) > 1 && utils.IsJSONObject(args[1]) {
				if err := json.Unmarshal([]byte(args[1]), &params); err != nil {
					return fmt.Errorf("invalid JSON params: %w", err)
				}
				deviceArgs = args[2:]
			}

			targets, err := utils.ResolveBatchTargets(groupName, all, deviceArgs)
			if err != nil {
				return err
			}
			return run(cmd.Context(), f, targets, method, params, timeout, concurrent, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&groupName, "group", "g", "", "Target device group")
	cmd.Flags().BoolVarP(&all, "all", "a", false, "Target all registered devices")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "json", "Output format: json, yaml")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, targets []string, method string, params map[string]any, timeout time.Duration, concurrent int, outputFormat string) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

	// Add all lines upfront
	for _, target := range targets {
		mw.AddLine(target, "pending")
	}

	// Results are collected by index to maintain order
	results := make([]model.BatchRPCResult, len(targets))

	// Create parent context with overall timeout
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(targets)))
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrent)

	for i, target := range targets {
		idx := i
		device := target // Capture for closure
		g.Go(func() error {
			mw.UpdateLine(device, iostreams.StatusRunning, method)

			// Per-device timeout
			deviceCtx, deviceCancel := context.WithTimeout(ctx, timeout)
			defer deviceCancel()

			resp, err := svc.RawRPC(deviceCtx, device, method, params)
			result := model.BatchRPCResult{Device: device}
			if err != nil {
				result.Error = err.Error()
				mw.UpdateLine(device, iostreams.StatusError, err.Error())
			} else {
				result.Response = resp
				mw.UpdateLine(device, iostreams.StatusSuccess, "done")
			}

			results[idx] = result
			return nil // Don't fail the whole batch on individual errors
		})
	}

	// Wait for all operations to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("batch operation failed: %w", err)
	}

	mw.Finalize()

	// For TTY, add a blank line before JSON/YAML output for clarity
	if ios.IsStdoutTTY() {
		ios.Printf("\n")
	}

	// Output results
	switch outputFormat {
	case "yaml":
		if err := output.PrintYAML(results); err != nil {
			return err
		}
	default:
		if err := output.PrintJSON(results); err != nil {
			return err
		}
	}

	// Print summary
	success, failed, _ := mw.Summary()
	if failed > 0 {
		ios.Warning("%d/%d devices failed", failed, len(targets))
		return fmt.Errorf("%d/%d devices failed", failed, len(targets))
	}
	ios.Info("Command sent to %d device(s)", success)
	return nil
}
