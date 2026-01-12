// Package coiot provides the debug coiot command.
package coiot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/debug"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/term"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	flags.OutputFlags
	Factory  *cmdutil.Factory
	Device   string
	Listen   bool
	Stream   bool
	Duration time.Duration
	Raw      bool
}

// NewCommand creates the debug coiot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "coiot [device]",
		Aliases: []string{"coap"},
		Short:   "Show CoIoT/CoAP status or listen for multicast updates",
		Long: `Show CoIoT (CoAP over Internet of Things) status for a device, or listen
for multicast updates from all Gen1 devices on the network.

CoIoT is used by Gen1 devices for local discovery and real-time status
updates via multicast UDP on 224.0.1.187:5683.

Without --listen, this command shows the CoIoT configuration for a specific device:
- CoIoT enabled/disabled status
- Multicast settings
- Peer configuration (for unicast mode)
- Update period settings

With --listen, it starts a multicast listener that receives and displays
CoIoT status broadcasts from all Gen1 devices on the network.`,
		Example: `  # Show CoIoT status for a device
  shelly debug coiot living-room

  # Output as JSON
  shelly debug coiot living-room -f json

  # Listen for CoIoT multicast updates for 30 seconds (default)
  shelly debug coiot --listen

  # Listen for 2 minutes
  shelly debug coiot --listen --duration 2m

  # Stream indefinitely (until Ctrl+C)
  shelly debug coiot --listen --stream

  # Stream with raw JSON output
  shelly debug coiot --listen --stream --raw`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Listen {
				if len(args) == 0 {
					return fmt.Errorf("device argument required (or use --listen to listen for multicast)")
				}
				opts.Device = args[0]
			}
			return run(cmd.Context(), opts)
		},
	}

	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")
	cmd.Flags().BoolVarP(&opts.Listen, "listen", "l", false, "Listen for CoIoT multicast updates from all Gen1 devices")
	cmd.Flags().BoolVarP(&opts.Stream, "stream", "s", false, "Stream indefinitely (until Ctrl+C)")
	cmd.Flags().DurationVar(&opts.Duration, "duration", 30*time.Second, "Listen duration (ignored if --stream)")
	cmd.Flags().BoolVar(&opts.Raw, "raw", false, "Output raw JSON events")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Handle listen mode
	if opts.Listen {
		listenerOpts := debug.CoIoTListenerOptions{
			Stream:   opts.Stream,
			Duration: opts.Duration,
			Raw:      opts.Raw,
		}
		if err := debug.RunCoIoTListener(ctx, ios, listenerOpts); err != nil {
			return fmt.Errorf("failed to start CoIoT listener: %w", err)
		}
		return nil
	}

	// Device status mode
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	svc := opts.Factory.ShellyService()

	var coiotStatus map[string]any
	var isGen1 bool
	err := svc.WithDevice(ctx, opts.Device, func(dev *shelly.DeviceClient) error {
		isGen1 = dev.IsGen1()

		if isGen1 {
			// Gen1: Fetch CoIoT settings from /settings endpoint
			result, err := dev.Gen1().GetSettings(ctx)
			if err != nil {
				return fmt.Errorf("failed to get settings: %w", err)
			}

			coiotStatus = make(map[string]any)
			coiotStatus["coiot"] = map[string]any{
				"enabled":       result.CoIoT.Enabled,
				"peer":          result.CoIoT.Peer,
				"update_period": result.CoIoT.UpdatePeriod,
			}

			// Add device info
			coiotStatus["device"] = map[string]any{
				"type": result.Device.Type,
				"mac":  result.Device.MAC,
			}
			return nil
		}

		// Gen2+ devices don't use CoIoT - they use WebSocket for real-time updates
		coiotStatus = map[string]any{
			"supported": false,
			"message":   "Gen2+ devices use WebSocket for real-time updates, not CoIoT",
		}
		return nil
	})
	if err != nil {
		return err
	}

	if opts.Format == "json" {
		jsonOutput, err := json.MarshalIndent(coiotStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonOutput))
		return nil
	}

	// Pretty print
	ios.Println(theme.Bold().Render("CoIoT/Device Configuration:"))
	ios.Println()

	if coiot, ok := coiotStatus["coiot"].(map[string]any); ok {
		ios.Println("  " + theme.Highlight().Render("CoIoT:"))
		term.DisplayMapSection(ios, coiot, "    ")
	} else {
		ios.Println("  " + theme.Dim().Render("CoIoT: not configured or not available"))
	}

	if device, ok := coiotStatus["device"].(map[string]any); ok {
		ios.Println()
		ios.Println("  " + theme.Highlight().Render("Device:"))
		term.DisplayMapSection(ios, device, "    ")
	}

	ios.Println()
	if !isGen1 {
		ios.Warning("Gen2+ devices do not use CoIoT.")
		ios.Info("Use 'shelly debug websocket <device>' for real-time event monitoring.")
		return nil
	}

	ios.Info("This is a Gen1 device with full CoIoT support.")
	coiot, ok := coiotStatus["coiot"].(map[string]any)
	if !ok {
		return nil
	}
	peer, hasPeer := coiot["peer"].(string)
	if !hasPeer || peer == "" {
		ios.Info("Multicast mode: device broadcasts to 224.0.1.187:5683")
	} else {
		ios.Info("Unicast mode: device sends to %s", peer)
	}

	return nil
}
