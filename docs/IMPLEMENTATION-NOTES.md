# Implementation Notes

This document contains session-specific implementation context extracted from PLAN.md. Reference this when you need historical context about completed phases.

---

## Phase 0: Architecture Refactoring

### Completion Summary
- **Phase 0.1-0.3:** Completed - IOStreams 92.3%, cmdutil 92.3%, package consolidation
- **Phase 0.4:** Completed - Context propagation fixed across 41+ files
- **Phase 0.5:** Completed - errgroup patterns, MultiWriter output
- **Phase 0.6:** Completed - DRY refactoring, cmdutil helpers
- **Phase 0.7:** Deferred to Phase 25 - IOStreams at 79.1% (interactive prompts gap)

### Key Decisions
- CLI restructured from `internal/cli/` to `internal/cmd/` with subdirectory-per-command
- All batch operations use `errgroup.SetLimit()` for concurrency
- Deleted `internal/ui/` and consolidated to `internal/iostreams/`
- Kept `internal/output/format.go` and `table.go` as separate formatters

---

## Phase 5: Configuration Commands

### Phase 5.1 - Device Configuration
**Approved nolint directives:**
- `MD5` - Required for Shelly digest authentication
- `G304` - User-provided file paths are intentional
- `G306` - 0o644 permissions for config exports are appropriate

**Service layer:** `internal/shelly/config.go`

### Phase 5.2 - Network Configuration
Created `internal/cmd/wifi/` and `internal/cmd/ethernet/` using cmdutil helpers:
- `RunDeviceStatus` for status commands
- `RunList` for scan/list commands
- `RunWithSpinner` for set commands

Added ethernet and WiFi AP service methods to `internal/shelly/config.go`.

### Phase 5.3-5.6 - Cloud, Auth, MQTT, Webhook
All completed with subcommand structure:
- `internal/cmd/cloud/` - status, enable, disable
- `internal/cmd/auth/` - status, set, disable
- `internal/cmd/mqtt/` - status, set, disable
- `internal/cmd/webhook/` - list, create, del, update

---

## Phase 6: Firmware Commands

**Service layer:** `internal/shelly/firmware.go` using shelly-go firmware package

**Features implemented:**
- Single device and fleet firmware checks
- Staged rollouts with `--staged` percentage
- Parallel updates with `--parallel`
- Beta and custom URL support
- Download firmware files locally

**Command alias:** `fw`

---

## Phase 7: Script Commands (Gen2+)

**Service layer:** `internal/shelly/script.go` using shelly-go `gen2/components/script`

**Commands:**
- `list`, `get` (code/status), `create`, `update`, `delete`
- `start`, `stop`, `eval` (execute code snippets)
- `upload`, `download` (file operations)

**Parent command alias:** `sc`

---

## Phase 8: Schedule Commands

**Service layer:** `internal/shelly/schedule.go` using shelly-go `gen2/components/schedule`

**Timespec format:** Supports cron-like syntax with `@sunrise`/`@sunset`

**Parent command alias:** `sched`

---

## Phase 9: Cloud Commands

**Service layer:** `internal/shelly/cloud.go` wrapping shelly-go cloud package

**Authentication:** Email/password with token storage in config

**Events command:** Uses gorilla/websocket for real-time cloud event streaming

---

## Phase 10: Backup & Restore

### Implementation Details (2025-12-12)
- Rewrote `internal/shelly/backup.go` to properly use shelly-go's `backup.Manager`
- Replaced weak bcrypt-only "encryption" with proper AES-256-GCM via shelly-go
- Added KVS (Key-Value Storage) backup support
- Created `internal/shelly/migrate.go` using shelly-go's `backup.Migrator`
- Updated commands to use wrapper API (`Device()` and `Encrypted()` methods)

---

## Phase 11: Monitoring Commands

### Phase 11.1 - Real-time Monitoring
**Service layer:** `internal/shelly/monitoring.go` with 8 new methods

**Commands:** `monitor status`, `monitor power`, `monitor events`, `monitor all`

### Phase 11.2 - Energy Monitoring (EM/EM1)
**shelly-go version:** Updated to v0.1.5 for EMData/EM1Data support

**Key fixes:**
- Fixed CSV URL methods to avoid unnecessary device connections
- All commands auto-detect EM vs EM1 component types

**Commands:** `energy list`, `energy status`, `energy history`, `energy export`, `energy reset`

### Phase 11.3 - Power Monitoring (PM/PM1)
**shelly-go version:** Updated to v0.1.6

**Key fixes:**
- Fixed CSV export type assertions (was using wrong types)
- Removed all `//nolint:errcheck` directives
- Replaced with proper `ios.DebugErr()` or `t.Logf()` logging

**Note:** PM/PM1 components do NOT have:
- Historical data storage (unlike EM/EM1 with EMData/EM1Data)
- Calibration method (only Cover and Thermostat have calibration)

The original plan items for `power history` and `power calibrate` were incorrect assumptions.

---

## Lessons Learned

### 1. Verify shelly-go API Before Planning
**Problem:** Plan items for `power history` and `power calibrate` were impossible.

**Solution:** Before adding ANY feature:
1. Check shelly-go components: `ls /db/appdata/shelly-go/gen2/components/`
2. Read the component file for available methods
3. Check Shelly API docs: https://shelly-api-docs.shelly.cloud/gen2/

### 2. Use Concrete Types, Never interface{}
**Problem:** export.go used `interface{}` and wrong type assertions, breaking CSV export.

**Solution:** Always use concrete types from shelly-go:
```go
// WRONG
func exportEMDataCSV(data interface{}) error

// CORRECT
func exportEMDataCSV(data *components.EMDataGetDataResult) error
```

### 3. Never Suppress Errors with nolint:errcheck
10+ `//nolint:errcheck` directives hid errors. Always use:
```go
if err := table.PrintTo(ios.Out); err != nil {
    ios.DebugErr("print table", err)
}
```

### 4. Check Existing Patterns First
Before writing command code, check:
- `internal/cmdutil/runner.go` - RunWithSpinner, RunStatus, RunList
- `internal/cmdutil/output.go` - FormatOutput, FormatTable, PrintListResult
- `internal/iostreams/` - DebugErr, Success, Info, Printf
- Similar existing commands (e.g., check energy/ before implementing power/)

### 5. Verify Completion Before Marking Complete
Phase 11.2 was marked complete but CSV export was broken.

Before marking ANY task complete:
1. Build and lint: `go build ./... && golangci-lint run ./...`
2. Run tests: `go test ./...`
3. Manually test user-facing features

### 6. Don't Run Bulk Formatting Without Approval
Running `gci` across 70+ files wasted context fixing import ordering that wasn't broken.

Only format files YOU changed, not the entire codebase.

---

## Service Layer Files

| File | Purpose | Added In |
|------|---------|----------|
| `internal/shelly/shelly.go` | Core service | Phase 0 |
| `internal/shelly/config.go` | Device/network/webhook config | Phase 5 |
| `internal/shelly/firmware.go` | Firmware operations | Phase 6 |
| `internal/shelly/script.go` | Script management | Phase 7 |
| `internal/shelly/schedule.go` | Schedule management | Phase 8 |
| `internal/shelly/cloud.go` | Cloud API wrapper | Phase 9 |
| `internal/shelly/backup.go` | Backup/restore | Phase 10 |
| `internal/shelly/migrate.go` | Device migration | Phase 10 |
| `internal/shelly/monitoring.go` | Energy/power monitoring | Phase 11 |

---

## shelly-go Version History

| Version | Phase | Changes |
|---------|-------|---------|
| v0.1.3 | Initial | Base dependency |
| v0.1.5 | 11.2 | EMData/EM1Data component support |
| v0.1.6 | 11.3 | CSV export type fixes |
