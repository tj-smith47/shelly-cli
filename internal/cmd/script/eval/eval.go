// Package eval provides the script eval subcommand.
package eval

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/term"
)

// NewCommand creates the script eval command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eval <device> <id> <code>",
		Aliases: []string{"exec"},
		Short:   "Evaluate JavaScript code",
		Long: `Evaluate a JavaScript expression in the context of a running script.

The script must be running for eval to work. The code argument can be
multiple words that will be joined together.`,
		Example: `  # Evaluate a simple expression
  shelly script eval living-room 1 "1 + 2"

  # Print a message
  shelly script eval living-room 1 "print('Hello from CLI!')"

  # Call a function defined in the script
  shelly script eval living-room 1 "myFunction()"`,
		Args:              cobra.MinimumNArgs(3),
		ValidArgsFunction: completion.DeviceThenScriptID(),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid script ID: %s", args[1])
			}
			code := strings.Join(args[2:], " ")
			return run(cmd.Context(), f, args[0], id, code)
		},
	}

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, id int, code string) error {
	ctx, cancel := f.WithDefaultTimeout(ctx)
	defer cancel()

	ios := f.IOStreams()
	svc := f.AutomationService()

	return cmdutil.RunWithSpinner(ctx, ios, "Evaluating code...", func(ctx context.Context) error {
		result, err := svc.EvalScript(ctx, device, id, code)
		if err != nil {
			return fmt.Errorf("failed to evaluate code: %w", err)
		}
		term.DisplayScriptEvalResult(ios, result)
		return nil
	})
}
