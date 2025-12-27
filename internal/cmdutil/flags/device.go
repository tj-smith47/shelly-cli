package flags

import "github.com/spf13/cobra"

// DeviceFilterFlags holds flags for filtering device lists.
// Embed this in your Options struct for commands that list or select devices.
//
// Usage:
//
//	type Options struct {
//	    flags.DeviceFilterFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddDeviceFilterFlags(cmd, &opts.DeviceFilterFlags)
//	    return cmd
//	}
type DeviceFilterFlags struct {
	Generation int    // Filter by generation (1, 2, or 3)
	DeviceType string // Filter by device type
	Platform   string // Filter by platform (e.g., shelly, tasmota)
}

// AddDeviceFilterFlags adds device filtering flags to a command.
func AddDeviceFilterFlags(cmd *cobra.Command, flags *DeviceFilterFlags) {
	cmd.Flags().IntVarP(&flags.Generation, "generation", "g", 0, "Filter by generation (1, 2, or 3)")
	cmd.Flags().StringVarP(&flags.DeviceType, "type", "t", "", "Filter by device type")
	cmd.Flags().StringVarP(&flags.Platform, "platform", "p", "", "Filter by platform (e.g., shelly, tasmota)")
}

// HasFilters returns true if any filter flag is set.
func (f *DeviceFilterFlags) HasFilters() bool {
	return f.Generation != 0 || f.DeviceType != "" || f.Platform != ""
}

// DeviceListFlags extends DeviceFilterFlags with sorting and display options
// for device listing commands.
//
// Usage:
//
//	type Options struct {
//	    flags.DeviceListFlags
//	    Factory *cmdutil.Factory
//	}
//
//	func NewCommand(f *cmdutil.Factory) *cobra.Command {
//	    opts := &Options{Factory: f}
//	    cmd := &cobra.Command{...}
//	    flags.AddDeviceListFlags(cmd, &opts.DeviceListFlags)
//	    return cmd
//	}
type DeviceListFlags struct {
	DeviceFilterFlags
	UpdatesFirst bool // Sort devices with available updates first
	ShowVersion  bool // Show firmware version information
}

// AddDeviceListFlags adds device listing flags to a command.
func AddDeviceListFlags(cmd *cobra.Command, flags *DeviceListFlags) {
	AddDeviceFilterFlags(cmd, &flags.DeviceFilterFlags)
	cmd.Flags().BoolVarP(&flags.UpdatesFirst, "updates-first", "u", false, "Sort devices with available updates first")
	cmd.Flags().BoolVarP(&flags.ShowVersion, "version", "V", false, "Show firmware version information")
}
