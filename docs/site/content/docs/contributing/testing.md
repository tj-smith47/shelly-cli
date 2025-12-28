---
title: "Testing Strategy"
description: "Testing approach and coverage requirements"
weight: 20
---


> **Target Coverage:** 90%+ overall, with specific targets per package.
> **Reference:** gh CLI achieves ~80%+ coverage; kubectl ~70%+; we target 90%+ as a premium CLI.

## Coverage Targets by Package

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| `internal/model/` | ~100% | 100% | Done |
| `internal/iostreams/` | 0% | 95%+ | Critical |
| `internal/cmdutil/` | 0% | 95%+ | Critical |
| `internal/config/` | ~60% | 90%+ | High |
| `internal/output/` | ~50% | 90%+ | High |
| `internal/client/` | ~5% | 85%+ | High |
| `internal/shelly/` | ~40% | 85%+ | High |
| `internal/plugins/` | ~30% | 80%+ | Medium |
| `internal/theme/` | ~50% | 80%+ | Medium |
| `internal/cmd/*/` | ~20% | 80%+ | High |
| `internal/tui/` | 0% | 75%+ | Medium |
| `pkg/api/` | 0% | 90%+ | High (public) |

## Unit Tests

### iostreams package (target: 95%+)
- Test IOStreams creation with various terminal states
- Test color detection (TTY, NO_COLOR, FORCE_COLOR)
- Test progress indicator start/stop
- Test writer variants (Out, ErrOut)
- Test MultiWriter concurrent updates

### cmdutil package (target: 95%+)
- Test Factory initialization and lazy loading
- Test RunWithSpinner with success/error/cancel
- Test RunBatch with concurrent execution
- Test PrintResult with all output formats
- Test flag helpers

### config package (target: 90%+)
- Test config loading from various sources
- Test alias CRUD operations
- Test device registry operations
- Test scene management
- Test group management

### output package (target: 90%+)
- Test JSON/YAML/Table/Text formatters
- Test template rendering
- Test table column alignment

### client package (target: 85%+)
- Mock HTTP responses
- Test connection management
- Test error handling

### command packages (target: 80%+)
- Table-driven tests for all commands
- Test flag parsing
- Test output formatting
- Test error conditions

## Integration Tests

### Mock Device Server
- Create mock Shelly device using shelly-go testutil
- Support Gen1 and Gen2 API responses
- Simulate various device states
- Support WebSocket connections for events

### Workflow Tests
- Discovery → Add → Control → Status flow
- Backup → Modify → Restore flow
- Scene create → Activate → Delete flow
- Group create → Add devices → Batch control flow

### TUI Tests
Use `charmbracelet/x/exp/teatest` for TUI testing:
- Test device list rendering
- Test keyboard navigation
- Test view transitions
- Test data refresh cycles

See: https://github.com/charmbracelet/x/tree/main/exp/teatest

Example:
```go
import (
    "testing"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/x/exp/teatest"
)

func TestDeviceListView(t *testing.T) {
    m := devicelist.New()
    tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

    // Wait for initial render
    teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
        return bytes.Contains(bts, []byte("Loading"))
    })

    // Send devices loaded message
    tm.Send(devicelist.DevicesLoadedMsg{
        Devices: []model.Device{{Name: "test", IP: "192.168.1.1"}},
    })

    // Verify table renders
    teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
        return bytes.Contains(bts, []byte("test"))
    })

    // Test keyboard navigation
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})

    // Verify quit
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
    tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
```

## E2E Tests

### CLI Invocation Tests
- Test all commands with `--help`
- Test output formats (`-o json`, `-o yaml`, `-o table`)
- Test verbose and quiet modes
- Test config file override

### Config File Tests
- Test default config path
- Test `--config` flag
- Test environment variable override

### Plugin Tests
- Test plugin discovery in PATH
- Test plugin execution with env vars
- Test plugin install/remove

### Completion Tests
- Verify bash completion script syntax
- Verify zsh completion script syntax
- Test dynamic completions

## Test Infrastructure

### `internal/testutil/` package

```go
// MockClient provides a mock Shelly device client
type MockClient struct {
    Responses map[string]any  // method -> response
    Errors    map[string]error
}

// MockServer creates an HTTP test server
func MockServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server

// NewTestIOStreams creates IOStreams for testing
func NewTestIOStreams() (*IOStreams, *bytes.Buffer, *bytes.Buffer)

// NewTestFactory creates a Factory with test dependencies
func NewTestFactory(t *testing.T) *cmdutil.Factory
```

### Test Fixtures
- Sample device responses (Gen1 and Gen2)
- Sample config files
- Sample backup files
- Sample scene definitions

### Assertion Helpers
```go
func AssertJSONEqual(t *testing.T, expected, actual string)
func AssertTableContains(t *testing.T, output string, columns []string, row []string)
func AssertCommandOutput(t *testing.T, cmd *cobra.Command, args []string, expected string)
```

### Go 1.25 Features
- Use `testing/synctest` for time-dependent tests
- Use `sync.WaitGroup.Go()` in concurrent test helpers

## CI Integration

- Run tests on every PR
- Generate coverage report
- Fail if coverage drops below threshold
- Badge showing coverage percentage
- Coverage report uploaded as artifact

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/cmdutil/...

# Run with race detection
go test -race ./...

# Run TUI tests (may require TTY)
go test ./internal/tui/...
```
