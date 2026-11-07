// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/config"
)

// AddCommandsToGroup adds multiple commands to a root command and assigns them to a group.
// This is a convenience function for organizing help output.
func AddCommandsToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.GroupID = groupID
		root.AddCommand(cmd)
	}
}

// SafeConfig returns the config, logging errors to debug if load fails.
// Use this when config is optional and the command can continue without it.
// Returns nil if config loading fails.
func SafeConfig(f *Factory) *config.Config {
	cfg, err := f.Config()
	if err != nil {
		f.IOStreams().DebugErr("load config", err)
		return nil
	}
	return cfg
}

// HideAliases marks alias-group commands as hidden so documentation generators
// skip them. Alias commands are shortcuts (e.g., "add" → "device add $@") that
// should appear in help output but not in generated markdown/man docs.
func HideAliases(root *cobra.Command) {
	for _, c := range root.Commands() {
		if c.GroupID == "alias" {
			c.Hidden = true
		}
	}
}

const aliasTextMaxLength = 50

// AddAliases creates cobra commands for each alias and adds them to the parent command.
// This makes aliases appear as real commands in help output under the specified group.
func AddAliases(parent *cobra.Command, groupID string) {
	// Collect all aliases (default + user-defined)
	allAliases := make(map[string]string)
	for name, alias := range config.DefaultAliases {
		allAliases[name] = alias.Command
	}
	for name, alias := range config.ListAliases() {
		allAliases[name] = alias.Command
	}

	// Create a command for each alias
	for aliasName, aliasValue := range allAliases {
		aliasCmd := newAliasCommand(aliasName, aliasValue)
		aliasCmd.GroupID = groupID
		parent.AddCommand(aliasCmd)
	}
}

// aliasAnnotation marks a command generated from a config alias.
const aliasAnnotation = "shelly.alias"

// IsAlias reports whether c was generated from a config alias and forwards
// its arguments to another command.
func IsAlias(c *cobra.Command) bool {
	return c.Annotations[aliasAnnotation] == annotationSet
}

// newAliasCommand creates a cobra command that represents an alias.
func newAliasCommand(aliasName, aliasValue string) *cobra.Command {
	return &cobra.Command{
		Use:         aliasName,
		Short:       fmt.Sprintf("Alias for %q", truncateText(aliasTextMaxLength, aliasValue)),
		Annotations: map[string]string{aliasAnnotation: annotationSet},
		RunE: func(c *cobra.Command, args []string) error {
			expandedArgs, err := expandAliasArgs(aliasValue, args)
			if err != nil {
				return err
			}
			root := c.Root()
			root.SetArgs(expandedArgs)
			return root.Execute()
		},
		DisableFlagParsing: true,
	}
}

// expandAliasArgs rewrites args according to an alias expansion: "$N" takes
// the Nth argument, "$@" stands for every argument no "$N" consumed, and with
// neither marker the arguments follow the expansion. Arguments are
// appended after the expansion is split, so one holding spaces stays one
// argument.
func expandAliasArgs(expansion string, args []string) ([]string, error) {
	extraArgs := []string{}
	for i, a := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		if strings.Contains(expansion, placeholder) {
			expansion = strings.ReplaceAll(expansion, placeholder, a)
			continue
		}
		extraArgs = append(extraArgs, a)
	}
	expansion = strings.ReplaceAll(expansion, "$@", "")

	lingeringRE := regexp.MustCompile(`\$\d`)
	if lingeringRE.MatchString(expansion) {
		return nil, fmt.Errorf("not enough arguments for alias: %s", expansion)
	}

	newArgs, err := shlex.Split(expansion)
	if err != nil {
		return nil, err
	}

	return append(newArgs, extraArgs...), nil
}

// truncateText shortens a string to maxLen characters, adding "..." if truncated.
func truncateText(maxLen int, s string) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
