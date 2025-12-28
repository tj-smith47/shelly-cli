package term

import (
	"strings"
	"testing"

	"github.com/tj-smith47/shelly-cli/internal/model"
)

func TestDisplayGroups(t *testing.T) {
	t.Parallel()

	t.Run("with groups", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		groups := []model.GroupInfo{
			{Name: "Living Room", DeviceCount: 3},
			{Name: "Kitchen", DeviceCount: 2},
			{Name: "Bedroom", DeviceCount: 1},
		}

		DisplayGroups(ios, groups)

		output := out.String()
		if !strings.Contains(output, "Living Room") {
			t.Error("output should contain 'Living Room'")
		}
		if !strings.Contains(output, "Kitchen") {
			t.Error("output should contain 'Kitchen'")
		}
		if !strings.Contains(output, "3 group") {
			t.Error("output should contain group count")
		}
	})

	t.Run("empty groups", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayGroups(ios, []model.GroupInfo{})

		output := out.String()
		if !strings.Contains(output, "0 group") {
			t.Error("output should contain '0 group'")
		}
	})

	t.Run("single group", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		groups := []model.GroupInfo{
			{Name: "Test Group", DeviceCount: 5},
		}

		DisplayGroups(ios, groups)

		output := out.String()
		if !strings.Contains(output, "Test Group") {
			t.Error("output should contain 'Test Group'")
		}
		if !strings.Contains(output, "1 group") {
			t.Error("output should contain '1 group'")
		}
	})
}

func TestDisplayGroupMembers(t *testing.T) {
	t.Parallel()

	t.Run("with members", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		devices := []string{"device1", "device2", "device3"}

		DisplayGroupMembers(ios, "Living Room", devices)

		output := out.String()
		if !strings.Contains(output, "Living Room") {
			t.Error("output should contain group name")
		}
		if !strings.Contains(output, "device1") {
			t.Error("output should contain 'device1'")
		}
		if !strings.Contains(output, "device2") {
			t.Error("output should contain 'device2'")
		}
		if !strings.Contains(output, "3 member") {
			t.Error("output should contain member count")
		}
	})

	t.Run("empty members", func(t *testing.T) {
		t.Parallel()

		ios, out, _ := testIOStreams()
		DisplayGroupMembers(ios, "Empty Group", []string{})

		output := out.String()
		if !strings.Contains(output, "Empty Group") {
			t.Error("output should contain group name")
		}
		if !strings.Contains(output, "0 member") {
			t.Error("output should contain '0 member'")
		}
	})
}

func TestGroupInfo_Fields(t *testing.T) {
	t.Parallel()

	info := model.GroupInfo{
		Name:        "Test Group",
		DeviceCount: 5,
	}

	if info.Name != "Test Group" {
		t.Errorf("got Name=%q, want Test Group", info.Name)
	}
	if info.DeviceCount != 5 {
		t.Errorf("got DeviceCount=%d, want 5", info.DeviceCount)
	}
}
