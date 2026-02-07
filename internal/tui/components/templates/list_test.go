package templates

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/keys"
)

func TestNewList(t *testing.T) {
	t.Parallel()
	deps := testDeps()
	m := NewList(deps)

	if m.ctx != deps.Ctx {
		t.Error("ctx not set")
	}
	if m.svc != deps.Svc {
		t.Error("svc not set")
	}
}

func TestNewList_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := testDeps()
	deps.Ctx = nil
	NewList(deps)
}

func TestNewList_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := testDeps()
	deps.Svc = nil
	NewList(deps)
}

func TestListDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    ListDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    testDeps(),
			wantErr: false,
		},
		{
			name: "nil ctx",
			deps: ListDeps{
				Ctx: nil,
				Svc: testDeps().Svc,
			},
			wantErr: true,
		},
		{
			name: "nil svc",
			deps: ListDeps{
				Ctx: context.Background(),
				Svc: nil,
			},
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

func TestListModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestList()

	m = m.SetSize(100, 50)

	if m.Width != 100 {
		t.Errorf("Width = %d, want 100", m.Width)
	}
	if m.Height != 50 {
		t.Errorf("Height = %d, want 50", m.Height)
	}
}

func TestListModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestList()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused should be true")
	}
}

func TestListModel_ScrollerNavigation(t *testing.T) {
	t.Parallel()
	m := newTestList()
	m.focused = true

	// Test down navigation (no templates, should stay at 0)
	m, _ = m.handleKey(tea.KeyPressMsg{Code: 'j'})
	if m.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", m.Cursor())
	}
}

func TestListModel_View_Empty(t *testing.T) {
	t.Parallel()
	m := newTestList()
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestListModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestList()

	if m.SelectedTemplate() != nil {
		t.Error("SelectedTemplate() should return nil with no templates")
	}

	if m.TemplateCount() != 0 {
		t.Error("TemplateCount() should return 0")
	}

	if m.Loading() {
		t.Error("Loading() should return false initially")
	}

	if m.Applying() {
		t.Error("Applying() should return false initially")
	}

	if m.Error() != nil {
		t.Error("Error() should return nil initially")
	}
}

func TestListModel_FooterText(t *testing.T) {
	t.Parallel()
	m := newTestList()

	footer := m.FooterText()
	expected := keys.FormatHints(footerHints, 0)
	if footer != expected {
		t.Errorf("FooterText() = %q, want %q", footer, expected)
	}
}

func TestDefaultListStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultListStyles()

	// Verify styles render without panic
	_ = styles.Name.Render("test")
	_ = styles.Description.Render("test")
	_ = styles.Model.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Cursor.Render("test")
}

func testDeps() ListDeps {
	return ListDeps{
		Ctx: context.Background(),
		Svc: &shelly.Service{},
	}
}

func newTestList() ListModel {
	return NewList(testDeps())
}
