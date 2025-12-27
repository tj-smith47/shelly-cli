// Package wizard provides the interactive setup wizard for first-time use.
package wizard

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/branding"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Options holds options for the init wizard.
type Options struct {
	// Device flags
	Devices     []string
	DevicesJSON []string

	// Discovery flags
	Discover        bool
	DiscoverTimeout time.Duration
	DiscoverModes   string
	Network         string

	// Completion flags
	Completions string
	Aliases     bool

	// Config flags
	Theme        string
	OutputFormat string
	NoColor      bool
	APIMode      string

	// Cloud flags
	CloudEmail    string
	CloudPassword string

	// Control flags
	Force bool
}

// IsNonInteractive returns true if non-interactive mode should be used.
func (o *Options) IsNonInteractive() bool {
	return len(o.Devices) > 0 ||
		len(o.DevicesJSON) > 0 ||
		o.Theme != "" ||
		o.OutputFormat != "" ||
		o.APIMode != "" ||
		o.NoColor ||
		o.CloudEmail != "" ||
		o.CloudPassword != "" ||
		o.Completions != "" ||
		o.Aliases ||
		o.Discover ||
		o.Force
}

// WantsCloudSetup returns true if cloud setup should be performed.
func (o *Options) WantsCloudSetup() bool {
	return o.CloudEmail != "" || o.CloudPassword != ""
}

// Run runs the init wizard.
func Run(ctx context.Context, f *cmdutil.Factory, rootCmd *cobra.Command, opts *Options) error {
	ios := f.IOStreams()

	PrintWelcome(ios)

	shouldContinue, err := CheckAndConfirmConfig(ios, opts)
	if err != nil {
		return err
	}
	if !shouldContinue {
		return nil
	}

	return runSetupSteps(ctx, f, rootCmd, opts)
}

// PrintWelcome prints the welcome banner.
func PrintWelcome(ios *iostreams.IOStreams) {
	ios.Println("")
	ios.Println(branding.StyledBanner())
	ios.Println("")
	ios.Println(theme.Title().Render("Welcome to Shelly CLI!"))
	ios.Println("")
	ios.Println("This wizard will help you set up shelly for the first time.")
	ios.Println("")
}

// PrintSummary prints the setup completion summary.
func PrintSummary(ios *iostreams.IOStreams) {
	ios.Println(theme.Dim().Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	ios.Success("Setup complete!")
	ios.Println("")
	ios.Println(theme.Bold().Render("Quick start commands:"))
	ios.Println(theme.Code().Render("  shelly device list") + "          " + theme.Dim().Render("# List your devices"))
	ios.Println(theme.Code().Render("  shelly switch status <name>") + " " + theme.Dim().Render("# Check switch status"))
	ios.Println(theme.Code().Render("  shelly dash") + "                 " + theme.Dim().Render("# Open TUI dashboard"))
	ios.Println("")
	ios.Println("Run " + theme.Code().Render("shelly --help") + " for all commands.")
	ios.Println("")
}
