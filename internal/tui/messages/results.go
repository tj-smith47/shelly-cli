// Package messages provides shared message types for TUI components.
package messages

// SaveResultMsg is a generic message for save operation results.
// Components can use this type directly or embed it in component-specific messages.
type SaveResultMsg struct {
	// ComponentID identifies the component that was saved.
	// Can be int (for indexed components like inputs, webhooks) or string (for keyed components like kvs, virtuals).
	ComponentID any

	// Success indicates whether the save operation succeeded.
	// When true, Err should be nil.
	Success bool

	// Err contains any error that occurred during the save operation.
	Err error
}

// NewSaveResult creates a SaveResultMsg for a successful save.
func NewSaveResult(componentID any) SaveResultMsg {
	return SaveResultMsg{
		ComponentID: componentID,
		Success:     true,
	}
}

// NewSaveError creates a SaveResultMsg for a failed save.
func NewSaveError(componentID any, err error) SaveResultMsg {
	return SaveResultMsg{
		ComponentID: componentID,
		Success:     false,
		Err:         err,
	}
}

// EditClosedMsg is a generic message for when an edit modal closes.
type EditClosedMsg struct {
	// Saved indicates whether changes were saved before closing.
	Saved bool
}

// EditOpenedMsg is a generic message for when an edit modal opens.
type EditOpenedMsg struct{}
