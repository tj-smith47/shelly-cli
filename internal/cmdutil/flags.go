// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"time"

	"github.com/spf13/cobra"
)

// Default values for common flags.
const (
	DefaultTimeout     = 10 * time.Second
	DefaultConcurrency = 5
	DefaultComponentID = 0
)

// AddComponentIDFlag adds a component ID flag (--id/-i) to a command.
// The componentName is used in the help text (e.g., "Switch", "Light", "Cover").
func AddComponentIDFlag(cmd *cobra.Command, target *int, componentName string) {
	cmd.Flags().IntVarP(target, "id", "i", DefaultComponentID,
		componentName+" component ID (default 0)")
}

// AddSwitchIDFlag adds a switch component ID flag (--switch/-s) to a command.
// This is used for batch operations that target switch components.
func AddSwitchIDFlag(cmd *cobra.Command, target *int) {
	cmd.Flags().IntVarP(target, "switch", "s", DefaultComponentID, "Switch component ID")
}

// AddTimeoutFlag adds a timeout flag (--timeout/-t) to a command.
func AddTimeoutFlag(cmd *cobra.Command, target *time.Duration) {
	cmd.Flags().DurationVarP(target, "timeout", "t", DefaultTimeout, "Timeout per device")
}

// AddConcurrencyFlag adds a concurrency flag (--concurrent/-c) to a command.
func AddConcurrencyFlag(cmd *cobra.Command, target *int) {
	cmd.Flags().IntVarP(target, "concurrent", "c", DefaultConcurrency, "Max concurrent operations")
}

// AddOutputFormatFlag adds an output format flag (--output/-o) to a command.
// Supports: table, json, yaml.
func AddOutputFormatFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "output", "o", "table", "Output format: table, json, yaml")
}

// AddYesFlag adds a confirmation bypass flag (--yes/-y) to a command.
func AddYesFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVarP(target, "yes", "y", false, "Skip confirmation prompt")
}

// AddConfirmFlag adds a double-confirmation flag (--confirm) to a command.
// Used for destructive operations like factory reset.
func AddConfirmFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVar(target, "confirm", false, "Double-confirm destructive operation")
}

// AddDryRunFlag adds a dry-run flag (--dry-run) to a command.
func AddDryRunFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVar(target, "dry-run", false, "Preview actions without executing")
}

// AddGroupFlag adds a device group flag (--group/-g) to a command.
func AddGroupFlag(cmd *cobra.Command, target *string) {
	cmd.Flags().StringVarP(target, "group", "g", "", "Target device group")
}

// AddAllFlag adds an all-devices flag (--all/-a) to a command.
func AddAllFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVarP(target, "all", "a", false, "Target all registered devices")
}

// AddNameFlag adds a name override flag (--name/-n) to a command.
func AddNameFlag(cmd *cobra.Command, target *string, usage string) {
	cmd.Flags().StringVarP(target, "name", "n", "", usage)
}

// AddOverwriteFlag adds an overwrite flag (--overwrite) to a command.
func AddOverwriteFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVar(target, "overwrite", false, "Overwrite existing resource")
}

// BatchFlags holds common flags for batch operations targeting multiple devices.
type BatchFlags struct {
	GroupName  string
	All        bool
	Timeout    time.Duration
	SwitchID   int
	Concurrent int
}

// AddBatchFlags adds all standard batch operation flags to a command.
// This is a convenience function for commands that support batch operations
// targeting multiple devices via groups or the --all flag.
func AddBatchFlags(cmd *cobra.Command, flags *BatchFlags) {
	AddGroupFlag(cmd, &flags.GroupName)
	AddAllFlag(cmd, &flags.All)
	AddTimeoutFlag(cmd, &flags.Timeout)
	AddSwitchIDFlag(cmd, &flags.SwitchID)
	AddConcurrencyFlag(cmd, &flags.Concurrent)
}

// SetBatchDefaults sets default values for batch flags.
// Call this after AddBatchFlags if the flags struct was not zero-initialized.
func SetBatchDefaults(flags *BatchFlags) {
	flags.Timeout = DefaultTimeout
	flags.Concurrent = DefaultConcurrency
	flags.SwitchID = DefaultComponentID
}

// SceneFlags holds common flags for scene activation operations.
type SceneFlags struct {
	Timeout    time.Duration
	Concurrent int
	DryRun     bool
}

// AddSceneFlags adds all standard scene operation flags to a command.
func AddSceneFlags(cmd *cobra.Command, flags *SceneFlags) {
	AddTimeoutFlag(cmd, &flags.Timeout)
	AddConcurrencyFlag(cmd, &flags.Concurrent)
	AddDryRunFlag(cmd, &flags.DryRun)
}

// SetSceneDefaults sets default values for scene flags.
func SetSceneDefaults(flags *SceneFlags) {
	flags.Timeout = DefaultTimeout
	flags.Concurrent = DefaultConcurrency
}
