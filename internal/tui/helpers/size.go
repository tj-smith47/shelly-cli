// Package helpers provides shared utilities for TUI components.
package helpers

import (
	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panel"
)

// LoaderBorderOffset is the standard offset for loader positioning within borders.
const LoaderBorderOffset = 4

// SizeConfig holds configuration for computing panel dimensions.
type SizeConfig struct {
	Width        int
	Height       int
	ScrollOffset int            // Header offset for scroller visible rows calculation
	Scroller     *panel.Scroller // Optional scroller to update
}

// SizeResult holds computed dimensions and updated loader.
type SizeResult struct {
	Width  int
	Height int
	Loader loading.Model
}

// ComputeSize calculates dimensions, updates scroller, and returns updated loader.
// Use this for components with a single loader and optional scroller.
// Returns SizeResult with Width, Height, and the resized Loader.
//
// Example usage:
//
//	func (m Model) SetSize(width, height int) Model {
//	    sr := helpers.ComputeSize(helpers.SizeConfig{
//	        Width: width, Height: height, ScrollOffset: 4, Scroller: m.scroller,
//	    }, m.loader)
//	    m.width, m.height, m.loader = sr.Width, sr.Height, sr.Loader
//	    return m
//	}
func ComputeSize(cfg SizeConfig, loader loading.Model) SizeResult {
	// Resize loader accounting for borders
	resizedLoader := loader.SetSize(cfg.Width-LoaderBorderOffset, cfg.Height-LoaderBorderOffset)

	// Update scroller if provided
	if cfg.Scroller != nil {
		visibleRows := cfg.Height - cfg.ScrollOffset
		if visibleRows < 1 {
			visibleRows = 1
		}
		cfg.Scroller.SetVisibleRows(visibleRows)
	}

	return SizeResult{
		Width:  cfg.Width,
		Height: cfg.Height,
		Loader: resizedLoader,
	}
}

// ComputeSizeMulti handles components with multiple loaders.
// Returns dimensions for the model and resized loaders in a slice.
//
// Example usage:
//
//	func (m Model) SetSize(width, height int) Model {
//	    loaders := helpers.ComputeSizeMulti(helpers.SizeConfig{
//	        Width: width, Height: height, ScrollOffset: 10, Scroller: m.scroller,
//	    }, m.checkLoader, m.updateLoader)
//	    m.width, m.height = width, height
//	    m.checkLoader, m.updateLoader = loaders[0], loaders[1]
//	    return m
//	}
func ComputeSizeMulti(cfg SizeConfig, loaders ...loading.Model) []loading.Model {
	result := make([]loading.Model, len(loaders))
	for i, loader := range loaders {
		result[i] = loader.SetSize(cfg.Width-LoaderBorderOffset, cfg.Height-LoaderBorderOffset)
	}

	// Update scroller if provided
	if cfg.Scroller != nil {
		visibleRows := cfg.Height - cfg.ScrollOffset
		if visibleRows < 1 {
			visibleRows = 1
		}
		cfg.Scroller.SetVisibleRows(visibleRows)
	}

	return result
}

// SetLoaderSize returns a loader with standard size accounting for panel borders.
// Standard offset is 4 (2 for borders on each side).
func SetLoaderSize(loader loading.Model, width, height int) loading.Model {
	return loader.SetSize(width-LoaderBorderOffset, height-LoaderBorderOffset)
}

// SetScrollerRows calculates and sets visible rows on a scroller.
// The headerOffset accounts for borders, title, footer, and any extra header content.
// Visible rows are clamped to a minimum of 1.
func SetScrollerRows(height, headerOffset int, scroller *panel.Scroller) {
	visibleRows := height - headerOffset
	if visibleRows < 1 {
		visibleRows = 1
	}
	scroller.SetVisibleRows(visibleRows)
}

// SetScrollerRowsReturn calculates visible rows and sets them on a scroller.
// Returns the clamped visible rows count for cases where the value is needed.
func SetScrollerRowsReturn(height, headerOffset int, scroller *panel.Scroller) int {
	visibleRows := height - headerOffset
	if visibleRows < 1 {
		visibleRows = 1
	}
	scroller.SetVisibleRows(visibleRows)
	return visibleRows
}
