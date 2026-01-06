// Package generics provides generic utilities for TUI components.
package generics

// Selectable defines an interface for items that can be selected.
type Selectable interface {
	IsSelected() bool
	SetSelected(bool)
}

// Filter returns a new slice containing only elements that match the predicate.
func Filter[T any](items []T, pred func(T) bool) []T {
	result := make([]T, 0, len(items)/2) // Reasonable default capacity
	for _, item := range items {
		if pred(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterSelected returns all items where IsSelected() returns true.
func FilterSelected[T Selectable](items []T) []T {
	return Filter(items, func(item T) bool {
		return item.IsSelected()
	})
}

// SelectAll sets Selected to true for all items.
func SelectAll[T Selectable](items []T) {
	for i := range items {
		items[i].SetSelected(true)
	}
}

// SelectNone sets Selected to false for all items.
func SelectNone[T Selectable](items []T) {
	for i := range items {
		items[i].SetSelected(false)
	}
}

// SelectWhere sets Selected to true for items matching the predicate.
func SelectWhere[T Selectable](items []T, pred func(T) bool) {
	for i := range items {
		if pred(items[i]) {
			items[i].SetSelected(true)
		}
	}
}

// ToggleAt toggles the Selected state at the given index.
// Returns false if index is out of bounds.
func ToggleAt[T Selectable](items []T, index int) bool {
	if index < 0 || index >= len(items) {
		return false
	}
	items[index].SetSelected(!items[index].IsSelected())
	return true
}

// CountSelected returns the number of selected items.
func CountSelected[T Selectable](items []T) int {
	count := 0
	for _, item := range items {
		if item.IsSelected() {
			count++
		}
	}
	return count
}

// Map transforms a slice of items using the provided function.
func Map[T, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// Merge combines two sorted slices using the less function for ordering.
// Both input slices must already be sorted according to less.
func Merge[T any](left, right []T, less func(a, b T) bool) []T {
	result := make([]T, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if less(left[i], right[j]) {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// Append remaining items
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

// Any returns true if any item matches the predicate.
func Any[T any](items []T, pred func(T) bool) bool {
	for _, item := range items {
		if pred(item) {
			return true
		}
	}
	return false
}

// All returns true if all items match the predicate.
func All[T any](items []T, pred func(T) bool) bool {
	for _, item := range items {
		if !pred(item) {
			return false
		}
	}
	return true
}

// Find returns the first item matching the predicate and true, or zero value and false.
func Find[T any](items []T, pred func(T) bool) (T, bool) {
	for _, item := range items {
		if pred(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first item matching the predicate, or -1.
func FindIndex[T any](items []T, pred func(T) bool) int {
	for i, item := range items {
		if pred(item) {
			return i
		}
	}
	return -1
}

// SelectableItem provides selection operations for items accessed by function.
// This is useful when working with value slices where interface methods won't work.

// ToggleAtFunc toggles the selection at index using the provided getter/setter.
// Returns false if index is out of bounds.
func ToggleAtFunc[T any](items []T, index int, get func(*T) bool, set func(*T, bool)) bool {
	if index < 0 || index >= len(items) {
		return false
	}
	set(&items[index], !get(&items[index]))
	return true
}

// SelectAllFunc sets all items as selected using the provided setter.
func SelectAllFunc[T any](items []T, set func(*T, bool)) {
	for i := range items {
		set(&items[i], true)
	}
}

// SelectNoneFunc sets all items as unselected using the provided setter.
func SelectNoneFunc[T any](items []T, set func(*T, bool)) {
	for i := range items {
		set(&items[i], false)
	}
}

// CountSelectedFunc returns the count of selected items using the provided getter.
func CountSelectedFunc[T any](items []T, get func(*T) bool) int {
	count := 0
	for i := range items {
		if get(&items[i]) {
			count++
		}
	}
	return count
}

// SelectWhereFunc sets items as selected where predicate returns true.
func SelectWhereFunc[T any](items []T, pred func(*T) bool, set func(*T, bool)) {
	for i := range items {
		if pred(&items[i]) {
			set(&items[i], true)
		}
	}
}

// CountWhereFunc returns the count of items matching predicate.
func CountWhereFunc[T any](items []T, pred func(*T) bool) int {
	count := 0
	for i := range items {
		if pred(&items[i]) {
			count++
		}
	}
	return count
}
