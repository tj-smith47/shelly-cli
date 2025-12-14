// Package set provides the mqtt set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	serverFlag      string
	userFlag        string
	passwordFlag    string
	topicPrefixFlag string
	enableFlag      bool
)

// NewCommand creates the mqtt set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set <device>",
		Aliases: []string{"configure", "config"},
		Short:   "Configure MQTT",
		Long: `Configure MQTT settings for a device.

Set the MQTT broker address, credentials, and topic prefix for
integration with home automation systems.`,
		Example: `  # Configure MQTT with server and credentials
  shelly mqtt set living-room --server "mqtt://broker:1883" --user user --password pass

  # Configure with custom topic prefix
  shelly mqtt set living-room --server "mqtt://broker:1883" --topic-prefix "home/shelly"

  # Enable MQTT with existing configuration
  shelly mqtt set living-room --enable`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}

	cmd.Flags().StringVar(&serverFlag, "server", "", "MQTT broker URL (e.g., mqtt://broker:1883)")
	cmd.Flags().StringVar(&userFlag, "user", "", "MQTT username")
	cmd.Flags().StringVar(&passwordFlag, "password", "", "MQTT password")
	cmd.Flags().StringVar(&topicPrefixFlag, "topic-prefix", "", "MQTT topic prefix")
	cmd.Flags().BoolVar(&enableFlag, "enable", false, "Enable MQTT")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string) error {
	// Validate - need at least one option
	if serverFlag == "" && userFlag == "" && passwordFlag == "" && topicPrefixFlag == "" && !enableFlag {
		return fmt.Errorf("specify at least one configuration option (--server, --user, --password, --topic-prefix) or --enable")
	}

	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	// Determine enable state
	var enable *bool
	if enableFlag {
		t := true
		enable = &t
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring MQTT...", func(ctx context.Context) error {
		if err := svc.SetMQTTConfig(ctx, device, enable, serverFlag, userFlag, passwordFlag, topicPrefixFlag); err != nil {
			return fmt.Errorf("failed to configure MQTT: %w", err)
		}
		ios.Success("MQTT configured on %s", device)
		if serverFlag != "" {
			ios.Printf("  Server: %s\n", serverFlag)
		}
		return nil
	})
}
