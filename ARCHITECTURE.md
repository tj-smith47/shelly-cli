# Shelly CLI Architecture

This document describes the architectural patterns and standards for the Shelly CLI codebase.

## Table of Contents

- [Foundational Patterns](#foundational-patterns)
- [Command Development Standards](#command-development-standards)
- [Factory Pattern](#factory-pattern)
- [IOStreams Usage](#iostreams-usage)
- [Error Handling](#error-handling)
- [Context Propagation](#context-propagation)
- [Import Organization](#import-organization)
- [Code Examples](#code-examples)

---

## Foundational Patterns

The Shelly CLI follows industry-standard patterns inspired by `gh` CLI and `kubectl`:

1. **Factory Pattern** - Dependency injection for testability
2. **IOStreams** - Unified I/O handling with TTY detection
3. **Context Propagation** - Signal-aware request cancellation
4. **Helper Functions** - DRY principles via `cmdutil` helpers
5. **Lazy Initialization** - Dependencies loaded only when needed

---

## Command Development Standards

### Constructor Naming

**Standard**: All command constructors must be named `NewCommand`.

```go
// ✅ Correct
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    return &cobra.Command{...}
}

// ❌ Incorrect
func NewCmd(f *cmdutil.Factory) *cobra.Command {
    return &cobra.Command{...}
}
```

**Rationale**: Consistency with Cobra conventions and better IDE autocomplete.

---

### Factory Parameter

**Required**: All command constructors MUST accept `*cmdutil.Factory` as the first parameter.

```go
// ✅ Correct - Factory-based
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    return &cobra.Command{
        Use:   "example <device>",
        Short: "Example command",
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cmd.Context(), f, args[0])
        },
    }
}

// ❌ Incorrect - No factory parameter
func NewCommand() *cobra.Command {
    return &cobra.Command{...}
}
```

**Rationale**:
- Enables dependency injection for testing
- Provides consistent access to IOStreams, Config, ShellyService, and Browser
- Prevents direct instantiation anti-pattern (`iostreams.System()`, `shelly.NewService()`)

---

### Parent-Child Command Structure

Parent commands create the factory once and pass it to all children.

```go
// Parent command
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "device",
        Short: "Device operations",
    }

    // Pass factory to all subcommands
    cmd.AddCommand(info.NewCommand(f))
    cmd.AddCommand(status.NewCommand(f))
    cmd.AddCommand(reboot.NewCommand(f))

    return cmd
}
```

**Never** create a new factory in child commands - always use the one passed from parent.

---

## Factory Pattern

### What the Factory Provides

The factory provides four core dependencies:

1. **IOStreams** - Terminal I/O with progress indicators, colors, prompts
2. **Config** - CLI configuration (devices, aliases, groups, scenes)
3. **ShellyService** - Business logic for device operations
4. **Browser** - Open device web UI in default browser

### Accessing Dependencies

```go
func run(ctx context.Context, f *cmdutil.Factory, device string) error {
    // Get dependencies from factory
    ios := f.IOStreams()
    svc := f.ShellyService()
    browser := f.Browser()

    // Use dependencies
    ios.StartProgress("Processing...")
    err := svc.DeviceReboot(ctx, device, 0)
    ios.StopProgress()

    if err != nil {
        return err
    }

    ios.Success("Device rebooted")

    // Optionally open device UI
    confirmed, _ := ios.Confirm("Open device web UI?", false)
    if confirmed {
        browser.Browse(fmt.Sprintf("http://%s", device))
    }

    return nil
}
```

### Why These Design Choices

**Q: Why doesn't the factory provide a raw Shelly HTTP client?**
**A:** Shelly clients are device-specific (require IP/hostname). The factory provides the `ShellyService` which handles device resolution from names/IPs.

**Q: Why doesn't the factory have an embedded context?**
**A:** Contexts are request-scoped (one per command execution), while the factory is application-scoped (singleton). Mixing lifetimes breaks cancellation semantics.

**Q: Why must factory be used in ALL commands?**
**A:** Consistency and testability. Direct instantiation (`iostreams.System()`) bypasses dependency injection and makes testing difficult.

---

## IOStreams Usage

### Standard Pattern

**Always** use factory IOStreams methods, **never** package-level functions.

```go
// ✅ Correct - Instance methods
func run(ctx context.Context, f *cmdutil.Factory, device string) error {
    ios := f.IOStreams()
    ios.StartProgress("Processing...")
    // ... work ...
    ios.StopProgress()
    ios.Success("Operation completed")
    return nil
}

// ❌ Incorrect - Package functions
func run(ctx context.Context, device string) error {
    spin := iostreams.NewSpinner("Processing...")
    spin.Start()
    // ... work ...
    spin.Stop()
    iostreams.Success("Operation completed") // ❌ Can't be mocked in tests
    return nil
}
```

### Progress Indicators

Use `StartProgress`/`StopProgress` instead of creating spinners directly.

```go
// ✅ Correct
ios.StartProgress("Rebooting device...")
err := svc.DeviceReboot(ctx, device, delay)
ios.StopProgress()

// ❌ Incorrect - Old pattern
spin := iostreams.NewSpinner("Rebooting device...")
spin.Start()
err := svc.DeviceReboot(ctx, device, delay)
spin.Stop()
```

### Available IOStreams Methods

- **Progress**: `StartProgress(msg)`, `StopProgress()`
- **Output**: `Printf()`, `Println()`, `Title()`, `Info()`, `Warning()`, `Error()`
- **Success/Failure**: `Success()`, `NoResults()`, `Added()`
- **Prompts**: `Confirm()`, `Prompt()`
- **Debug**: `DebugErr()`

---

## Error Handling

### Standard Pattern

Use separate declaration for readability and debugging.

```go
// ✅ Correct
err := svc.DeviceReboot(ctx, device, delay)
ios.StopProgress()
if err != nil {
    return fmt.Errorf("failed to reboot device: %w", err)
}

// ❌ Avoid - Inline pattern (except for simple parsing)
if err := svc.DeviceReboot(ctx, device, delay); err != nil {
    return fmt.Errorf("failed to reboot device: %w", err)
}
```

**Exception**: Inline error handling is acceptable for `fmt.Sscanf` and simple parsing operations.

### Error Wrapping

Always wrap errors with context using `%w` verb for error chains.

```go
return fmt.Errorf("failed to reboot device: %w", err)
```

---

## Context Propagation

### Context Flow

```
Root Command (creates signal-aware context)
    ↓
cmd.Context() passed to RunE
    ↓
run(ctx, f, args...)
    ↓
Service calls (svc.DeviceReboot(ctx, ...))
```

### Rules

1. **Root command** creates context with `signal.NotifyContext` for Ctrl+C handling
2. **All commands** use `cmd.Context()`, **never** `context.Background()`
3. **Command timeouts** wrap the passed context: `ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)`
4. **Always defer cancel()** to prevent context leaks

```go
func run(ctx context.Context, f *cmdutil.Factory, device string) error {
    // Wrap context with timeout
    ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
    defer cancel()

    svc := f.ShellyService()
    return svc.DeviceReboot(ctx, device, 0) // ✅ Context propagates
}
```

---

## Import Organization

### gci-Compliant Ordering

Imports must be organized in three groups with blank lines between:

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "strings"

    // 2. Third-party packages
    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    // 3. Internal packages
    "github.com/tj-smith47/shelly-cli/internal/cmdutil"
    "github.com/tj-smith47/shelly-cli/internal/iostreams"
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)
```

**Enforcement**: `golangci-lint` with `gci` linter enforces this automatically.

---

## Code Examples

### Complete Command Example

```go
// Package reboot provides the device reboot subcommand.
package reboot

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/tj-smith47/shelly-cli/internal/cmdutil"
    "github.com/tj-smith47/shelly-cli/internal/shelly"
)

// NewCommand creates the device reboot command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
    var delay int

    cmd := &cobra.Command{
        Use:   "reboot <device>",
        Short: "Reboot a device",
        Long:  `Reboot a Shelly device with optional delay.`,
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return run(cmd.Context(), f, args[0], delay)
        },
    }

    cmd.Flags().IntVarP(&delay, "delay", "d", 0, "Delay in seconds")

    return cmd
}

func run(ctx context.Context, f *cmdutil.Factory, device string, delay int) error {
    ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
    defer cancel()

    svc := f.ShellyService()
    ios := f.IOStreams()

    ios.StartProgress("Rebooting device...")

    err := svc.DeviceReboot(ctx, device, delay)
    ios.StopProgress()

    if err != nil {
        return fmt.Errorf("failed to reboot device: %w", err)
    }

    ios.Success("Device rebooted successfully")
    return nil
}
```

### Using cmdutil Helpers

For simple operations, use helper functions to reduce boilerplate:

```go
func run(ctx context.Context, f *cmdutil.Factory, device string, lightID int) error {
    ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout)
    defer cancel()

    ios := f.IOStreams()
    svc := f.ShellyService()

    return cmdutil.RunSimple(ctx, ios, svc, device, lightID,
        "Turning light on...",
        fmt.Sprintf("Light %d turned on", lightID),
        func(ctx context.Context, svc *shelly.Service, device string, id int) error {
            return svc.LightOn(ctx, device, id)
        })
}
```

**Available Helpers**:
- `RunSimple` - Simple operations with spinner and success message
- `RunStatus` - Status commands with fetcher + display
- `RunWithSpinner` - Generic wrapper with custom logic
- `RunList` - List commands with consistent formatting
- `RunBatch` - Batch operations with errgroup

---

## Testing Guidelines (Phase 25)

While comprehensive testing is scheduled for Phase 25, the factory pattern enables easy testing:

```go
func TestRebootCommand(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    mockService := &mockShellyService{}

    f := cmdutil.NewFactory().
        SetIOStreams(ios).
        SetShellyService(mockService)

    cmd := NewCommand(f)
    cmd.SetArgs([]string{"device1"})

    err := cmd.Execute()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify output
    output := stdout.String()
    if !strings.Contains(output, "rebooted successfully") {
        t.Errorf("expected success message, got: %s", output)
    }
}
```

---

## Anti-Patterns to Avoid

### 1. Direct Instantiation

```go
// ❌ Never do this
func run(ctx context.Context, device string) error {
    ios := iostreams.System()  // Bypasses factory
    svc := shelly.NewService() // Bypasses factory
    // ...
}

// ✅ Always use factory
func run(ctx context.Context, f *cmdutil.Factory, device string) error {
    ios := f.IOStreams()
    svc := f.ShellyService()
    // ...
}
```

### 2. Creating Context in Commands

```go
// ❌ Never create context.Background() in commands
func run(device string) error {
    ctx := context.Background() // Breaks Ctrl+C handling
    // ...
}

// ✅ Always use passed context
func run(ctx context.Context, f *cmdutil.Factory, device string) error {
    // ctx comes from cmd.Context()
}
```

### 3. Package-Level IOStreams Calls

```go
// ❌ Avoid package functions
iostreams.Success("Done")
iostreams.Warning("Watch out")

// ✅ Use instance methods
ios := f.IOStreams()
ios.Success("Done")
ios.Warning("Watch out")
```

### 4. Manual Spinner Management

```go
// ❌ Old pattern - manual spinner
spin := iostreams.NewSpinner("Processing...")
spin.Start()
// work
spin.Stop()

// ✅ New pattern - factory IOStreams
ios := f.IOStreams()
ios.StartProgress("Processing...")
// work
ios.StopProgress()
```

---

## Migration Checklist

When creating a new command or updating an existing one:

- [ ] Command constructor named `NewCommand(f *cmdutil.Factory)`
- [ ] Factory passed to all subcommands
- [ ] Dependencies accessed via factory (`f.IOStreams()`, `f.ShellyService()`, `f.Browser()`)
- [ ] Context from `cmd.Context()`, not `context.Background()`
- [ ] Progress indicators use `ios.StartProgress()/StopProgress()`
- [ ] No package-level iostreams calls
- [ ] Imports organized in gci format (stdlib, third-party, internal)
- [ ] Errors wrapped with `%w` for context
- [ ] Helper functions used where applicable (DRY principle)

---

## Reference Implementations

**Well-architected examples to study**:
- `internal/cmd/energy/status/status.go` - Factory pattern, auto-detection logic
- `internal/cmd/backup/create/create.go` - Complex operations, multiple dependencies
- `internal/cmd/scene/activate/activate.go` - Batch operations with errgroup
- `internal/cmd/discover/ble/ble.go` - Context-aware discovery

**Helper usage examples**:
- `internal/cmd/light/on/on.go` - RunSimple helper
- `internal/cmd/light/status/status.go` - RunStatus helper
- `internal/cmd/batch/command/command.go` - RunBatch helper

---

## Questions?

For questions about architecture or patterns, see:
- `internal/cmdutil/` - Factory and helper implementations
- `internal/iostreams/` - I/O handling
- `PLAN.md` - Project roadmap and phase definitions
- This document - Architectural standards
