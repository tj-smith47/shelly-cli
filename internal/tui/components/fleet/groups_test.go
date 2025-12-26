package fleet

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNewGroups(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := GroupsDeps{Ctx: ctx}

	m := NewGroups(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
}

func TestNewGroups_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := GroupsDeps{Ctx: nil}
	NewGroups(deps)
}

func TestGroupsDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    GroupsDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    GroupsDeps{Ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    GroupsDeps{Ctx: nil},
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

func TestGroupsModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestGroups()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestGroupsModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestGroups()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestGroupsModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestGroups()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused should be true")
	}
}

func TestGroupsModel_ScrollerNavigation(t *testing.T) {
	t.Parallel()
	m := newTestGroups()
	m.focused = true

	// Test down navigation (no groups, should stay at 0)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}
}

func TestGroupsModel_View_NoFleet(t *testing.T) {
	t.Parallel()
	m := newTestGroups()
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestGroupsModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestGroups()

	if m.SelectedGroup() != nil {
		t.Error("SelectedGroup() should return nil with no groups")
	}

	if len(m.Groups()) != 0 {
		t.Error("Groups() should return empty slice")
	}

	if m.GroupCount() != 0 {
		t.Error("GroupCount() should return 0")
	}

	if m.Loading() {
		t.Error("Loading() should return false initially")
	}

	if m.Error() != nil {
		t.Error("Error() should return nil initially")
	}
}

func TestDefaultGroupsStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultGroupsStyles()

	_ = styles.Name.Render("test")
	_ = styles.Count.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestGroups() GroupsModel {
	ctx := context.Background()
	deps := GroupsDeps{Ctx: ctx}
	return NewGroups(deps)
}
