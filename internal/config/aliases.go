// Package config manages the CLI configuration using Viper.
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// NamedAlias combines an alias name with its definition for listing/display.
type NamedAlias struct {
	Name string
	Alias
}

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
func ListAliasesSorted() []NamedAlias {
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

// ExpandAliasArgs checks if the first argument is an alias and expands it.
// Returns the expanded args and whether it's a shell alias.
func ExpandAliasArgs(args []string) (expandedArgs []string, isShell bool) {
	if len(args) == 0 {
		return args, false
	}

	// Check if first arg is an alias
	aliasObj, ok := GetAlias(args[0])
	if !ok {
		return args, false
	}

	// Expand the alias with remaining arguments
	expanded := ExpandAlias(aliasObj, args[1:])

	if aliasObj.Shell {
		return []string{expanded}, true
	}

	// Split expanded command into args
	expandedArgs = strings.Fields(expanded)
	return expandedArgs, false
}

// ExecuteShellAlias runs a shell alias command.
// Returns the exit code from the shell command.
func ExecuteShellAlias(ctx context.Context, args []string) int {
	if len(args) == 0 {
		return 0
	}

	// Execute via shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	//nolint:gosec // G204: args are from user-defined aliases in their own config
	cmd := exec.CommandContext(ctx, shell, "-c", args[0])
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error executing shell alias: %v\n", err)
		return 1
	}

	return 0
}

// =============================================================================
// Manager Alias Methods
// =============================================================================

// AddAlias adds or updates an alias.
func (m *Manager) AddAlias(name, command string, shell bool) error {
	if err := ValidateAliasName(name); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Aliases == nil {
		m.config.Aliases = make(map[string]Alias)
	}

	m.config.Aliases[name] = Alias{
		Command: command,
		Shell:   shell,
	}
	return m.saveWithoutLock()
}

// RemoveAlias removes an alias by name.
func (m *Manager) RemoveAlias(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.config.Aliases[name]; !exists {
		return fmt.Errorf("alias %q not found", name)
	}
	delete(m.config.Aliases, name)
	return m.saveWithoutLock()
}

// GetAlias returns an alias by name.
func (m *Manager) GetAlias(name string) (Alias, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	alias, ok := m.config.Aliases[name]
	return alias, ok
}

// ListAliases returns all aliases sorted by name.
func (m *Manager) ListAliases() []NamedAlias {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]NamedAlias, 0, len(m.config.Aliases))
	for name, alias := range m.config.Aliases {
		result = append(result, NamedAlias{Name: name, Alias: alias})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListAliasesMap returns all aliases as a map.
func (m *Manager) ListAliasesMap() map[string]Alias {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Alias, len(m.config.Aliases))
	for k, v := range m.config.Aliases {
		result[k] = v
	}
	return result
}

// IsAlias checks if a command name is an alias.
func (m *Manager) IsAlias(name string) bool {
	_, ok := m.GetAlias(name)
	return ok
}

// ImportAliases imports aliases from a YAML file.
// Returns the number of imported aliases, skipped aliases, and any error.
// If merge is true, existing aliases are not overwritten.
func (m *Manager) ImportAliases(filename string, merge bool) (imported, skipped int, err error) {
	data, err := afero.ReadFile(m.Fs(), filename)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read file: %w", err)
	}

	var af aliasFile
	if err := yaml.Unmarshal(data, &af); err != nil {
		return 0, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	// Validate all alias names before acquiring lock
	for name := range af.Aliases {
		if err := ValidateAliasName(name); err != nil {
			return 0, 0, fmt.Errorf("invalid alias %q: %w", name, err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Aliases == nil {
		m.config.Aliases = make(map[string]Alias)
	}

	for name, command := range af.Aliases {
		if merge {
			if _, exists := m.config.Aliases[name]; exists {
				skipped++
				continue
			}
		}

		// Detect shell aliases (prefixed with !)
		shell := false
		if command != "" && command[0] == '!' {
			shell = true
			command = command[1:]
		}

		m.config.Aliases[name] = Alias{
			Command: command,
			Shell:   shell,
		}
		imported++
	}

	if err := m.saveWithoutLock(); err != nil {
		return 0, 0, err
	}
	return imported, skipped, nil
}

// ExportAliases exports all aliases to a YAML file.
// If filename is empty, returns the YAML data as a string.
func (m *Manager) ExportAliases(filename string) (string, error) {
	aliases := m.ListAliases()

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
		return "", fmt.Errorf("failed to marshal aliases: %w", err)
	}

	if filename == "" {
		return string(data), nil
	}

	if err := afero.WriteFile(m.Fs(), filename, data, 0o600); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return "", nil
}
