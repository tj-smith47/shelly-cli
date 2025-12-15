// Package test provides the auth test subcommand.
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
)

// Options holds the command options.
type Options struct {
	User     string
	Password string
	Timeout  time.Duration
}

// NewCommand creates the auth test command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Timeout: 10 * time.Second,
	}

	cmd := &cobra.Command{
		Use:     "test <device>",
		Aliases: []string{"verify", "check"},
		Short:   "Test authentication credentials",
		Long: `Test authentication credentials against a device.

This command verifies that the provided credentials are valid
by attempting to connect to the device.

Exit codes:
  0 - Authentication successful
  1 - Authentication failed or error`,
		Example: `  # Test with provided credentials
  shelly auth test living-room --user admin --password secret

  # Test with configured credentials
  shelly auth test living-room

  # Quick test with short timeout
  shelly auth test living-room --timeout 5s`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.User, "user", "", "Username to test")
	cmd.Flags().StringVar(&opts.Password, "password", "", "Password to test")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Second, "Connection timeout")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, opts *Options) error {
	ios := f.IOStreams()
	svc := f.ShellyService()

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	ios.Info("Testing authentication for %s...", device)

	// Try to connect and get device info
	conn, err := svc.Connect(ctx, device)
	if err != nil {
		ios.Error("Connection failed: %v", err)
		return fmt.Errorf("authentication test failed")
	}
	defer iostreams.CloseWithDebug("closing auth test connection", conn)

	// Try to make an authenticated call
	rawResult, err := conn.Call(ctx, "Shelly.GetDeviceInfo", nil)
	if err != nil {
		ios.Error("Authentication failed: %v", err)
		return fmt.Errorf("authentication test failed")
	}

	// Parse result
	jsonBytes, err := json.Marshal(rawResult)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var result struct {
		ID     string  `json:"id"`
		MAC    string  `json:"mac"`
		Auth   bool    `json:"auth_en"`
		Domain *string `json:"auth_domain"`
	}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	// Success
	ios.Success("Authentication successful!")
	ios.Println("")

	ios.Printf("Device: %s\n", device)
	ios.Printf("ID: %s\n", result.ID)
	ios.Printf("MAC: %s\n", result.MAC)

	if result.Auth {
		ios.Info("Authentication is enabled on this device")
		if result.Domain != nil && *result.Domain != "" {
			ios.Printf("Auth domain: %s\n", *result.Domain)
		}
	} else {
		ios.Warning("Authentication is not enabled on this device")
	}

	return nil
}
