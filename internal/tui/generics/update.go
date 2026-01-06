// Package generics provides generic utilities for TUI components.
package generics

import (
	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
)

// IsPanelCacheMsg returns true if the message is a panelcache message.
// These messages should typically pass through loader forwarding.
func IsPanelCacheMsg(msg tea.Msg) bool {
	switch msg.(type) {
	case panelcache.CacheHitMsg, panelcache.CacheMissMsg, panelcache.RefreshCompleteMsg:
		return true
	}
	return false
}

// UpdateLoaderResult holds the result of UpdateLoader.
type UpdateLoaderResult struct {
	Loader   loading.Model
	Cmd      tea.Cmd
	Consumed bool
}

// UpdateLoader forwards a message to a loader during async operations.
// The passthrough predicate returns true for messages that should not consume the update
// (i.e., messages that should be processed by the main Update switch).
//
// Returns:
//   - Loader: the updated loader
//   - Cmd: any command from the loader update
//   - Consumed: true if the message was consumed by the loader (caller should return early)
//
// Usage:
//
//	result := generics.UpdateLoader(m.loader, msg, func(msg tea.Msg) bool {
//	    switch msg.(type) {
//	    case LoadedMsg, ActionMsg:
//	        return true
//	    }
//	    return generics.IsPanelCacheMsg(msg)
//	})
//	m.loader = result.Loader
//	if result.Consumed {
//	    return m, result.Cmd
//	}
func UpdateLoader(loader loading.Model, msg tea.Msg, passthrough func(tea.Msg) bool) UpdateLoaderResult {
	updated, cmd := loader.Update(msg)
	if passthrough(msg) {
		return UpdateLoaderResult{Loader: updated, Cmd: nil, Consumed: false}
	}
	if cmd != nil {
		return UpdateLoaderResult{Loader: updated, Cmd: cmd, Consumed: true}
	}
	return UpdateLoaderResult{Loader: updated, Cmd: nil, Consumed: false}
}
