package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/itchyny/gojq"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil/flags"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/mock"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
	testfactory "github.com/tj-smith47/shelly-cli/internal/testutil/factory"
)

// bareDevice is a Gen2 mock device with no components, scripts, schedules or
// stored keys: every per-device list command has nothing to list on it.
const bareDevice = "bare"

// Gen1 mock devices: a relay device with one input, and a white bulb.
const (
	gen1Relay = "gen1-relay"
	gen1Bulb  = "gen1-bulb"
)

// gen1Only lists the commands that work on Gen1 devices only; their examples
// run against gen1Relay whatever device name the example shows.
var gen1Only = map[string]bool{"shelly action list": true}

// advertisedPipeline is a `shelly ... | jq '<expr>'` pipeline shown to users.
type advertisedPipeline struct {
	source string
	shelly []string
	exprs  []string
	// stdinFed marks a command that reads device names from the pipe before it.
	stdinFed bool
}

var (
	// assignedInvocation matches `var=$(shelly ...)` in a script, so a later
	// `"${var}" | jq` is checked against the command that filled the variable.
	assignedInvocation = regexp.MustCompile(`(\w+)=\$\((shelly [^|)]*)`)
	variablePiped      = regexp.MustCompile(`\$\{?(\w+)\}?"?\s*\|\s*(jq\s.*)`)
	deviceArgument     = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	goFieldName        = regexp.MustCompile(`^[A-Z]`)
)

// splitPipeline splits one command line on unquoted "|" and stops at the first
// operator that ends the pipeline.
func splitPipeline(text string) []string {
	var segments []string
	var cur strings.Builder
	var quote rune
	runes := []rune(text)
	for i, r := range runes {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			cur.WriteRune(r)
		case r == '|' && i+1 < len(runes) && runes[i+1] == '|',
			r == ';', r == '&', r == ')', r == '>', r == '\n':
			return append(segments, cur.String())
		case r == '|':
			segments = append(segments, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(segments, cur.String())
}

// jqExpression returns the filter of a `jq [flags] '<expr>'` segment.
func jqExpression(segment string) (string, bool) {
	tokens := shellSplit(strings.TrimSpace(segment))
	if len(tokens) == 0 || tokens[0] != "jq" {
		return "", false
	}
	for _, tok := range tokens[1:] {
		if !strings.HasPrefix(tok, "-") {
			return tok, true
		}
	}
	return "", false
}

// pipelinesInText returns every shelly invocation in text that is piped into
// jq, directly or through a shell variable.
func pipelinesInText(text string) []advertisedPipeline {
	// A trailing backslash continues the pipeline on the next line.
	text = strings.ReplaceAll(text, "\\\n", " ")
	var out []advertisedPipeline
	assigned := map[string][]string{}

	for line := range strings.SplitSeq(text, "\n") {
		for _, m := range assignedInvocation.FindAllStringSubmatch(line, -1) {
			if tokens := shellSplit(m[2]); len(tokens) > 1 {
				assigned[m[1]] = tokens[1:]
			}
		}
		if m := variablePiped.FindStringSubmatch(line); m != nil && assigned[m[1]] != nil {
			if expr, ok := jqExpression(splitPipeline(m[2])[0]); ok {
				out = append(out, advertisedPipeline{shelly: assigned[m[1]], exprs: []string{expr}})
			}
		}

		rest := line
		for {
			loc := commandHead.FindStringSubmatchIndex(rest)
			if loc == nil {
				break
			}
			segments := splitPipeline(rest[loc[2]:])
			tokens := shellSplit(segments[0])
			p := advertisedPipeline{stdinFed: strings.HasSuffix(strings.TrimSpace(rest[:loc[2]]), "|")}
			if len(tokens) > 1 {
				p.shelly = tokens[1:]
			}
			for _, seg := range segments[1:] {
				expr, ok := jqExpression(seg)
				if !ok {
					break
				}
				p.exprs = append(p.exprs, expr)
			}
			if len(p.shelly) > 0 && len(p.exprs) > 0 {
				out = append(out, p)
			}
			rest = rest[loc[2]+len("shelly"):]
		}
	}
	return out
}

// jqFieldReads returns the object fields a jq expression reads: `.name`,
// `.a.b`, `.[].id` and the shorthand `{name, id}`. Keys an expression creates
// (`{position: .current_pos}`) are not reads.
func jqFieldReads(expr string) ([]string, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return
			}
			switch node := v.Interface().(type) {
			case *gojq.Index:
				if node.Name != "" {
					seen[node.Name] = true
				}
			case *gojq.ObjectKeyVal:
				if node.Val == nil && node.Key != "" && !strings.HasPrefix(node.Key, "$") {
					seen[node.Key] = true
				}
			}
			visit(v.Elem())
		case reflect.Struct:
			for i := range v.NumField() {
				visit(v.Field(i))
			}
		case reflect.Slice:
			for i := range v.Len() {
				visit(v.Index(i))
			}
		default:
		}
	}
	visit(reflect.ValueOf(query))

	fields := make([]string, 0, len(seen))
	for name := range seen {
		fields = append(fields, name)
	}
	sort.Strings(fields)
	return fields, nil
}

// collectKeys records every object key in a decoded JSON value, at any depth.
func collectKeys(v any, keys map[string]bool) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			keys[k] = true
			collectKeys(child, keys)
		}
	case []any:
		for _, child := range val {
			collectKeys(child, keys)
		}
	}
}

// asYAML returns args with the -o/--output json value replaced by yaml.
func asYAML(args []string) []string {
	out := append([]string{}, args...)
	for i := 1; i < len(out); i++ {
		if out[i] == "json" && (out[i-1] == "-o" || out[i-1] == "--output") {
			out[i] = "yaml"
		}
	}
	return out
}

// decodeYAML parses a YAML document into the shapes encoding/json produces:
// a mapping whose keys YAML reads as numbers or booleans gets string keys, so
// the document compares key for key with the command's JSON.
func decodeYAML(text string) (any, error) {
	var doc any
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		return nil, err
	}
	var normalize func(v any) any
	normalize = func(v any) any {
		switch val := v.(type) {
		case map[string]any:
			for k, child := range val {
				val[k] = normalize(child)
			}
			return val
		case map[any]any:
			m := make(map[string]any, len(val))
			for k, child := range val {
				m[fmt.Sprint(k)] = normalize(child)
			}
			return m
		case []any:
			for i, child := range val {
				val[i] = normalize(child)
			}
			return val
		default:
			return v
		}
	}
	return normalize(doc), nil
}

// yamlKeyMismatch runs args with -o yaml and describes how its keys differ
// from those of jsonDoc, the same command's JSON output; "" when they match.
func (h *contractHarness) yamlKeyMismatch(args []string, jsonDoc any) string {
	stdout, err := h.run(asYAML(args))
	if err != nil {
		return fmt.Sprintf("-o yaml failed: %v", err)
	}
	doc, err := decodeYAML(stdout)
	if err != nil {
		return fmt.Sprintf("-o yaml stdout is not YAML (%v):\n%s", err, firstLines(stdout, 4))
	}
	jsonKeys, yamlKeys := map[string]bool{}, map[string]bool{}
	collectKeys(jsonDoc, jsonKeys)
	collectKeys(doc, yamlKeys)
	var missing, extra []string
	for k := range jsonKeys {
		if !yamlKeys[k] {
			missing = append(missing, k)
		}
	}
	for k := range yamlKeys {
		if !jsonKeys[k] {
			extra = append(extra, k)
		}
	}
	if len(missing)+len(extra) == 0 {
		return ""
	}
	sort.Strings(missing)
	sort.Strings(extra)
	return fmt.Sprintf("-o yaml keys differ from -o json: missing %v, extra %v", missing, extra)
}

// goNamedKeys returns the keys of v that are Go field names: a struct printed
// without json tags. Keys below a "response", "config", "status" or "params"
// key are device data passed through unchanged and are not checked.
func goNamedKeys(v any, path string, out *[]string) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if goFieldName.MatchString(k) {
				*out = append(*out, path+"."+k)
			}
			if k == "response" || k == "config" || k == "status" || k == "params" || k == "value" {
				continue
			}
			goNamedKeys(child, path+"."+k, out)
		}
	case []any:
		for _, child := range val {
			goNamedKeys(child, path+"[]", out)
		}
	}
}

// contractHarness runs commands of the real tree against a mock device server.
type contractHarness struct {
	io *testfactory.TestIO
}

// contractState is the component state every populated mock device reports.
func contractState() mock.DeviceState {
	return mock.DeviceState{
		"switch:0": map[string]any{"id": 0, "output": true, "apower": 45.2, "voltage": 230.1, "current": 0.2,
			"aenergy": map[string]any{"total": 1234.5}, "source": "button"},
		"input:0": map[string]any{"id": 0, "state": true},
		"light:0": map[string]any{"id": 0, "output": true, "brightness": 75, "apower": 8.1},
		"cover:0": map[string]any{"id": 0, "state": "open", "current_pos": 100, "apower": 0.0},
		"rgb:0": map[string]any{"id": 0, "output": true, "brightness": 80, "rgb": []any{255, 100, 50},
			"apower": 3.3},
		"rgbw:0": map[string]any{"id": 0, "output": true, "brightness": 80, "rgb": []any{255, 100, 50},
			"white": 20, "apower": 3.3},
		"pm1:0": map[string]any{"id": 0, "apower": 12.5, "voltage": 230.0, "current": 0.05},
		"em:0": map[string]any{"id": 0, "a_act_power": 10.0, "b_act_power": 11.0, "c_act_power": 12.0,
			"total_act_power": 33.0, "a_voltage": 230.0, "b_voltage": 230.0, "c_voltage": 230.0,
			"a_current": 0.1, "b_current": 0.1, "c_current": 0.1, "total_current": 0.3},
		"thermostat:0": map[string]any{"id": 0, "enable": true, "target_C": 21.5, "current_C": 20.0,
			"output": true},
		"bthomedevice:200": map[string]any{"id": 200, "rssi": -60, "battery": 90, "last_updated_ts": 1700000000},
		"boolean:200":      map[string]any{"value": true},
		"number:201":       map[string]any{"value": 21.5},
		"script:1":         map[string]any{"id": 1, "running": true},
		"schedule:1": map[string]any{"id": 1, "enable": true, "timespec": "0 0 7 * * *",
			"calls": []any{map[string]any{"method": "Switch.Set", "params": map[string]any{"id": 0, "on": true}}}},
		"webhooks": []any{
			map[string]any{"id": 1, "cid": 0, "enable": true, "event": "switch.on", "name": "on",
				"urls": []any{"http://example.invalid/on"}},
		},
		"sys": map[string]any{"uptime": 86400},
		"kvs": map[string]any{"greeting": "hello", "my_key": "value"},
	}
}

// newContractHarness points the command tree's factory at a mock server that
// serves a populated device under every name in devices, plus bareDevice.
func newContractHarness(t *testing.T, devices []string, withGen1 bool) *contractHarness {
	t.Helper()
	memFs := testfactory.SetupTestFs(t)

	fixtures := &mock.Fixtures{Version: "1", DeviceStates: map[string]mock.DeviceState{}}
	add := func(name string, gen int, model string) {
		// Devices sharing a MAC are treated as one device that changed address.
		mac := fmt.Sprintf("AA:BB:CC:DD:EE:%02X", len(fixtures.Config.Devices))
		fixtures.Config.Devices = append(fixtures.Config.Devices, mock.DeviceFixture{
			Name: name, Address: "192.0.2.10", MAC: mac, Type: model, Model: model, Generation: gen,
		})
	}
	for _, name := range devices {
		add(name, 2, "SNSW-001P16EU")
		fixtures.DeviceStates[name] = contractState()
	}
	add(bareDevice, 2, "SNSW-001P16EU")
	fixtures.DeviceStates[bareDevice] = mock.DeviceState{}
	if withGen1 {
		add(gen1Relay, 1, "SHSW-1")
		fixtures.DeviceStates[gen1Relay] = mock.DeviceState{
			"relays": []any{map[string]any{"name": "Porch", "ison": true, "btn_type": "toggle"}},
			"relay":  map[string]any{"ison": true},
			"inputs": []any{map[string]any{"input": 1}},
		}
		add(gen1Bulb, 1, "SHBDUO-1")
		fixtures.DeviceStates[gen1Bulb] = mock.DeviceState{
			"lights": []any{map[string]any{"name": "Bulb", "ison": true, "brightness": 60}},
		}
	}
	if len(devices) > 0 {
		fixtures.Config.Groups = []mock.GroupFixture{{Name: "living-room", Devices: devices[:1]}}
		fixtures.Config.Scenes = []mock.SceneFixture{{Name: "movie-night", Actions: []mock.SceneActionFixture{
			{Device: devices[0], Method: "Switch.Set", Params: map[string]any{"id": 0, "on": false}},
		}}}
		fixtures.Config.Aliases = []mock.AliasFixture{{Name: "lights", Command: "light list"}}
	}

	demo, err := mock.StartWithFixtures(fixtures)
	if err != nil {
		t.Fatalf("StartWithFixtures: %v", err)
	}
	t.Cleanup(demo.Cleanup)

	tio := testfactory.NewTestIOStreams()
	factory.SetIOStreams(tio.IOStreams)
	// A cache built by an earlier test points at that test's filesystem.
	factory.SetFileCache(nil)
	demo.InjectIntoFactory(factory)
	if len(devices) > 0 {
		seedContractFiles(t, memFs, demo)
	}

	// The real hook loads the user's config file; only the output-format
	// mapping matters here.
	saved := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return flags.ApplyGlobalOutputFormat(cmd)
	}
	t.Cleanup(func() {
		rootCmd.PersistentPreRunE = saved
		factory.SetIOStreams(nil)
		factory.SetConfigManager(nil)
		factory.SetShellyService(nil)
		factory.SetFileCache(nil)
	})
	return &contractHarness{io: tio}
}

// seedContractFiles gives the populated harness one backup file, one device
// template and one installed plugin, for the commands that list those.
func seedContractFiles(t *testing.T, memFs afero.Fs, demo *mock.Demo) {
	t.Helper()
	dir, err := config.Dir()
	if err != nil {
		t.Fatalf("config.Dir: %v", err)
	}
	files := map[string]string{
		filepath.Join(dir, "backups", "living-room.json"): `{"version": 1, "created_at": "2026-01-02T03:04:05Z",
			"device_info": {"id": "shellyplus1pm-aabbcc", "model": "SNSW-001P16EU", "ver": "1.4.4"}, "config": {"sys": {}}}`,
		filepath.Join(dir, "plugins", "shelly-demo", "shelly-demo"): "#!/bin/sh\n",
		filepath.Join(dir, "plugins", "shelly-demo", plugins.ManifestFileName): `{"schema_version": "1", "name": "demo",
			"version": "1.0.0", "installed_at": "2026-01-02T03:04:05Z", "source": {"type": "local"},
			"binary": {"name": "shelly-demo"}}`,
	}
	for path, content := range files {
		if err := memFs.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := afero.WriteFile(memFs, path, []byte(content), 0o755); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	err = demo.ConfigMgr.CreateDeviceTemplate("my-config", "", "SNSW-001P16EU", "", 2,
		map[string]any{"sys": map[string]any{"device": map[string]any{"name": "template"}}}, "")
	if err != nil {
		t.Fatalf("CreateDeviceTemplate: %v", err)
	}
}

// run executes `shelly <args>` and returns what it wrote to stdout.
func (h *contractHarness) run(args []string) (string, error) {
	h.io.Reset()
	var usage bytes.Buffer
	rootCmd.SetOut(&usage)
	rootCmd.SetErr(&usage)
	rootCmd.SetArgs(args)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// cobra keeps the context of a command's previous run, which is cancelled.
	walkCommands(rootCmd, func(c *cobra.Command) { c.SetContext(ctx) })

	err := rootCmd.ExecuteContext(ctx)

	// The tree is shared, so a flag set by this run would leak into the next.
	walkCommands(rootCmd, func(c *cobra.Command) {
		reset := func(f *pflag.Flag) {
			if !f.Changed {
				return
			}
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				var def []string
				if trimmed := strings.Trim(f.DefValue, "[]"); trimmed != "" {
					def = strings.Split(trimmed, ",")
				}
				if replaceErr := sv.Replace(def); replaceErr != nil {
					panic(replaceErr)
				}
			} else if setErr := f.Value.Set(f.DefValue); setErr != nil {
				panic(setErr)
			}
			f.Changed = false
		}
		c.Flags().VisitAll(reset)
		c.PersistentFlags().VisitAll(reset)
	})
	viper.Set("output", nil)
	return h.io.OutString(), err
}

// examplePipelines returns the jq pipelines of every command's help text and of
// every hand-written document, script and alias in the repository.
func examplePipelines(t *testing.T) []advertisedPipeline {
	t.Helper()
	var out []advertisedPipeline
	walkCommands(rootCmd, func(c *cobra.Command) {
		for _, text := range []string{c.Example, c.Long} {
			for _, p := range pipelinesInText(text) {
				p.source = "help of " + c.CommandPath()
				out = append(out, p)
			}
		}
	})
	for rel, text := range handWrittenTexts(t) {
		for _, p := range pipelinesInText(text) {
			p.source = rel
			out = append(out, p)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].source < out[j].source })
	return out
}

// handWrittenTexts returns the shell text of every hand-written document,
// script and example alias, keyed by path. Alias files are decoded first, so a
// jq filter is read as the shell receives it, without the YAML escaping.
func handWrittenTexts(t *testing.T) map[string]string {
	t.Helper()
	texts := map[string]string{}
	repo := os.DirFS(repoRoot)
	err := fs.WalkDir(repo, ".", func(rel string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			// docs/site/content/examples is copied from examples/ by
			// scripts/migrate-examples.sh; the originals are read instead.
			if advertisedSkipDirs[rel] || advertisedSkipDirs[d.Name()] && !strings.Contains(rel, "/") ||
				rel == "docs/site/content/examples" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(rel)
		isAlias := strings.HasPrefix(rel, "examples/aliases/")
		if advertisedSkipFiles[rel] || (ext != ".md" && ext != ".sh" && !isAlias) {
			return nil
		}
		data, readErr := fs.ReadFile(repo, rel)
		if readErr != nil {
			return readErr
		}
		if !isAlias {
			texts[rel] = string(data)
			return nil
		}
		var af aliasFile
		if yamlErr := yaml.Unmarshal(data, &af); yamlErr != nil {
			return fmt.Errorf("%s: %w", rel, yamlErr)
		}
		names := make([]string, 0, len(af.Aliases))
		for name := range af.Aliases {
			names = append(names, name)
		}
		sort.Strings(names)
		var b strings.Builder
		for _, name := range names {
			if shell, ok := strings.CutPrefix(af.Aliases[name], "!"); ok {
				b.WriteString(shell + "\n")
			}
		}
		texts[rel] = b.String()
		return nil
	})
	if err != nil {
		t.Fatalf("walk repository: %v", err)
	}
	return texts
}

// resolveCommand returns the command an invocation names and its positional
// arguments, or nil when the first word is not a built-in command.
func resolveCommand(tokens []string) (*cobra.Command, []string) {
	sub, rest, err := rootCmd.Find(tokens)
	if err != nil || sub == rootCmd {
		return nil, nil
	}
	var positional []string
	for i := 0; i < len(rest); i++ {
		tok := rest[i]
		if !strings.HasPrefix(tok, "-") {
			positional = append(positional, tok)
			continue
		}
		name := strings.TrimLeft(tok, "-")
		f := sub.Flags().Lookup(name)
		if f == nil {
			f = sub.InheritedFlags().Lookup(name)
		}
		if f == nil && len(name) == 1 {
			if f = sub.Flags().ShorthandLookup(name); f == nil {
				f = sub.InheritedFlags().ShorthandLookup(name)
			}
		}
		if f != nil && f.Value.Type() != "bool" && f.NoOptDefVal == "" && !strings.Contains(tok, "=") {
			i++
		}
	}
	return sub, positional
}

// onGen1Device returns args with the device argument replaced by gen1Relay
// when the command works on Gen1 devices only.
func onGen1Device(sub *cobra.Command, args []string) []string {
	if !gen1Only[sub.CommandPath()] {
		return args
	}
	out := append([]string{}, args...)
	depth := len(strings.Fields(sub.CommandPath())) - 1
	for i := depth; i < len(out); i++ {
		if deviceArgument.MatchString(out[i]) {
			out[i] = gen1Relay
			break
		}
	}
	return out
}

// jsonExamples returns every concrete invocation with `-o json` in a command's
// help text, keyed by the invocation.
func jsonExamples() map[string][]string {
	out := map[string][]string{}
	walkCommands(rootCmd, func(c *cobra.Command) {
		for _, inv := range invocationsInScript(c.Example) {
			joined := strings.Join(inv.tokens, " ")
			if !strings.Contains(joined, "-o json") && !strings.Contains(joined, "--output json") {
				continue
			}
			placeholderArg := false
			for _, tok := range inv.tokens {
				placeholderArg = placeholderArg || isPlaceholder(tok)
			}
			if !placeholderArg {
				out[joined] = inv.tokens
			}
		}
	})
	return out
}

// deviceNames returns the first positional argument of each invocation when it
// looks like a device name, so the mock can serve a device under that name.
func deviceNames(invocations [][]string) []string {
	seen := map[string]bool{}
	for _, tokens := range invocations {
		sub, positional := resolveCommand(tokens)
		if sub == nil {
			continue
		}
		for _, arg := range positional {
			if deviceArgument.MatchString(arg) && arg != bareDevice {
				seen[arg] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// jsonUncheckable lists the commands whose `-o json` examples the mock cannot
// serve, with the reason. Every other advertised example must print JSON.
var jsonUncheckable = map[string]string{
	"shelly discover":          discoverReason,
	"shelly discover ble":      discoverReason,
	"shelly discover coiot":    discoverReason,
	"shelly discover http":     discoverReason,
	"shelly discover mdns":     discoverReason,
	"shelly monitor events":    "streams events until interrupted",
	"shelly cloud devices":     "needs a Shelly Cloud login",
	"shelly cloud events":      "streams events from a Shelly Cloud login; TestEventPrinter_OneDocumentPerEvent checks its JSON and YAML documents",
	"shelly fleet health":      fleetReason,
	"shelly fleet stats":       fleetReason,
	"shelly fleet status":      fleetReason,
	"shelly provision inspect": "joins the device's access point over the host's WiFi",
}

const fleetReason = "needs Shelly integrator credentials"

const discoverReason = "scans the host's network; TestDiscoverCommands_HonourOutputFormat checks its output path"

// pipelineUncheckable lists the commands whose jq pipelines cannot be checked
// against mock output, with the reason.
var pipelineUncheckable = map[string]string{}

// emptyListUncheckable lists the list commands that cannot be run against an
// empty mock, with the reason.
var emptyListUncheckable = map[string]string{
	"shelly mcp claude list":      ophisReason,
	"shelly mcp cursor list":      ophisReason,
	"shelly mcp vscode list":      ophisReason,
	"shelly profile list":         "the device profiles are compiled in, so the list is never empty",
	"shelly script template list": "the built-in script templates are compiled in, so the list is never empty",
	"shelly zigbee list":          "the mock reports Zigbee on every device",
}

const ophisReason = "provided by the ophis library, which prints text to the process stdout and has no -o flag"

// emptyListObjects lists the list commands whose JSON document is an object
// keyed by name; with nothing to list they print an object, never a sentence.
var emptyListObjects = map[string]bool{"shelly action list": true}

// TestAdvertisedJSONExamplesPrintJSON runs every `-o json` example of every
// command against the mock and fails when stdout is not a JSON document, or
// when the document carries Go field names in place of snake_case keys.
func TestAdvertisedJSONExamplesPrintJSON(t *testing.T) {
	// Commands resolve paths under the home directory; none may reach the real one.
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	examples := jsonExamples()
	if len(examples) < 80 {
		t.Fatalf("only %d -o json examples found; the collector is not reading the command tree", len(examples))
	}
	invocations := make([][]string, 0, len(examples))
	keys := make([]string, 0, len(examples))
	for key, tokens := range examples {
		invocations = append(invocations, tokens)
		keys = append(keys, key)
	}
	sort.Strings(keys)
	h := newContractHarness(t, deviceNames(invocations), true)

	used := map[string]bool{}
	for _, key := range keys {
		tokens := examples[key]
		sub, _ := resolveCommand(tokens)
		if sub == nil {
			t.Errorf("shelly %s: not a built-in command", key)
			continue
		}
		if reason, ok := jsonUncheckable[sub.CommandPath()]; ok {
			used[sub.CommandPath()] = true
			t.Logf("not run: shelly %s (%s)", key, reason)
			continue
		}
		stdout, err := h.run(onGen1Device(sub, tokens))
		if err != nil {
			t.Errorf("shelly %s: %v", key, err)
			continue
		}
		var doc any
		if jsonErr := json.Unmarshal([]byte(stdout), &doc); jsonErr != nil {
			t.Errorf("shelly %s: stdout is not JSON (%v):\n%s", key, jsonErr, firstLines(stdout, 4))
			continue
		}
		var bad []string
		goNamedKeys(doc, "", &bad)
		if len(bad) > 0 {
			sort.Strings(bad)
			t.Errorf("shelly %s: keys are Go field names, want snake_case: %s", key, strings.Join(bad, ", "))
		}
		if msg := h.yamlKeyMismatch(onGen1Device(sub, tokens), doc); msg != "" {
			t.Errorf("shelly %s: %s", key, msg)
		}
	}
	for path := range jsonUncheckable {
		if !used[path] {
			t.Errorf("jsonUncheckable lists %q, which has no -o json example", path)
		}
	}
}

// TestAdvertisedJQPipelinesReadRealFields runs the shelly side of every
// advertised `shelly ... | jq '<expr>'` pipeline against the mock and fails
// when the expression reads a field the command's JSON does not carry, or does
// not run at all on that JSON.
func TestAdvertisedJQPipelinesReadRealFields(t *testing.T) {
	// Commands resolve paths under the home directory; none may reach the real one.
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	pipelines := examplePipelines(t)
	if len(pipelines) < 60 {
		t.Fatalf("only %d advertised jq pipelines found; the collector is not reading the repository", len(pipelines))
	}
	invocations := make([][]string, 0, len(pipelines))
	for _, p := range pipelines {
		invocations = append(invocations, p.shelly)
	}
	names := deviceNames(invocations)
	// Gen1 devices reject the RPC methods the `batch command --all` examples send.
	h := newContractHarness(t, names, false)

	used := map[string]bool{}
	seen := map[string]bool{}
	outputs := map[string]string{}
	for _, p := range pipelines {
		label := fmt.Sprintf("%s: shelly %s | jq '%s'", p.source, strings.Join(p.shelly, " "), strings.Join(p.exprs, "' | jq '"))
		if seen[label] {
			continue
		}
		seen[label] = true

		sub, _ := resolveCommand(p.shelly)
		if sub == nil {
			t.Errorf("%s\n      not a built-in command", label)
			continue
		}
		if reason, ok := pipelineUncheckable[sub.CommandPath()]; ok {
			used[sub.CommandPath()] = true
			t.Logf("not checked: %s (%s)", label, reason)
			continue
		}

		cmdline := strings.Join(p.shelly, "\x00")
		stdout, ok := outputs[cmdline]
		if !ok {
			args := onGen1Device(sub, p.shelly)
			if p.stdinFed {
				// Piped names and names given as arguments select devices the
				// same way, and arguments need no process-level stdin.
				args = append(append([]string{}, args...), names...)
			}
			var err error
			stdout, err = h.run(args)
			if err != nil {
				t.Errorf("%s\n      command failed: %v", label, err)
				continue
			}
			outputs[cmdline] = stdout
		}
		if problem := checkPipeline(stdout, p.exprs[0]); problem != "" {
			t.Errorf("%s\n      %s", label, problem)
		}
	}
	for path := range pipelineUncheckable {
		if !used[path] {
			t.Errorf("pipelineUncheckable lists %q, which no advertised pipeline uses", path)
		}
	}
}

// checkPipeline reports why expr does not work on the JSON in stdout, or "".
func checkPipeline(stdout, expr string) string {
	var doc any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		return fmt.Sprintf("stdout is not JSON (%v): %s", err, firstLines(stdout, 2))
	}
	fields, err := jqFieldReads(expr)
	if err != nil {
		return fmt.Sprintf("jq expression does not parse: %v", err)
	}
	keys := map[string]bool{}
	collectKeys(doc, keys)
	if len(fields) > 0 && len(keys) == 0 {
		return "the mock returned no items, so the fields cannot be checked"
	}
	var missing []string
	for _, f := range fields {
		if !keys[f] {
			missing = append(missing, "."+f)
		}
	}
	if len(missing) > 0 {
		have := make([]string, 0, len(keys))
		for k := range keys {
			have = append(have, k)
		}
		sort.Strings(have)
		return fmt.Sprintf("jq reads %s, the JSON has only: %s", strings.Join(missing, ", "), strings.Join(have, ", "))
	}

	query, err := gojq.Parse(expr)
	if err != nil {
		return fmt.Sprintf("jq expression does not parse: %v", err)
	}
	iter := query.Run(doc)
	for {
		v, more := iter.Next()
		if !more {
			return ""
		}
		if runErr, isErr := v.(error); isErr {
			return fmt.Sprintf("jq expression fails on the output: %v", runErr)
		}
	}
}

// TestListCommandsPrintEmptyListAsJSON runs every `list` command against a
// device with nothing to list (or an empty config for commands that take no
// device) and requires `-o json` to print exactly `[]`.
func TestListCommandsPrintEmptyListAsJSON(t *testing.T) {
	// Commands resolve paths under the home directory; none may reach the real one.
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	h := newContractHarness(t, nil, true)
	// With the bare device registered, config-wide lists would not be empty.
	var lists []*cobra.Command
	walkCommands(rootCmd, func(c *cobra.Command) {
		if c.Name() == "list" && c.Runnable() {
			lists = append(lists, c)
		}
	})
	if len(lists) < 30 {
		t.Fatalf("only %d list commands found", len(lists))
	}

	used := map[string]bool{}
	for _, c := range lists {
		path := c.CommandPath()
		if reason, ok := emptyListUncheckable[path]; ok {
			used[path] = true
			t.Logf("not run: %s (%s)", path, reason)
			continue
		}
		args := strings.Fields(strings.TrimPrefix(path, "shelly "))
		if c.Args != nil && c.Args(c, nil) != nil {
			args = onGen1Device(c, append(args, bareDevice))
		}
		args = append(args, emptyListArgs[path]...)
		args = append(args, "-o", "json")
		stdout, err := h.run(args)
		if err != nil {
			t.Errorf("shelly %s: %v", strings.Join(args, " "), err)
			continue
		}
		if emptyListObjects[path] {
			var obj map[string]any
			if jsonErr := json.Unmarshal([]byte(stdout), &obj); jsonErr != nil {
				t.Errorf("shelly %s: stdout = %q, want a JSON object", strings.Join(args, " "), firstLines(stdout, 3))
				continue
			}
			if msg := h.yamlKeyMismatch(args, obj); msg != "" {
				t.Errorf("shelly %s: %s", strings.Join(args, " "), msg)
			}
			continue
		}
		if strings.TrimSpace(stdout) != "[]" {
			t.Errorf("shelly %s: stdout = %q, want []", strings.Join(args, " "), firstLines(stdout, 3))
		}
		yamlArgs := asYAML(args)
		stdout, err = h.run(yamlArgs)
		if err != nil {
			t.Errorf("shelly %s: %v", strings.Join(yamlArgs, " "), err)
			continue
		}
		if strings.TrimSpace(stdout) != "[]" {
			t.Errorf("shelly %s: stdout = %q, want []", strings.Join(yamlArgs, " "), firstLines(stdout, 3))
		}
	}
	for path := range emptyListUncheckable {
		if !used[path] {
			t.Errorf("emptyListUncheckable lists %q, which is not a list command", path)
		}
	}
}

// emptyListArgs are the extra arguments that make a list command's result
// empty when the mock always has something to list.
var emptyListArgs = map[string][]string{
	"shelly device list": {"--platform", "no-such-platform"},
	"shelly theme list":  {"--filter", "no-such-theme"},
	// Without --values the document is {"keys": [], "rev": N}.
	"shelly kvs list": {"--values"},
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(strings.TrimSpace(s), "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func TestJQFieldReads(t *testing.T) {
	t.Parallel()

	tests := map[string][]string{
		`.[] | select(.output == true)`:                              {"output"},
		`.[] | {id, position: .current_pos}`:                         {"current_pos", "id"},
		`.[] | {id, r: .rgb.r}`:                                      {"id", "r", "rgb"},
		`.[] | "\(.name): \(.power // 0)W"`:                          {"name", "power"},
		`group_by(.event) | map({event: .[0].event, count: length})`: {"event"},
		`.[].urls[]`: {"urls"},
		`length`:     {},
	}
	for expr, want := range tests {
		got, err := jqFieldReads(expr)
		if err != nil {
			t.Errorf("jqFieldReads(%q): %v", expr, err)
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("jqFieldReads(%q) = %v, want %v", expr, got, want)
		}
	}
}

func TestPipelinesInText(t *testing.T) {
	t.Parallel()

	text := "updates=$(shelly firmware check --all -o json 2>/dev/null || printf '[]')\n" +
		"printf '%s' \"${updates}\" | jq -r '.[] | .name'\n" +
		"shelly device list -o json | jq -r '.[].name' | \\\n" +
		"  shelly batch command \"Shelly.GetStatus\" | jq '.[] | {device}'\n"
	got := pipelinesInText(text)
	want := []advertisedPipeline{
		{shelly: []string{"firmware", "check", "--all", "-o", "json"}, exprs: []string{".[] | .name"}},
		{shelly: []string{"device", "list", "-o", "json"}, exprs: []string{".[].name"}},
		{shelly: []string{"batch", "command", "Shelly.GetStatus"}, exprs: []string{".[] | {device}"}, stdinFed: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pipelinesInText = %+v, want %+v", got, want)
	}
}

// TestComponentListsWorkOnGen1 runs every component list command against Gen1
// mock devices: a relay and a bulb list their components, and a component type
// the device does not have lists as empty.
func TestComponentListsWorkOnGen1(t *testing.T) {
	// Commands resolve paths under the home directory; none may reach the real one.
	t.Setenv("HOME", "/testhome")
	rootCmdMu.Lock()
	defer rootCmdMu.Unlock()

	h := newContractHarness(t, nil, true)
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"switch", "list", gen1Relay}, `[{"id":0,"name":"Porch","output":true,"power":0}]`},
		{[]string{"input", "list", gen1Relay}, `[{"id":0,"name":"","state":true,"type":"toggle"}]`},
		{[]string{"light", "list", gen1Bulb}, `[{"brightness":60,"id":0,"name":"Bulb","output":true,"power":0}]`},
		{[]string{"light", "list", gen1Relay}, `[]`},
		{[]string{"switch", "list", gen1Bulb}, `[]`},
		{[]string{"cover", "list", gen1Relay}, `[]`},
		{[]string{"rgb", "list", gen1Relay}, `[]`},
		{[]string{"rgbw", "list", gen1Relay}, `[]`},
	}
	for _, tt := range tests {
		args := append(append([]string{}, tt.args...), "-o", "json")
		stdout, err := h.run(args)
		if err != nil {
			t.Errorf("shelly %s: %v", strings.Join(args, " "), err)
			continue
		}
		var doc any
		if jsonErr := json.Unmarshal([]byte(stdout), &doc); jsonErr != nil {
			t.Errorf("shelly %s: stdout is not JSON: %s", strings.Join(args, " "), firstLines(stdout, 3))
			continue
		}
		got, marshalErr := json.Marshal(doc)
		if marshalErr != nil {
			t.Fatalf("marshal: %v", marshalErr)
		}
		if string(got) != tt.want {
			t.Errorf("shelly %s = %s, want %s", strings.Join(args, " "), got, tt.want)
		}
		yamlArgs := asYAML(args)
		stdout, err = h.run(yamlArgs)
		if err != nil {
			t.Errorf("shelly %s: %v", strings.Join(yamlArgs, " "), err)
			continue
		}
		yamlDoc, yamlErr := decodeYAML(stdout)
		if yamlErr != nil {
			t.Errorf("shelly %s: stdout is not YAML: %s", strings.Join(yamlArgs, " "), firstLines(stdout, 3))
			continue
		}
		if got, marshalErr = json.Marshal(yamlDoc); marshalErr != nil {
			t.Fatalf("marshal: %v", marshalErr)
		}
		if string(got) != tt.want {
			t.Errorf("shelly %s = %s, want %s", strings.Join(yamlArgs, " "), got, tt.want)
		}
	}
}
