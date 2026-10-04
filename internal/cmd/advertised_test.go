package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// repoRoot is the repository root relative to this package.
const repoRoot = "../.."

// advertisedSkipDirs are directories whose files are not hand-written
// advertisements: dependencies, build output, and pages generated from the
// commands' own Example fields (those fields are checked directly).
var advertisedSkipDirs = map[string]bool{
	".git": true, "vendor": true, "bin": true, "dist": true, "node_modules": true,
	".claude": true, ".superpowers": true, ".github": true, "scripts": true,
	"docs/commands": true, "docs/man": true, "docs/site/content/docs/commands": true,
	"docs/site/public": true, "docs/site/themes": true, "docs/site/resources": true,
}

// advertisedSkipFiles are single files outside the population: the changelog
// records what past releases shipped, including commands since removed, and
// the untracked work list describes commands that are not built yet.
var advertisedSkipFiles = map[string]bool{"CHANGELOG.md": true, "PLAN.md": true}

// userDefinedCommands are first words that documentation uses for commands a
// user creates (aliases, plugins), so they are not part of the built-in tree.
// Alias names from examples/aliases and plugin names from examples/plugins are
// added at run time.
var userDefinedCommands = map[string]string{
	"myext":       "placeholder plugin name in the plugin guide",
	"myextension": "placeholder plugin name in the plugin guide",
	"myplugin":    "placeholder plugin name in the plugin guide",
	"hello":       "plugin built in the plugin tutorial",
}

var (
	// inlineSpan matches an invocation quoted in prose with backticks or single
	// quotes. Double quotes are left out: they wrap phrases ("shelly data
	// export"), not commands.
	inlineSpan = regexp.MustCompile("[`'](shelly [^`']+)[`']")
	// commandHead matches "shelly" where a shell would run it: at the start of
	// the line or after an operator or keyword, so prose inside a quoted string
	// ("shelly CLI not found") and arguments ("mv shelly /usr/local/bin") are
	// not read as invocations.
	commandHead = regexp.MustCompile(`(?:^|&&|\|\||[|;!(]|\b(?:if|then|do|else|sudo|exec|time)\s)\s*(shelly\s+[^\n]*)`)
	placeholder = regexp.MustCompile(`^<[^<>]*>`)
)

type advertised struct {
	source string
	tokens []string
	// loose marks an invocation whose arguments are not all written out: a
	// command named in prose, or an alias body that receives the caller's
	// arguments. The command path and flags are checked, the argument count is not.
	loose bool
}

// shellSplit splits a command line into words, honouring quotes, and stops at
// the first shell operator or comment so only one invocation is returned.
func shellSplit(line string) []string {
	var tokens []string
	var cur strings.Builder
	var quote rune
	started := false
	flush := func() {
		if started {
			tokens = append(tokens, cur.String())
			cur.Reset()
			started = false
		}
	}
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			started = true
		case r == '$' && i+1 < len(runes) && runes[i+1] == '(':
			depth := 0
			for ; i < len(runes); i++ {
				cur.WriteRune(runes[i])
				if runes[i] == '(' {
					depth++
				}
				if runes[i] == ')' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			started = true
		case r == ' ' || r == '\t':
			flush()
		case r == '<' && placeholder.MatchString(string(runes[i:])):
			// "<device>" is a usage placeholder, not an input redirect.
			end := i + len([]rune(placeholder.FindString(string(runes[i:]))))
			cur.WriteString(string(runes[i:end]))
			i = end - 1
			started = true
		case r == '(' && !started:
			// "(on/off/status)" lists alternatives in a synopsis.
			cur.WriteRune(r)
			started = true
		case r == '|' || r == ';' || r == '&' || r == '>' || r == '<' || r == ')' || r == '`' || r == '\\':
			// A redirect such as "2>" leaves its file descriptor as the word in progress.
			if r == '>' && (cur.String() == "2" || cur.String() == "1") {
				cur.Reset()
				started = false
			}
			flush()
			return tokens
		case r == '#' && !started:
			return tokens
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	flush()
	return tokens
}

// invocationsInCode returns every shelly invocation on a line of shell code.
func invocationsInCode(line string) []advertised {
	var out []advertised
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return invocationsInProse(line)
	}
	trimmed = strings.TrimPrefix(trimmed, "$ ")
	// A backticked command inside code is a mention (a tree diagram, a comment),
	// read with the looser prose rules.
	out = append(out, invocationsInProse(trimmed)...)
	rest := trimmed
	for {
		loc := commandHead.FindStringSubmatchIndex(rest)
		if loc == nil {
			break
		}
		cmdText := rest[loc[2]:loc[3]]
		tokens := shellSplit(cmdText)
		if len(tokens) > 0 {
			out = append(out, advertised{tokens: tokens[1:]})
		}
		// Continue after this "shelly" so chained invocations on the line are found.
		rest = rest[loc[2]+len("shelly"):]
	}
	return out
}

// invocationsInProse returns the invocations quoted inside prose or a comment.
func invocationsInProse(line string) []advertised {
	var out []advertised
	for _, m := range inlineSpan.FindAllStringSubmatch(line, -1) {
		if tokens := shellSplit(m[1]); len(tokens) > 0 {
			out = append(out, advertised{tokens: tokens[1:], loose: namesCommandOnly(tokens[1:])})
		}
	}
	return out
}

// namesCommandOnly reports whether tokens are bare words with no flag, value
// or placeholder: prose that names a command ("run `shelly device add`").
func namesCommandOnly(tokens []string) bool {
	for _, tok := range tokens {
		if strings.HasPrefix(tok, "-") || strings.ContainsAny(tok, "<>[]$.:/=") {
			return false
		}
	}
	return true
}

func invocationsInMarkdown(text string) []advertised {
	var out []advertised
	inFence := false
	for line := range strings.SplitSeq(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			out = append(out, invocationsInCode(line)...)
		} else {
			out = append(out, invocationsInProse(line)...)
		}
	}
	return out
}

func invocationsInScript(text string) []advertised {
	var out []advertised
	for line := range strings.SplitSeq(text, "\n") {
		out = append(out, invocationsInCode(line)...)
	}
	return out
}

// aliasFile is the import format of `shelly alias import`.
type aliasFile struct {
	Aliases map[string]string `yaml:"aliases"`
}

// collectAdvertised gathers every invocation the repository shows to users
// outside the generated command reference, plus the names of example aliases
// and plugins.
func collectAdvertised(t *testing.T) (found []advertised, userDefined map[string]bool) {
	t.Helper()
	userDefined = map[string]bool{}
	for name := range userDefinedCommands {
		userDefined[name] = true
	}

	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if advertisedSkipDirs[rel] || advertisedSkipDirs[d.Name()] && !strings.Contains(rel, "/") {
				return filepath.SkipDir
			}
			if strings.HasPrefix(rel, "examples/plugins/shelly-") && strings.Count(rel, "/") == 2 {
				userDefined[strings.TrimPrefix(d.Name(), "shelly-")] = true
			}
			return nil
		}
		if advertisedSkipFiles[rel] {
			return nil
		}
		ext := filepath.Ext(rel)
		inExamples := strings.HasPrefix(rel, "examples/")
		isScript := ext == ".sh" || (inExamples && ext == "")
		if ext != ".md" && ext != ".yaml" && ext != ".yml" && !isScript {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(data)

		var invocations []advertised
		switch {
		case ext == ".md":
			invocations = invocationsInMarkdown(text)
		case strings.HasPrefix(rel, "examples/aliases/"):
			var af aliasFile
			if yamlErr := yaml.Unmarshal(data, &af); yamlErr != nil {
				return fmt.Errorf("%s: %w", rel, yamlErr)
			}
			for name, command := range af.Aliases {
				userDefined[name] = true
				if shell, ok := strings.CutPrefix(command, "!"); ok {
					invocations = append(invocations, invocationsInScript(shell)...)
				} else {
					invocations = append(invocations, advertised{tokens: shellSplit(command), loose: true})
				}
			}
			for line := range strings.SplitSeq(text, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					invocations = append(invocations, invocationsInProse(line)...)
				}
			}
		case ext == ".yaml" || ext == ".yml":
			// Other YAML only mentions commands inside comments and descriptions.
			for line := range strings.SplitSeq(text, "\n") {
				invocations = append(invocations, invocationsInProse(line)...)
			}
		default:
			invocations = invocationsInScript(text)
		}
		for _, inv := range invocations {
			inv.source = rel
			found = append(found, inv)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}
	return found, userDefined
}

// isPlaceholder reports whether a word stands for a value in a usage synopsis
// rather than being a literal argument.
func isPlaceholder(tok string) bool {
	return strings.ContainsAny(tok, "<>[]") || strings.Contains(tok, "...") ||
		strings.HasPrefix(tok, "(") || strings.HasPrefix(tok, "$")
}

// checkInvocation resolves tokens against the command tree and returns a
// description of the first thing the tree does not accept, or "".
func checkInvocation(root *cobra.Command, inv advertised, userDefined map[string]bool) string {
	tokens := inv.tokens
	cur := root
	var positional []string
	synopsis := inv.loose
	descending := true

	lookup := func(tok string) (*pflag.Flag, bool) {
		name, _, _ := strings.Cut(strings.TrimLeft(tok, "-"), "=")
		if strings.HasPrefix(tok, "--") {
			if name == "help" || name == "version" {
				return nil, true
			}
			f := cur.Flag(name)
			if f == nil {
				f = cur.InheritedFlags().Lookup(name)
			}
			return f, f != nil
		}
		short := name[:1]
		if short == "h" {
			return nil, true
		}
		f := cur.Flags().ShorthandLookup(short)
		if f == nil {
			f = cur.InheritedFlags().ShorthandLookup(short)
		}
		if f != nil && len(name) > 1 && f.Value.Type() != "bool" {
			// "-n5" style: the rest of the word is the value.
			return nil, true
		}
		return f, f != nil
	}

	var flagTokens []string
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if isPlaceholder(tok) {
			synopsis = true
		}
		if tok == "--" {
			positional = append(positional, tokens[i+1:]...)
			break
		}
		isFlag := len(tok) > 1 && tok[0] == '-' && !isPlaceholder(tok) &&
			(tok[1] < '0' || tok[1] > '9')
		if !isFlag {
			if descending && !isPlaceholder(tok) {
				if sub, _, err := cur.Find([]string{tok}); err == nil && sub != cur {
					cur = sub
					continue
				}
			}
			descending = false
			positional = append(positional, tok)
			continue
		}
		flagTokens = append(flagTokens, tok)
		if f, ok := lookup(tok); ok && f != nil && f.Value.Type() != "bool" &&
			f.NoOptDefVal == "" && !strings.Contains(tok, "=") && i+1 < len(tokens) {
			i++
		}
	}

	if cur == root && len(positional) > 0 {
		if userDefined[positional[0]] || isPlaceholder(positional[0]) {
			return ""
		}
		return fmt.Sprintf("unknown command %q", positional[0])
	}
	if len(positional) > 0 && cur.HasSubCommands() && !cur.Runnable() && !isPlaceholder(positional[0]) {
		return fmt.Sprintf("%q has no subcommand %q", cur.CommandPath(), positional[0])
	}
	for _, tok := range flagTokens {
		if cur.DisableFlagParsing {
			// The command hands every word to something else (a plugin).
			break
		}
		if _, ok := lookup(tok); !ok {
			return fmt.Sprintf("%q has no flag %s", cur.CommandPath(), strings.SplitN(tok, "=", 2)[0])
		}
	}
	if !synopsis && cur.Args != nil && cur.Runnable() {
		if err := cur.Args(cur, positional); err != nil {
			return fmt.Sprintf("%q rejects these arguments: %v", cur.CommandPath(), err)
		}
	}
	return ""
}

// TestAdvertisedInvocationsExist fails when a command example, the README, a
// guide, an example script or an example alias shows a `shelly` invocation the
// command tree does not accept: a command or subcommand that does not exist, a
// flag the command does not have, or an argument count it rejects.
func TestAdvertisedInvocationsExist(t *testing.T) {
	t.Parallel()

	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	found, userDefined := collectAdvertised(t)
	walkCommands(rootCmd, func(c *cobra.Command) {
		for _, inv := range invocationsInScript(c.Example) {
			inv.source = "Example of " + c.CommandPath()
			found = append(found, inv)
		}
		for _, inv := range invocationsInScript(c.Long) {
			inv.source = "Long of " + c.CommandPath()
			found = append(found, inv)
		}
	})
	if len(found) < 500 {
		t.Fatalf("only %d advertised invocations found; the collector is not reading the repository", len(found))
	}

	seen := map[string]bool{}
	var problems []string
	for _, inv := range found {
		problem := checkInvocation(rootCmd, inv, userDefined)
		if problem == "" {
			continue
		}
		line := fmt.Sprintf("%s: shelly %s\n      %s", inv.source, strings.Join(inv.tokens, " "), problem)
		if !seen[line] {
			seen[line] = true
			problems = append(problems, line)
		}
	}
	sort.Strings(problems)
	for _, p := range problems {
		t.Error(p)
	}
}
