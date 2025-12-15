// Package status provides the matter status command.
package status

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the matter status command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "status <device>",
		Aliases: []string{"st", "info"},
		Short:   "Show Matter status",
		Long: `Show Matter connectivity status for a Shelly device.

Displays:
- Whether Matter is enabled
- Commissionable status (can be added to a fabric)
- Number of paired fabrics
- Network information when connected`,
		Example: `  # Show Matter status
  shelly matter status living-room

  # Output as JSON
  shelly matter status living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

// MatterStatus represents full Matter status.
type MatterStatus struct {
	Enabled        bool `json:"enabled"`
	Commissionable bool `json:"commissionable"`
	FabricsCount   int  `json:"fabrics_count"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	var status MatterStatus

	// Get Matter config
	cfg, err := svc.MatterGetConfig(ctx, opts.Device)
	if err != nil {
		ios.Debug("Matter.GetConfig failed: %v", err)
		return fmt.Errorf("matter not available on this device: %w", err)
	}
	status.Enabled = cfg.Enable

	// Get Matter status
	statusMap, err := svc.MatterGetStatus(ctx, opts.Device)
	if err != nil {
		ios.Debug("Matter.GetStatus failed: %v", err)
		// Config succeeded, show partial info
	} else {
		if commissionable, ok := statusMap["commissionable"].(bool); ok {
			status.Commissionable = commissionable
		}
		if fabricsCount, ok := statusMap["fabrics_count"].(float64); ok {
			status.FabricsCount = int(fabricsCount)
		}
	}

	if opts.JSON {
		output, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	ios.Println(theme.Bold().Render("Matter Status:"))
	ios.Println()

	enabledStr := theme.Dim().Render("Disabled")
	if status.Enabled {
		enabledStr = theme.StatusOK().Render("Enabled")
	}
	ios.Printf("  Enabled: %s\n", enabledStr)

	if status.Enabled {
		commissionStr := theme.Dim().Render("Not Commissionable")
		if status.Commissionable {
			commissionStr = theme.StatusOK().Render("Commissionable")
		}
		ios.Printf("  Status: %s\n", commissionStr)
		ios.Printf("  Paired Fabrics: %d\n", status.FabricsCount)

		if status.Commissionable {
			ios.Println()
			ios.Info("Device is ready to be added to a Matter fabric.")
			ios.Info("Use 'shelly matter code %s' to get the pairing code.", opts.Device)
		}
	} else {
		ios.Println()
		ios.Info("Enable Matter with: shelly matter enable %s", opts.Device)
	}

	return nil
}
