// Package set provides the mqtt set subcommand.
package set

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	Device      string
	Enable      bool
	Password    string
	Server      string
	TopicPrefix string
	User        string
}

// NewCommand creates the mqtt set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

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
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Server, "server", "", "MQTT broker URL (e.g., mqtt://broker:1883)")
	cmd.Flags().StringVar(&opts.User, "user", "", "MQTT username")
	cmd.Flags().StringVar(&opts.Password, "password", "", "MQTT password")
	cmd.Flags().StringVar(&opts.TopicPrefix, "topic-prefix", "", "MQTT topic prefix")
	cmd.Flags().BoolVar(&opts.Enable, "enable", false, "Enable MQTT")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	// Validate - need at least one option
	if opts.Server == "" && opts.User == "" && opts.Password == "" && opts.TopicPrefix == "" && !opts.Enable {
		return fmt.Errorf("specify at least one configuration option (--server, --user, --password, --topic-prefix) or --enable")
	}

	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Determine enable state
	var enable *bool
	if opts.Enable {
		t := true
		enable = &t
	}

	return cmdutil.RunWithSpinner(ctx, ios, "Configuring MQTT...", func(ctx context.Context) error {
		if err := svc.SetMQTTConfig(ctx, opts.Device, enable, opts.Server, opts.User, opts.Password, opts.TopicPrefix); err != nil {
			return fmt.Errorf("failed to configure MQTT: %w", err)
		}
		ios.Success("MQTT configured on %s", opts.Device)
		if opts.Server != "" {
			ios.Printf("  Server: %s\n", opts.Server)
		}
		return nil
	})
}
