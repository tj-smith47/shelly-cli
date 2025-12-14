// Package bulk provides bulk device provisioning.
package bulk

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// Config represents the bulk provisioning configuration file.
type Config struct {
	WiFi    *WiFiConfig    `yaml:"wifi,omitempty"`
	Devices []DeviceConfig `yaml:"devices"`
}

// WiFiConfig represents shared WiFi settings.
type WiFiConfig struct {
	SSID     string `yaml:"ssid"`
	Password string `yaml:"password"`
}

// DeviceConfig represents per-device settings.
type DeviceConfig struct {
	Name    string      `yaml:"name"`
	Address string      `yaml:"address,omitempty"`
	WiFi    *WiFiConfig `yaml:"wifi,omitempty"`
	DevName string      `yaml:"device_name,omitempty"`
}

// Options holds command options.
type Options struct {
	ConfigFile string
	Parallel   int
	DryRun     bool
	Factory    *cmdutil.Factory
}

// NewCommand creates the provision bulk command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "bulk <config-file>",
		Aliases: []string{"batch", "mass"},
		Short:   "Bulk provision from config file",
		Long: `Provision multiple devices from a YAML configuration file.

The config file specifies WiFi credentials and device-specific settings.
Provisioning is performed in parallel for efficiency.

Config file format:
  wifi:
    ssid: "MyNetwork"
    password: "secret"
  devices:
    - name: living-room
      address: 192.168.1.100  # optional, uses registered device if omitted
      device_name: "Living Room Light"  # optional device name to set
    - name: bedroom
      wifi:  # optional per-device WiFi override
        ssid: "OtherNetwork"
        password: "other-secret"`,
		Example: `  # Provision devices from config file
  shelly provision bulk devices.yaml

  # Dry run to validate config
  shelly provision bulk devices.yaml --dry-run

  # Limit parallel operations
  shelly provision bulk devices.yaml --parallel 2`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ConfigFile = args[0]
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVar(&opts.Parallel, "parallel", 5, "Maximum parallel provisioning operations")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Validate config without provisioning")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()

	// Load config file
	data, err := os.ReadFile(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if len(cfg.Devices) == 0 {
		return fmt.Errorf("no devices specified in config file")
	}

	ios.Info("Found %d devices to provision", len(cfg.Devices))

	if opts.DryRun {
		ios.Info("Dry run - validating configuration:")
		for _, d := range cfg.Devices {
			wifi := cfg.WiFi
			if d.WiFi != nil {
				wifi = d.WiFi
			}
			if wifi == nil {
				ios.Warning("  %s: no WiFi config", d.Name)
			} else {
				ios.Info("  %s: SSID=%s", d.Name, wifi.SSID)
			}
		}
		return nil
	}

	// Provision devices in parallel
	return provisionDevices(ctx, ios, opts, &cfg)
}

func provisionDevices(ctx context.Context, ios *iostreams.IOStreams, opts *Options, cfg *Config) error {
	svc := opts.Factory.ShellyService()

	type result struct {
		Device string
		Err    error
	}

	results := make(chan result, len(cfg.Devices))
	sem := make(chan struct{}, opts.Parallel)

	var wg sync.WaitGroup
	for _, device := range cfg.Devices {
		wg.Add(1)
		go func(d DeviceConfig) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			err := provisionDevice(ctx, svc, cfg, d)
			results <- result{Device: d.Name, Err: err}
		}(device)
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var failed []string
	for r := range results {
		if r.Err != nil {
			ios.Error("Failed to provision %s: %v", r.Device, r.Err)
			failed = append(failed, r.Device)
		} else {
			ios.Success("Provisioned %s", r.Device)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("%d devices failed to provision", len(failed))
	}

	ios.Success("All %d devices provisioned successfully", len(cfg.Devices))
	return nil
}

func provisionDevice(ctx context.Context, svc *shelly.Service, cfg *Config, device DeviceConfig) error {
	// Get WiFi config (device-specific or global)
	wifi := cfg.WiFi
	if device.WiFi != nil {
		wifi = device.WiFi
	}

	if wifi == nil {
		return fmt.Errorf("no WiFi configuration")
	}

	// Apply WiFi settings
	enable := true
	if err := svc.SetWiFiConfig(ctx, device.Name, wifi.SSID, wifi.Password, &enable); err != nil {
		return fmt.Errorf("failed to set WiFi: %w", err)
	}

	return nil
}
