package migration

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}

	w := New(deps)

	if w.ctx != ctx {
		t.Error("ctx not set")
	}
	if w.svc != svc {
		t.Error("svc not set")
	}
	if w.step != StepSourceSelect {
		t.Errorf("step = %v, want StepSourceSelect", w.step)
	}
}

func TestNew_PanicOnNilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil ctx")
		}
	}()

	deps := Deps{Ctx: nil, Svc: &shelly.Service{}}
	New(deps)
}

func TestNew_PanicOnNilSvc(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil svc")
		}
	}()

	deps := Deps{Ctx: context.Background(), Svc: nil}
	New(deps)
}

func TestDeps_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		deps    Deps
		wantErr bool
	}{
		{
			name:    "valid",
			deps:    Deps{Ctx: context.Background(), Svc: &shelly.Service{}},
			wantErr: false,
		},
		{
			name:    "nil ctx",
			deps:    Deps{Ctx: nil, Svc: &shelly.Service{}},
			wantErr: true,
		},
		{
			name:    "nil svc",
			deps:    Deps{Ctx: context.Background(), Svc: nil},
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

func TestWizard_SetSize(t *testing.T) {
	t.Parallel()
	w := newTestWizard()

	w = w.SetSize(100, 50)

	if w.Width != 100 {
		t.Errorf("Width = %d, want 100", w.Width)
	}
	if w.Height != 50 {
		t.Errorf("Height = %d, want 50", w.Height)
	}
}

func TestWizard_SetFocused(t *testing.T) {
	t.Parallel()
	w := newTestWizard()

	w = w.SetFocused(true)

	if !w.focused {
		t.Error("focused should be true")
	}
}

func TestWizard_Update_DevicesLoaded(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.loading = true

	msg := devicesLoadedMsg{
		Devices: []DeviceInfo{
			{Name: "device1", Model: "SHSW-25"},
			{Name: "device2", Model: "SHSW-25"},
		},
	}

	updated, _ := w.Update(msg)

	if updated.loading {
		t.Error("loading should be false")
	}
	if len(updated.devices) != 2 {
		t.Errorf("devices len = %d, want 2", len(updated.devices))
	}
}

func TestWizard_Update_DevicesLoaded_Error(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.loading = true

	msg := devicesLoadedMsg{
		Err: errors.New("load failed"),
	}

	updated, _ := w.Update(msg)

	if updated.loading {
		t.Error("loading should be false")
	}
	if updated.err == nil {
		t.Error("err should be set")
	}
}

func TestWizard_Update_CaptureComplete(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.loading = true
	w.step = StepSourceSelect
	w.devices = []DeviceInfo{{Name: "device1"}}

	msg := CaptureCompleteMsg{
		SourceConfig: map[string]any{"key": "value"},
		SourceModel:  "SHSW-25",
	}

	updated, _ := w.Update(msg)

	if updated.loading {
		t.Error("loading should be false")
	}
	if updated.step != StepTargetSelect {
		t.Errorf("step = %v, want StepTargetSelect", updated.step)
	}
	if updated.sourceConfig == nil {
		t.Error("sourceConfig should be set")
	}
}

func TestWizard_Update_CompareComplete(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.loading = true
	w.step = StepTargetSelect

	msg := CompareCompleteMsg{
		Diffs: []model.ConfigDiff{
			{Path: "name", OldValue: "old", NewValue: "new", DiffType: model.DiffChanged},
		},
	}

	updated, _ := w.Update(msg)

	if updated.loading {
		t.Error("loading should be false")
	}
	if updated.step != StepPreview {
		t.Errorf("step = %v, want StepPreview", updated.step)
	}
	if len(updated.diffs) != 1 {
		t.Errorf("diffs len = %d, want 1", len(updated.diffs))
	}
}

func TestWizard_Update_ApplyComplete(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.applying = true
	w.step = StepPreview

	msg := ApplyCompleteMsg{Success: true}

	updated, _ := w.Update(msg)

	if updated.applying {
		t.Error("applying should be false")
	}
	if updated.step != StepComplete {
		t.Errorf("step = %v, want StepComplete", updated.step)
	}
}

func TestWizard_HandleAction_Navigation(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.focused = true
	w.devices = []DeviceInfo{
		{Name: "device1"},
		{Name: "device2"},
		{Name: "device3"},
	}
	w.Scroller.SetItemCount(len(w.devices))

	// Move down
	updated, _ := w.Update(messages.NavigationMsg{Direction: messages.NavDown})
	if updated.Scroller.Cursor() != 1 {
		t.Errorf("cursor after NavDown = %d, want 1", updated.Scroller.Cursor())
	}

	// Move up
	updated, _ = updated.Update(messages.NavigationMsg{Direction: messages.NavUp})
	if updated.Scroller.Cursor() != 0 {
		t.Errorf("cursor after NavUp = %d, want 0", updated.Scroller.Cursor())
	}
}

func TestWizard_HandleKey_ToggleWiFi(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.focused = true
	w.step = StepSourceSelect

	if w.includeWiFi {
		t.Error("includeWiFi should be false initially")
	}

	updated, _ := w.Update(tea.KeyPressMsg{Code: 'w'})

	if !updated.includeWiFi {
		t.Error("includeWiFi should be true after 'w'")
	}
}

func TestWizard_HandleKey_Escape_FromTargetSelect(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.focused = true
	w.step = StepTargetSelect
	w.sourceConfig = map[string]any{"key": "value"}

	updated, _ := w.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.step != StepSourceSelect {
		t.Errorf("step = %v, want StepSourceSelect", updated.step)
	}
	if updated.sourceConfig != nil {
		t.Error("sourceConfig should be nil after escape")
	}
}

func TestWizard_HandleKey_Escape_FromPreview(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.focused = true
	w.step = StepPreview
	w.diffs = []model.ConfigDiff{{Path: "test"}}

	updated, _ := w.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if updated.step != StepTargetSelect {
		t.Errorf("step = %v, want StepTargetSelect", updated.step)
	}
	if updated.diffs != nil {
		t.Error("diffs should be nil after escape")
	}
}

func TestWizard_HandleAction_NotFocused(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.focused = false
	w.devices = []DeviceInfo{{Name: "device1"}}
	w.Scroller.SetItemCount(1)

	updated, _ := w.Update(messages.NavigationMsg{Direction: messages.NavDown})

	if updated.Scroller.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestWizard_View_SourceSelect(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.step = StepSourceSelect
	w.devices = []DeviceInfo{{Name: "device1", Model: "SHSW-25"}}
	w = w.SetSize(80, 24)

	view := w.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestWizard_View_Loading(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.loading = true
	w = w.SetSize(80, 24)

	view := w.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestWizard_View_Complete(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.step = StepComplete
	w.devices = []DeviceInfo{
		{Name: "source", Model: "SHSW-25"},
		{Name: "target", Model: "SHSW-25"},
	}
	w.sourceIdx = 0
	w.targetIdx = 1
	w = w.SetSize(80, 24)

	view := w.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestWizard_Accessors(t *testing.T) {
	t.Parallel()
	w := newTestWizard()
	w.step = StepPreview
	w.loading = true
	w.applying = true
	w.err = errors.New("test error")

	if w.Step() != StepPreview {
		t.Errorf("Step() = %v, want StepPreview", w.Step())
	}
	if !w.Loading() {
		t.Error("Loading() should be true")
	}
	if !w.Applying() {
		t.Error("Applying() should be true")
	}
	if w.Error() == nil {
		t.Error("Error() should not be nil")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles render without panic
	_ = styles.Title.Render("test")
	_ = styles.Step.Render("test")
	_ = styles.Selected.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Success.Render("test")
	_ = styles.Warning.Render("test")
	_ = styles.DiffAdd.Render("test")
	_ = styles.DiffRemove.Render("test")
	_ = styles.DiffChange.Render("test")
	_ = styles.DeviceName.Render("test")
	_ = styles.DeviceModel.Render("test")
}

func TestWizard_FooterText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		step WizardStep
		want string
	}{
		{StepSourceSelect, "j/k:nav enter:select w:toggle-wifi"},
		{StepTargetSelect, "j/k:nav enter:select esc:back"},
		{StepPreview, "j/k:scroll spc:toggle a:all n:none enter:apply esc:back"},
		{StepComplete, "R:new migration"},
	}

	for _, tt := range tests {
		w := newTestWizard()
		w.step = tt.step

		got := w.FooterText()
		if got != tt.want {
			t.Errorf("FooterText() for step %v = %q, want %q", tt.step, got, tt.want)
		}
	}
}

func TestWizard_SelectiveMigration(t *testing.T) {
	t.Parallel()

	w := newTestWizard()
	w.step = StepPreview
	w.focused = true
	w.diffs = []model.ConfigDiff{
		{Path: "sys.device.name", DiffType: model.DiffChanged, OldValue: "old", NewValue: "new"},
		{Path: "switch:0.name", DiffType: model.DiffChanged, OldValue: "s1", NewValue: "s2"},
		{Path: "wifi.sta.ssid", DiffType: model.DiffChanged, OldValue: "net1", NewValue: "net2"},
	}
	w.selectedDiffs = []bool{true, true, true}
	w.diffScroller.SetItemCount(3)

	// Test toggle - should unselect first diff
	updated := w.withDiffToggled()
	if updated.selectedDiffs[0] {
		t.Error("first diff should be unselected after toggle")
	}

	// Test select none
	updated = w.withNoDiffsSelected()
	for i, sel := range updated.selectedDiffs {
		if sel {
			t.Errorf("diff %d should be unselected", i)
		}
	}

	// Test select all
	updated = w.withAllDiffsSelected()
	for i, sel := range updated.selectedDiffs {
		if !sel {
			t.Errorf("diff %d should be selected", i)
		}
	}
}

func TestWizard_SelectedDiffCount(t *testing.T) {
	t.Parallel()

	w := newTestWizard()
	w.selectedDiffs = []bool{true, false, true, false, true}

	if got := w.selectedDiffCount(); got != 3 {
		t.Errorf("selectedDiffCount() = %d, want 3", got)
	}
}

func TestAreModelsCompatible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		source string
		target string
		want   bool
	}{
		{"SHSW-25", "SHSW-25", true},          // Exact match
		{"Plus 1PM", "Plus 1PM", true},        // Exact match Gen2
		{"SHSW-25", "SHSW-1", true},           // Same Gen1 family (SHSW)
		{"Plus 1PM", "Plus 2PM", true},        // Same Gen2 family (Plus)
		{"Plus 1PM", "Pro 1PM", false},        // Different product lines
		{"Shelly Plus 1PM", "Plus 1PM", true}, // Same family after prefix strip
		{"Plus 1PM", "i4", false},             // Completely different families
	}

	for _, tt := range tests {
		got := areModelsCompatible(tt.source, tt.target)
		if got != tt.want {
			t.Errorf("areModelsCompatible(%q, %q) = %v, want %v", tt.source, tt.target, got, tt.want)
		}
	}
}

func TestExtractModelFamily(t *testing.T) {
	t.Parallel()

	tests := []struct {
		model string
		want  string
	}{
		{"SHSW-25", "SHSW"},
		{"Plus 1PM", "Plus"},
		{"Pro 2PM", "Pro"},
		{"Shelly Plus 1PM", "Plus"},
		{"1PM", "1PM"}, // No separator
	}

	for _, tt := range tests {
		got := extractModelFamily(tt.model)
		if got != tt.want {
			t.Errorf("extractModelFamily(%q) = %q, want %q", tt.model, got, tt.want)
		}
	}
}

func TestExtractTopLevelKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want string
	}{
		{"sys.device.name", "sys"},
		{"switch:0.name", "switch:0"},
		{"wifi", "wifi"},
	}

	for _, tt := range tests {
		got := extractTopLevelKey(tt.path)
		if got != tt.want {
			t.Errorf("extractTopLevelKey(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestWizard_FilterSelectedConfig(t *testing.T) {
	t.Parallel()

	w := newTestWizard()
	w.sourceConfig = map[string]any{
		"sys":      map[string]any{"device": map[string]any{"name": "test"}},
		"switch:0": map[string]any{"name": "switch1"},
		"wifi":     map[string]any{"sta": map[string]any{"ssid": "network"}},
	}
	w.diffs = []model.ConfigDiff{
		{Path: "sys.device.name", DiffType: model.DiffChanged},
		{Path: "switch:0.name", DiffType: model.DiffChanged},
		{Path: "wifi.sta.ssid", DiffType: model.DiffChanged},
	}
	// Only select sys and wifi, not switch:0
	w.selectedDiffs = []bool{true, false, true}

	filtered := w.filterSelectedConfig()

	if _, ok := filtered["sys"]; !ok {
		t.Error("filtered config should include sys")
	}
	if _, ok := filtered["wifi"]; !ok {
		t.Error("filtered config should include wifi")
	}
	if _, ok := filtered["switch:0"]; ok {
		t.Error("filtered config should NOT include switch:0")
	}
}

func newTestWizard() Wizard {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
