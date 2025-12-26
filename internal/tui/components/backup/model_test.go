package backup

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}

	m := New(deps)

	if m.ctx != ctx {
		t.Error("ctx not set")
	}
	if m.svc != svc {
		t.Error("svc not set")
	}
	if m.exporting {
		t.Error("should not be exporting initially")
	}
	if m.importing {
		t.Error("should not be importing initially")
	}
	if m.mode != ModeExport {
		t.Errorf("mode = %v, want ModeExport", m.mode)
	}
	if m.scroller == nil {
		t.Error("scroller should be initialized")
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

func TestMode_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeExport, "Export"},
		{ModeImport, "Import"},
		{Mode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModel_SetSize(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetSize(100, 50)

	if updated.width != 100 {
		t.Errorf("width = %d, want 100", updated.width)
	}
	if updated.height != 50 {
		t.Errorf("height = %d, want 50", updated.height)
	}
	// visible rows = height - 10
	if updated.scroller.VisibleRows() != 40 {
		t.Errorf("scroller.VisibleRows() = %d, want 40", updated.scroller.VisibleRows())
	}
}

func TestModel_SetFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	if m.focused {
		t.Error("should not be focused initially")
	}

	updated := m.SetFocused(true)

	if !updated.focused {
		t.Error("should be focused after SetFocused(true)")
	}
}

func TestModel_SetBackupDir(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	updated := m.SetBackupDir("/tmp/backups")

	if updated.backupDir != "/tmp/backups" {
		t.Errorf("backupDir = %q, want /tmp/backups", updated.backupDir)
	}
}

func TestModel_Update_ExportCompleteMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.exporting = true
	m.devices = []DeviceBackup{
		{Name: "device1", Exporting: true},
		{Name: "device2", Exporting: true},
	}
	results := []ExportResult{
		{Name: "device1", Success: true, FilePath: "/tmp/backup1.json"},
		{Name: "device2", Success: false, Err: errors.New("failed")},
	}
	msg := ExportCompleteMsg{Results: results}

	updated, _ := m.Update(msg)

	if updated.exporting {
		t.Error("should not be exporting after ExportCompleteMsg")
	}
	if !updated.devices[0].Exported {
		t.Error("device1 should be exported")
	}
	if updated.devices[0].FilePath != "/tmp/backup1.json" {
		t.Errorf("device1 FilePath = %q, want /tmp/backup1.json", updated.devices[0].FilePath)
	}
	if updated.devices[1].Exported {
		t.Error("device2 should not be exported")
	}
	if updated.devices[1].Err == nil {
		t.Error("device2 should have error")
	}
}

func TestModel_Update_ImportCompleteMsg(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.importing = true

	// Success case
	msg := ImportCompleteMsg{Name: "backup.json", Success: true}
	updated, _ := m.Update(msg)

	if updated.importing {
		t.Error("should not be importing after ImportCompleteMsg")
	}
	if updated.err != nil {
		t.Error("should not have error on success")
	}

	// Failure case
	m.importing = true
	testErr := errors.New("import failed")
	msg = ImportCompleteMsg{Name: "backup.json", Success: false, Err: testErr}
	updated, _ = m.Update(msg)

	if updated.err == nil {
		t.Error("should have error on failure")
	}
}

func TestModel_HandleKey_Navigation(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceBackup{
		{Name: "device0"},
		{Name: "device1"},
		{Name: "device2"},
	}
	m.scroller.SetItemCount(3)

	// Move down
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	if updated.scroller.Cursor() != 1 {
		t.Errorf("cursor after j = %d, want 1", updated.scroller.Cursor())
	}

	// Move up
	updated, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	if updated.scroller.Cursor() != 0 {
		t.Errorf("cursor after k = %d, want 0", updated.scroller.Cursor())
	}
}

func TestModel_HandleKey_Selection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: false},
		{Name: "device1", Selected: false},
	}
	m.scroller.SetItemCount(2)

	// Toggle selection with space
	updated, _ := m.Update(tea.KeyPressMsg{Code: ' '})
	if !updated.devices[0].Selected {
		t.Error("device0 should be selected after space")
	}

	// Toggle again
	updated, _ = updated.Update(tea.KeyPressMsg{Code: ' '})
	if updated.devices[0].Selected {
		t.Error("device0 should be unselected after second space")
	}
}

func TestModel_HandleKey_SelectAll(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: false},
		{Name: "device1", Selected: false},
	}

	// Select all with 'a'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'a'})
	if !updated.devices[0].Selected || !updated.devices[1].Selected {
		t.Error("all devices should be selected after 'a'")
	}
}

func TestModel_HandleKey_SelectNone(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: true},
	}

	// Deselect all with 'n'
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	if updated.devices[0].Selected || updated.devices[1].Selected {
		t.Error("all devices should be deselected after 'n'")
	}
}

func TestModel_HandleKey_ModeSwitch(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport

	// Switch to import mode
	updated, _ := m.Update(tea.KeyPressMsg{Code: '2'})
	if updated.mode != ModeImport {
		t.Errorf("mode = %v, want ModeImport", updated.mode)
	}

	// Switch back to export mode
	updated, _ = updated.Update(tea.KeyPressMsg{Code: '1'})
	if updated.mode != ModeExport {
		t.Errorf("mode = %v, want ModeExport", updated.mode)
	}
}

func TestModel_HandleKey_Export_NoSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: false},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x'})

	if cmd != nil {
		t.Error("should not return command when no devices selected")
	}
	if updated.err == nil {
		t.Error("should set error when no devices selected")
	}
}

func TestModel_HandleKey_Export_WithSelection(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: true},
	}

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x'})

	if !updated.exporting {
		t.Error("should be exporting after 'x'")
	}
	if cmd == nil {
		t.Error("should return command when devices selected")
	}
	if !updated.devices[0].Exporting {
		t.Error("selected device should have Exporting=true")
	}
}

func TestModel_HandleKey_NotFocused(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = false
	m.devices = []DeviceBackup{{Name: "device0"}}
	m.scroller.SetItemCount(1)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'j'})

	if updated.scroller.Cursor() != 0 {
		t.Error("cursor should not change when not focused")
	}
}

func TestModel_ScrollerCursorBounds(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.focused = true
	m.devices = []DeviceBackup{
		{Name: "device0"},
		{Name: "device1"},
	}
	m.scroller.SetItemCount(2)

	// Can't go below 0
	m.scroller.CursorUp()
	if m.scroller.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (can't go below)", m.scroller.Cursor())
	}

	// Can't exceed list length
	m.scroller.SetCursor(1)
	m.scroller.CursorDown()
	if m.scroller.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (can't exceed list)", m.scroller.Cursor())
	}
}

func TestModel_ScrollerVisibleRows(t *testing.T) {
	t.Parallel()
	m := newTestModel()

	m = m.SetSize(80, 20)
	if rows := m.scroller.VisibleRows(); rows != 10 {
		t.Errorf("scroller.VisibleRows() = %d, want 10", rows)
	}

	m = m.SetSize(80, 5)
	if rows := m.scroller.VisibleRows(); rows != 1 {
		t.Errorf("scroller.VisibleRows() with small height = %d, want 1", rows)
	}
}

func TestModel_SelectedDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: false},
		{Name: "device2", Selected: true},
	}

	selected := m.selectedDevices()

	if len(selected) != 2 {
		t.Errorf("selectedDevices() len = %d, want 2", len(selected))
	}
}

func TestModel_View_NoDevices(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ExportMode(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeExport
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: false, Exported: true, FilePath: "/tmp/backup.json"},
	}
	m.scroller.SetItemCount(2)
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_ImportMode(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeImport
	m = m.SetSize(80, 30)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Exporting(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceBackup{{Name: "device0", Exporting: true}}
	m.exporting = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_View_Importing(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeImport
	m.importing = true
	m = m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("View() should not return empty string")
	}
}

func TestModel_Accessors(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.devices = []DeviceBackup{
		{Name: "device0", Selected: true},
		{Name: "device1", Selected: false},
	}
	m.mode = ModeImport
	m.exporting = true
	m.importing = true
	m.err = errors.New("test error")
	m.scroller.SetItemCount(5)
	m.scroller.SetCursor(2)
	m.backupDir = "/custom/backups"

	if len(m.Devices()) != 2 {
		t.Errorf("Devices() len = %d, want 2", len(m.Devices()))
	}
	if m.Mode() != ModeImport {
		t.Errorf("Mode() = %v, want ModeImport", m.Mode())
	}
	if !m.Exporting() {
		t.Error("Exporting() should be true")
	}
	if !m.Importing() {
		t.Error("Importing() should be true")
	}
	if m.Error() == nil {
		t.Error("Error() should not be nil")
	}
	if m.Cursor() != 2 {
		t.Errorf("Cursor() = %d, want 2", m.Cursor())
	}
	if m.SelectedCount() != 1 {
		t.Errorf("SelectedCount() = %d, want 1", m.SelectedCount())
	}
	if m.BackupDir() != "/custom/backups" {
		t.Errorf("BackupDir() = %q, want /custom/backups", m.BackupDir())
	}
}

func TestModel_ExportSelected_AlreadyExporting(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.exporting = true
	m.devices = []DeviceBackup{{Name: "device0", Selected: true}}

	updated, cmd := m.ExportSelected()

	if cmd != nil {
		t.Error("should not return command when already exporting")
	}
	if !updated.exporting {
		t.Error("should still be exporting")
	}
}

func TestModel_ExportSelected_WrongMode(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeImport
	m.devices = []DeviceBackup{{Name: "device0", Selected: true}}

	updated, cmd := m.ExportSelected()

	if cmd != nil {
		t.Error("should not return command in wrong mode")
	}
	if updated.exporting {
		t.Error("should not be exporting in wrong mode")
	}
}

func TestModel_ImportSelected_WrongMode(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeExport
	m.backupFiles = []File{{Name: "backup.json"}}

	updated, cmd := m.ImportSelected()

	if cmd != nil {
		t.Error("should not return command in wrong mode")
	}
	if updated.importing {
		t.Error("should not be importing in wrong mode")
	}
}

func TestModel_ImportSelected_NoFiles(t *testing.T) {
	t.Parallel()
	m := newTestModel()
	m.mode = ModeImport
	m.backupFiles = nil

	updated, cmd := m.ImportSelected()

	if cmd != nil {
		t.Error("should not return command with no files")
	}
	if updated.err == nil {
		t.Error("should set error with no files")
	}
}

func TestDefaultStyles(t *testing.T) {
	t.Parallel()
	styles := DefaultStyles()

	// Verify styles are created without panic
	_ = styles.Selected.Render("test")
	_ = styles.Unselected.Render("test")
	_ = styles.Cursor.Render("test")
	_ = styles.Success.Render("test")
	_ = styles.Failure.Render("test")
	_ = styles.InProgress.Render("test")
	_ = styles.Label.Render("test")
	_ = styles.Error.Render("test")
	_ = styles.Muted.Render("test")
	_ = styles.Button.Render("test")
	_ = styles.ModeActive.Render("test")
}

func newTestModel() Model {
	ctx := context.Background()
	svc := &shelly.Service{}
	deps := Deps{Ctx: ctx, Svc: svc}
	return New(deps)
}
