package generics

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/components/loading"
	"github.com/tj-smith47/shelly-cli/internal/tui/panelcache"
)

func TestIsPanelCacheMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		msg      tea.Msg
		expected bool
	}{
		{
			name:     "CacheHitMsg",
			msg:      panelcache.CacheHitMsg{},
			expected: true,
		},
		{
			name:     "CacheMissMsg",
			msg:      panelcache.CacheMissMsg{},
			expected: true,
		},
		{
			name:     "RefreshCompleteMsg",
			msg:      panelcache.RefreshCompleteMsg{},
			expected: true,
		},
		{
			name:     "KeyPressMsg",
			msg:      tea.KeyPressMsg{},
			expected: false,
		},
		{
			name:     "nil",
			msg:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsPanelCacheMsg(tt.msg)
			if got != tt.expected {
				t.Errorf("IsPanelCacheMsg() = %v, want %v", got, tt.expected)
			}
		})
	}
}

type testLoadedMsg struct{}

func TestUpdateLoader_Passthrough(t *testing.T) {
	t.Parallel()

	loader := loading.New(loading.WithMessage("test"))
	msg := panelcache.CacheHitMsg{}

	result := UpdateLoader(loader, msg, IsPanelCacheMsg)

	if result.Consumed {
		t.Error("expected Consumed to be false for passthrough message")
	}
	if result.Cmd != nil {
		t.Error("expected Cmd to be nil for passthrough message")
	}
}

func TestUpdateLoader_CustomPassthrough(t *testing.T) {
	t.Parallel()

	loader := loading.New(loading.WithMessage("test"))
	msg := testLoadedMsg{}

	result := UpdateLoader(loader, msg, func(msg tea.Msg) bool {
		if _, ok := msg.(testLoadedMsg); ok {
			return true
		}
		return IsPanelCacheMsg(msg)
	})

	if result.Consumed {
		t.Error("expected Consumed to be false for custom passthrough message")
	}
}

func TestUpdateLoader_NonPassthrough(t *testing.T) {
	t.Parallel()

	loader := loading.New(loading.WithMessage("test"))
	msg := tea.KeyPressMsg{}

	result := UpdateLoader(loader, msg, IsPanelCacheMsg)

	// For non-passthrough, non-tick messages, Consumed should be false
	// (only tick messages produce commands from loader)
	if result.Consumed {
		t.Error("expected Consumed to be false for non-passthrough, non-tick message")
	}
}
