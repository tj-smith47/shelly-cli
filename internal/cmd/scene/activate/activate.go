// Package activate provides the scene activate subcommand.
package activate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// NewCommand creates the scene activate command.
func NewCommand() *cobra.Command {
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
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0], timeout, concurrent, dryRun)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Timeout per device")
	cmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Max concurrent operations")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview actions without executing")

	return cmd
}

// Result holds the result of a scene action execution.
type Result struct {
	Index  int
	Device string
	Method string
	Error  error
}

func run(ctx context.Context, name string, timeout time.Duration, concurrent int, dryRun bool) error {
	scene, exists := config.GetScene(name)
	if !exists {
		return fmt.Errorf("scene %q not found", name)
	}

	if len(scene.Actions) == 0 {
		iostreams.Warning("Scene %q has no actions", name)
		return nil
	}

	if dryRun {
		iostreams.Info("Dry run - would execute %d action(s):", len(scene.Actions))
		for i, action := range scene.Actions {
			iostreams.Info("  %d. %s %s %s",
				i+1,
				theme.Bold().Render(action.Device),
				theme.Highlight().Render(action.Method),
				formatParams(action.Params),
			)
		}
		return nil
	}

	iostreams.Info("Activating scene %q (%d actions)...", theme.Bold().Render(name), len(scene.Actions))

	svc := shelly.NewService()

	// Use mutex to protect results collection
	var mu sync.Mutex
	var results []Result

	// Create parent context with overall timeout
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Duration(len(scene.Actions)))
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(concurrent)

	for i, action := range scene.Actions {
		idx := i
		act := action // Capture for closure
		g.Go(func() error {
			// Per-action timeout
			actionCtx, actionCancel := context.WithTimeout(ctx, timeout)
			defer actionCancel()

			_, err := svc.RawRPC(actionCtx, act.Device, act.Method, act.Params)

			mu.Lock()
			results = append(results, Result{
				Index:  idx,
				Device: act.Device,
				Method: act.Method,
				Error:  err,
			})
			mu.Unlock()

			return nil // Don't fail the whole batch on individual errors
		})
	}

	// Wait for all operations to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("scene activation failed: %w", err)
	}

	// Report results
	var succeeded, failed int
	for _, result := range results {
		if result.Error != nil {
			iostreams.Warning("%s %s: %v",
				theme.Bold().Render(result.Device),
				result.Method,
				result.Error,
			)
			failed++
		} else {
			iostreams.Success("%s %s",
				theme.Bold().Render(result.Device),
				theme.Dim().Render(result.Method),
			)
			succeeded++
		}
	}

	// Print summary
	if failed > 0 {
		iostreams.Warning("Scene %q: %d/%d actions failed",
			name, failed, len(scene.Actions))
		return fmt.Errorf("%d/%d actions failed", failed, len(scene.Actions))
	}

	iostreams.Success("Scene %q activated (%d actions)", name, succeeded)
	return nil
}

func formatParams(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	result := "{"
	first := true
	for k, v := range params {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s=%v", k, v)
		first = false
	}
	result += "}"
	return theme.Dim().Render(result)
}
