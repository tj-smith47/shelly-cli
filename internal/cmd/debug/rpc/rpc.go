// Package rpc provides the debug rpc command.
package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	Method  string
	Params  string
	Raw     bool
}

// NewCommand creates the debug rpc command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "rpc <device> <method> [params_json]",
		Aliases: []string{"call", "invoke"},
		Short:   "Execute a raw RPC call",
		Long: `Execute a raw RPC call on a Shelly device.

This command allows you to call any RPC method supported by the device.
Use 'shelly debug methods <device>' to list available methods.

Parameters should be provided as a JSON object. If no parameters are
needed, omit the params argument or use '{}'.`,
		Example: `  # Get device info
  shelly debug rpc living-room Shelly.GetDeviceInfo

  # Get switch status with ID parameter
  shelly debug rpc living-room Switch.GetStatus '{"id":0}'

  # Set switch state
  shelly debug rpc living-room Switch.Set '{"id":0,"on":true}'

  # Get all methods
  shelly debug rpc living-room Shelly.ListMethods

  # Raw output (no formatting)
  shelly debug rpc living-room Shelly.GetStatus --raw`,
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

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Parse params if provided
	var params map[string]any
	if opts.Params != "" {
		opts.Params = strings.TrimSpace(opts.Params)
		if err := json.Unmarshal([]byte(opts.Params), &params); err != nil {
			return fmt.Errorf("invalid JSON params: %w", err)
		}
	}

	ios.Debug("calling RPC method %s on %s with params: %v", opts.Method, opts.Device, params)

	result, err := svc.RawRPC(ctx, opts.Device, opts.Method, params)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}

	// Format output
	var output []byte
	if opts.Raw {
		output, err = json.Marshal(result)
	} else {
		output, err = json.MarshalIndent(result, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}

	ios.Println(string(output))
	return nil
}
