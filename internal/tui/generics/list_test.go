package generics

import (
	"testing"
)

// testItem implements Selectable for testing.
type testItem struct {
	Name     string
	Value    int
	selected bool
}

func (t *testItem) IsSelected() bool   { return t.selected }
func (t *testItem) SetSelected(v bool) { t.selected = v }

func TestFilter(t *testing.T) {
	t.Parallel()
	items := []int{1, 2, 3, 4, 5, 6}
	evens := Filter(items, func(n int) bool { return n%2 == 0 })

	if len(evens) != 3 {
		t.Errorf("expected 3 evens, got %d", len(evens))
	}
	if evens[0] != 2 || evens[1] != 4 || evens[2] != 6 {
		t.Errorf("expected [2,4,6], got %v", evens)
	}
}

func TestFilterEmpty(t *testing.T) {
	t.Parallel()
	var items []int
	result := Filter(items, func(n int) bool { return true })
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestFilterSelected(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", selected: true},
		{Name: "b", selected: false},
		{Name: "c", selected: true},
	}

	selected := FilterSelected(items)
	if len(selected) != 2 {
		t.Errorf("expected 2 selected, got %d", len(selected))
	}
	if selected[0].Name != "a" || selected[1].Name != "c" {
		t.Errorf("expected [a,c], got %v", selected)
	}
}

func TestSelectAll(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", selected: false},
		{Name: "b", selected: false},
		{Name: "c", selected: true},
	}

	SelectAll(items)

	for _, item := range items {
		if !item.IsSelected() {
			t.Errorf("expected all selected, %s was not", item.Name)
		}
	}
}

func TestSelectNone(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", selected: true},
		{Name: "b", selected: true},
		{Name: "c", selected: false},
	}

	SelectNone(items)

	for _, item := range items {
		if item.IsSelected() {
			t.Errorf("expected none selected, %s was", item.Name)
		}
	}
}

func TestSelectWhere(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", Value: 1, selected: false},
		{Name: "b", Value: 2, selected: false},
		{Name: "c", Value: 3, selected: false},
	}

	SelectWhere(items, func(item *testItem) bool {
		return item.Value > 1
	})

	if items[0].IsSelected() {
		t.Error("expected a not selected")
	}
	if !items[1].IsSelected() {
		t.Error("expected b selected")
	}
	if !items[2].IsSelected() {
		t.Error("expected c selected")
	}
}

func TestToggleAt(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", selected: false},
		{Name: "b", selected: true},
	}

	// Toggle first item
	if !ToggleAt(items, 0) {
		t.Error("expected ToggleAt to return true")
	}
	if !items[0].IsSelected() {
		t.Error("expected a to be selected after toggle")
	}

	// Toggle second item
	if !ToggleAt(items, 1) {
		t.Error("expected ToggleAt to return true")
	}
	if items[1].IsSelected() {
		t.Error("expected b to be unselected after toggle")
	}

	// Out of bounds
	if ToggleAt(items, -1) {
		t.Error("expected ToggleAt to return false for negative index")
	}
	if ToggleAt(items, 2) {
		t.Error("expected ToggleAt to return false for out of bounds index")
	}
}

func TestCountSelected(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{selected: true},
		{selected: false},
		{selected: true},
		{selected: true},
	}

	count := CountSelected(items)
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestMap(t *testing.T) {
	t.Parallel()
	items := []int{1, 2, 3}
	doubled := Map(items, func(n int) int { return n * 2 })

	if len(doubled) != 3 {
		t.Errorf("expected 3 items, got %d", len(doubled))
	}
	if doubled[0] != 2 || doubled[1] != 4 || doubled[2] != 6 {
		t.Errorf("expected [2,4,6], got %v", doubled)
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()
	left := []int{1, 3, 5}
	right := []int{2, 4, 6}

	merged := Merge(left, right, func(a, b int) bool { return a < b })

	expected := []int{1, 2, 3, 4, 5, 6}
	if len(merged) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(merged))
	}
	for i, v := range expected {
		if merged[i] != v {
			t.Errorf("at index %d: expected %d, got %d", i, v, merged[i])
		}
	}
}

func TestMergeEmptySlices(t *testing.T) {
	t.Parallel()
	var left, right []int
	merged := Merge(left, right, func(a, b int) bool { return a < b })
	if len(merged) != 0 {
		t.Errorf("expected empty, got %v", merged)
	}

	left = []int{1, 2, 3}
	merged = Merge(left, right, func(a, b int) bool { return a < b })
	if len(merged) != 3 {
		t.Errorf("expected 3, got %d", len(merged))
	}
}

func TestAny(t *testing.T) {
	t.Parallel()
	items := []int{1, 2, 3, 4, 5}

	if !Any(items, func(n int) bool { return n == 3 }) {
		t.Error("expected Any to return true for 3")
	}
	if Any(items, func(n int) bool { return n == 10 }) {
		t.Error("expected Any to return false for 10")
	}
	if Any([]int{}, func(n int) bool { return true }) {
		t.Error("expected Any to return false for empty slice")
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	items := []int{2, 4, 6, 8}

	if !All(items, func(n int) bool { return n%2 == 0 }) {
		t.Error("expected All to return true for all evens")
	}
	if All(items, func(n int) bool { return n > 5 }) {
		t.Error("expected All to return false when some are <= 5")
	}
	if !All([]int{}, func(n int) bool { return false }) {
		t.Error("expected All to return true for empty slice")
	}
}

func TestFind(t *testing.T) {
	t.Parallel()
	items := []*testItem{
		{Name: "a", Value: 1},
		{Name: "b", Value: 2},
		{Name: "c", Value: 3},
	}

	found, ok := Find(items, func(item *testItem) bool { return item.Value == 2 })
	if !ok {
		t.Error("expected Find to return true")
	}
	if found.Name != "b" {
		t.Errorf("expected 'b', got '%s'", found.Name)
	}

	_, ok = Find(items, func(item *testItem) bool { return item.Value == 10 })
	if ok {
		t.Error("expected Find to return false for missing item")
	}
}

func TestFindIndex(t *testing.T) {
	t.Parallel()
	items := []string{"a", "b", "c", "d"}

	idx := FindIndex(items, func(s string) bool { return s == "c" })
	if idx != 2 {
		t.Errorf("expected 2, got %d", idx)
	}

	idx = FindIndex(items, func(s string) bool { return s == "z" })
	if idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

// testValueItem is a value type for testing Func-based selection helpers.
type testValueItem struct {
	Name     string
	Selected bool
}

func getSelected(t *testValueItem) bool    { return t.Selected }
func setSelected(t *testValueItem, v bool) { t.Selected = v }

func TestToggleAtFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItem{
		{Name: "a", Selected: false},
		{Name: "b", Selected: true},
	}

	if !ToggleAtFunc(items, 0, getSelected, setSelected) {
		t.Error("expected ToggleAtFunc to return true")
	}
	if !items[0].Selected {
		t.Error("expected a to be selected after toggle")
	}

	if !ToggleAtFunc(items, 1, getSelected, setSelected) {
		t.Error("expected ToggleAtFunc to return true")
	}
	if items[1].Selected {
		t.Error("expected b to be unselected after toggle")
	}

	if ToggleAtFunc(items, -1, getSelected, setSelected) {
		t.Error("expected ToggleAtFunc to return false for negative index")
	}
	if ToggleAtFunc(items, 2, getSelected, setSelected) {
		t.Error("expected ToggleAtFunc to return false for out of bounds")
	}
}

func TestSelectAllFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItem{
		{Name: "a", Selected: false},
		{Name: "b", Selected: false},
		{Name: "c", Selected: true},
	}

	SelectAllFunc(items, setSelected)

	for _, item := range items {
		if !item.Selected {
			t.Errorf("expected all selected, %s was not", item.Name)
		}
	}
}

func TestSelectNoneFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItem{
		{Name: "a", Selected: true},
		{Name: "b", Selected: true},
		{Name: "c", Selected: false},
	}

	SelectNoneFunc(items, setSelected)

	for _, item := range items {
		if item.Selected {
			t.Errorf("expected none selected, %s was", item.Name)
		}
	}
}

func TestCountSelectedFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItem{
		{Selected: true},
		{Selected: false},
		{Selected: true},
		{Selected: true},
	}

	count := CountSelectedFunc(items, getSelected)
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

// testValueItemWithValue is a value type for testing conditional selection.
type testValueItemWithValue struct {
	Name     string
	Value    int
	Selected bool
}

func setValueSelected(t *testValueItemWithValue, v bool) { t.Selected = v }
func hasHighValue(t *testValueItemWithValue) bool        { return t.Value > 5 }

func TestSelectWhereFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItemWithValue{
		{Name: "a", Value: 3, Selected: false},
		{Name: "b", Value: 7, Selected: false},
		{Name: "c", Value: 10, Selected: false},
	}

	SelectWhereFunc(items, hasHighValue, setValueSelected)

	if items[0].Selected {
		t.Error("expected a not selected (value <= 5)")
	}
	if !items[1].Selected {
		t.Error("expected b selected (value > 5)")
	}
	if !items[2].Selected {
		t.Error("expected c selected (value > 5)")
	}
}

func TestCountWhereFunc(t *testing.T) {
	t.Parallel()
	items := []testValueItemWithValue{
		{Name: "a", Value: 3},
		{Name: "b", Value: 7},
		{Name: "c", Value: 10},
		{Name: "d", Value: 2},
	}

	count := CountWhereFunc(items, hasHighValue)
	if count != 2 {
		t.Errorf("expected 2 items with value > 5, got %d", count)
	}
}
