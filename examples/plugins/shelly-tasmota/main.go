// shelly-tasmota is a plugin for shelly-cli that provides Tasmota device support.
//
// This plugin implements the detection and status hooks for Tasmota devices,
// allowing shelly-cli to discover and monitor Tasmota devices alongside native
// Shelly devices.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/examples/plugins/shelly-tasmota/cmd"
)

// Version is set at build time.
var Version = "1.0.0-dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "shelly-tasmota",
		Short: "Tasmota device support for shelly-cli",
		Long: `shelly-tasmota is a plugin for shelly-cli that provides integration
with Tasmota-based smart home devices.

This plugin allows shelly-cli to:
  - Detect Tasmota devices during network discovery
  - Monitor device status (power state, sensors, energy)
  - Control relays and switches (control hook)
  - Check and apply firmware updates (firmware hooks)

The plugin communicates with Tasmota devices using their HTTP API.`,
		Version: Version,
	}

	// Add subcommands for plugin hooks
	rootCmd.AddCommand(cmd.NewDetectCmd())
	rootCmd.AddCommand(cmd.NewStatusCmd())
	rootCmd.AddCommand(cmd.NewControlCmd())
	rootCmd.AddCommand(cmd.NewCheckUpdatesCmd())
	rootCmd.AddCommand(cmd.NewApplyUpdateCmd())

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("shelly-tasmota version %s\n", Version)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
