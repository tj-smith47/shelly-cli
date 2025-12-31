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

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	Concurrent int
	DryRun     bool
	Name       string
	Timeout    time.Duration
}

// NewCommand creates the scene activate command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{
		Factory:    f,
		Concurrent: 5,
		Timeout:    10 * time.Second,
	}

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
			opts.Name = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().DurationVarP(&opts.Timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&opts.Concurrent, "concurrent", "c", 5, "Max concurrent operations")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview actions without executing")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Cap concurrency to global rate limit
	concurrent := cmdutil.CapConcurrency(ios, opts.Concurrent)

	scene, exists := config.GetScene(opts.Name)
	if !exists {
		return fmt.Errorf("scene %q not found", opts.Name)
	}

	if len(scene.Actions) == 0 {
		ios.Warning("Scene %q has no actions", opts.Name)
		return nil
	}

	if opts.DryRun {
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

	ios.Info("Activating scene %q (%d actions)...", theme.Bold().Render(opts.Name), len(scene.Actions))

	svc := opts.Factory.ShellyService()

	// Create MultiWriter for progress tracking
	mw := iostreams.NewMultiWriter(ios.Out, ios.IsStdoutTTY())

	// Add all actions upfront (use device:method as identifier for clarity)
	for _, action := range scene.Actions {
		lineID := fmt.Sprintf("%s:%s", action.Device, action.Method)
		mw.AddLine(lineID, "pending")
	}

	// Create parent context with overall timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout*time.Duration(len(scene.Actions)))
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
			actionCtx, actionCancel := context.WithTimeout(ctx, opts.Timeout)
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
		ios.Warning("Scene %q: %d/%d actions failed", opts.Name, failed, len(scene.Actions))
		return fmt.Errorf("%d/%d actions failed", failed, len(scene.Actions))
	}

	ios.Success("Scene %q activated (%d actions)", opts.Name, succeeded)
	return nil
}
