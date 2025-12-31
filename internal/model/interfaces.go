// Package model defines domain models and interfaces for the CLI.
package model

// Listable represents a component that can be displayed in a list table.
// Types implementing this interface can be used with the generic DisplayList function.
type Listable interface {
	// ListHeaders returns the column headers for the table.
	ListHeaders() []string

	// ListRow returns the formatted row values for the table.
	ListRow() []string
}
