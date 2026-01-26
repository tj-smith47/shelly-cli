package diagram

import (
	"fmt"
	"sort"
	"strings"
)

// catalog is the static registry of all known Shelly device models.
var catalog = []DeviceModel{
	// Gen1
	{Slug: "1", Name: "Shelly 1", Generation: 1, Topology: SingleRelay, AltSlugs: []string{"shelly1"},
		Specs: DeviceSpecs{Voltage: "110-240V AC / 24-60V DC", MaxAmps: 16, NeutralRequired: false, Notes: "Dry contact relay"}},
	{Slug: "1pm", Name: "Shelly 1PM", Generation: 1, Topology: SingleRelay, AltSlugs: []string{"shelly1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "With power metering"}},
	{Slug: "1l", Name: "Shelly 1L", Generation: 1, Topology: SingleRelay, AltSlugs: []string{"shelly1l"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 4.1, NeutralRequired: false, Notes: "No neutral required"}},
	{Slug: "2.5", Name: "Shelly 2.5", Generation: 1, Topology: DualRelay, AltSlugs: []string{"shelly25"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 10, MaxAmpsTotal: 20, PowerMonitoring: true, NeutralRequired: true, Notes: "Dual relay with power metering"}},
	{Slug: "dimmer-2", Name: "Shelly Dimmer 2", Generation: 1, Topology: Dimmer, AltSlugs: []string{"dimmer2"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 1.1, NeutralRequired: false, Notes: "No neutral required, trailing edge"}},
	{Slug: "i3", Name: "Shelly i3", Generation: 1, Topology: InputOnly, AltSlugs: []string{"shellyi3"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", NeutralRequired: true, Notes: "3 digital inputs, no relay"}},
	{Slug: "em", Name: "Shelly EM", Generation: 1, Topology: EnergyMonitor, AltSlugs: []string{"shellyem"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", PowerMonitoring: true, NeutralRequired: true, Notes: "2-channel energy meter with CT clamps"}},
	{Slug: "3em", Name: "Shelly 3EM", Generation: 1, Topology: EnergyMonitor, AltSlugs: []string{"shelly3em"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", PowerMonitoring: true, NeutralRequired: true, Notes: "3-phase energy meter with CT clamps"}},
	{Slug: "plug-s", Name: "Shelly Plug S", Generation: 1, Topology: Plug, AltSlugs: []string{"shellyplug-s", "plugs"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 12, PowerMonitoring: true, Notes: "Wi-Fi smart plug with power metering"}},
	{Slug: "rgbw2", Name: "Shelly RGBW2", Generation: 1, Topology: RGBW, AltSlugs: []string{"shellyrgbw2"},
		Specs: DeviceSpecs{Voltage: "12/24V DC", MaxAmps: 3.75, MaxAmpsTotal: 12, NeutralRequired: true, Notes: "4-channel LED controller"}},

	// Gen2 Plus
	{Slug: "plus-1", Name: "Shelly Plus 1", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellyplus1"},
		Specs: DeviceSpecs{Voltage: "110-240V AC / 24-48V DC", MaxAmps: 16, NeutralRequired: false, Notes: "ESP32 with Bluetooth"}},
	{Slug: "plus-1pm", Name: "Shelly Plus 1PM", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellyplus1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "With power metering"}},
	{Slug: "plus-2pm", Name: "Shelly Plus 2PM", Generation: 2, Topology: DualRelay, AltSlugs: []string{"shellyplus2pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 10, MaxAmpsTotal: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "Dual relay with power metering"}},
	{Slug: "plus-i4", Name: "Shelly Plus i4", Generation: 2, Topology: InputOnly, AltSlugs: []string{"shellyplusi4"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", NeutralRequired: true, Notes: "4 digital inputs, no relay"}},
	{Slug: "plus-dimmer", Name: "Shelly Plus Wall Dimmer", Generation: 2, Topology: Dimmer, AltSlugs: []string{"shellyplusdimmer"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 1.1, NeutralRequired: false, Notes: "Trailing edge dimmer"}},
	{Slug: "plus-plug-s", Name: "Shelly Plus Plug S", Generation: 2, Topology: Plug, AltSlugs: []string{"shellyplusplug-s"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 12, PowerMonitoring: true, Notes: "Wi-Fi smart plug with Bluetooth"}},
	{Slug: "plus-rgbw-pm", Name: "Shelly Plus RGBW PM", Generation: 2, Topology: RGBW, AltSlugs: []string{"shellyplusrgbwpm"},
		Specs: DeviceSpecs{Voltage: "12/24V DC", MaxAmps: 4, MaxAmpsTotal: 10, PowerMonitoring: true, NeutralRequired: true, Notes: "4-channel LED with power metering"}},
	{Slug: "plus-1-mini", Name: "Shelly Plus 1 Mini", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellyplus1mini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, NeutralRequired: true, Notes: "Compact form factor"}},
	{Slug: "plus-1pm-mini", Name: "Shelly Plus 1PM Mini", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellyplus1pmmini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, PowerMonitoring: true, NeutralRequired: true, Notes: "Compact with power metering"}},

	// Gen2 Pro
	{Slug: "pro-1", Name: "Shelly Pro 1", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellypro1"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, NeutralRequired: true, Notes: "DIN rail mount, LAN+Wi-Fi"}},
	{Slug: "pro-1pm", Name: "Shelly Pro 1PM", Generation: 2, Topology: SingleRelay, AltSlugs: []string{"shellypro1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail with power metering"}},
	{Slug: "pro-2pm", Name: "Shelly Pro 2PM", Generation: 2, Topology: DualRelay, AltSlugs: []string{"shellypro2pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, MaxAmpsTotal: 25, PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail dual relay"}},
	{Slug: "pro-4pm", Name: "Shelly Pro 4PM", Generation: 2, Topology: QuadRelay, AltSlugs: []string{"shellypro4pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, MaxAmpsTotal: 40, PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail quad relay with power metering"}},
	{Slug: "pro-dimmer-1pm", Name: "Shelly Pro Dimmer 1PM", Generation: 2, Topology: Dimmer, AltSlugs: []string{"shellyprodimmer1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 1.22, PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail dimmer"}},
	{Slug: "pro-dimmer-2pm", Name: "Shelly Pro Dimmer 2PM", Generation: 2, Topology: Dimmer, AltSlugs: []string{"shellyprodimmer2pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 1.22, PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail dual dimmer"}},
	{Slug: "pro-3em", Name: "Shelly Pro 3EM", Generation: 2, Topology: EnergyMonitor, AltSlugs: []string{"shellypro3em"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", PowerMonitoring: true, NeutralRequired: true, Notes: "DIN rail 3-phase energy meter"}},

	// Gen3
	{Slug: "1-gen3", Name: "Shelly 1 Gen3", Generation: 3, Topology: SingleRelay, AltSlugs: []string{"shelly1gen3", "1"},
		Specs: DeviceSpecs{Voltage: "110-240V AC / 24-48V DC", MaxAmps: 16, NeutralRequired: false, Notes: "Gen3 with improved connectivity"}},
	{Slug: "1pm-gen3", Name: "Shelly 1PM Gen3", Generation: 3, Topology: SingleRelay, AltSlugs: []string{"shelly1pmgen3", "1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "Gen3 with power metering"}},
	{Slug: "1-mini-gen3", Name: "Shelly 1 Mini Gen3", Generation: 3, Topology: SingleRelay, AltSlugs: []string{"shelly1minigen3", "1-mini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, NeutralRequired: true, Notes: "Gen3 compact form factor"}},
	{Slug: "1pm-mini-gen3", Name: "Shelly 1PM Mini Gen3", Generation: 3, Topology: SingleRelay, AltSlugs: []string{"shelly1pmminigen3", "1pm-mini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, PowerMonitoring: true, NeutralRequired: true, Notes: "Gen3 compact with power metering"}},

	// Gen4
	{Slug: "1-gen4", Name: "Shelly 1 Gen4", Generation: 4, Topology: SingleRelay, AltSlugs: []string{"shelly1gen4", "1"},
		Specs: DeviceSpecs{Voltage: "110-240V AC / 24-48V DC", MaxAmps: 16, NeutralRequired: false, Notes: "Gen4 latest generation"}},
	{Slug: "1pm-gen4", Name: "Shelly 1PM Gen4", Generation: 4, Topology: SingleRelay, AltSlugs: []string{"shelly1pmgen4", "1pm"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "Gen4 with power metering"}},
	{Slug: "2pm-gen4", Name: "Shelly 2PM Gen4", Generation: 4, Topology: DualRelay, AltSlugs: []string{"shelly2pmgen4"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 10, MaxAmpsTotal: 16, PowerMonitoring: true, NeutralRequired: true, Notes: "Gen4 dual relay with power metering"}},
	{Slug: "1-mini-gen4", Name: "Shelly 1 Mini Gen4", Generation: 4, Topology: SingleRelay, AltSlugs: []string{"shelly1minigen4", "1-mini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, NeutralRequired: true, Notes: "Gen4 compact form factor"}},
	{Slug: "1pm-mini-gen4", Name: "Shelly 1PM Mini Gen4", Generation: 4, Topology: SingleRelay, AltSlugs: []string{"shelly1pmminigen4", "1pm-mini"},
		Specs: DeviceSpecs{Voltage: "110-240V AC", MaxAmps: 8, PowerMonitoring: true, NeutralRequired: true, Notes: "Gen4 compact with power metering"}},
}

// LookupModel finds a device model by slug, optionally filtered by generation.
// If generation is 0, it matches all generations. When the slug matches multiple
// generations and no generation filter is provided, the latest generation is
// returned (highest generation number).
func LookupModel(slug string, generation int) (DeviceModel, error) {
	var matches []DeviceModel
	for _, m := range catalog {
		if !matchesSlug(m, slug) {
			continue
		}
		if generation > 0 && m.Generation != generation {
			continue
		}
		matches = append(matches, m)
	}

	switch len(matches) {
	case 0:
		if generation > 0 {
			return DeviceModel{}, fmt.Errorf("model %q not found for generation %d", slug, generation)
		}
		return DeviceModel{}, fmt.Errorf("model %q not found; use 'shelly diagram -m <model>' with a valid model slug", slug)
	case 1:
		return matches[0], nil
	default:
		return latestGeneration(matches), nil
	}
}

// latestGeneration returns the model with the highest generation number.
func latestGeneration(matches []DeviceModel) DeviceModel {
	best := matches[0]
	for _, m := range matches[1:] {
		if m.Generation > best.Generation {
			best = m
		}
	}
	return best
}

// ModelSlugs returns all unique model slugs, optionally filtered by generation.
// If generation is 0, all slugs are returned.
func ModelSlugs(generation int) []string {
	seen := make(map[string]bool)
	var slugs []string
	for _, m := range catalog {
		if generation > 0 && m.Generation != generation {
			continue
		}
		if !seen[m.Slug] {
			seen[m.Slug] = true
			slugs = append(slugs, m.Slug)
		}
	}
	sort.Strings(slugs)
	return slugs
}

// ListModels returns all device models, optionally filtered by generation.
// If generation is 0, all models are returned.
func ListModels(generation int) []DeviceModel {
	if generation == 0 {
		result := make([]DeviceModel, len(catalog))
		copy(result, catalog)
		return result
	}
	var result []DeviceModel
	for _, m := range catalog {
		if m.Generation == generation {
			result = append(result, m)
		}
	}
	return result
}

// ModelGenerations returns all generation numbers that match the given slug.
// Used to determine if a slug spans multiple generations.
func ModelGenerations(slug string) []int {
	seen := make(map[int]bool)
	var gens []int
	for _, m := range catalog {
		if matchesSlug(m, slug) && !seen[m.Generation] {
			seen[m.Generation] = true
			gens = append(gens, m.Generation)
		}
	}
	sort.Ints(gens)
	return gens
}

// matchesSlug checks if a model matches a given slug, including alt slugs.
func matchesSlug(m DeviceModel, slug string) bool {
	if strings.EqualFold(m.Slug, slug) {
		return true
	}
	for _, alt := range m.AltSlugs {
		if strings.EqualFold(alt, slug) {
			return true
		}
	}
	return false
}
