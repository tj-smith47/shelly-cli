package editmodal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

func TestBase_StartSave(t *testing.T) {
	t.Parallel()
	b := Base{
		Err: fmt.Errorf("old error"),
	}
	b.StartSave()

	if !b.Saving {
		t.Error("Saving = false, want true")
	}
	if b.Err != nil {
		t.Errorf("Err = %v, want nil", b.Err)
	}
}

func TestBase_SaveCmd_Success(t *testing.T) {
	t.Parallel()
	b := Base{Ctx: context.Background()}
	cmd := b.SaveCmd(func(_ context.Context) error {
		return nil
	})

	msg := cmd()
	result, ok := msg.(messages.SaveResultMsg)
	if !ok {
		t.Fatalf("expected SaveResultMsg, got %T", msg)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.Err != nil {
		t.Errorf("Err = %v, want nil", result.Err)
	}
}

func TestBase_SaveCmd_Error(t *testing.T) {
	t.Parallel()
	b := Base{Ctx: context.Background()}
	testErr := fmt.Errorf("save failed")
	cmd := b.SaveCmd(func(_ context.Context) error {
		return testErr
	})

	msg := cmd()
	result, ok := msg.(messages.SaveResultMsg)
	if !ok {
		t.Fatalf("expected SaveResultMsg, got %T", msg)
	}
	if result.Success {
		t.Error("Success = true, want false")
	}
	if !errors.Is(result.Err, testErr) {
		t.Errorf("Err = %v, want %v", result.Err, testErr)
	}
}

func TestBase_SaveCmdWithID(t *testing.T) {
	t.Parallel()
	b := Base{Ctx: context.Background()}
	cmd := b.SaveCmdWithID("myID", func(_ context.Context) error {
		return nil
	})

	msg := cmd()
	result, ok := msg.(messages.SaveResultMsg)
	if !ok {
		t.Fatalf("expected SaveResultMsg, got %T", msg)
	}
	if result.ComponentID != "myID" {
		t.Errorf("ComponentID = %v, want %q", result.ComponentID, "myID")
	}
}

func TestBase_HandleSaveResult_Success(t *testing.T) {
	t.Parallel()
	b := Base{Saving: true}
	b.Show("dev", 3)

	msg := messages.NewSaveResult(nil)
	saved, cmd := b.HandleSaveResult(msg)

	if !saved {
		t.Error("saved = false, want true")
	}
	if b.Saving {
		t.Error("Saving = true after success, want false")
	}
	if b.Visible() {
		t.Error("Visible() = true after success, want false")
	}
	if cmd == nil {
		t.Fatal("cmd = nil, want EditClosedMsg cmd")
	}
	closeMsg := cmd()
	closed, ok := closeMsg.(messages.EditClosedMsg)
	if !ok {
		t.Fatalf("expected EditClosedMsg, got %T", closeMsg)
	}
	if !closed.Saved {
		t.Error("EditClosedMsg.Saved = false, want true")
	}
}

func TestBase_HandleSaveResult_Error(t *testing.T) {
	t.Parallel()
	b := Base{Saving: true}
	b.Show("dev", 3)

	testErr := fmt.Errorf("network error")
	msg := messages.NewSaveError(nil, testErr)
	saved, cmd := b.HandleSaveResult(msg)

	if saved {
		t.Error("saved = true, want false")
	}
	if cmd != nil {
		t.Error("cmd should be nil on error")
	}
	if b.Saving {
		t.Error("Saving = true after error, want false")
	}
	if !errors.Is(b.Err, testErr) {
		t.Errorf("Err = %v, want %v", b.Err, testErr)
	}
	if !b.Visible() {
		t.Error("Visible() = false after error, want true (stay open)")
	}
}
