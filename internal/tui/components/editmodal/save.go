package editmodal

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tj-smith47/shelly-cli/internal/tui/messages"
)

// SaveTimeout is the default timeout for save and async operations.
const SaveTimeout = 30 * time.Second

// StartSave sets the modal into saving state and clears any previous error.
func (b *Base) StartSave() {
	b.Saving = true
	b.Err = nil
}

// SaveCmd creates an async tea.Cmd that executes fn with a 30s timeout
// and returns a SaveResultMsg. The ComponentID is nil.
func (b *Base) SaveCmd(fn func(ctx context.Context) error) tea.Cmd {
	return b.SaveCmdWithID(nil, fn)
}

// SaveCmdWithID creates an async tea.Cmd that executes fn with a 30s timeout
// and returns a SaveResultMsg with the given componentID.
func (b *Base) SaveCmdWithID(componentID any, fn func(ctx context.Context) error) tea.Cmd {
	ctx := b.Ctx
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, SaveTimeout)
		defer cancel()

		if err := fn(ctx); err != nil {
			return messages.NewSaveError(componentID, err)
		}
		return messages.NewSaveResult(componentID)
	}
}

// HandleSaveResult processes a SaveResultMsg, updating modal state.
// On error: sets Saving=false, sets Err, returns (false, nil).
// On success: sets Saving=false, hides modal, returns (true, closeCmd).
// The closeCmd sends an EditClosedMsg{Saved: true}.
func (b *Base) HandleSaveResult(msg messages.SaveResultMsg) (saved bool, closeCmd tea.Cmd) {
	b.Saving = false
	if msg.Err != nil {
		b.Err = msg.Err
		return false, nil
	}
	b.Hide()
	return true, func() tea.Msg {
		return messages.EditClosedMsg{Saved: true}
	}
}
