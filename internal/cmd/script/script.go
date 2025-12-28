// Package script provides the script management command group.
package script

import (
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmd/script/create"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/del"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/download"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/eval"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/get"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/list"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/start"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/stop"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/template"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/update"
	"github.com/tj-smith47/shelly-cli/internal/cmd/script/upload"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

// NewCommand creates the script command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "script",
		Aliases: []string{"sc"},
		Short:   "Manage device scripts",
		Long: `Manage JavaScript scripts on Gen2+ Shelly devices.

Scripts allow you to extend device functionality with custom JavaScript code.
Scripts can respond to events, automate actions, and interact with sensors.

Note: Scripts are only available on Gen2+ devices.`,
		Example: `  # List scripts on a device
  shelly script list living-room

  # Get script code
  shelly script get living-room 1

  # Create a new script
  shelly script create living-room --name "My Script" --file script.js

  # Start/stop a script
  shelly script start living-room 1
  shelly script stop living-room 1

  # Evaluate code on a running script
  shelly script eval living-room 1 "print('Hello!')"`,
	}

	cmd.AddCommand(list.NewCommand(f))
	cmd.AddCommand(get.NewCommand(f))
	cmd.AddCommand(create.NewCommand(f))
	cmd.AddCommand(update.NewCommand(f))
	cmd.AddCommand(del.NewCommand(f))
	cmd.AddCommand(start.NewCommand(f))
	cmd.AddCommand(stop.NewCommand(f))
	cmd.AddCommand(eval.NewCommand(f))
	cmd.AddCommand(upload.NewCommand(f))
	cmd.AddCommand(download.NewCommand(f))
	cmd.AddCommand(template.NewCommand(f))

	return cmd
}
