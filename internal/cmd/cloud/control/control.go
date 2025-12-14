// Package control provides the cloud control subcommand.
package control

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var channelFlag int

// NewCommand creates the cloud control command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "control <device-id> <action>",
		Aliases: []string{"ctrl", "cmd"},
		Short:   "Control a device via cloud",
		Long: `Control a Shelly device through the Shelly Cloud API.

Supported actions:
  switch:  on, off, toggle
  cover:   open, close, stop, position=<0-100>
  light:   on, off, toggle, brightness=<0-100>

This command requires authentication with 'shelly cloud login'.`,
		Example: `  # Turn on a switch
  shelly cloud control abc123 on

  # Turn off switch on channel 1
  shelly cloud control abc123 off --channel 1

  # Toggle a switch
  shelly cloud control abc123 toggle

  # Set cover to 50%
  shelly cloud control abc123 position=50

  # Open cover
  shelly cloud control abc123 open

  # Set light brightness
  shelly cloud control abc123 brightness=75`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], args[1])
		},
	}

	cmd.Flags().IntVar(&channelFlag, "channel", 0, "Device channel/relay number")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, deviceID, action string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2)
	defer cancel()

	ios := f.IOStreams()

	// Check if logged in
	cfg := config.Get()
	if cfg.Cloud.AccessToken == "" {
		ios.Error("Not logged in to Shelly Cloud")
		ios.Info("Use 'shelly cloud login' to authenticate")
		return fmt.Errorf("not logged in")
	}

	// Create cloud client
	client := shelly.NewCloudClient(cfg.Cloud.AccessToken)

	return cmdutil.RunWithSpinner(ctx, ios, "Sending command...", func(ctx context.Context) error {
		return executeAction(ctx, ios, client, deviceID, action)
	})
}

func executeAction(ctx context.Context, ios *iostreams.IOStreams, client *shelly.CloudClient, deviceID, action string) error {
	actionLower := strings.ToLower(action)

	// Handle simple actions
	if err := handleSimpleAction(ctx, ios, client, deviceID, actionLower); err == nil {
		return nil
	} else if !errors.Is(err, errNotSimpleAction) {
		return err
	}

	// Handle parameterized actions
	return handleParameterizedAction(ctx, ios, client, deviceID, actionLower)
}

var errNotSimpleAction = fmt.Errorf("not a simple action")

type actionHandler struct {
	fn      func(ctx context.Context, client *shelly.CloudClient, deviceID string, channel int) error
	success string
	errMsg  string
}

func buildActionHandlers() map[string]actionHandler {
	return map[string]actionHandler{
		"on": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.SetSwitch(ctx, d, ch, true)
			},
			success: "Switch turned on",
			errMsg:  "failed to turn on switch",
		},
		"off": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.SetSwitch(ctx, d, ch, false)
			},
			success: "Switch turned off",
			errMsg:  "failed to turn off switch",
		},
		"toggle": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.ToggleSwitch(ctx, d, ch)
			},
			success: "Switch toggled",
			errMsg:  "failed to toggle switch",
		},
		"open": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.OpenCover(ctx, d, ch)
			},
			success: "Cover opening",
			errMsg:  "failed to open cover",
		},
		"close": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.CloseCover(ctx, d, ch)
			},
			success: "Cover closing",
			errMsg:  "failed to close cover",
		},
		"stop": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.StopCover(ctx, d, ch)
			},
			success: "Cover stopped",
			errMsg:  "failed to stop cover",
		},
		"light-on": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.SetLight(ctx, d, ch, true)
			},
			success: "Light turned on",
			errMsg:  "failed to turn on light",
		},
		"light-off": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.SetLight(ctx, d, ch, false)
			},
			success: "Light turned off",
			errMsg:  "failed to turn off light",
		},
		"light-toggle": {
			fn: func(ctx context.Context, c *shelly.CloudClient, d string, ch int) error {
				return c.ToggleLight(ctx, d, ch)
			},
			success: "Light toggled",
			errMsg:  "failed to toggle light",
		},
	}
}

func handleSimpleAction(ctx context.Context, ios *iostreams.IOStreams, client *shelly.CloudClient, deviceID, action string) error {
	handlers := buildActionHandlers()

	handler, ok := handlers[action]
	if !ok {
		return errNotSimpleAction
	}

	if err := handler.fn(ctx, client, deviceID, channelFlag); err != nil {
		return fmt.Errorf("%s: %w", handler.errMsg, err)
	}

	ios.Success("%s", handler.success)
	return nil
}

func handleParameterizedAction(ctx context.Context, ios *iostreams.IOStreams, client *shelly.CloudClient, deviceID, action string) error {
	switch {
	case strings.HasPrefix(action, "position="):
		return handlePositionAction(ctx, ios, client, deviceID, action)

	case strings.HasPrefix(action, "brightness="):
		return handleBrightnessAction(ctx, ios, client, deviceID, action)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func handlePositionAction(ctx context.Context, ios *iostreams.IOStreams, client *shelly.CloudClient, deviceID, action string) error {
	var position int
	if _, err := fmt.Sscanf(action, "position=%d", &position); err != nil {
		return fmt.Errorf("invalid position format: %s", action)
	}
	if position < 0 || position > 100 {
		return fmt.Errorf("position must be 0-100, got %d", position)
	}
	if err := client.SetCoverPosition(ctx, deviceID, channelFlag, position); err != nil {
		return fmt.Errorf("failed to set cover position: %w", err)
	}
	ios.Success("Cover position set to %d%%", position)
	return nil
}

func handleBrightnessAction(ctx context.Context, ios *iostreams.IOStreams, client *shelly.CloudClient, deviceID, action string) error {
	var brightness int
	if _, err := fmt.Sscanf(action, "brightness=%d", &brightness); err != nil {
		return fmt.Errorf("invalid brightness format: %s", action)
	}
	if brightness < 0 || brightness > 100 {
		return fmt.Errorf("brightness must be 0-100, got %d", brightness)
	}
	if err := client.SetLightBrightness(ctx, deviceID, channelFlag, brightness); err != nil {
		return fmt.Errorf("failed to set light brightness: %w", err)
	}
	ios.Success("Light brightness set to %d%%", brightness)
	return nil
}
