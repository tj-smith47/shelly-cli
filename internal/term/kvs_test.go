package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/shelly/kvs"
)

func TestDisplayKVSRaw(t *testing.T) {
	t.Parallel()

	t.Run("string value", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := &kvs.GetResult{
			Value: "test string",
			Etag:  "etag123",
		}

		DisplayKVSRaw(ios, result)

		output := out.String()
		if !strings.Contains(output, "test string") {
			t.Errorf("output should contain 'test string', got %q", output)
		}
	})

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := &kvs.GetResult{
			Value: nil,
			Etag:  "etag123",
		}

		DisplayKVSRaw(ios, result)

		output := out.String()
		if !strings.Contains(output, "null") {
			t.Errorf("output should contain 'null', got %q", output)
		}
	})

	t.Run("number value", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := &kvs.GetResult{
			Value: 42,
			Etag:  "etag123",
		}

		DisplayKVSRaw(ios, result)

		output := out.String()
		if !strings.Contains(output, "42") {
			t.Errorf("output should contain '42', got %q", output)
		}
	})

	t.Run("bool value", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := &kvs.GetResult{
			Value: true,
			Etag:  "etag123",
		}

		DisplayKVSRaw(ios, result)

		output := out.String()
		if !strings.Contains(output, "true") {
			t.Errorf("output should contain 'true', got %q", output)
		}
	})
}

func TestDisplayKVSResult(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	result := &kvs.GetResult{
		Value: "test value",
		Etag:  "etag123",
	}

	DisplayKVSResult(ios, "mykey", result)

	output := out.String()
	if !strings.Contains(output, "mykey") {
		t.Error("output should contain key name")
	}
	if !strings.Contains(output, "test value") {
		t.Error("output should contain value")
	}
	if !strings.Contains(output, "etag123") {
		t.Error("output should contain etag")
	}
}

func TestDisplayKVSKeys(t *testing.T) {
	t.Parallel()

	t.Run("with keys", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		result := &kvs.ListResult{
			Keys: []string{"key1", "key2", "key3"},
			Rev:  5,
		}

		DisplayKVSKeys(ios, result)

		output := out.String()
		if !strings.Contains(output, "key1") {
			t.Error("output should contain 'key1'")
		}
		if !strings.Contains(output, "key2") {
			t.Error("output should contain 'key2'")
		}
		if !strings.Contains(output, "3 key(s)") {
			t.Error("output should contain key count")
		}
		if !strings.Contains(output, "revision 5") {
			t.Error("output should contain revision")
		}
	})

	t.Run("no keys", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		result := &kvs.ListResult{
			Keys: []string{},
			Rev:  0,
		}

		DisplayKVSKeys(ios, result)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "No keys stored") {
			t.Errorf("output should contain 'No keys stored', got %q", allOutput)
		}
	})
}

func TestDisplayKVSItems(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	items := []kvs.Item{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: 42},
		{Key: "key3", Value: true},
	}

	DisplayKVSItems(ios, items)

	output := out.String()
	if !strings.Contains(output, "key1") {
		t.Error("output should contain 'key1'")
	}
	if !strings.Contains(output, "3 key(s)") {
		t.Error("output should contain key count")
	}
}

func TestDisplayKVSImportPreview(t *testing.T) {
	t.Parallel()

	ios, out, _ := testIOStreams()
	data := &kvs.Export{
		Items: []kvs.Item{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: "value2"},
		},
	}

	DisplayKVSImportPreview(ios, data)

	output := out.String()
	if !strings.Contains(output, "2 key(s) to import") {
		t.Error("output should contain key count")
	}
	if !strings.Contains(output, "key1") {
		t.Error("output should contain 'key1'")
	}
}

func TestDisplayKVSDryRun(t *testing.T) {
	t.Parallel()

	t.Run("with overwrite", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayKVSDryRun(ios, 5, true)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "5 key(s)") {
			t.Errorf("output should contain key count, got %q", allOutput)
		}
		if !strings.Contains(allOutput, "overwrite") {
			t.Errorf("output should contain 'overwrite', got %q", allOutput)
		}
	})

	t.Run("without overwrite", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayKVSDryRun(ios, 3, false)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "3 key(s)") {
			t.Errorf("output should contain key count, got %q", allOutput)
		}
		if !strings.Contains(allOutput, "skipped") {
			t.Errorf("output should contain 'skipped', got %q", allOutput)
		}
	})
}

func TestDisplayKVSImportResults(t *testing.T) {
	t.Parallel()

	t.Run("all imported", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayKVSImportResults(ios, 5, 0)

		output := out.String()
		if !strings.Contains(output, "5 imported") {
			t.Errorf("output should contain '5 imported', got %q", output)
		}
	})

	t.Run("some skipped", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayKVSImportResults(ios, 3, 2)

		output := out.String()
		if !strings.Contains(output, "3 imported") {
			t.Errorf("output should contain '3 imported', got %q", output)
		}
		if !strings.Contains(output, "2 skipped") {
			t.Errorf("output should contain '2 skipped', got %q", output)
		}
	})

	t.Run("all skipped", func(t *testing.T) {
		t.Parallel()

		ios, out, errOut := testIOStreams()
		DisplayKVSImportResults(ios, 0, 3)

		allOutput := out.String() + errOut.String()
		if !strings.Contains(allOutput, "3 skipped") {
			t.Errorf("output should contain '3 skipped', got %q", allOutput)
		}
	})
}

func TestGetResult_Fields(t *testing.T) {
	t.Parallel()

	result := kvs.GetResult{
		Value: "test",
		Etag:  "abc123",
	}

	if result.Value != "test" {
		t.Errorf("got Value=%v, want test", result.Value)
	}
	if result.Etag != "abc123" {
		t.Errorf("got Etag=%q, want abc123", result.Etag)
	}
}

func TestListResult_Fields(t *testing.T) {
	t.Parallel()

	result := kvs.ListResult{
		Keys: []string{"key1", "key2"},
		Rev:  10,
	}

	if len(result.Keys) != 2 {
		t.Errorf("got %d keys, want 2", len(result.Keys))
	}
	if result.Rev != 10 {
		t.Errorf("got Rev=%d, want 10", result.Rev)
	}
}

func TestItem_Fields(t *testing.T) {
	t.Parallel()

	item := kvs.Item{
		Key:   "testkey",
		Value: "testvalue",
	}

	if item.Key != "testkey" {
		t.Errorf("got Key=%q, want testkey", item.Key)
	}
	if item.Value != "testvalue" {
		t.Errorf("got Value=%v, want testvalue", item.Value)
	}
}

func TestExport_Fields(t *testing.T) {
	t.Parallel()

	export := kvs.Export{
		Items: []kvs.Item{
			{Key: "key1", Value: "value1"},
		},
	}

	if len(export.Items) != 1 {
		t.Errorf("got %d items, want 1", len(export.Items))
	}
	if export.Items[0].Key != "key1" {
		t.Errorf("got Key=%q, want key1", export.Items[0].Key)
	}
}
