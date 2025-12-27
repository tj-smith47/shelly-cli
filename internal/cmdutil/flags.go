// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
)

// Default flag values re-exported from flags package.
const (
	// DefaultTimeout is the default timeout for device operations.
	DefaultTimeout = flags.DefaultTimeout
	// DefaultConcurrency is the default max concurrent operations.
	DefaultConcurrency = flags.DefaultConcurrency
	// DefaultComponentID is the default component ID (0).
	DefaultComponentID = flags.DefaultComponentID
)

// BatchFlags holds common flags for batch operations.
// Re-exported from flags package for compatibility.
type BatchFlags = flags.BatchFlags

// SceneFlags holds common flags for scene operations.
// Re-exported from flags package for compatibility.
type SceneFlags = flags.SceneFlags

// ComponentFlags holds flags for targeting specific components.
// Re-exported from flags package for compatibility.
type ComponentFlags = flags.ComponentFlags

// OutputFlags holds flags for controlling output format.
// Re-exported from flags package for compatibility.
type OutputFlags = flags.OutputFlags

// ConfirmFlags holds flags for confirmation handling.
// Re-exported from flags package for compatibility.
type ConfirmFlags = flags.ConfirmFlags

// AddComponentIDFlag adds a component ID flag. See flags.AddComponentIDFlag.
func AddComponentIDFlag(cmd *cobra.Command, target *int, componentName string) {
	flags.AddComponentIDFlag(cmd, target, componentName)
}

// AddSwitchIDFlag adds a switch ID flag. See flags.AddSwitchIDFlag.
func AddSwitchIDFlag(cmd *cobra.Command, target *int) {
	flags.AddSwitchIDFlag(cmd, target)
}

// AddTimeoutFlag adds a timeout flag. See flags.AddTimeoutFlag.
func AddTimeoutFlag(cmd *cobra.Command, target *time.Duration) {
	flags.AddTimeoutFlag(cmd, target)
}

// AddConcurrencyFlag adds a concurrency flag. See flags.AddConcurrencyFlag.
func AddConcurrencyFlag(cmd *cobra.Command, target *int) {
	flags.AddConcurrencyFlag(cmd, target)
}

// AddOutputFormatFlag adds an output format flag. See flags.AddOutputFormatFlag.
func AddOutputFormatFlag(cmd *cobra.Command, target *string) {
	flags.AddOutputFormatFlag(cmd, target)
}

// AddYesFlag adds a yes flag. See flags.AddYesFlag.
func AddYesFlag(cmd *cobra.Command, target *bool) {
	flags.AddYesFlag(cmd, target)
}

// AddConfirmFlag adds a confirm flag. See flags.AddConfirmFlag.
func AddConfirmFlag(cmd *cobra.Command, target *bool) {
	flags.AddConfirmFlag(cmd, target)
}

// AddDryRunFlag adds a dry-run flag. See flags.AddDryRunFlag.
func AddDryRunFlag(cmd *cobra.Command, target *bool) {
	flags.AddDryRunFlag(cmd, target)
}

// AddGroupFlag adds a group flag. See flags.AddGroupFlag.
func AddGroupFlag(cmd *cobra.Command, target *string) {
	flags.AddGroupFlag(cmd, target)
}

// AddAllFlag adds an all flag. See flags.AddAllFlag.
func AddAllFlag(cmd *cobra.Command, target *bool) {
	flags.AddAllFlag(cmd, target)
}

// AddNameFlag adds a name flag. See flags.AddNameFlag.
func AddNameFlag(cmd *cobra.Command, target *string, usage string) {
	flags.AddNameFlag(cmd, target, usage)
}

// AddOverwriteFlag adds an overwrite flag. See flags.AddOverwriteFlag.
func AddOverwriteFlag(cmd *cobra.Command, target *bool) {
	flags.AddOverwriteFlag(cmd, target)
}

// AddBatchFlags adds batch operation flags. See flags.AddBatchFlags.
func AddBatchFlags(cmd *cobra.Command, f *BatchFlags) {
	flags.AddBatchFlags(cmd, f)
}

// SetBatchDefaults sets batch flag defaults. See flags.SetBatchDefaults.
func SetBatchDefaults(f *BatchFlags) {
	flags.SetBatchDefaults(f)
}

// AddSceneFlags adds scene operation flags. See flags.AddSceneFlags.
func AddSceneFlags(cmd *cobra.Command, f *SceneFlags) {
	flags.AddSceneFlags(cmd, f)
}

// SetSceneDefaults sets scene flag defaults. See flags.SetSceneDefaults.
func SetSceneDefaults(f *SceneFlags) {
	flags.SetSceneDefaults(f)
}

// AddComponentFlags adds component flags. See flags.AddComponentFlags.
func AddComponentFlags(cmd *cobra.Command, f *ComponentFlags, componentName string) {
	flags.AddComponentFlags(cmd, f, componentName)
}

// AddOutputFlags adds output flags. See flags.AddOutputFlags.
func AddOutputFlags(cmd *cobra.Command, f *OutputFlags) {
	flags.AddOutputFlags(cmd, f)
}

// SetOutputDefaults sets output flag defaults. See flags.SetOutputDefaults.
func SetOutputDefaults(f *OutputFlags) {
	flags.SetOutputDefaults(f)
}

// AddConfirmFlags adds confirmation flags. See flags.AddConfirmFlags.
func AddConfirmFlags(cmd *cobra.Command, f *ConfirmFlags) {
	flags.AddConfirmFlags(cmd, f)
}

// AddYesOnlyFlag adds only the yes flag. See flags.AddYesOnlyFlag.
func AddYesOnlyFlag(cmd *cobra.Command, f *ConfirmFlags) {
	flags.AddYesOnlyFlag(cmd, f)
}
