// Package create provides the alert create subcommand.
package create

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/utils"
)

// Options holds the command options.
type Options struct {
	Factory     *cmdutil.Factory
	Name        string
	Device      string
	Condition   string
	Action      string
	Description string
}

// NewCommand creates the alert create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"add", "new"},
		Short:   "Create a monitoring alert",
		Long: `Create a new monitoring alert for device conditions.

Conditions can be:
  - offline: Device becomes unreachable
  - online: Device becomes reachable
  - power>N: Power consumption exceeds N watts
  - temperature>N: Temperature exceeds N degrees

Actions can be:
  - notify: Desktop notification (default)
  - webhook:URL: Send HTTP POST to URL
  - command:CMD: Execute shell command`,
		Example: `  # Alert when device goes offline
  shelly alert create kitchen-offline --device kitchen --condition offline

  # Alert on high power consumption
  shelly alert create high-power --device heater --condition "power>2000"

  # Alert with webhook action
  shelly alert create temp-alert --device sensor --condition "temperature>30" \
    --action "webhook:http://example.com/alert"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Device, "device", "d", "", "Device to monitor (required)")
	cmd.Flags().StringVarP(&opts.Condition, "condition", "c", "", "Alert condition (required)")
	cmd.Flags().StringVarP(&opts.Action, "action", "a", "notify", "Action when alert triggers")
	cmd.Flags().StringVar(&opts.Description, "description", "", "Alert description")

	utils.Must(cmd.MarkFlagRequired("device"))
	utils.Must(cmd.MarkFlagRequired("condition"))

	return cmd
}

func run(_ context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Check if alert already exists
	if _, exists := cfg.Alerts[opts.Name]; exists {
		return fmt.Errorf("alert %q already exists", opts.Name)
	}

	// Create alert
	alert := config.Alert{
		Name:        opts.Name,
		Description: opts.Description,
		Device:      opts.Device,
		Condition:   opts.Condition,
		Action:      opts.Action,
		Enabled:     true,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	cfg.Alerts[opts.Name] = alert

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	ios.Success("Created alert %q", opts.Name)
	ios.Printf("  Device: %s\n", opts.Device)
	ios.Printf("  Condition: %s\n", opts.Condition)
	ios.Printf("  Action: %s\n", opts.Action)

	return nil
}
