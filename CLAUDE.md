# CLAUDE.md - Session Context

> **Read this FIRST before doing ANY work.** This is the single source of truth for all sessions.

## Project Overview

**Repository:** `github.com/tj-smith47/shelly-cli`
**Command:** `shelly`
**Goal:** Production-ready Cobra CLI for Shelly smart home devices, targeting adoption by ALLTERCO Robotics as the official Shelly CLI.

**Built on:** `shelly-go` library at `/db/appdata/shelly-go`

---

## Core Rules

### 1. Factory Pattern is MANDATORY

Every command receives `*cmdutil.Factory` - no direct instantiation.

```go
// CORRECT
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "mycommand <device>",
        Aliases: []string{"mc"},           // REQUIRED
        Example: `  shelly mycommand dev`, // REQUIRED
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cmd.Context(), f, args)
        },
    }
    return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, args []string) error {
    ios := f.IOStreams()      // Use factory
    svc := f.ShellyService()  // Use factory
    cfg, _ := f.Config()      // Use factory
    // ...
}

// WRONG - Direct instantiation
ios := iostreams.System()        // NO
svc := shelly.NewService()       // NO
exec.Command("open", url)        // NO (use browser package)
```

### 2. All Commands MUST Have Aliases and Examples

```go
cmd := &cobra.Command{
    Use:     "status <device>",
    Aliases: []string{"st", "s"},     // REQUIRED - at least one
    Example: `  # Basic usage
  shelly status kitchen

  # With output format
  shelly status kitchen -o json`,    // REQUIRED - usage examples
}
```

### 3. Context from cmd.Context(), NEVER context.Background()

```go
// CORRECT
RunE: func(cmd *cobra.Command, args []string) error {
    return run(cmd.Context(), f, args)  // Use cmd.Context()
}

func run(ctx context.Context, f *cmdutil.Factory, ...) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()  // Always defer cancel
}

// WRONG - Breaks Ctrl+C handling
ctx := context.Background()
```

### 4. All Output via IOStreams Methods

```go
ios := f.IOStreams()

// Output methods
ios.Print("message")
ios.Success("Done")
ios.Error("Failed")
ios.Warning("Warning")
ios.Info("Info")

// Progress
ios.StartProgress("Working...")
ios.StopProgress()

// NEVER use these in commands:
fmt.Println("message")     // NO
spinner.New()              // NO
color.Green("success")     // NO
```

### 5. Error Handling - Never Suppress

```go
// CORRECT - Log non-propagating errors
if err := table.PrintTo(ios.Out); err != nil {
    ios.DebugErr("print table", err)
}

// CORRECT - In tests
if err := os.Unsetenv("VAR"); err != nil {
    t.Logf("warning: %v", err)
}

// WRONG - Silent suppression
_ = err                    // NO
//nolint:errcheck          // NO (without approval)
```

---

## Available Helpers (Use These!)

Before implementing ANYTHING, check `internal/cmdutil/`:

```go
// Runners
cmdutil.RunWithSpinner(ctx, ios, "message", func(ctx context.Context) error { ... })
cmdutil.RunBatch(ctx, ios, svc, targets, concurrent, func(...) { ... })

// Output
cmdutil.PrintResult(ios, data, displayFunc)
cmdutil.PrintListResult(ios, items, displayFunc)
output.FormatOutput(w, data)  // For JSON/YAML output

// Flags
cmdutil.AddTimeoutFlag(cmd, &timeout)
cmdutil.AddOutputFormatFlag(cmd, &format)
cmdutil.AddComponentIDFlag(cmd, &id, "Switch")
```

---

## Import Organization (gci format)

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "time"

    // 2. Third-party (blank line before)
    "github.com/spf13/cobra"
    "golang.org/x/sync/errgroup"

    // 3. Internal (blank line before)
    "github.com/tj-smith47/shelly-cli/internal/cmdutil"
    "github.com/tj-smith47/shelly-cli/internal/iostreams"
)
```

---

## Anti-Patterns Cheat Sheet

| NEVER | INSTEAD |
|-------|---------|
| `iostreams.System()` | `f.IOStreams()` |
| `shelly.NewService()` | `f.ShellyService()` |
| `exec.Command("open", ...)` | `browser.New().Browse(ctx, url)` |
| `context.Background()` in commands | `cmd.Context()` |
| `fmt.Println()` in commands | `ios.Print()` |
| `spinner.New()` | `ios.StartProgress()` |
| `_ = err` | `ios.DebugErr("context", err)` |
| Missing `Aliases` field | At least one alias |
| Missing `Example` field | Usage examples required |

---

## Directory Structure

```
internal/
├── cmd/                    # ONLY command definitions
│   ├── root.go
│   ├── device/
│   │   ├── device.go       # Parent command
│   │   ├── list/           # `shelly device list`
│   │   ├── info/           # `shelly device info`
│   │   └── reboot/         # `shelly device reboot`
│   └── ...
│
├── cmdutil/                # Command utilities (NOT under cmd/)
│   ├── factory.go          # Dependency injection
│   ├── runner.go           # RunWithSpinner, RunBatch
│   ├── helpers.go          # Helper functions
│   └── flags.go            # Flag helpers
│
├── iostreams/              # I/O abstraction
│   ├── iostreams.go        # IOStreams struct
│   ├── color.go            # Color/theme handling
│   └── progress.go         # Progress indicators
│
├── shelly/                 # Business logic service layer
│   ├── shelly.go           # Core service + Gen1/Gen2 detection
│   ├── quick.go            # Quick commands (QuickOn/Off/Toggle)
│   ├── devicedata.go       # Device data collection
│   └── ...                 # Component services
│
├── browser/                # Cross-platform URL opening
├── config/                 # Configuration management
├── helpers/                # Device discovery helpers
├── model/                  # Domain models
├── output/                 # Output formatters (JSON, YAML, table)
├── plugins/                # Plugin system
├── theme/                  # Theming (bubbletint)
└── tui/                    # TUI components (BubbleTea)
```

---

## Reference Implementations

Study these before writing new commands:

- `internal/cmd/device/reboot/` - Simple device operation
- `internal/cmd/config/get/` - Config with output formatting
- `internal/cmd/energy/status/` - Status with auto-detection
- `internal/cmd/batch/command/` - Batch operations
- `internal/cmd/scene/activate/` - Concurrent operations

---

## Lessons Learned

1. **Verify shelly-go API before planning** - Check `/db/appdata/shelly-go/gen2/components/` before assuming features exist
2. **Use concrete types, never interface{}** - Wrong type assertions break at runtime
3. **Never suppress errors** - Use `ios.DebugErr()` instead of `//nolint:errcheck`
4. **Check existing patterns first** - Search cmdutil/helpers before reimplementing
5. **Verify before marking complete** - Build, lint, test, then manual verification
6. **No bulk formatting without approval** - Only format files you changed

---

## Gen1/Gen2 Auto-Detection

Gen1 support is consolidated in the service layer - commands don't need to handle generations directly:

```go
// Service layer handles both generations automatically
svc.QuickOn(ctx, device)      // Works for Gen1 and Gen2
svc.QuickOff(ctx, device)     // Works for Gen1 and Gen2
svc.QuickToggle(ctx, device)  // Works for Gen1 and Gen2

// Detection available if needed
isGen1, device, err := svc.IsGen1Device(ctx, identifier)
```

---

## Workflow Requirements

1. **Lint after every file change** - `golangci-lint run <file>`
2. **Tests with every change** - Write tests immediately, not as separate phase
3. **Commit frequently** - After each major section
4. **No nolint without approval** - Ask user first
5. **No bulk gci/gofmt** - Only format files you changed
6. **Use verbose logging for errors** - `ios.DebugErr()` for non-propagating errors

---

## Verification Commands

```bash
# Build and lint
go build ./...
golangci-lint run ./...

# Run tests
go test ./...

# Check for anti-patterns (should find NONE in commands)
grep -r "iostreams.System()" internal/cmd/
grep -r "context.Background()" internal/cmd/
grep -r "fmt.Println" internal/cmd/

# Check factory usage (should be ALL commands)
grep -r "func NewCommand(f \*cmdutil.Factory)" internal/cmd/ | wc -l
```

---

## Critical Session Rules

> These rules are NON-NEGOTIABLE.

1. **NEVER defer work without explicit user approval** - No "future improvement", "nice to have", or "optional"
2. **NEVER mark tasks complete prematurely** - Build + lint + test + manual verify FIRST
3. **NEVER change scope without approval** - If the task seems too large, ask
4. **Complete tasks fully** - Don't stop mid-task claiming context limits or time constraints

---

## File Reference

| File | Purpose |
|------|---------|
| `CLAUDE.md` | This file - session context (read first) |
| `PLAN.md` | Incomplete work only - task tracking |
| `.claude/COMPLETED.md` | Archive of verified completed work |
| `.claude/SESSION-START.md` | Quick pre-flight checklist |
| `.claude/COMPACT-INSTRUCTIONS.md` | Compaction rules for context |
| `.claude/IMPLEMENTATION-NOTES.md` | Historical implementation details |
| `docs/architecture.md` | Full architectural patterns |
| `docs/testing.md` | Testing strategy |

---

## Quick Start Checklist

Before writing ANY code:

- [ ] Read this entire file
- [ ] Factory pattern understood (IOStreams, ShellyService, Config from factory)
- [ ] IOStreams methods known (StartProgress, Success, Error, etc.)
- [ ] cmdutil helpers checked (RunWithSpinner, RunBatch, flags)
- [ ] Similar existing command found for reference
- [ ] Context from cmd.Context(), not context.Background()
