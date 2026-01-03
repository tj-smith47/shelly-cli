package cachestatus

import (
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	m := New()

	if m.IsRefreshing() {
		t.Error("New model should not be refreshing")
	}

	if !m.UpdatedAt().IsZero() {
		t.Error("New model should have zero UpdatedAt")
	}
}

func TestSetUpdatedAt(t *testing.T) {
	t.Parallel()

	m := New()
	now := time.Now()

	m = m.SetUpdatedAt(now)

	if !m.UpdatedAt().Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", m.UpdatedAt(), now)
	}
}

func TestSetRefreshing(t *testing.T) {
	t.Parallel()

	m := New()

	m = m.SetRefreshing(true)
	if !m.IsRefreshing() {
		t.Error("SetRefreshing(true) should set refreshing to true")
	}

	m = m.SetRefreshing(false)
	if m.IsRefreshing() {
		t.Error("SetRefreshing(false) should set refreshing to false")
	}
}

func TestStartRefresh(t *testing.T) {
	t.Parallel()

	m := New()

	m, cmd := m.StartRefresh()

	if !m.IsRefreshing() {
		t.Error("StartRefresh should set refreshing to true")
	}
	if cmd == nil {
		t.Error("StartRefresh should return a tick command")
	}
}

func TestStopRefresh(t *testing.T) {
	t.Parallel()

	m := New()
	m = m.SetRefreshing(true)

	beforeStop := time.Now()
	m = m.StopRefresh()

	if m.IsRefreshing() {
		t.Error("StopRefresh should set refreshing to false")
	}
	if m.UpdatedAt().Before(beforeStop) {
		t.Error("StopRefresh should set UpdatedAt to current time")
	}
}

func TestViewEmpty(t *testing.T) {
	t.Parallel()

	m := New()

	view := m.View()

	if view != "" {
		t.Errorf("View() with zero UpdatedAt should be empty, got %q", view)
	}
}

func TestViewRefreshing(t *testing.T) {
	t.Parallel()

	m := New()
	m = m.SetRefreshing(true)

	view := m.View()

	if !strings.Contains(view, "Refreshing") {
		t.Errorf("View() while refreshing should contain 'Refreshing', got %q", view)
	}
}

func TestViewWithUpdatedAt(t *testing.T) {
	t.Parallel()

	m := New()
	m = m.SetUpdatedAt(time.Now().Add(-5 * time.Minute))

	view := m.View()

	if !strings.Contains(view, "Updated") {
		t.Errorf("View() should contain 'Updated', got %q", view)
	}
	if !strings.Contains(view, "minute") {
		t.Errorf("View() should contain 'minute', got %q", view)
	}
}

func TestViewCompactEmpty(t *testing.T) {
	t.Parallel()

	m := New()

	view := m.ViewCompact()

	if view != "" {
		t.Errorf("ViewCompact() with zero UpdatedAt should be empty, got %q", view)
	}
}

func TestViewCompactRefreshing(t *testing.T) {
	t.Parallel()

	m := New()
	m = m.SetRefreshing(true)

	view := m.ViewCompact()

	// Compact view shows just the spinner when refreshing.
	if view == "" {
		t.Error("ViewCompact() while refreshing should not be empty")
	}
}

func TestViewCompactWithUpdatedAt(t *testing.T) {
	t.Parallel()

	m := New()
	m = m.SetUpdatedAt(time.Now().Add(-5 * time.Minute))

	view := m.ViewCompact()

	if !strings.Contains(view, "5m") {
		t.Errorf("ViewCompact() should contain '5m', got %q", view)
	}
}

func TestFormatAge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		age      time.Duration
		contains string
	}{
		{"just now", 30 * time.Second, "just now"},
		{"1 minute", time.Minute, "1 minute ago"},
		{"5 minutes", 5 * time.Minute, "5 minutes ago"},
		{"1 hour", time.Hour, "1 hour ago"},
		{"3 hours", 3 * time.Hour, "3 hours ago"},
		{"1 day", 25 * time.Hour, "1 day ago"},
		{"2 days", 50 * time.Hour, "2 days ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatAge(time.Now().Add(-tt.age))
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatAge(%v ago) = %q, want to contain %q", tt.age, result, tt.contains)
			}
		})
	}
}

func TestFormatAgeZero(t *testing.T) {
	t.Parallel()

	result := formatAge(time.Time{})
	if result != "never" {
		t.Errorf("formatAge(zero) = %q, want 'never'", result)
	}
}

func TestFormatAgeCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		age  time.Duration
		want string
	}{
		{"just now", 30 * time.Second, "now"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"1 day", 25 * time.Hour, "1d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := formatAgeCompact(time.Now().Add(-tt.age))
			if result != tt.want {
				t.Errorf("formatAgeCompact(%v ago) = %q, want %q", tt.age, result, tt.want)
			}
		})
	}
}

func TestFormatAgeCompactZero(t *testing.T) {
	t.Parallel()

	result := formatAgeCompact(time.Time{})
	if result != "" {
		t.Errorf("formatAgeCompact(zero) = %q, want ''", result)
	}
}

func TestWithStyles(t *testing.T) {
	t.Parallel()

	customStyles := Styles{
		Muted:   DefaultStyles().Muted,
		Spinner: DefaultStyles().Spinner,
	}

	m := New(WithStyles(customStyles))

	// Just verify it doesn't panic - styles are applied.
	_ = m.View()
}

func TestInit(t *testing.T) {
	t.Parallel()

	// Not refreshing - should return nil.
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() when not refreshing should return nil")
	}

	// Refreshing - should return tick command.
	m = m.SetRefreshing(true)
	cmd = m.Init()
	if cmd == nil {
		t.Error("Init() when refreshing should return tick command")
	}
}
