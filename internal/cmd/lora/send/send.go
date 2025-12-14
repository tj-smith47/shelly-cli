// Package send provides the lora send command.
package send

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	ID      int
	Data    string
	Hex     bool
}

// NewCommand creates the lora send command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "send <device> <data>",
		Aliases: []string{"tx", "transmit"},
		Short:   "Send data over LoRa",
		Long: `Send data over LoRa RF on a Shelly device.

Transmits data through the LoRa add-on. Data can be provided as:
- Plain text (default)
- Hexadecimal bytes with --hex flag

The data is base64-encoded before transmission as required by
the LoRa.SendBytes API.`,
		Example: `  # Send a text message
  shelly lora send living-room "Hello World"

  # Send hex data
  shelly lora send living-room "48656c6c6f" --hex

  # Specify component ID
  shelly lora send living-room "test" --id 100`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Data = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.ID, "id", 100, "LoRa component ID")
	cmd.Flags().BoolVar(&opts.Hex, "hex", false, "Data is hexadecimal")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Convert data to bytes
	var data []byte
	if opts.Hex {
		// Parse hex string
		hexStr := strings.ReplaceAll(opts.Data, " ", "")
		hexStr = strings.ReplaceAll(hexStr, ":", "")
		if len(hexStr)%2 != 0 {
			return fmt.Errorf("invalid hex string: odd number of characters")
		}
		data = make([]byte, len(hexStr)/2)
		for i := 0; i < len(hexStr); i += 2 {
			var b byte
			if _, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b); err != nil {
				return fmt.Errorf("invalid hex at position %d: %w", i, err)
			}
			data[i/2] = b
		}
	} else {
		data = []byte(opts.Data)
	}

	// Base64 encode for API
	b64Data := base64.StdEncoding.EncodeToString(data)

	err := svc.WithConnection(ctx, opts.Device, func(conn *client.Client) error {
		params := map[string]any{
			"id":   opts.ID,
			"data": b64Data,
		}

		_, err := conn.Call(ctx, "LoRa.SendBytes", params)
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	ios.Success("Data sent over LoRa.")
	ios.Info("  Size: %d bytes", len(data))

	return nil
}
