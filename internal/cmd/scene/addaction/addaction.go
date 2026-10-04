// Package addaction provides the scene add-action subcommand.
package addaction

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
)

// Options holds the options for the add-action command.
type Options struct {
	Factory *cmdutil.Factory
	Scene   string
	Device  string
	Method  string
	Params  string
}

// NewCommand creates the scene add-action command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "add-action <scene> <device> <method> [params]",
		Aliases: []string{"add", "action"},
		Short:   "Add a device action to a scene",
		Long: `Add a device action to a scene.

An action is one RPC call that 'shelly scene activate' sends to a device.
The method is the RPC method name, and params is its parameters as a JSON
object. Actions run in the order they were added.

The device does not have to be reachable when the action is added: nothing is
sent until the scene is activated.`,
		Example: `  # Turn a switch on when the scene is activated
  shelly scene add-action movie-night living-room Switch.Set '{"id":0,"on":true}'

  # Dim a light
  shelly scene add-action movie-night lamp Light.Set '{"id":0,"on":true,"brightness":20}'

  # A method that takes no parameters
  shelly scene add-action bedtime hallway Shelly.Reboot

  # Using alias
  shelly scene add movie-night tv-plug Switch.Set '{"id":0,"on":false}'`,
		Args:              cobra.RangeArgs(3, 4),
		ValidArgsFunction: completion.SceneNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			opts.Scene, opts.Device, opts.Method = args[0], args[1], args[2]
			if len(args) == 4 {
				opts.Params = args[3]
			}
			return run(opts)
		},
	}

	return cmd
}

func run(opts *Options) error {
	ios := opts.Factory.IOStreams()

	action := config.SceneAction{Device: opts.Device, Method: opts.Method}
	if opts.Params != "" {
		if err := json.Unmarshal([]byte(opts.Params), &action.Params); err != nil {
			return fmt.Errorf("params must be a JSON object, for example '{\"id\":0,\"on\":true}': %w", err)
		}
	}

	if err := config.AddActionToScene(opts.Scene, action); err != nil {
		return err
	}

	ios.Success("Added %s on %s to scene %q", opts.Method, opts.Device, opts.Scene)
	return nil
}
