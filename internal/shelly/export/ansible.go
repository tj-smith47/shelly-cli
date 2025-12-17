// Package export provides export format builders for device data.
package export

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

// AnsibleInventory represents an Ansible inventory structure.
type AnsibleInventory struct {
	All AnsibleGroup `yaml:"all"`
}

// AnsibleGroup represents an Ansible group.
type AnsibleGroup struct {
	Hosts    map[string]AnsibleHost  `yaml:"hosts,omitempty"`
	Children map[string]AnsibleGroup `yaml:"children,omitempty"`
}

// AnsibleHost represents an Ansible host entry.
type AnsibleHost struct {
	AnsibleHost string `yaml:"ansible_host"`
	ShellyModel string `yaml:"shelly_model"`
	ShellyGen   int    `yaml:"shelly_generation"`
	ShellyApp   string `yaml:"shelly_app,omitempty"`
}

// BuildAnsibleInventory builds an Ansible inventory from device data.
// The groupName parameter sets the top-level group name (default: "shelly").
// Devices are grouped by model under the main group.
func BuildAnsibleInventory(devices []shelly.DeviceData, groupName string) (*AnsibleInventory, []byte, error) {
	// Group by model
	hostsByModel := make(map[string]map[string]AnsibleHost)
	for _, d := range devices {
		host := AnsibleHost{
			AnsibleHost: d.Address,
			ShellyModel: d.Model,
			ShellyGen:   d.Generation,
			ShellyApp:   d.App,
		}
		model := d.Model
		if hostsByModel[model] == nil {
			hostsByModel[model] = make(map[string]AnsibleHost)
		}
		hostsByModel[model][d.Name] = host
	}

	// Build inventory structure
	inventory := &AnsibleInventory{
		All: AnsibleGroup{
			Children: make(map[string]AnsibleGroup),
		},
	}

	// Create group with all hosts
	allHosts := make(map[string]AnsibleHost)
	for _, hosts := range hostsByModel {
		for name, host := range hosts {
			allHosts[name] = host
		}
	}

	mainGroup := AnsibleGroup{
		Hosts:    allHosts,
		Children: make(map[string]AnsibleGroup),
	}

	// Add subgroups by model
	for model, hosts := range hostsByModel {
		subGroupName := NormalizeGroupName(model)
		mainGroup.Children[subGroupName] = AnsibleGroup{Hosts: hosts}
	}

	inventory.All.Children[groupName] = mainGroup

	// Serialize to YAML
	data, err := yaml.Marshal(inventory)
	if err != nil {
		return nil, nil, err
	}

	return inventory, data, nil
}

// NormalizeGroupName converts a model name to a valid Ansible group name.
func NormalizeGroupName(model string) string {
	name := strings.ToLower(model)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
