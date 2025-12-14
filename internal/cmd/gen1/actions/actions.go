// Package actions provides the gen1 actions command.
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds command options.
type Options struct {
	Factory *cmdutil.Factory
	Device  string
	JSON    bool
}

// NewCommand creates the gen1 actions command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "actions <device>",
		Aliases: []string{"urls", "webhooks"},
		Short:   "List Gen1 action URLs",
		Long: `List the action URLs configured on a Gen1 Shelly device.

Gen1 devices use action URLs (webhooks) to trigger external services
when events occur. These are configured per-relay or per-input and
triggered on events like:
- Relay on/off
- Short/long button press
- Input state change

Note: Gen2+ devices use webhooks instead. See 'shelly webhook' commands.`,
		Example: `  # List action URLs
  shelly gen1 actions living-room

  # Output as JSON
  shelly gen1 actions living-room --json`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: cmdutil.CompleteDeviceNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Device = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Output as JSON")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	settings, err := fetchSettings(ctx, ios, opts.Device)
	if err != nil {
		return err
	}

	actionsData := collectActions(settings)

	if opts.JSON {
		output, err := json.MarshalIndent(actionsData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		ios.Println(string(output))
		return nil
	}

	return displayActions(ios, actionsData, opts.Device)
}

func fetchSettings(ctx context.Context, ios *iostreams.IOStreams, device string) (map[string]any, error) {
	devCfg, err := config.ResolveDevice(device)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve device: %w", err)
	}

	address := devCfg.Address
	if address == "" {
		return nil, fmt.Errorf("device %s has no address configured", device)
	}

	// Ensure http:// prefix
	if len(address) < 7 || address[:7] != "http://" {
		address = "http://" + address
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address+"/settings", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if devCfg.Auth != nil && devCfg.Auth.Username != "" {
		req.SetBasicAuth(devCfg.Auth.Username, devCfg.Auth.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			ios.Debug("failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return settings, nil
}

func collectActions(settings map[string]any) map[string]any {
	actionsData := make(map[string]any)

	// Check for actions in different locations
	if actions, ok := settings["actions"].(map[string]any); ok {
		actionsData["actions"] = actions
	}

	collectRelayActions(settings, actionsData)
	collectInputActions(settings, actionsData)

	return actionsData
}

func collectRelayActions(settings, actionsData map[string]any) {
	relays, ok := settings["relays"].([]any)
	if !ok {
		return
	}

	relayEvents := []string{"btn_on_url", "btn_off_url", "out_on_url", "out_off_url", "shortpush_url", "longpush_url"}

	for i, r := range relays {
		relay, ok := r.(map[string]any)
		if !ok {
			continue
		}

		relayActions := make(map[string]string)
		for _, event := range relayEvents {
			if url, ok := relay[event].(string); ok && url != "" {
				relayActions[event] = url
			}
		}

		if len(relayActions) > 0 {
			actionsData[fmt.Sprintf("relay_%d", i)] = relayActions
		}
	}
}

func collectInputActions(settings, actionsData map[string]any) {
	inputs, ok := settings["inputs"].([]any)
	if !ok {
		return
	}

	inputEvents := []string{"btn_on_url", "btn_off_url", "shortpush_url", "longpush_url", "double_shortpush_url", "triple_shortpush_url"}

	for i, inp := range inputs {
		input, ok := inp.(map[string]any)
		if !ok {
			continue
		}

		inputActions := make(map[string]string)
		for _, event := range inputEvents {
			if url, ok := input[event].(string); ok && url != "" {
				inputActions[event] = url
			}
		}

		if len(inputActions) > 0 {
			actionsData[fmt.Sprintf("input_%d", i)] = inputActions
		}
	}
}

func displayActions(ios *iostreams.IOStreams, actionsData map[string]any, device string) error {
	ios.Println(theme.Bold().Render("Gen1 Action URLs:"))
	ios.Println()

	if len(actionsData) == 0 {
		devCfg, err := config.ResolveDevice(device)
		ios.Info("No action URLs configured.")
		if err == nil && devCfg.Address != "" {
			ios.Info("Configure actions in the device web UI at http://%s", devCfg.Address)
		}
		return nil
	}

	for component, actions := range actionsData {
		ios.Printf("  %s:\n", theme.Highlight().Render(component))

		switch v := actions.(type) {
		case map[string]string:
			for event, url := range v {
				ios.Printf("    %s: %s\n", event, url)
			}
		case map[string]any:
			for event, url := range v {
				ios.Printf("    %s: %v\n", event, url)
			}
		}
		ios.Println()
	}

	return nil
}
