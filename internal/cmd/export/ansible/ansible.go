// Package ansible provides the export ansible subcommand.
package ansible

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// yamlExtensions defines valid YAML file extensions.
var yamlExtensions = []string{".yaml", ".yml"}

// Options holds command options.
type Options struct {
	Devices   []string
	File      string
	GroupName string
	Factory   *cmdutil.Factory
}

// NewCommand creates the export ansible command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "ansible <devices...> [file]",
		Aliases: []string{"ans"},
		Short:   "Export devices as Ansible inventory",
		Long: `Export devices as an Ansible inventory YAML file.

Creates an Ansible-compatible inventory with device groups based on
model type. Use @all to export all registered devices.`,
		Example: `  # Export to stdout
  shelly export ansible @all

  # Export to file
  shelly export ansible @all inventory.yaml

  # Export specific devices
  shelly export ansible living-room bedroom inventory.yaml

  # Specify group name
  shelly export ansible @all --group-name shelly_devices`,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: cmdutil.CompleteDevicesWithGroups(),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Devices, opts.File = cmdutil.SplitDevicesAndFile(args, yamlExtensions)
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.GroupName, "group-name", "shelly", "Ansible group name for devices")

	return cmd
}

// Inventory represents an Ansible inventory structure.
type Inventory struct {
	All Group `yaml:"all"`
}

// Group represents an Ansible group.
type Group struct {
	Hosts    map[string]Host  `yaml:"hosts,omitempty"`
	Children map[string]Group `yaml:"children,omitempty"`
}

// Host represents an Ansible host entry.
type Host struct {
	AnsibleHost string `yaml:"ansible_host"`
	ShellyModel string `yaml:"shelly_model"`
	ShellyGen   int    `yaml:"shelly_generation"`
	ShellyApp   string `yaml:"shelly_app,omitempty"`
}

func run(ctx context.Context, opts *Options) error {
	ctx, cancel := context.WithTimeout(ctx, 2*shelly.DefaultTimeout)
	defer cancel()

	ios := opts.Factory.IOStreams()
	svc := opts.Factory.ShellyService()

	// Expand @all to all registered devices
	devices := cmdutil.ExpandDeviceArgs(opts.Devices)
	if len(devices) == 0 {
		return fmt.Errorf("no devices specified")
	}

	// Collect device data using shared helper
	var deviceData []cmdutil.DeviceData
	err := cmdutil.RunWithSpinner(ctx, ios, "Fetching device data...", func(ctx context.Context) error {
		deviceData = cmdutil.CollectDeviceData(ctx, svc, devices)
		return nil
	})
	if err != nil {
		return err
	}

	// Group by model
	hostsByModel := make(map[string]map[string]Host)
	for _, d := range deviceData {
		host := Host{
			AnsibleHost: d.Address,
			ShellyModel: d.Model,
			ShellyGen:   d.Generation,
			ShellyApp:   d.App,
		}
		model := d.Model
		if hostsByModel[model] == nil {
			hostsByModel[model] = make(map[string]Host)
		}
		hostsByModel[model][d.Name] = host
	}

	// Build inventory structure
	inventory := Inventory{
		All: Group{
			Children: make(map[string]Group),
		},
	}

	// Create group with all hosts
	allHosts := make(map[string]Host)
	for _, hosts := range hostsByModel {
		for name, host := range hosts {
			allHosts[name] = host
		}
	}

	mainGroup := Group{
		Hosts:    allHosts,
		Children: make(map[string]Group),
	}

	// Add subgroups by model
	for model, hosts := range hostsByModel {
		groupName := normalizeGroupName(model)
		mainGroup.Children[groupName] = Group{Hosts: hosts}
	}

	inventory.All.Children[opts.GroupName] = mainGroup

	// Serialize to YAML
	data, err := yaml.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("failed to serialize inventory: %w", err)
	}

	// Output
	if opts.File == "" {
		ios.Printf("%s", string(data))
		return nil
	}

	if err := os.WriteFile(opts.File, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	ios.Success("Exported %d devices to %s", len(allHosts), opts.File)
	return nil
}

// normalizeGroupName converts a model name to a valid Ansible group name.
func normalizeGroupName(model string) string {
	name := strings.ToLower(model)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
