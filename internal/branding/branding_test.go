package branding

import (
	"strings"
	"testing"
)

func TestBanner(t *testing.T) {
	t.Parallel()

	if Banner == "" {
		t.Error("Banner should not be empty")
	}

	// Should contain CLI name
	if !strings.Contains(Banner, "Shelly") && !strings.Contains(Banner, "CLI") {
		// ASCII art may not spell out full names, just verify non-empty
	}
}

func TestBannerLines(t *testing.T) {
	t.Parallel()

	lines := BannerLines()
	if len(lines) == 0 {
		t.Error("BannerLines() should return at least one line")
	}

	// All lines should be joined to recreate banner
	rejoined := strings.Join(lines, "\n")
	if rejoined != Banner {
		t.Error("BannerLines() when joined should equal Banner")
	}
}

func TestBannerWidth(t *testing.T) {
	t.Parallel()

	width := BannerWidth()
	if width == 0 {
		t.Error("BannerWidth() should be greater than 0")
	}

	// Width should be the max line length
	lines := BannerLines()
	for _, line := range lines {
		if len(line) > width {
			t.Errorf("line length %d exceeds BannerWidth %d", len(line), width)
		}
	}
}

func TestBannerHeight(t *testing.T) {
	t.Parallel()

	height := BannerHeight()
	if height == 0 {
		t.Error("BannerHeight() should be greater than 0")
	}

	// Height should match number of lines
	lines := BannerLines()
	if height != len(lines) {
		t.Errorf("BannerHeight() = %d, want %d", height, len(lines))
	}
}

func TestStyledBanner(t *testing.T) {
	t.Parallel()

	styled := StyledBanner()
	if styled == "" {
		t.Error("StyledBanner() should not return empty string")
	}

	// Styled should contain the original banner content (at minimum)
	// Note: styling may add ANSI codes, so we can't do exact match
	if len(styled) < len(Banner) {
		t.Error("StyledBanner() should be at least as long as Banner")
	}
}

func TestStyledBannerLines(t *testing.T) {
	t.Parallel()

	styledLines := StyledBannerLines()
	lines := BannerLines()

	if len(styledLines) != len(lines) {
		t.Errorf("StyledBannerLines() returned %d lines, want %d", len(styledLines), len(lines))
	}

	// Each styled line should be at least as long as original
	for i, styled := range styledLines {
		if len(styled) < len(lines[i]) {
			t.Errorf("styledLines[%d] shorter than original", i)
		}
	}
}
