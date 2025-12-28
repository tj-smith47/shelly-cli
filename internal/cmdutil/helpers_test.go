package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func TestAddCommandsToGroup(t *testing.T) {
	t.Parallel()

	t.Run("adds commands with group ID", func(t *testing.T) {
		t.Parallel()

		root := &cobra.Command{Use: "root"}
		root.AddGroup(&cobra.Group{ID: "test-group", Title: "Test"})

		cmd1 := &cobra.Command{Use: "cmd1"}
		cmd2 := &cobra.Command{Use: "cmd2"}
		cmd3 := &cobra.Command{Use: "cmd3"}

		cmdutil.AddCommandsToGroup(root, "test-group", cmd1, cmd2, cmd3)

		// Verify all commands are added
		if len(root.Commands()) != 3 {
			t.Errorf("root has %d commands, want 3", len(root.Commands()))
		}

		// Verify group IDs are set
		for _, cmd := range root.Commands() {
			if cmd.GroupID != "test-group" {
				t.Errorf("cmd %s GroupID = %q, want %q", cmd.Use, cmd.GroupID, "test-group")
			}
		}
	})

	t.Run("empty commands list", func(t *testing.T) {
		t.Parallel()

		root := &cobra.Command{Use: "root"}

		cmdutil.AddCommandsToGroup(root, "test-group")

		if len(root.Commands()) != 0 {
			t.Errorf("root has %d commands, want 0", len(root.Commands()))
		}
	})

	t.Run("single command", func(t *testing.T) {
		t.Parallel()

		root := &cobra.Command{Use: "root"}
		root.AddGroup(&cobra.Group{ID: "core", Title: "Core"})

		cmd := &cobra.Command{Use: "test"}

		cmdutil.AddCommandsToGroup(root, "core", cmd)

		if len(root.Commands()) != 1 {
			t.Errorf("root has %d commands, want 1", len(root.Commands()))
		}
		if root.Commands()[0].GroupID != "core" {
			t.Errorf("cmd GroupID = %q, want %q", root.Commands()[0].GroupID, "core")
		}
	})
}
