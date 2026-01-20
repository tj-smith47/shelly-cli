package fleet

import (
	"context"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestNewOperations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	deps := OperationsDeps{Ctx: ctx}

	m := NewOperations(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.operation != OpAllOn {
		t.Errorf("operation = %v, want OpAllOn", m.operation)
	}
}

func TestNewOperations_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := OperationsDeps{Ctx: nil}
	NewOperations(deps)
}

func TestOperationsDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    OperationsDeps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    OperationsDeps{Ctx: context.Background()},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    OperationsDeps{Ctx: nil},
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

func TestOperationsModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestOperations()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestOperationsModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestOperations()

	m = m.SetSize(100, 50)

	if m.width != 100 {
		t.Errorf("width = %d, want 100", m.width)
	}
	if m.height != 50 {
		t.Errorf("height = %d, want 50", m.height)
	}
}

func TestOperationsModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestOperations()

	m = m.SetFocused(true)
	if !m.focused {
		t.Error("focused should be true")
	}
}

func TestOperationsModel_HandleAction_SelectOperation(t *testing.T) {
	t.Parallel()
	m := newTestOperations()
	m.focused = true

	// Select All On (mode 1)
	m, _ = m.Update(messages.ModeSelectMsg{Mode: 1})
	if m.operation != OpAllOn {
		t.Errorf("operation = %v, want OpAllOn", m.operation)
	}

	// Select All Off (mode 2)
	m, _ = m.Update(messages.ModeSelectMsg{Mode: 2})
	if m.operation != OpAllOff {
		t.Errorf("operation = %v, want OpAllOff", m.operation)
	}
}

func TestOperationsModel_HandleAction_Navigate(t *testing.T) {
	t.Parallel()
	m := newTestOperations()
	m.focused = true
	m.operation = OpAllOn

	// Navigate right
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavRight})
	if m.operation != OpAllOff {
		t.Errorf("operation = %v, want OpAllOff", m.operation)
	}

	// Navigate left
	m, _ = m.Update(messages.NavigationMsg{Direction: messages.NavLeft})
	if m.operation != OpAllOn {
		t.Errorf("operation = %v, want OpAllOn", m.operation)
	}
}

func TestOperationsModel_View_NoFleet(t *testing.T) {
	t.Parallel()
	m := newTestOperations()
	m = m.SetSize(80, 24)

	view := m.View()
	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestOperationsModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestOperations()

	if m.Operation() != OpAllOn {
		t.Errorf("Operation() = %v, want OpAllOn", m.Operation())
	}

	if m.Executing() {
		t.Error("Executing() should return false initially")
	}

	if len(m.LastResults()) != 0 {
		t.Error("LastResults() should return empty slice initially")
	}

	if m.LastError() != nil {
		t.Error("LastError() should return nil initially")
	}
}

func TestOp_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		op   Op
		want string
	}{
		{OpAllOn, "All Relays On"},
		{OpAllOff, "All Relays Off"},
		{OpGroupOn, "Group On"},
		{OpGroupOff, "Group Off"},
		{Op(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.op.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.op, got, tt.want)
		}
	}
}

func TestDefaultOperationsStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultOperationsStyles()

	_ = styles.Button.Render("test")
	_ = styles.ButtonActive.Render("test")
	_ = styles.Success.Render("test")
	_ = styles.Failure.Render("test")
	_ = styles.Muted.Render("test")
}

func newTestOperations() OperationsModel {
	ctx := context.Background()
	deps := OperationsDeps{Ctx: ctx}
	return NewOperations(deps)
}
