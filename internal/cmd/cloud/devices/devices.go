// Package devices provides the cloud devices subcommand.
package devices

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the cloud devices command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "devices",
		Aliases: []string{"ls", "list"},
		Short:   "List cloud-registered devices",
		Long: `List all devices registered with your Shelly Cloud account.

Shows device ID, name, model, firmware version, and online status.`,
		Example: `  # List all cloud devices
  shelly cloud devices

  # Output as JSON
  shelly cloud devices -o json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(f, cmd.Context())
		},
	}

	return cmd
}

func run(f *cmdutil.Factory, ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*2) // Longer timeout for cloud
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

	return cmdutil.RunWithSpinner(ctx, ios, "Fetching devices from cloud...", func(ctx context.Context) error {
		devices, err := client.GetAllDevices(ctx)
		if err != nil {
			return fmt.Errorf("failed to get devices: %w", err)
		}

		displayDevices(ios, devices)
		return nil
	})
}

func displayDevices(ios *iostreams.IOStreams, devices []shelly.CloudDevice) {
	if len(devices) == 0 {
		ios.Info("No devices found in your Shelly Cloud account")
		return
	}

	// Sort by ID for consistent display
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})

	table := output.NewTable("ID", "Model", "Gen", "Online")

	for _, d := range devices {
		model := d.Model
		if model == "" {
			model = output.FormatPlaceholder("unknown")
		}

		gen := output.FormatPlaceholder("-")
		if d.Generation > 0 {
			gen = fmt.Sprintf("%d", d.Generation)
		}

		table.AddRow(d.ID, model, gen, output.RenderYesNo(d.Online, output.CaseLower, theme.FalseError))
	}

	ios.Printf("Found %d device(s):\n\n", len(devices))
	if err := table.PrintTo(ios.Out); err != nil {
		ios.DebugErr("print table", err)
	}
}
