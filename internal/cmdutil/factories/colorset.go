// Package factories provides command factory functions for creating standard CLI commands.
package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// ColorSetOpts configures a color set command (RGB or RGBW).
type ColorSetOpts struct {
	// Component is the display name (e.g., "RGB", "RGBW").
	Component string

	// HasWhite enables the --white flag for RGBW components.
	HasWhite bool

	// SetFunc performs the set operation with the collected parameters.
	// The white value is -1 (unset) when HasWhite is false.
	SetFunc func(ctx context.Context, svc *cmdutil.Factory, device string, id, red, green, blue, white, brightness int, on bool) error
}

// colorSetOptions holds the runtime options for a color set command.
type colorSetOptions struct {
	flags.ComponentFlags
	Factory    *cmdutil.Factory
	Device     string
	Red        int
	Green      int
	Blue       int
	White      int
	Brightness int
	On         bool
	Config     ColorSetOpts
}

// NewColorSetCommand creates a color set command for RGB or RGBW components.
// This factory consolidates the common pattern across rgb/set and rgbw/set.
func NewColorSetCommand(f *cmdutil.Factory, cfg ColorSetOpts) *cobra.Command {
	componentLower := toLower(cfg.Component)

	opts := &colorSetOptions{
		Factory:    f,
		Red:        -1,
		Green:      -1,
		Blue:       -1,
		White:      -1,
		Brightness: -1,
		Config:     cfg,
	}

	whiteDesc := ""
	if cfg.HasWhite {
		whiteDesc = ", white channel,"
	}

	long := fmt.Sprintf(`Set parameters of an %s light component on the specified device.

You can set color values (red, green, blue)%s brightness, and on/off state.
Values not specified will be left unchanged.`, cfg.Component, whiteDesc)

	var examples string
	if cfg.HasWhite {
		examples = fmt.Sprintf(`  # Set %s color to red
  shelly %s set living-room --red 255 --green 0 --blue 0

  # Set %s with white channel
  shelly %s set living-room -r 255 -g 200 -b 150 --white 128

  # Set %s with brightness
  shelly %s color living-room -r 0 -g 255 -b 128 --brightness 75

  # Set only white channel
  shelly %s set living-room --white 200`,
			cfg.Component, componentLower, cfg.Component, componentLower,
			cfg.Component, componentLower, componentLower)
	} else {
		examples = fmt.Sprintf(`  # Set %s color to red
  shelly %s set living-room --red 255 --green 0 --blue 0

  # Set %s with brightness
  shelly %s color living-room -r 0 -g 255 -b 128 --brightness 75`,
			cfg.Component, componentLower, cfg.Component, componentLower)
	}

	cmd := &cobra.Command{
		Use:               "set <device>",
		Aliases:           []string{"color", "c"},
		Short:             fmt.Sprintf("Set %s parameters", cfg.Component),
		Long:              long,
		Example:           examples,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return runColorSet(cmd.Context(), opts)
		},
	}

	flags.AddComponentFlags(cmd, &opts.ComponentFlags, cfg.Component)
	cmd.Flags().IntVarP(&opts.Red, "red", "r", -1, "Red value (0-255)")
	cmd.Flags().IntVarP(&opts.Green, "green", "g", -1, "Green value (0-255)")
	cmd.Flags().IntVarP(&opts.Blue, "blue", "b", -1, "Blue value (0-255)")
	if cfg.HasWhite {
		cmd.Flags().IntVarP(&opts.White, "white", "w", -1, "White channel value (0-100)")
	}
	cmd.Flags().IntVar(&opts.Brightness, "brightness", -1, "Brightness (0-100)")
	cmd.Flags().BoolVar(&opts.On, "on", false, "Turn on")

	return cmd
}

func runColorSet(ctx context.Context, opts *colorSetOptions) error {
	f := opts.Factory
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()

	err := cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("Setting %s parameters...", opts.Config.Component), func(ctx context.Context) error {
		return opts.Config.SetFunc(ctx, f, opts.Device, opts.ID, opts.Red, opts.Green, opts.Blue, opts.White, opts.Brightness, opts.On)
	})
	if err != nil {
		return fmt.Errorf("failed to set %s parameters: %w", opts.Config.Component, err)
	}

	ios.Success("%s %d parameters set", opts.Config.Component, opts.ID)
	return nil
}

// toLower is a simple lowercase helper to avoid importing strings for one call.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}
