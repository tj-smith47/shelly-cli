// Package activate provides the scene activate subcommand.
package activate

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the scene activate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		timeout    time.Duration
		concurrent int
		dryRun     bool
	)

	cmd := &cobra.Command{
		Use:     "activate <name>",
		Aliases: []string{"run", "exec", "play"},
		Short:   "Activate a scene",
		Long: `Execute all actions defined in a scene.

Actions are executed concurrently for faster execution.
Use --dry-run to preview actions without executing them.`,
		Example: `  # Activate a scene
  shelly scene activate movie-night

  # Preview without executing
  shelly scene activate movie-night --dry-run

  # Using aliases
  shelly scene run bedtime
  shelly scene play morning-routine

  # Short form
  shelly sc activate party-mode`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.SceneNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0], timeout, concurrent, dryRun)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview actions without executing")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, name string, timeout time.Duration, concurrent int, dryRun bool) error {
	ios := f.IOStreams()

	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	if len(scene.Actions) == 0 {
		ios.Warning("Scene %q has no actions", name)
		return nil
	}

	if dryRun {
		ios.Info("Dry run - would execute %d action(s):", len(scene.Actions))
		for i, action := range scene.Actions {
			params := output.FormatParamsInline(action.Params)
			if params != "" {
				params = theme.Dim().Render("{" + params + "}")
			}
			ios.Info("  %d. %s %s %s",
				i+1,
				theme.Bold().Render(action.Device),
				theme.Highlight().Render(action.Method),
				params,
			)
		}
		return nil
	}

	ios.Info("Activating scene %q (%d actions)...", theme.Bold().Render(name), len(scene.Actions))

	svc := f.ShellyService()

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

	// Add all actions upfront (use device:method as identifier for clarity)
	for _, action := range scene.Actions {
		lineID := fmt.Sprintf("%s:%s", action.Device, action.Method)
		mw.AddLine(lineID, "pending")
	}

	// Create parent context with overall timeout
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(scene.Actions)))
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrent)

	for _, action := range scene.Actions {
		act := action // Capture for closure
		lineID := fmt.Sprintf("%s:%s", act.Device, act.Method)
		g.Go(func() error {
			params := output.FormatParamsInline(act.Params)
			if params != "" {
				params = theme.Dim().Render("{" + params + "}")
			}
			mw.UpdateLine(lineID, iostreams.StatusRunning, params)

			// Per-action timeout
			actionCtx, actionCancel := context.WithTimeout(ctx, timeout)
			defer actionCancel()

			_, err := svc.RawRPC(actionCtx, act.Device, act.Method, act.Params)
			if err != nil {
				mw.UpdateLine(lineID, iostreams.StatusError, err.Error())
			} else {
				mw.UpdateLine(lineID, iostreams.StatusSuccess, "done")
			}

			return nil // Don't fail the whole batch on individual errors
		})
	}

	// Wait for all operations to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("scene activation failed: %w", err)
	}

	mw.Finalize()

	// Get summary from MultiWriter
	succeeded, failed, _ := mw.Summary()

	// Print summary
	if failed > 0 {
		ios.Warning("Scene %q: %d/%d actions failed", name, failed, len(scene.Actions))
		return fmt.Errorf("%d/%d actions failed", failed, len(scene.Actions))
	}

	ios.Success("Scene %q activated (%d actions)", name, succeeded)
	return nil
}
