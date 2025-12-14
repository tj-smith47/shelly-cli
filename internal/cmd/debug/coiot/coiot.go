// Package coiot provides the debug coiot command.
package coiot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the debug coiot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "coiot <device>",
		Aliases: []string{"coap"},
		Short:   "Show CoIoT/CoAP status",
		Long: `Show CoIoT (CoAP over Internet of Things) status for a device.

CoIoT is used by Gen1 and some Gen2 devices for local discovery and
real-time status updates via multicast UDP.

This command shows:
- CoIoT enabled/disabled status
- Multicast settings
- Peer configuration (for unicast mode)
- Update period settings`,
		Example: `  # Show CoIoT status
  shelly debug coiot living-room

  # Output as JSON
  shelly debug coiot living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var coiotStatus map[string]any
	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		// Try to get CoIoT config via Sys.GetConfig for Gen2+
		result, err := conn.Call(ctx, "Sys.GetConfig", nil)
		if err != nil {
			ios.Debug("Sys.GetConfig failed: %v", err)
			return fmt.Errorf("failed to get system config: %w", err)
		}

		// Extract CoIoT/device settings from config
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		var config map[string]any
		if err := json.Unmarshal(jsonBytes, &config); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		// Build CoIoT status from various config sections
		coiotStatus = make(map[string]any)

		// Check for coiot section (Gen2+)
		if coiot, ok := config["coiot"].(map[string]any); ok {
			coiotStatus["coiot"] = coiot
		}

		// Check for device section
		if device, ok := config["device"].(map[string]any); ok {
			coiotStatus["device"] = device
		}

		// Check for sys section
		if sys, ok := config["sys"].(map[string]any); ok {
			coiotStatus["sys"] = sys
		}

		return nil
	})
	if err != nil {
		return err
	}

	if opts.JSON {
		output, err := json.MarshalIndent(coiotStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	// Pretty print
	ios.Println(theme.Bold().Render("CoIoT/Device Configuration:"))
	ios.Println()

	if coiot, ok := coiotStatus["coiot"].(map[string]any); ok {
		ios.Println("  " + theme.Highlight().Render("CoIoT:"))
		printMapSection(ios, coiot, "    ")
	} else {
		ios.Println("  " + theme.Dim().Render("CoIoT: not configured or not available"))
	}

	if device, ok := coiotStatus["device"].(map[string]any); ok {
		ios.Println()
		ios.Println("  " + theme.Highlight().Render("Device:"))
		printMapSection(ios, device, "    ")
	}

	ios.Println()
	ios.Info("Note: CoIoT is primarily used by Gen1 devices.")
	ios.Info("Gen2+ devices may have limited CoIoT support.")

	return nil
}

// printMapSection prints a map section with indentation.
func printMapSection(ios *iostreams.IOStreams, m map[string]any, indent string) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			ios.Println(indent + k + ":")
			printMapSection(ios, val, indent+"  ")
		default:
			ios.Printf("%s%s: %v\n", indent, k, v)
		}
	}
}
