package client

import (
	"encoding/json"
	"testing"
)

func TestAsObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		result any
		wantOK bool
	}{
		{"raw object, as Call returns it", json.RawMessage(`{"enable":true}`), true},
		{"decoded map", map[string]any{"enable": true}, true},
		{"raw array", json.RawMessage(`[1,2]`), false},
		{"raw null", json.RawMessage(`null`), false},
		{"nil", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			obj, ok := AsObject(tt.result)
			if ok != tt.wantOK {
				t.Fatalf("AsObject() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && obj["enable"] != true {
				t.Errorf("AsObject() = %v, want enable=true", obj)
			}
		})
	}
}
