package fleet

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewHealth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := HealthDeps{Ctx: ctx}

	m := NewHealth(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
}

func TestNewHealth_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := HealthDeps{Ctx: nil}
	NewHealth(deps)
}

func TestHealthDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    HealthDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    HealthDeps{Ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    HealthDeps{Ctx: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.deps.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHealthModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestHealth()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestHealthModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestHealth()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestHealthModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestHealth()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused should be true")
	}
}

func TestHealthModel_KeyHandler(t *testing.T) {
	t.Parallel()
	m := newTestHealth()
	m.focused = true

	// Refresh key (no fleet, so should be no-op)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'r'})
	if m.loading {
		t.Error("loading should be false when no fleet")
	}
}

func TestHealthModel_View_NoFleet(t *testing.T) {
	t.Parallel()
	m := newTestHealth()
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestHealthModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestHealth()

	if m.Stats() != nil {
		t.Error("Stats() should return nil initially")
	}

	if m.Loading() {
		t.Error("Loading() should return false initially")
	}

	if m.Error() != nil {
		t.Error("Error() should return nil initially")
	}
}

func TestDefaultHealthStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultHealthStyles()

	_ = styles.Online.Render("test")
	_ = styles.Offline.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Value.Render("test")
	_ = styles.StatGood.Render("test")
	_ = styles.StatBad.Render("test")
}

func newTestHealth() HealthModel {
	ctx := context.Background()
	deps := HealthDeps{Ctx: ctx}
	return NewHealth(deps)
}
