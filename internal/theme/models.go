package theme

// Export represents an exported theme configuration.
type Export struct {
	Name           string            `yaml:"name" json:"name"`
	ColorOverrides map[string]string `yaml:"color_overrides,omitempty" json:"color_overrides,omitempty"`
	RenderedColors RenderedColors    `yaml:"rendered_colors" json:"rendered_colors"`
}

// Import represents an imported theme configuration.
type Import struct {
	Name   string            `yaml:"name" json:"name,omitempty"`
	Colors map[string]string `yaml:"colors" json:"colors,omitempty"`
}

// RenderedColors represents the actual color values being used (base + overrides).
type RenderedColors struct {
	Foreground  string `yaml:"foreground" json:"foreground"`
	Background  string `yaml:"background" json:"background"`
	Green       string `yaml:"green" json:"green"`
	Red         string `yaml:"red" json:"red"`
	Yellow      string `yaml:"yellow" json:"yellow"`
	Blue        string `yaml:"blue" json:"blue"`
	Cyan        string `yaml:"cyan" json:"cyan"`
	Purple      string `yaml:"purple" json:"purple"`
	BrightBlack string `yaml:"bright_black" json:"bright_black"`
}
