// Package config manages the CLI configuration using Viper.
package config

import (
	"fmt"
	"strings"
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
// Supports $1, $2, etc. for explicit argument interpolation.
// Any remaining arguments not consumed by $N placeholders are auto-appended,
// unless $@ is used (which consumes all remaining arguments explicitly).
func ExpandAlias(alias Alias, args []string) string {
	cmd := alias.Command
	hasExplicitAll := strings.Contains(cmd, "$@")

	// Track which args were consumed by $N placeholders
	consumed := make([]bool, len(args))

	// Replace $N with corresponding argument
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		if strings.Contains(cmd, placeholder) {
			cmd = strings.ReplaceAll(cmd, placeholder, arg)
			consumed[i] = true
		}
	}

	// Replace $@ with all remaining arguments (marks all as consumed)
	if hasExplicitAll {
		cmd = strings.ReplaceAll(cmd, "$@", strings.Join(args, " "))
		return cmd
	}

	// Auto-append unconsumed arguments
	var remaining []string
	for i, arg := range args {
		if !consumed[i] {
			remaining = append(remaining, arg)
		}
	}
	if len(remaining) > 0 {
		cmd = cmd + " " + strings.Join(remaining, " ")
	}

	return cmd
}

// aliasFile represents the structure for import/export files.
type aliasFile struct {
	Aliases map[string]string `yaml:"aliases"`
}

// Package-level functions delegate to the default manager.

// AddAlias adds or updates an alias.
func AddAlias(name, command string, shell bool) error {
	return getDefaultManager().AddAlias(name, command, shell)
}

// RemoveAlias removes an alias by name.
func RemoveAlias(name string) error {
	return getDefaultManager().RemoveAlias(name)
}

// GetAlias returns an alias by name.
func GetAlias(name string) (Alias, bool) {
	return getDefaultManager().GetAlias(name)
}

// ListAliases returns all aliases as a map.
func ListAliases() map[string]Alias {
	return getDefaultManager().ListAliasesMap()
}

// ListAliasesSorted returns all aliases sorted by name.
func ListAliasesSorted() []Alias {
	return getDefaultManager().ListAliases()
}

// IsAlias checks if a command name is an alias.
func IsAlias(name string) bool {
	return getDefaultManager().IsAlias(name)
}

// ImportAliases imports aliases from a YAML file.
func ImportAliases(filename string, merge bool) (imported, skipped int, err error) {
	return getDefaultManager().ImportAliases(filename, merge)
}

// ExportAliases exports all aliases to a YAML file or returns YAML string if filename is empty.
func ExportAliases(filename string) (string, error) {
	return getDefaultManager().ExportAliases(filename)
}
