// Package set provides the virtual set command.
package set

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Options holds command options.
type Options struct {
	Device  string
	Key     string
	Value   string
	Toggle  bool
	Factory *cmdutil.Factory
}

// NewCommand creates the virtual set command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "set <device> <key> <value>",
		Aliases: []string{"update"},
		Short:   "Set a virtual component value",
		Long: `Set the value of a virtual component.

The key format is "type:id", for example "boolean:200" or "number:201".

For boolean components, use --toggle to flip the current value.
For button components, use "trigger" as the value to press the button.`,
		Example: `  # Set a boolean to true
  shelly virtual set kitchen boolean:200 true

  # Toggle a boolean
  shelly virtual set kitchen boolean:200 --toggle

  # Set a number
  shelly virtual set kitchen number:201 25.5

  # Set text
  shelly virtual set kitchen text:202 "Hello World"

  # Set enum value
  shelly virtual set kitchen enum:203 "option1"

  # Trigger a button
  shelly virtual set kitchen button:204 trigger`,
		Args:              cobra.RangeArgs(2, 3),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			opts.Key = args[1]
			if len(args) > 2 {
				opts.Value = args[2]
			}
			if !opts.Toggle && opts.Value == "" {
				return fmt.Errorf("value required (or use --toggle for booleans)")
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Toggle, "toggle", "t", false, "Toggle boolean value instead of setting")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Parse key
	compType, id, err := shelly.ParseVirtualKey(opts.Key)
	if err != nil {
		return err
	}

	err = cmdutil.RunWithSpinner(ctx, ios, "Setting virtual component...", func(ctx context.Context) error {
		switch compType {
		case "boolean":
			if opts.Toggle {
				return svc.ToggleVirtualBoolean(ctx, opts.Device, id)
			}
			val, parseErr := strconv.ParseBool(opts.Value)
			if parseErr != nil {
				return fmt.Errorf("invalid boolean value %q", opts.Value)
			}
			return svc.SetVirtualBoolean(ctx, opts.Device, id, val)

		case "number":
			val, parseErr := strconv.ParseFloat(opts.Value, 64)
			if parseErr != nil {
				return fmt.Errorf("invalid number value %q: %w", opts.Value, parseErr)
			}
			return svc.SetVirtualNumber(ctx, opts.Device, id, val)

		case "text":
			return svc.SetVirtualText(ctx, opts.Device, id, opts.Value)

		case "enum":
			return svc.SetVirtualEnum(ctx, opts.Device, id, opts.Value)

		case "button":
			if !strings.EqualFold(opts.Value, "trigger") {
				return fmt.Errorf("use 'trigger' to press a virtual button")
			}
			return svc.TriggerVirtualButton(ctx, opts.Device, id)

		case "group":
			return fmt.Errorf("group components cannot be set directly")

		default:
			return fmt.Errorf("unknown component type %q", compType)
		}
	})
	if err != nil {
		return err
	}

	switch {
	case opts.Toggle:
		ios.Success("Toggled %s", opts.Key)
	case compType == "button":
		ios.Success("Triggered %s", opts.Key)
	default:
		ios.Success("Set %s to %s", opts.Key, opts.Value)
	}

	return nil
}
