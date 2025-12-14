# Shelly CLI - Mandatory Rules

> ⛔ **STOP.** Read this ENTIRE file before writing ANY code.

This document contains mandatory rules that MUST be followed for every code change. Violations will break the codebase architecture and waste context.

---

## Pre-Flight Checklist

Before writing ANY code, verify you understand:

- [ ] **Factory Pattern** - All commands receive `*cmdutil.Factory`, no direct instantiation
- [ ] **IOStreams** - All output via `f.IOStreams()` methods, never `fmt.Println`
- [ ] **Context** - Always `cmd.Context()`, never `context.Background()`
- [ ] **Helpers** - Check `internal/cmdutil/` before reimplementing patterns
- [ ] **Reference Commands** - Study `device/reboot`, `config/get`, `energy/status`

---

## ⛔ CRITICAL RULES (Never Violate)

### 1. Factory Pattern is MANDATORY

```go
// ✅ CORRECT - Every command MUST use factory
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "mycommand <device>",
        Aliases: []string{"mc"},  // REQUIRED
        Example: `  shelly mycommand device1`,  // REQUIRED
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cmd.Context(), f, args)
        },
    }
    return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, args []string) error {
    ios := f.IOStreams()       // ✅ Use factory
    svc := f.ShellyService()   // ✅ Use factory
    browser := f.Browser()     // ✅ Use factory
}

// ❌ WRONG - Direct instantiation
func run(ctx context.Context, device string) error {
    ios := iostreams.System()        // ❌ NO
    svc := shelly.NewService()       // ❌ NO
    exec.Command("open", url)        // ❌ NO (use browser)
}
```

### 2. All Commands MUST Have Aliases and Examples

```go
cmd := &cobra.Command{
    Use:     "status <device>",
    Aliases: []string{"st", "s"},  // ✅ REQUIRED - at least one
    Short:   "Show device status",
    Example: `  # Basic usage
  shelly status kitchen

  # With output format
  shelly status kitchen -o json`,  // ✅ REQUIRED - usage examples
}
```

### 3. Context MUST Come from cmd.Context()

```go
// ✅ CORRECT
RunE: func(cmd *cobra.Command, args []string) error {
    return run(cmd.Context(), f, args)  // ✅ Use cmd.Context()
}

func run(ctx context.Context, f *cmdutil.Factory, ...) error {
    ctx, cancel := context.WithTimeout(ctx, time.Second*30)
    defer cancel()  // ✅ Always defer cancel
}

// ❌ WRONG
func run() error {
    ctx := context.Background()  // ❌ Breaks Ctrl+C handling
}
```

### 4. All Output via IOStreams Methods

```go
ios := f.IOStreams()

// ✅ CORRECT
ios.Print("message")
ios.Success("Done")
ios.Error("Failed")
ios.Warning("Warning")
ios.Info("Info")
ios.StartProgress("Working...")
ios.StopProgress()

// ❌ WRONG
fmt.Println("message")                  // ❌ NO
spinner.New()                           // ❌ NO
color.Green("success")                  // ❌ NO
```

### 5. Error Handling - Never Suppress

```go
// ✅ CORRECT - Log non-propagating errors
if err := table.PrintTo(ios.Out); err != nil {
    ios.DebugErr("print table", err)
}

// ✅ CORRECT - In tests
if err := os.Unsetenv("VAR"); err != nil {
    t.Logf("warning: %v", err)
}

// ❌ WRONG - Silent suppression
_ = err                    // ❌ NO
//nolint:errcheck          // ❌ NO (without approval)
```

---

## ⚠️ IMPORTANT RULES

### 6. Use Existing Helpers

Before implementing ANYTHING, check `internal/cmdutil/`:

```go
// Available helpers - USE THESE
cmdutil.RunWithSpinner(ctx, ios, "message", func(ctx context.Context) error { ... })
cmdutil.RunBatch(ctx, ios, targets, func(...) { ... })
cmdutil.FormatOutput(ios, format, data, tableFn)
cmdutil.AddTimeoutFlag(cmd, &timeout)
cmdutil.AddOutputFormatFlag(cmd, &format)
cmdutil.AddComponentIDFlag(cmd, &id, "Switch")
```

### 7. Import Organization (gci format)

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
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)
```

### 8. Lint and Test After Every Change

```bash
# After EVERY file change:
golangci-lint run path/to/file.go
go test ./path/to/...

# Before commit:
golangci-lint run ./...
go test ./...
```

### 9. Commit Frequently

Commit after:
- Completing each command or major feature
- Before running bulk operations (sed, gci, find -exec)
- Before any operation that modifies multiple files

### 10. Never Run Bulk Formatters Without Approval

```bash
# ❌ NEVER without explicit user approval:
gci write .
gofmt -w .
go fmt ./...
find . -exec sed ...
```

Only format files YOU changed in this session.

---

## Workflow Requirements

1. **Lint after every file change** - `golangci-lint run <file>`
2. **Tests for every file** - Write tests immediately, not as separate phase
3. **Commit frequently** - After each major section
4. **Commit at end of every phase** - Always commit when completing a PLAN.md phase before moving to the next
5. **No nolint without approval** - Ask user first
6. **No gci/gofmt without approval** - Only format your changes
7. **Use verbose logging for errors** - `ios.DebugErr()` for non-propagating errors

---

## Anti-Patterns Cheat Sheet

| ❌ NEVER | ✅ INSTEAD |
|----------|-----------|
| `iostreams.System()` | `f.IOStreams()` |
| `shelly.NewService()` | `f.ShellyService()` |
| `exec.Command("open", ...)` | `f.Browser().Browse(url)` |
| `context.Background()` | `cmd.Context()` |
| `fmt.Println()` in commands | `ios.Print()` |
| `spinner.New()` | `ios.StartProgress()` |
| `_ = err` | `ios.DebugErr("context", err)` |
| Missing `Aliases` field | At least one alias |
| Missing `Example` field | Usage examples required |

---

## Reference Implementations

Study these well-architected commands before writing new ones:

- `internal/cmd/device/reboot/` - Simple device operation
- `internal/cmd/config/get/` - Config with output formatting
- `internal/cmd/energy/status/` - Status with auto-detection
- `internal/cmd/batch/command/` - Batch operations
- `internal/cmd/scene/activate/` - Concurrent operations

---

## Verification Commands

Run these to verify compliance:

```bash
# Check for anti-patterns (should find NONE)
grep -r "iostreams.System()" internal/cmd/
grep -r "context.Background()" internal/cmd/
grep -r "fmt.Println" internal/cmd/
grep -r "spinner.New()" internal/cmd/

# Check factory usage (should be ALL commands)
grep -r "func NewCommand(f \*cmdutil.Factory)" internal/cmd/ | wc -l

# Verify imports
golangci-lint run --disable-all -E gci
```

---

## Quick Start Prompt

New sessions should use this prompt:

```
I'm continuing work on the Shelly CLI. Before starting:
1. I've read RULES.md for mandatory rules
2. I've read docs/architecture.md for patterns
3. Factory provides: IOStreams, ShellyService, Browser, Config
4. I will use cmdutil helpers instead of reimplementing
5. I will study existing commands before writing new ones

What would you like me to work on?
```
