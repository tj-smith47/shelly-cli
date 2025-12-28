package config

import (
	"testing"
	"time"
)

func TestAlert_IsSnoozed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		snoozedTime string
		want        bool
	}{
		{
			name:        "empty snooze time",
			snoozedTime: "",
			want:        false,
		},
		{
			name:        "future snooze time",
			snoozedTime: time.Now().Add(time.Hour).Format(time.RFC3339),
			want:        true,
		},
		{
			name:        "past snooze time",
			snoozedTime: time.Now().Add(-time.Hour).Format(time.RFC3339),
			want:        false,
		},
		{
			name:        "invalid time format",
			snoozedTime: "invalid-time",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := Alert{
				Name:         "test",
				SnoozedUntil: tt.snoozedTime,
			}
			if got := a.IsSnoozed(); got != tt.want {
				t.Errorf("IsSnoozed() = %v, want %v", got, tt.want)
			}
		})
	}
}
