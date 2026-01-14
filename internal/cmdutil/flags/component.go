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

// ComponentNameFlags holds flags for targeting components by name or ID.
// The Name flag takes precedence over ID if both are specified.
// Embed this in your Options struct for commands that support name-based targeting.
//
// Usage:
//
//	type Options struct {
//	    flags.ComponentNameFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func run(ctx context.Context, opts *Options) error {
//	    id, err := opts.ResolveID(ctx, opts.Factory, device, "switch")
//	    if err != nil { return err }
//	    // use id
//	}
type ComponentNameFlags struct {
	ID   int
	Name string
}

// HasName returns true if a component name was specified.
func (f *ComponentNameFlags) HasName() bool {
	return f.Name != ""
}

// AddComponentNameFlags adds both --id/-i and --name/-n flags to a command.
// The componentName is used in the help text (e.g., "Switch", "Light", "Cover").
func AddComponentNameFlags(cmd *cobra.Command, flags *ComponentNameFlags, componentName string) {
	AddComponentIDFlag(cmd, &flags.ID, componentName)
	cmd.Flags().StringVarP(&flags.Name, "name", "n", "", componentName+" name (alternative to --id)")
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

// QuickComponentNameFlags holds flags for quick commands that support
// targeting by name or ID, with "all components" as default.
// Name takes precedence over ID if both are specified.
type QuickComponentNameFlags struct {
	ID   int
	Name string
}

// HasName returns true if a component name was specified.
func (f *QuickComponentNameFlags) HasName() bool {
	return f.Name != ""
}

// ComponentIDPointer returns nil if ID is -1 (meaning "all components"),
// otherwise returns a pointer to the ID.
func (f *QuickComponentNameFlags) ComponentIDPointer() *int {
	if f.ID < 0 {
		return nil
	}
	return &f.ID
}

// AddQuickComponentNameFlags adds --id and --name flags with "all" as default.
// Used by quick commands that support name-based targeting.
func AddQuickComponentNameFlags(cmd *cobra.Command, flags *QuickComponentNameFlags) {
	cmd.Flags().IntVar(&flags.ID, "id", -1, "Component ID to control (omit to control all)")
	cmd.Flags().StringVarP(&flags.Name, "name", "n", "", "Component name (alternative to --id)")
}
