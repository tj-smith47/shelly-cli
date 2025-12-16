// Package branding provides CLI branding assets like ASCII banners.
package branding

import (
	"strings"

	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// Banner is the ASCII art banner for Shelly CLI.
// This can be used in init, dash TUI loading screen, or other places.
const Banner = `   _____ _          _ _         _____ _      _____
  / ____| |        | | |       / ____| |    |_   _|
 | (___ | |__   ___| | |_   _ | |    | |      | |
  \___ \| '_ \ / _ \ | | | | || |    | |      | |
  ____) | | | |  __/ | | |_| || |____| |____ _| |_
 |_____/|_| |_|\___|_|_|\__, | \_____|______|_____|
                         __/ |
                        |___/                      `

// BannerLines returns the banner split into lines for line-by-line rendering.
func BannerLines() []string {
	return strings.Split(Banner, "\n")
}

// BannerWidth returns the width of the banner in characters.
func BannerWidth() int {
	lines := BannerLines()
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	return maxWidth
}

// BannerHeight returns the height of the banner in lines.
func BannerHeight() int {
	return len(BannerLines())
}

// StyledBanner returns the banner with theme styling applied.
func StyledBanner() string {
	return theme.Title().Render(Banner)
}

// StyledBannerLines returns styled banner lines for TUI rendering.
func StyledBannerLines() []string {
	lines := BannerLines()
	styled := make([]string, len(lines))
	style := theme.Title()
	for i, line := range lines {
		styled[i] = style.Render(line)
	}
	return styled
}
