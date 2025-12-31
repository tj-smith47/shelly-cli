// Package feedback provides the feedback command for reporting issues.
package feedback

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/github"
)

// Options holds the command options.
type Options struct {
	Factory    *cmdutil.Factory
	AttachLog  bool
	Device     string
	DryRun     bool
	OpenIssues bool
	Title      string
	Type       string
}

// NewCommand creates the feedback command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{Factory: f}

	cmd := &cobra.Command{
		Use:     "feedback",
		Aliases: []string{"issue", "report-bug", "bug"},
		Short:   "Report issues or request features",
		Long: `Report issues, bugs, or request features via GitHub.

This command helps create well-formatted GitHub issues with automatic
system information collection.

Issue types:
  bug     - Report a bug or unexpected behavior
  feature - Request a new feature
  device  - Report a device compatibility issue

The command opens your browser to create a GitHub issue with
pre-populated system information.`,
		Example: `  # Report a bug
  shelly feedback --type bug

  # Request a feature
  shelly feedback --type feature --title "Add XYZ support"

  # Report device compatibility issue
  shelly feedback --type device --device kitchen-light

  # Preview the issue without opening browser
  shelly feedback --type bug --dry-run

  # View existing issues
  shelly feedback --issues`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "", "Issue type: bug, feature, or device")
	cmd.Flags().StringVar(&opts.Title, "title", "", "Issue title")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Device name/IP for device compatibility issues")
	cmd.Flags().BoolVar(&opts.AttachLog, "attach-log", false, "Include CLI log info in report")
	flags.AddDryRunFlag(cmd, &opts.DryRun)
	cmd.Flags().BoolVar(&opts.OpenIssues, "issues", false, "Open GitHub issues page instead")

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	ios := opts.Factory.IOStreams()
	br := opts.Factory.Browser()

	// If --issues flag, just open issues page
	if opts.OpenIssues {
		ios.Info("Opening GitHub issues...")
		if err := br.Browse(ctx, github.IssuesURL()); err != nil {
			var clipErr *browser.ClipboardFallbackError
			if errors.As(err, &clipErr) {
				ios.Warning("Could not open browser. URL copied to clipboard: %s", clipErr.URL)
				return nil
			}
			return fmt.Errorf("failed to open browser: %w", err)
		}
		return nil
	}

	// Determine issue type
	issueType := opts.Type
	if issueType == "" {
		// Default based on context
		if opts.Device != "" {
			issueType = github.IssueTypeDevice
		} else {
			issueType = github.IssueTypeBug
		}
	}

	// Get config for system info (optional, so ignore error)
	cfg, err := opts.Factory.Config()
	ios.DebugErr("load config for issue info", err)

	// Build issue body
	issueOpts := github.IssueOpts{
		Type:      issueType,
		Title:     opts.Title,
		Device:    opts.Device,
		AttachLog: opts.AttachLog,
	}
	body := github.BuildIssueBody(cfg, issueOpts)

	// Build URL
	issueURL := github.BuildIssueURL(issueType, opts.Title, body)

	if opts.DryRun {
		ios.Info("Issue Preview")
		ios.Println("")
		ios.Printf("Type: %s\n", issueType)
		if opts.Title != "" {
			ios.Printf("Title: %s\n", opts.Title)
		}
		ios.Println("")
		ios.Printf("Body:\n%s\n", body)
		ios.Println("")
		ios.Printf("URL: %s\n", issueURL)
		return nil
	}

	ios.Info("Opening GitHub issue form...")
	if err := br.Browse(ctx, issueURL); err != nil {
		var clipErr *browser.ClipboardFallbackError
		if errors.As(err, &clipErr) {
			ios.Warning("Could not open browser. URL copied to clipboard: %s", clipErr.URL)
			ios.Info("Please open the URL manually to submit your issue")
			return nil
		}
		return fmt.Errorf("failed to open browser: %w", err)
	}

	ios.Success("Issue form opened in browser")
	ios.Info("Please fill in the description and submit the issue")

	return nil
}
