// Package get provides the config get subcommand.
package get

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"

	"github.com/tj-smith47/shelly-cli/internal/output"
)

// NewCommand creates the config get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <device> [component]",
		Short: "Get device configuration",
		Long: `Get configuration for a device or specific component.

Without a component argument, returns the full device configuration.
With a component argument (e.g., "switch:0", "sys", "wifi"), returns
only that component's configuration.`,
		Example: `  # Get full device configuration
  shelly config get living-room

  # Get switch:0 configuration
  shelly config get living-room switch:0

  # Get system configuration
  shelly config get living-room sys

  # Get WiFi configuration
  shelly config get living-room wifi

  # Output as JSON
  shelly config get living-room -o json

  # Output as YAML
  shelly config get living-room -o yaml`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			device := args[0]
			component := ""
			if len(args) > 1 {
				component = args[1]
			}
			return run(cmd.Context(), f, device, component)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device, component string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	svc := f.ShellyService()

	spin := iostreams.NewSpinner("Getting configuration...")
	spin.Start()

	config, err := svc.GetConfig(ctx, device)
	spin.Stop()

	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	// If a specific component was requested, extract just that
	var result any = config
	if component != "" {
		compConfig, ok := config[component]
		if !ok {
			return fmt.Errorf("component %q not found in configuration", component)
		}
		result = map[string]any{component: compConfig}
	}

	// Output based on format
	format := viper.GetString("output")
	switch format {
	case "json":
		return output.PrintJSON(result)
	case "yaml":
		return output.PrintYAML(result)
	default:
		return printConfigTable(result)
	}
}

// printConfigTable prints configuration as a formatted table.
func printConfigTable(config any) error {
	configMap, ok := config.(map[string]any)
	if !ok {
		return output.PrintJSON(config)
	}

	for component, cfg := range configMap {
		iostreams.Title("%s", component)

		cfgMap, ok := cfg.(map[string]any)
		if !ok {
			// If it's not a map, just print it as JSON
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				iostreams.DebugErr("marshaling config component", err)
			} else {
				fmt.Println(string(data))
			}
			fmt.Println()
			continue
		}

		table := output.NewTable("Setting", "Value")
		for key, value := range cfgMap {
			table.AddRow(key, formatValue(value))
		}
		table.Print()
		fmt.Println()
	}

	return nil
}

// formatValue formats a configuration value for display.
func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "<not set>"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		// Check if it's an integer
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		if val == "" {
			return "<empty>"
		}
		return val
	case map[string]any, []any:
		data, err := json.Marshal(val)
		if err != nil {
			iostreams.DebugErr("marshaling config value", err)
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}
