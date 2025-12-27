package flags

import "github.com/spf13/cobra"

// ComponentFlags holds flags for targeting specific device components.
// Embed this in your Options struct for commands that operate on components.
// Default ID is 0 (first component).
//
// Usage:
//
//	type Options struct {
//	    flags.ComponentFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddComponentFlags(cmd, &opts.ComponentFlags, "Switch")
//	    return cmd
//	}
type ComponentFlags struct {
	ID int
}

// AddComponentFlags adds a component ID flag (--id/-i) to a command.
// The componentName is used in the help text (e.g., "Switch", "Light", "Cover").
func AddComponentFlags(cmd *cobra.Command, flags *ComponentFlags, componentName string) {
	AddComponentIDFlag(cmd, &flags.ID, componentName)
}

// QuickComponentFlags holds flags for quick commands (on/off/toggle) that
// support targeting all components when no ID is specified.
// Default ID is -1 (meaning "all components").
//
// Usage:
//
//	type Options struct {
//	    flags.QuickComponentFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func run(ctx context.Context, opts *Options) error {
//	    componentID := opts.ComponentIDPointer() // Returns nil for "all", or pointer to ID
//	    result, err := svc.QuickOn(ctx, device, componentID)
//	}
type QuickComponentFlags struct {
	ID int
}

// ComponentIDPointer returns nil if ID is -1 (meaning "all components"),
// otherwise returns a pointer to the ID.
func (f *QuickComponentFlags) ComponentIDPointer() *int {
	if f.ID < 0 {
		return nil
	}
	return &f.ID
}

// AddQuickComponentFlags adds a component ID flag with "all" as default.
// Used by quick commands (on/off/toggle) that support operating on all components.
func AddQuickComponentFlags(cmd *cobra.Command, flags *QuickComponentFlags) {
	cmd.Flags().IntVar(&flags.ID, "id", -1, "Component ID to control (omit to control all)")
}
