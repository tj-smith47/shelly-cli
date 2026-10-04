package shelly

import "github.com/tj-smith47/shelly-go/zwave"

// ZWaveConfigParameter is a Z-Wave configuration parameter as printed by
// `shelly zwave config`. The SDK type carries no JSON tags, so it would print
// with Go field names.
type ZWaveConfigParameter struct {
	Number       int    `json:"number" yaml:"number"`
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description" yaml:"description"`
	Size         int    `json:"size" yaml:"size"`
	DefaultValue int    `json:"default_value" yaml:"default_value"`
	MinValue     int    `json:"min_value" yaml:"min_value"`
	MaxValue     int    `json:"max_value" yaml:"max_value"`
	CurrentValue *int   `json:"current_value,omitempty" yaml:"current_value,omitempty"`
}

// ZWaveConfigParameters converts SDK configuration parameters for printing.
func ZWaveConfigParameters(params []zwave.ConfigurationParameter) []ZWaveConfigParameter {
	out := make([]ZWaveConfigParameter, 0, len(params))
	for _, p := range params {
		out = append(out, ZWaveConfigParameter{
			Number:       p.Number,
			Name:         p.Name,
			Description:  p.Description,
			Size:         p.Size,
			DefaultValue: p.DefaultValue,
			MinValue:     p.MinValue,
			MaxValue:     p.MaxValue,
			CurrentValue: p.CurrentValue,
		})
	}
	return out
}
