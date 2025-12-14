// Package get provides the kvs get subcommand.
package get

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Device  string
	Key     string
	Raw     bool // Output raw value only
	Factory *cmdutil.Factory
}

// NewCommand creates the kvs get command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "get <device> <key>",
		Aliases: []string{"g", "read"},
		Short:   "Get a KVS value",
		Long: `Retrieve a value from the device's Key-Value Storage.

Returns the value, its type, and the etag (version identifier).
Use --raw to output only the value without formatting.`,
		Example: `  # Get a value
  shelly kvs get living-room my_key

  # Get raw value only (for scripting)
  shelly kvs get living-room my_key --raw

  # Output as JSON
  shelly kvs get living-room my_key -o json`,
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: completeDeviceThenKey(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Raw, "raw", "r", false, "Output raw value only")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	return cmdutil.RunDeviceStatus(ctx, ios, svc, opts.Device,
		fmt.Sprintf("Getting key %q...", opts.Key),
		func(ctx context.Context, svc *shelly.Service, device string) (*shelly.KVSGetResult, error) {
			return svc.GetKVS(ctx, device, opts.Key)
		},
		func(ios *iostreams.IOStreams, result *shelly.KVSGetResult) {
			if opts.Raw {
				displayRaw(ios, result)
			} else {
				displayResult(ios, opts.Key, result)
			}
		})
}

func displayRaw(ios *iostreams.IOStreams, result *shelly.KVSGetResult) {
	switch v := result.Value.(type) {
	case string:
		ios.Println(v)
	case nil:
		ios.Println("null")
	default:
		// For other types (numbers, bools), use JSON encoding
		data, err := json.Marshal(v)
		if err != nil {
			ios.Printf("%v\n", v)
			return
		}
		ios.Println(string(data))
	}
}

func displayResult(ios *iostreams.IOStreams, key string, result *shelly.KVSGetResult) {
	label := theme.Highlight()

	ios.Printf("%s: %s\n", label.Render("Key"), key)
	ios.Printf("%s: %s\n", label.Render("Value"), formatValue(result.Value))
	ios.Printf("%s: %s\n", label.Render("Type"), valueType(result.Value))
	ios.Printf("%s: %s\n", label.Render("Etag"), result.Etag)
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func valueType(v any) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "bool"
	case float64:
		return "number"
	default:
		return "unknown"
	}
}

// completeDeviceThenKey provides completion for device and key arguments.
func completeDeviceThenKey() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			// First argument: complete device names
			devices := config.ListDevices()
			var completions []string
			for name := range devices {
				completions = append(completions, name)
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
		// Second argument: key (no completion - would require device query)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
