package cmdutil

import (
	"testing"

	"github.com/spf13/cobra"
)

// buildTree wires a parent/child cobra tree so the ancestor walk in
// IncludeCommandAsMCPTool has a real CommandPath to traverse.
func buildTree(child *cobra.Command, ancestors ...*cobra.Command) *cobra.Command {
	cur := child
	for _, parent := range ancestors {
		parent.AddCommand(cur)
		cur = parent
	}
	return child
}

func TestIncludeCommandAsMCPTool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  *cobra.Command
		want bool
	}{
		{
			name: "ordinary leaf is included",
			cmd:  buildTree(&cobra.Command{Use: "get"}, &cobra.Command{Use: "config"}, &cobra.Command{Use: "shelly"}),
			want: true,
		},
		{
			// Regression: the binary "shelly" contains the substring "sh"; an exact
			// match must NOT treat the root as excluded (the old substring filter did,
			// which dropped every tool).
			name: "binary root containing sh substring is included",
			cmd:  &cobra.Command{Use: "shelly"},
			want: true,
		},
		{
			// Regression: "switch" contains the letter "i"; exact-match must include it.
			name: "command containing i substring is included",
			cmd:  buildTree(&cobra.Command{Use: "on"}, &cobra.Command{Use: "switch"}, &cobra.Command{Use: "shelly"}),
			want: true,
		},
		{
			name: "excluded command by name is dropped",
			cmd:  buildTree(&cobra.Command{Use: "monitor"}, &cobra.Command{Use: "shelly"}),
			want: false,
		},
		{
			name: "child of an excluded parent is dropped",
			cmd:  buildTree(&cobra.Command{Use: "scan"}, &cobra.Command{Use: cmdNameWiFi}, &cobra.Command{Use: "shelly"}),
			want: false,
		},
		{
			// The command's own name is clean but an alias is excluded — this is the
			// only shape that exercises the alias check (a command excluded by NAME
			// returns before the alias loop is ever reached).
			name: "excluded by alias while name is clean",
			cmd: buildTree(
				&cobra.Command{Use: "watch", Aliases: []string{"mon"}},
				&cobra.Command{Use: "shelly"},
			),
			want: false,
		},
		{
			name: "excluded by short alias",
			cmd: buildTree(
				&cobra.Command{Use: "interactive", Aliases: []string{"i", "repl"}},
				&cobra.Command{Use: "shelly"},
			),
			want: false,
		},
		{
			name: "child of a parent excluded only by alias is dropped",
			cmd: buildTree(
				&cobra.Command{Use: "open"},
				&cobra.Command{Use: "shell", Aliases: []string{"sh"}},
				&cobra.Command{Use: "shelly"},
			),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IncludeCommandAsMCPTool(tt.cmd); got != tt.want {
				t.Errorf("IncludeCommandAsMCPTool(%q) = %v, want %v", tt.cmd.CommandPath(), got, tt.want)
			}
		})
	}
}
