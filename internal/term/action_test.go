package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-go/gen1"
)

func TestDisplayGen1Actions_Nil(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	DisplayGen1Actions(ios, nil)

	output := out.String()
	if !strings.Contains(output, "No actions configured") {
		t.Error("expected 'No actions configured' message for nil actions")
	}
}

func TestDisplayGen1Actions_Empty(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	actions := &gen1.ActionSettings{
		Actions: []gen1.Action{},
	}
	DisplayGen1Actions(ios, actions)

	output := out.String()
	if !strings.Contains(output, "No actions configured") {
		t.Error("expected 'No actions configured' message for empty actions")
	}
}

func TestDisplayGen1Actions_WithActions(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	actions := &gen1.ActionSettings{
		Actions: []gen1.Action{
			{
				Index:   0,
				Event:   gen1.ActionOutputOn,
				Enabled: true,
				URLs:    []string{"http://example.com/on"},
			},
			{
				Index:   1,
				Event:   gen1.ActionOutputOff,
				Enabled: false,
				URLs:    []string{"http://example.com/off1", "http://example.com/off2"},
			},
			{
				Index:   2,
				Event:   gen1.ActionShortpush,
				Enabled: true,
				URLs:    []string{},
			},
		},
	}
	DisplayGen1Actions(ios, actions)

	output := out.String()
	if !strings.Contains(output, "INDEX") {
		t.Error("expected table header INDEX")
	}
	if !strings.Contains(output, "EVENT") {
		t.Error("expected table header EVENT")
	}
	if !strings.Contains(output, "ENABLED") {
		t.Error("expected table header ENABLED")
	}
	if !strings.Contains(output, "URLS") {
		t.Error("expected table header URLS")
	}
	if !strings.Contains(output, "yes") {
		t.Error("expected 'yes' for enabled action")
	}
	if !strings.Contains(output, "no") {
		t.Error("expected 'no' for disabled action")
	}
	if !strings.Contains(output, "+1 more") {
		t.Error("expected '+1 more' for multiple URLs")
	}
}

func TestDisplayGen1ActionURLs_WithURLs(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	action := gen1.Action{
		Index:   0,
		Event:   gen1.ActionOutputOn,
		Enabled: true,
		URLs:    []string{"http://example.com/url1", "http://example.com/url2"},
	}
	DisplayGen1ActionURLs(ios, action)

	output := out.String()
	if !strings.Contains(output, "Action: out_on") {
		t.Error("expected action event name")
	}
	if !strings.Contains(output, "index 0") {
		t.Error("expected index")
	}
	if !strings.Contains(output, "Enabled: true") {
		t.Error("expected enabled status")
	}
	if !strings.Contains(output, "[0]") {
		t.Error("expected URL index [0]")
	}
	if !strings.Contains(output, "[1]") {
		t.Error("expected URL index [1]")
	}
	if !strings.Contains(output, "http://example.com/url1") {
		t.Error("expected first URL")
	}
}

func TestDisplayGen1ActionURLs_NoURLs(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	action := gen1.Action{
		Index:   1,
		Event:   gen1.ActionOutputOff,
		Enabled: false,
		URLs:    []string{},
	}
	DisplayGen1ActionURLs(ios, action)

	output := out.String()
	if !strings.Contains(output, "URLs: none") {
		t.Error("expected 'URLs: none' for empty URLs")
	}
	if !strings.Contains(output, "Enabled: false") {
		t.Error("expected enabled status false")
	}
}
