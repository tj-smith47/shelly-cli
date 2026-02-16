package factories

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
)

// InputToggleOpts configures an input enable or disable command.
type InputToggleOpts struct {
	// Enable is true for enable commands, false for disable commands.
	Enable bool

	// Long is the long description.
	Long string

	// Example is the example usage text.
	Example string
}

// NewInputToggleCommand creates an input enable or disable command.
// This factory consolidates the get-config → check-state → modify → set-config pattern
// used by both input enable and input disable commands.
func NewInputToggleCommand(f *cmdutil.Factory, opts InputToggleOpts) *cobra.Command {
	v := newToggleVerbs(opts.Enable)

	var (
		cflags flags.ComponentFlags
		device string
	)

	cmd := &cobra.Command{
		Use:               fmt.Sprintf("%s <device>", v.Verb),
		Aliases:           []string{v.Alias},
		Short:             fmt.Sprintf("%s input component", capitalize(v.Verb)),
		Long:              opts.Long,
		Example:           opts.Example,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			device = args[0]
			return runInputToggle(cmd.Context(), f, device, cflags.ID, opts.Enable, v.Verb, v.Gerund, v.PastVerb)
		},
	}

	flags.AddComponentFlags(cmd, &cflags, "Input")

	return cmd
}

func runInputToggle(ctx context.Context, f *cmdutil.Factory, device string, id int, enable bool, verb, gerund, pastVerb string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	cfg, err := svc.InputGetConfig(ctx, device, id)
	if err != nil {
		return fmt.Errorf("failed to get input config: %w", err)
	}

	if cfg.Enable == enable {
		ios.Info("Input %d is already %s", id, pastVerb)
		return nil
	}

	cfg.Enable = enable

	err = cmdutil.RunWithSpinner(ctx, ios, fmt.Sprintf("%s input...", gerund), func(ctx context.Context) error {
		return svc.InputSetConfig(ctx, device, id, cfg)
	})
	if err != nil {
		return fmt.Errorf("failed to %s input: %w", verb, err)
	}

	ios.Success("Input %d %s", id, pastVerb)
	return nil
}
