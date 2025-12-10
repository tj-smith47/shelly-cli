// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ReservedCommands are built-in commands that cannot be aliased.
var ReservedCommands = map[string]bool{
	"help":       true,
	"version":    true,
	"completion": true,
	"alias":      true,
	"extension":  true,
	"config":     true,
	"device":     true,
	"discover":   true,
	"switch":     true,
	"cover":      true,
	"light":      true,
	"rgb":        true,
	"input":      true,
	"batch":      true,
	"scene":      true,
	"firmware":   true,
	"script":     true,
	"schedule":   true,
	"cloud":      true,
	"backup":     true,
	"monitor":    true,
	"dash":       true,
	"theme":      true,
	"on":         true,
	"off":        true,
	"toggle":     true,
	"status":     true,
	"reboot":     true,
	"reset":      true,
}

// AddAlias adds or updates an alias.
func (c *Config) AddAlias(name, command string, shell bool) error {
	if err := ValidateAliasName(name); err != nil {
		return err
	}

	cfgMu.Lock()
	defer cfgMu.Unlock()

	if c.Aliases == nil {
		c.Aliases = make(map[string]Alias)
	}

	c.Aliases[name] = Alias{
		Name:    name,
		Command: command,
		Shell:   shell,
	}

	return nil
}

// RemoveAlias removes an alias by name.
func (c *Config) RemoveAlias(name string) {
	cfgMu.Lock()
	defer cfgMu.Unlock()

	delete(c.Aliases, name)
}

// GetAlias returns an alias by name, or nil if not found.
func (c *Config) GetAlias(name string) *Alias {
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	if alias, ok := c.Aliases[name]; ok {
		return &alias
	}
	return nil
}

// ListAliases returns all aliases sorted by name.
func (c *Config) ListAliases() []Alias {
	cfgMu.RLock()
	defer cfgMu.RUnlock()

	result := make([]Alias, 0, len(c.Aliases))
	for _, v := range c.Aliases {
		result = append(result, v)
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// ValidateAliasName checks if an alias name is valid.
func ValidateAliasName(name string) error {
	if name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}

	if strings.ContainsAny(name, " \t\n") {
		return fmt.Errorf("alias name cannot contain whitespace")
	}

	if ReservedCommands[name] {
		return fmt.Errorf("alias name %q conflicts with built-in command", name)
	}

	return nil
}

// ExpandAlias expands an alias command with arguments.
// Supports $1, $2, etc. for argument interpolation.
func ExpandAlias(alias Alias, args []string) string {
	cmd := alias.Command

	// Replace $N with corresponding argument
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		cmd = strings.ReplaceAll(cmd, placeholder, arg)
	}

	// Replace $@ with all remaining arguments
	cmd = strings.ReplaceAll(cmd, "$@", strings.Join(args, " "))

	return cmd
}

// IsAlias checks if a command name is an alias.
func (c *Config) IsAlias(name string) bool {
	return c.GetAlias(name) != nil
}

// aliasFile represents the structure for import/export files.
type aliasFile struct {
	Aliases map[string]string `yaml:"aliases"`
}

// ImportAliases imports aliases from a YAML file.
// Returns the number of imported aliases, skipped aliases, and any error.
// If merge is true, existing aliases are not overwritten.
func (c *Config) ImportAliases(filename string, merge bool) (imported int, skipped int, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	var af aliasFile
	if err := yaml.Unmarshal(data, &af); err != nil {
		return 0, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	for name, command := range af.Aliases {
		// Check if alias exists and we're in merge mode
		if merge && c.GetAlias(name) != nil {
			skipped++
			continue
		}

		// Validate the alias name
		if err := ValidateAliasName(name); err != nil {
			return imported, skipped, fmt.Errorf("invalid alias %q: %w", name, err)
		}

		// Detect shell aliases (prefixed with !)
		shell := strings.HasPrefix(command, "!")
		if shell {
			command = strings.TrimPrefix(command, "!")
		}

		cfgMu.Lock()
		if c.Aliases == nil {
			c.Aliases = make(map[string]Alias)
		}
		c.Aliases[name] = Alias{
			Name:    name,
			Command: command,
			Shell:   shell,
		}
		cfgMu.Unlock()

		imported++
	}

	return imported, skipped, nil
}

// ExportAliases exports all aliases to a YAML file.
// If filename is empty, exports to stdout.
func (c *Config) ExportAliases(filename string) error {
	aliases := c.ListAliases()

	af := aliasFile{
		Aliases: make(map[string]string, len(aliases)),
	}

	for _, a := range aliases {
		cmd := a.Command
		if a.Shell {
			cmd = "!" + cmd
		}
		af.Aliases[a.Name] = cmd
	}

	data, err := yaml.Marshal(&af)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	if filename == "" {
		fmt.Print(string(data))
		return nil
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Package-level functions for backward compatibility

// AddAlias adds or updates an alias (package-level).
func AddAlias(name, command string, shell bool) error {
	c := Get()
	if err := c.AddAlias(name, command, shell); err != nil {
		return err
	}
	return Save()
}

// RemoveAlias removes an alias by name (package-level).
func RemoveAlias(name string) error {
	c := Get()
	if c.GetAlias(name) == nil {
		return fmt.Errorf("alias %q not found", name)
	}
	c.RemoveAlias(name)
	return Save()
}

// GetAlias returns an alias by name (package-level).
func GetAlias(name string) (Alias, bool) {
	c := Get()
	alias := c.GetAlias(name)
	if alias == nil {
		return Alias{}, false
	}
	return *alias, true
}

// ListAliases returns all aliases (package-level).
func ListAliases() map[string]Alias {
	c := Get()
	aliases := c.ListAliases()

	result := make(map[string]Alias, len(aliases))
	for _, a := range aliases {
		result[a.Name] = a
	}
	return result
}

// IsAlias checks if a command name is an alias (package-level).
func IsAlias(name string) bool {
	return Get().IsAlias(name)
}
