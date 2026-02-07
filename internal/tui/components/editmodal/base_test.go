package editmodal

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestBase_Show(t *testing.T) {
	t.Parallel()
	b := Base{
		Ctx:    context.Background(),
		Styles: DefaultStyles(),
		Cursor: 5,
		Saving: true,
		Err:    fmt.Errorf("old error"),
	}

	b.Show("device1", 3)

	if b.Device != "device1" {
		t.Errorf("Device = %q, want %q", b.Device, "device1")
	}
	if b.FieldCount != 3 {
		t.Errorf("FieldCount = %d, want 3", b.FieldCount)
	}
	if b.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", b.Cursor)
	}
	if b.Saving {
		t.Error("Saving = true, want false")
	}
	if b.Err != nil {
		t.Errorf("Err = %v, want nil", b.Err)
	}
	if !b.Visible() {
		t.Error("Visible() = false, want true")
	}
}

func TestBase_Hide(t *testing.T) {
	t.Parallel()
	b := Base{}
	b.Show("dev", 2)
	b.Hide()

	if b.Visible() {
		t.Error("Visible() = true after Hide, want false")
	}
}

func TestBase_SetSize(t *testing.T) {
	t.Parallel()
	b := Base{}
	b.SetSize(80, 40)

	if b.Width != 80 {
		t.Errorf("Width = %d, want 80", b.Width)
	}
	if b.Height != 40 {
		t.Errorf("Height = %d, want 40", b.Height)
	}
}

func TestBase_SetErr(t *testing.T) {
	t.Parallel()
	b := Base{}
	err := fmt.Errorf("test error")
	b.SetErr(err)

	if !errors.Is(b.Err, err) {
		t.Errorf("Err = %v, want %v", b.Err, err)
	}

	b.ClearErr()
	if b.Err != nil {
		t.Errorf("Err = %v after ClearErr, want nil", b.Err)
	}
}

func TestBase_InputWidth(t *testing.T) {
	t.Parallel()
	b := Base{Width: 80}
	w := b.InputWidth()
	// ModalInputWidth(80) = 80 - 20 = 60
	if w != 60 {
		t.Errorf("InputWidth() = %d, want 60", w)
	}
}

func TestBase_InputWidth_Minimum(t *testing.T) {
	t.Parallel()
	b := Base{Width: 30}
	w := b.InputWidth()
	// ModalInputWidth(30) = max(30-20, 40) = 40
	if w != 40 {
		t.Errorf("InputWidth() = %d, want 40", w)
	}
}

func TestBase_ContentHeight(t *testing.T) {
	t.Parallel()
	b := Base{Height: 30}
	h := b.ContentHeight()
	// 30 - 4 = 26
	if h != 26 {
		t.Errorf("ContentHeight() = %d, want 26", h)
	}
}

func TestBase_ContentHeight_Small(t *testing.T) {
	t.Parallel()
	b := Base{Height: 2}
	h := b.ContentHeight()
	if h != 0 {
		t.Errorf("ContentHeight() = %d, want 0", h)
	}
}
