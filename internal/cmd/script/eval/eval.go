// Package eval provides the script eval subcommand.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
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
		Args: cobra.MinimumNArgs(3),
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
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
	defer cancel()

	ios := f.IOStreams()
	svc := f.ShellyService()

	return cmdutil.RunWithSpinner(ctx, ios, "Evaluating code...", func(ctx context.Context) error {
		result, err := svc.EvalScript(ctx, device, id, code)
		if err != nil {
			return fmt.Errorf("failed to evaluate code: %w", err)
		}
		displayResult(ios, result)
		return nil
	})
}

func displayResult(ios *iostreams.IOStreams, result any) {
	if result == nil {
		ios.Info("(no result)")
		return
	}

	// Try to pretty-print JSON for complex types
	switch v := result.(type) {
	case string:
		ios.Println(v)
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			ios.Printf("%d\n", int64(v))
		} else {
			ios.Printf("%v\n", v)
		}
	case bool:
		ios.Printf("%t\n", v)
	default:
		// Try to marshal as JSON for complex types
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			ios.Printf("%v\n", result)
		} else {
			ios.Println(string(jsonBytes))
		}
	}
}
