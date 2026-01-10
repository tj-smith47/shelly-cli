// Package methods provides the api methods subcommand.
package methods

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/completion"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds options for the methods subcommand.
type Options struct {
	flags.OutputFlags
	Factory *cmdutil.Factory
	Device  string
	Filter  string
}

// NewCommand creates the api methods subcommand.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "methods <device>",
		Aliases: []string{"list-methods", "lm"},
		Short:   "List available RPC methods (Gen2+ only)",
		Long: `List all RPC methods available on a Shelly device.

This shows the methods you can call using 'shelly api <device> <Method>'.
Use --filter to search for specific methods by name.

Note: This command only works with Gen2+ devices as Gen1 devices don't
support RPC method introspection.`,
		Example: `  # List all methods
  shelly api methods living-room

  # Filter methods containing "Switch"
  shelly api methods living-room --filter Switch

  # Output as JSON
  shelly api methods living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.DeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter methods by name (case-insensitive)")
	flags.AddOutputFlagsCustom(cmd, &opts.OutputFlags, "text", "text", "json")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := opts.Factory.WithDefaultTimeout(ctx)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	result, err := svc.RawRPC(ctx, opts.Device, "Shelly.ListMethods", nil)
	if err != nil {
		return fmt.Errorf("failed to list methods: %w", err)
	}

	// Parse the result
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	var resp struct {
		Methods []string `json:"methods"`
	}
	if err := json.Unmarshal(jsonBytes, &resp); err != nil {
		return fmt.Errorf("failed to parse methods: %w", err)
	}

	// Sort methods alphabetically
	sort.Strings(resp.Methods)

	// Filter if requested
	methods := resp.Methods
	if opts.Filter != "" {
		filter := strings.ToLower(opts.Filter)
		filtered := make([]string, 0, len(methods))
		for _, m := range methods {
			if strings.Contains(strings.ToLower(m), filter) {
				filtered = append(filtered, m)
			}
		}
		methods = filtered
	}

	// Output
	if opts.Format == "json" {
		jsonOutput, err := json.MarshalIndent(methods, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(jsonOutput))
		return nil
	}

	ios.Println(theme.Bold().Render(fmt.Sprintf("Available RPC Methods (%d):", len(methods))))
	ios.Println()

	// Group methods by namespace
	namespaces := make(map[string][]string)
	for _, m := range methods {
		parts := strings.SplitN(m, ".", 2)
		ns := parts[0]
		namespaces[ns] = append(namespaces[ns], m)
	}

	// Get sorted namespace keys
	nsKeys := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		nsKeys = append(nsKeys, ns)
	}
	sort.Strings(nsKeys)

	// Print by namespace
	for _, ns := range nsKeys {
		ios.Println("  " + theme.Highlight().Render(ns+":"))
		for _, method := range namespaces[ns] {
			// Just the method name without namespace
			parts := strings.SplitN(method, ".", 2)
			if len(parts) == 2 {
				ios.Println("    " + parts[1])
			} else {
				ios.Println("    " + method)
			}
		}
		ios.Println()
	}

	return nil
}
