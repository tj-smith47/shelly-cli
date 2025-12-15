// Package feedback provides the feedback command for reporting issues.
package feedback

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/browser"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

const (
	repoURL         = "https://github.com/tj-smith47/shelly-cli"
	issueTypeDevice = "device"
	issueTypeBug    = "bug"
)

// Options holds the command options.
type Options struct {
	Type       string
	Title      string
	Device     string
	AttachLog  bool
	DryRun     bool
	OpenIssues bool
}

// NewCommand creates the feedback command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	opts := &Options{}

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
			return run(cmd.Context(), f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Type, "type", "t", "", "Issue type: bug, feature, or device")
	cmd.Flags().StringVar(&opts.Title, "title", "", "Issue title")
	cmd.Flags().StringVar(&opts.Device, "device", "", "Device name/IP for device compatibility issues")
	cmd.Flags().BoolVar(&opts.AttachLog, "attach-log", false, "Include CLI log info in report")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview issue without opening browser")
	cmd.Flags().BoolVar(&opts.OpenIssues, "issues", false, "Open GitHub issues page instead")

	return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, opts *Options) error {
	ios := f.IOStreams()
	br := browser.New()

	// If --issues flag, just open issues page
	if opts.OpenIssues {
		issuesURL := repoURL + "/issues"
		ios.Info("Opening GitHub issues...")
		return br.Browse(ctx, issuesURL)
	}

	// Determine issue type
	issueType := opts.Type
	if issueType == "" {
		// Default based on context
		if opts.Device != "" {
			issueType = issueTypeDevice
		} else {
			issueType = issueTypeBug
		}
	}

	// Build issue body
	body := buildIssueBody(f, opts, issueType)

	// Build URL
	issueURL := buildIssueURL(issueType, opts.Title, body)

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
		return fmt.Errorf("failed to open browser: %w", err)
	}

	ios.Success("Issue form opened in browser")
	ios.Info("Please fill in the description and submit the issue")

	return nil
}

func buildIssueBody(f *cmdutil.Factory, opts *Options, issueType string) string {
	var sb strings.Builder

	switch issueType {
	case "bug":
		sb.WriteString("## Bug Description\n\n")
		sb.WriteString("<!-- Describe the bug clearly and concisely -->\n\n")
		sb.WriteString("## Steps to Reproduce\n\n")
		sb.WriteString("1. \n2. \n3. \n\n")
		sb.WriteString("## Expected Behavior\n\n")
		sb.WriteString("<!-- What did you expect to happen? -->\n\n")
		sb.WriteString("## Actual Behavior\n\n")
		sb.WriteString("<!-- What actually happened? -->\n\n")

	case "feature":
		sb.WriteString("## Feature Request\n\n")
		sb.WriteString("<!-- Describe the feature you'd like -->\n\n")
		sb.WriteString("## Use Case\n\n")
		sb.WriteString("<!-- Why would this be useful? -->\n\n")
		sb.WriteString("## Proposed Solution\n\n")
		sb.WriteString("<!-- Optional: How do you think this could work? -->\n\n")

	case "device":
		sb.WriteString("## Device Compatibility Issue\n\n")
		if opts.Device != "" {
			sb.WriteString(fmt.Sprintf("**Device:** %s\n\n", opts.Device))
		}
		sb.WriteString("**Device Model:** <!-- e.g., Shelly Plus 1PM -->\n")
		sb.WriteString("**Device Firmware:** <!-- e.g., 1.0.8 -->\n\n")
		sb.WriteString("## Issue Description\n\n")
		sb.WriteString("<!-- What's not working? -->\n\n")
		sb.WriteString("## Command Used\n\n")
		sb.WriteString("```bash\n# The command you ran\nshelly \n```\n\n")
		sb.WriteString("## Error Message\n\n")
		sb.WriteString("```\n<!-- Paste any error output here -->\n```\n\n")
	}

	// System info section
	sb.WriteString("## Environment\n\n")
	sb.WriteString(formatSystemInfo(f, opts))

	return sb.String()
}

func formatSystemInfo(f *cmdutil.Factory, opts *Options) string {
	info := version.Get()

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("CLI Version: %s\n", info.Version))
	sb.WriteString(fmt.Sprintf("Commit: %s\n", info.Commit))
	sb.WriteString(fmt.Sprintf("OS: %s\n", runtime.GOOS))
	sb.WriteString(fmt.Sprintf("Arch: %s\n", runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("Go: %s\n", runtime.Version()))

	// Add config info if available
	if cfg, err := f.Config(); err == nil && cfg != nil {
		if cfg.Theme != "" {
			sb.WriteString(fmt.Sprintf("Theme: %s\n", cfg.Theme))
		}
		deviceCount := len(cfg.Devices)
		sb.WriteString(fmt.Sprintf("Configured Devices: %d\n", deviceCount))
	}

	if opts.AttachLog {
		sb.WriteString("\n[Log info requested - please attach relevant log output]\n")
	}

	sb.WriteString("```\n")

	return sb.String()
}

func buildIssueURL(issueType, title, body string) string {
	baseURL := repoURL + "/issues/new"

	params := url.Values{}

	// Set labels based on type
	switch issueType {
	case "bug":
		params.Set("labels", "bug")
		if title == "" {
			params.Set("title", "[Bug] ")
		}
	case "feature":
		params.Set("labels", "enhancement")
		if title == "" {
			params.Set("title", "[Feature] ")
		}
	case "device":
		params.Set("labels", "device-compatibility")
		if title == "" {
			params.Set("title", "[Device] ")
		}
	}

	if title != "" {
		params.Set("title", title)
	}

	params.Set("body", body)

	return baseURL + "?" + params.Encode()
}
