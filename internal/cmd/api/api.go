// Package apicmd provides direct API access to Shelly devices.
package apicmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/api/methods"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Method  string
	Params  string
	Raw     bool
}

// NewCommand creates the api command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "api <device> <method|path> [params_json]",
		Aliases: []string{"rpc", "call"},
		Short:   "Execute API calls on Shelly devices",
		Long: `Execute API calls on Shelly devices.

This command provides direct access to the device API using either:
  - RPC methods: "Shelly.GetDeviceInfo", "Switch.Set" (Gen2+ only)
  - REST paths: "/status", "/relay/0?turn=on", "/rpc/Shelly.GetStatus" (all generations)

The command auto-detects the call type based on input format:
  - Starts with "/" → HTTP GET to that path (works for all generations)
  - Contains "." → JSON-RPC call via WebSocket (Gen2+ only)

For Gen2+ devices, you can use either format:
  - RPC: Shelly.GetStatus (uses WebSocket JSON-RPC)
  - Path: /rpc/Shelly.GetStatus (uses HTTP GET)

Use 'shelly api methods <device>' to list available RPC methods (Gen2+ only).`,
		Example: `  # Gen2+ RPC methods (JSON-RPC via WebSocket)
  shelly api living-room Shelly.GetDeviceInfo
  shelly api living-room Switch.GetStatus '{"id":0}'
  shelly api living-room Switch.Set '{"id":0,"on":true}'

  # Path-based HTTP calls (all generations)
  shelly api backyard /status                      # Gen1 status
  shelly api backyard /relay/0?turn=on             # Gen1 relay control
  shelly api living-room /rpc/Shelly.GetStatus     # Gen2+ via HTTP
  shelly api living-room /rpc/Switch.Set?id=0&on=true  # Gen2+ via HTTP

  # Raw output (no formatting)
  shelly api living-room Shelly.GetStatus --raw

  # Using 'rpc' alias
  shelly rpc living-room Shelly.GetDeviceInfo`,
		Args:              cobra.RangeArgs(2, 3),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Method = args[1]
			if len(args) > 2 {
				opts.Params = args[2]
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "Output raw JSON without formatting")

	cmd.AddCommand(methods.NewCommand(f))

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Path-based calls (starting with /) use HTTP GET - works for all generations
	if strings.HasPrefix(opts.Method, "/") {
		if opts.Params != "" {
			ios.Warning("JSON params are ignored for HTTP path calls. Use query parameters in the path instead.")
		}
		ios.Debug("calling HTTP path %s on %s", opts.Method, opts.Device)

		httpResult, err := svc.RawHTTPCall(ctx, opts.Device, opts.Method)
		if err != nil {
			return fmt.Errorf("HTTP call failed: %w", err)
		}

		// Try to parse as JSON for pretty printing
		var result any
		if json.Unmarshal(httpResult, &result) != nil {
			// Not JSON, output raw text
			ios.Println(string(httpResult))
			return nil
		}
		return term.PrintAPIResult(ios, result, opts.Raw)
	}

	// RPC method call (containing .) - must be Gen2+ device
	isGen1, _, err := svc.IsGen1Device(ctx, opts.Device)
	if err != nil {
		return fmt.Errorf("failed to detect device generation: %w", err)
	}
	if isGen1 {
		return fmt.Errorf("RPC methods are for Gen2+ devices only\n"+
			"Device %q is Gen1. Use REST paths instead, e.g.: /status",
			opts.Device)
	}

	// Parse params if provided
	var params map[string]any
	if opts.Params != "" {
		if err := json.Unmarshal([]byte(strings.TrimSpace(opts.Params)), &params); err != nil {
			return fmt.Errorf("invalid JSON params: %w", err)
		}
	}

	ios.Debug("calling RPC method %s on %s with params: %v", opts.Method, opts.Device, params)

	result, err := svc.RawRPC(ctx, opts.Device, opts.Method, params)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}

	return term.PrintAPIResult(ios, result, opts.Raw)
}
