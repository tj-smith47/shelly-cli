// Package table provides table rendering with styled output.
package table

// Builder provides a fluent API for constructing tables.
// It separates the construction phase from the final Table object.
//
// Example:
//
//	t := table.NewBuilder("Name", "Status", "Power").
//	    AddRow("switch:0", "ON", "45.2W").
//	    AddRow("switch:1", "off", "0W").
//	    WithStyle(table.PlainStyle()).
//	    Build()
//	t.PrintTo(w)

// Builder provides a fluent API for constructing tables.
type Builder struct {
	headers      []string
	rows         [][]string
	separators   map[int]int // row index -> starting column for separator
	style        *Style
	hideHeaders  bool
	rowLines     bool
	mergeHeaders bool
}

// NewBuilder creates a new table builder with the specified headers.
func NewBuilder(headers ...string) *Builder {
	return &Builder{
		headers:    headers,
		rows:       [][]string{},
		separators: make(map[int]int),
	}
}

// AddRow adds a row to the builder.
// Returns the builder for method chaining.
func (b *Builder) AddRow(values ...string) *Builder {
	// Ensure row has same number of columns as headers
	row := make([]string, len(b.headers))
	for i := range row {
		if i < len(values) {
			row[i] = values[i]
		}
	}
	b.rows = append(b.rows, row)
	return b
}

// AddRows adds multiple rows to the builder.
// Returns the builder for method chaining.
func (b *Builder) AddRows(rows [][]string) *Builder {
	for _, row := range rows {
		b.AddRow(row...)
	}
	return b
}

// AddSeparator adds a full-width separator line between sections.
// Returns the builder for method chaining.
func (b *Builder) AddSeparator() *Builder {
	return b.AddSeparatorAt(0)
}

// AddSeparatorAt adds a separator line starting at the given column.
// Columns before startCol will not have the separator.
// Returns the builder for method chaining.
func (b *Builder) AddSeparatorAt(startCol int) *Builder {
	b.separators[len(b.rows)] = startCol
	return b
}

// WithStyle sets the table style.
// Returns the builder for method chaining.
func (b *Builder) WithStyle(style Style) *Builder {
	b.style = &style
	return b
}

// WithPlainStyle sets plain mode styling (no borders, aligned columns).
// Returns the builder for method chaining.
func (b *Builder) WithPlainStyle() *Builder {
	style := PlainStyle()
	b.style = &style
	return b
}

// WithNoColorStyle sets no-color styling (ASCII borders, no ANSI codes).
// Returns the builder for method chaining.
func (b *Builder) WithNoColorStyle() *Builder {
	style := NoColorStyle()
	b.style = &style
	return b
}

// WithModeStyle sets styling based on the ModeChecker (IOStreams).
// This automatically selects plain/no-color/default style.
// Returns the builder for method chaining.
func (b *Builder) WithModeStyle(ios ModeChecker) *Builder {
	style := GetStyle(ios)
	b.style = &style
	return b
}

// HideHeaders hides the header row in the output.
// Returns the builder for method chaining.
func (b *Builder) HideHeaders() *Builder {
	b.hideHeaders = true
	return b
}

// WithRowLines enables lines between rows for better visual separation.
// Returns the builder for method chaining.
func (b *Builder) WithRowLines() *Builder {
	b.rowLines = true
	return b
}

// MergeEmptyHeaders merges empty header cells with the previous non-empty header.
// This creates a visual spanning effect in the header row.
// Returns the builder for method chaining.
func (b *Builder) MergeEmptyHeaders() *Builder {
	b.mergeHeaders = true
	return b
}

// Build constructs and returns the final Table.
func (b *Builder) Build() *Table {
	t := New(b.headers...)

	// Apply rows
	for _, row := range b.rows {
		t.AddRow(row...)
	}

	// Apply style if set
	if b.style != nil {
		t.SetStyle(*b.style)
	}

	// Apply header visibility
	if b.hideHeaders {
		t.HideHeaders()
	}

	// Apply row lines
	if b.rowLines {
		t.WithRowLines()
	}

	// Apply header merging
	if b.mergeHeaders {
		t.MergeEmptyHeaders()
	}

	// Apply separators
	t.separators = b.separators

	return t
}

// RowCount returns the number of rows added to the builder.
func (b *Builder) RowCount() int {
	return len(b.rows)
}

// Empty returns true if the builder has no rows.
func (b *Builder) Empty() bool {
	return len(b.rows) == 0
}
