package cmd

import (
	"sort"
	"testing"

	"github.com/spf13/cobra"
)

// TestSiblingNamesAndAliasesAreUnique fails when two subcommands of one parent
// answer to the same word. Cobra resolves the word to whichever it meets first,
// so the other command cannot be reached by that name: `group delete` once
// carried the alias "remove", which hid `group remove <group> <device>`.
func TestSiblingNamesAndAliasesAreUnique(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	var problems []string
	walkCommands(rootCmd, func(parent *cobra.Command) {
		owner := map[string]string{}
		for _, child := range parent.Commands() {
			words := append([]string{child.Name()}, child.Aliases...)
			for _, w := range words {
				if other, taken := owner[w]; taken && other != child.Name() {
					problems = append(problems, parent.CommandPath()+": \""+w+"\" names both \""+other+"\" and \""+child.Name()+"\"")
					continue
				}
				owner[w] = child.Name()
			}
		}
	})
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}
