// Package automation provides script, schedule, and event automation for Shelly devices.
package automation

import (
	"testing"
)

func TestBuiltInScriptTemplates(t *testing.T) {
	t.Parallel()

	templates := BuiltInScriptTemplates()

	if templates == nil {
		t.Fatal("expected non-nil templates map")
	}

	// Check expected templates exist
	expectedTemplates := []string{
		"motion-light",
		"power-monitor",
		"schedule-helper",
		"toggle-sync",
		"energy-logger",
	}

	for _, name := range expectedTemplates {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tpl, ok := templates[name]
			if !ok {
				t.Errorf("expected template %q to exist", name)
				return
			}
			if tpl.Name != name {
				t.Errorf("got Name=%q, want %q", tpl.Name, name)
			}
			if tpl.Description == "" {
				t.Error("expected non-empty Description")
			}
			if tpl.Category == "" {
				t.Error("expected non-empty Category")
			}
			if tpl.MinGen < 2 {
				t.Errorf("expected MinGen >= 2, got %d", tpl.MinGen)
			}
			if !tpl.BuiltIn {
				t.Error("expected BuiltIn to be true")
			}
			if tpl.Author == "" {
				t.Error("expected non-empty Author")
			}
			if tpl.Version == "" {
				t.Error("expected non-empty Version")
			}
			if tpl.Code == "" {
				t.Error("expected non-empty Code")
			}
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		code   string
		values map[string]any
		want   string
	}{
		{
			name:   "string substitution",
			code:   "let name = NAME;",
			values: map[string]any{"NAME": "test"},
			want:   `let name = "test";`,
		},
		{
			name:   "number substitution",
			code:   "let id = SWITCH_ID;",
			values: map[string]any{"SWITCH_ID": 0},
			want:   "let id = 0;",
		},
		{
			name:   "boolean substitution",
			code:   "let enabled = ENABLED;",
			values: map[string]any{"ENABLED": true},
			want:   "let enabled = true;",
		},
		{
			name:   "nil substitution",
			code:   "let value = VALUE;",
			values: map[string]any{"VALUE": nil},
			want:   "let value = null;",
		},
		{
			name:   "multiple substitutions",
			code:   "let a = A; let b = B;",
			values: map[string]any{"A": 1, "B": 2},
			want:   "let a = 1; let b = 2;",
		},
		{
			name:   "no substitutions",
			code:   "let x = 5;",
			values: map[string]any{},
			want:   "let x = 5;",
		},
		{
			name:   "variable not in code",
			code:   "let x = 5;",
			values: map[string]any{"UNUSED": "value"},
			want:   "let x = 5;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := SubstituteVariables(tt.code, tt.values)

			if got != tt.want {
				t.Errorf("SubstituteVariables(%q, %v) = %q, want %q",
					tt.code, tt.values, got, tt.want)
			}
		})
	}
}

func TestGetScriptTemplate(t *testing.T) {
	t.Parallel()

	// Test getting a built-in template
	tpl, ok := GetScriptTemplate("motion-light")
	if !ok {
		t.Error("expected to find motion-light template")
	}
	if tpl.Name != "motion-light" {
		t.Errorf("got Name=%q, want %q", tpl.Name, "motion-light")
	}

	// Test getting a non-existent template
	_, ok = GetScriptTemplate("non-existent-template")
	if ok {
		t.Error("expected non-existent template to not be found")
	}
}

func TestListAllScriptTemplates(t *testing.T) {
	t.Parallel()

	templates := ListAllScriptTemplates()

	if templates == nil {
		t.Fatal("expected non-nil templates map")
	}

	// Should have at least the built-in templates
	if len(templates) < 5 {
		t.Errorf("expected at least 5 templates, got %d", len(templates))
	}

	// Check built-in templates are included
	expectedTemplates := []string{"motion-light", "power-monitor", "schedule-helper", "toggle-sync", "energy-logger"}
	for _, name := range expectedTemplates {
		if _, ok := templates[name]; !ok {
			t.Errorf("expected template %q to be in list", name)
		}
	}
}

func TestBuiltInTemplates_Variables(t *testing.T) {
	t.Parallel()

	templates := BuiltInScriptTemplates()

	tests := []struct {
		templateName     string
		expectedVars     []string
		expectedRequired []string
	}{
		{
			templateName:     "motion-light",
			expectedVars:     []string{"LIGHT_ID", "INPUT_ID", "TIMEOUT_SEC"},
			expectedRequired: []string{"LIGHT_ID", "INPUT_ID"},
		},
		{
			templateName:     "power-monitor",
			expectedVars:     []string{"SWITCH_ID", "THRESHOLD_W", "CHECK_INTERVAL_SEC"},
			expectedRequired: []string{"SWITCH_ID"},
		},
		{
			templateName:     "schedule-helper",
			expectedVars:     []string{"SWITCH_ID", "ON_HOUR", "OFF_HOUR"},
			expectedRequired: []string{"SWITCH_ID"},
		},
		{
			templateName:     "toggle-sync",
			expectedVars:     []string{"MASTER_ID", "SLAVE_ID"},
			expectedRequired: []string{"MASTER_ID", "SLAVE_ID"},
		},
		{
			templateName:     "energy-logger",
			expectedVars:     []string{"SWITCH_ID"},
			expectedRequired: []string{"SWITCH_ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.templateName, func(t *testing.T) {
			t.Parallel()

			tpl, ok := templates[tt.templateName]
			if !ok {
				t.Fatalf("template %q not found", tt.templateName)
			}

			// Check expected variables exist
			varMap := make(map[string]bool)
			for _, v := range tpl.Variables {
				varMap[v.Name] = true
			}

			for _, expectedVar := range tt.expectedVars {
				if !varMap[expectedVar] {
					t.Errorf("expected variable %q in template %q", expectedVar, tt.templateName)
				}
			}

			// Check required variables
			for _, v := range tpl.Variables {
				for _, requiredName := range tt.expectedRequired {
					if v.Name == requiredName && !v.Required {
						t.Errorf("expected variable %q to be required", requiredName)
					}
				}
			}
		})
	}
}
