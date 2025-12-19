// Package github provides GitHub integration for releases and issues.
package github

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/version"
)

const (
	// RepoURL is the GitHub repository URL for shelly-cli.
	RepoURL = "https://github.com/tj-smith47/shelly-cli"

	// IssueTypeBug is the issue type for bug reports.
	IssueTypeBug = "bug"
	// IssueTypeFeature is the issue type for feature requests.
	IssueTypeFeature = "feature"
	// IssueTypeDevice is the issue type for device compatibility issues.
	IssueTypeDevice = "device"
)

// IssueOpts holds options for building an issue.
type IssueOpts struct {
	Type      string
	Title     string
	Device    string
	AttachLog bool
}

// BuildIssueBody builds the GitHub issue body text.
func BuildIssueBody(cfg *config.Config, opts IssueOpts) string {
	var sb strings.Builder

	switch opts.Type {
	case IssueTypeBug:
		sb.WriteString("## Bug Description\n\n")
		sb.WriteString("<!-- Describe the bug clearly and concisely -->\n\n")
		sb.WriteString("## Steps to Reproduce\n\n")
		sb.WriteString("1. \n2. \n3. \n\n")
		sb.WriteString("## Expected Behavior\n\n")
		sb.WriteString("<!-- What did you expect to happen? -->\n\n")
		sb.WriteString("## Actual Behavior\n\n")
		sb.WriteString("<!-- What actually happened? -->\n\n")

	case IssueTypeFeature:
		sb.WriteString("## Feature Request\n\n")
		sb.WriteString("<!-- Describe the feature you'd like -->\n\n")
		sb.WriteString("## Use Case\n\n")
		sb.WriteString("<!-- Why would this be useful? -->\n\n")
		sb.WriteString("## Proposed Solution\n\n")
		sb.WriteString("<!-- Optional: How do you think this could work? -->\n\n")

	case IssueTypeDevice:
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
	sb.WriteString(FormatSystemInfo(cfg, opts.AttachLog))

	return sb.String()
}

// FormatSystemInfo formats system information for issue reports.
func FormatSystemInfo(cfg *config.Config, attachLog bool) string {
	info := version.Get()

	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("CLI Version: %s\n", info.Version))
	sb.WriteString(fmt.Sprintf("Commit: %s\n", info.Commit))
	sb.WriteString(fmt.Sprintf("OS: %s\n", runtime.GOOS))
	sb.WriteString(fmt.Sprintf("Arch: %s\n", runtime.GOARCH))
	sb.WriteString(fmt.Sprintf("Go: %s\n", runtime.Version()))

	// Add config info if available
	if cfg != nil {
		if cfg.Theme != "" {
			sb.WriteString(fmt.Sprintf("Theme: %s\n", cfg.Theme))
		}
		deviceCount := len(cfg.Devices)
		sb.WriteString(fmt.Sprintf("Configured Devices: %d\n", deviceCount))
	}

	if attachLog {
		sb.WriteString("\n[Log info requested - please attach relevant log output]\n")
	}

	sb.WriteString("```\n")

	return sb.String()
}

// BuildIssueURL builds the GitHub new issue URL with prefilled data.
func BuildIssueURL(issueType, title, body string) string {
	baseURL := RepoURL + "/issues/new"

	params := url.Values{}

	// Set labels based on type
	switch issueType {
	case IssueTypeBug:
		params.Set("labels", "bug")
		if title == "" {
			params.Set("title", "[Bug] ")
		}
	case IssueTypeFeature:
		params.Set("labels", "enhancement")
		if title == "" {
			params.Set("title", "[Feature] ")
		}
	case IssueTypeDevice:
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

// IssuesURL returns the URL to the GitHub issues page.
func IssuesURL() string {
	return RepoURL + "/issues"
}
