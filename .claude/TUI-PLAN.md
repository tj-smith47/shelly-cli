# TUI Superfile Alignment Plan - Comprehensive Checklist

> **GOAL**: Make this TUI as immaculate as superfile, while exposing ALL shelly-go capabilities.

---

## MANDATORY INSTRUCTIONS - READ BEFORE EVERY TASK

**CRITICAL: This plan MUST be executed with ZERO compromises.**

**QUALITY OVER SPEED.** Take the time to do things right. Rushing leads to bugs, incomplete features, and technical debt.

**DO NOT EVER SKIP TASKS WITHOUT EXPLICIT USER APPROVAL.** If a task seems unnecessary, ask first. Never assume something can be skipped.

The following are COMPLETELY UNACCEPTABLE and will result in rejection:
- `// TODO: implement later` comments
- `// FIXME` comments
- Stub functions that return nil or empty values
- Placeholder implementations
- "Good enough for now" shortcuts
- Partial implementations with promises to "finish later"
- Skipping error handling
- Deferring functionality to "future work"
- Marking tasks complete without verification
- `//nolint` directives without explicit user approval

**BEFORE marking ANY task complete:**
- Manual verification confirms the feature works as described
  - Describe the feature to the user and ask to confirm - tests passing is not confirmatioon.
- All edge cases are handled

**AT THE END OF EVERY PHASE:**
- Ensure all code touched this session is properly integrated
- Remove any dead code that is no longer needed
- Update imports to remove unused dependencies
- Run the convention audit: `./scripts/audit-conventions.sh` - EVERYTHING MUST PASS, whether you think you wrote it or not (you did).

The audit script runs build, lint, tests, doc generation, and checks for stubs, TODOs, deferred work, factory pattern compliance, IOStreams usage, context usage, and architecture separation of concerns.

**REMINDER on failure:** No issue is pre-existing. All code was written by Claude. Fix all errors before marking a phase complete.

**If you cannot fully implement something, STOP and ask. Do not proceed with partial work.**

---

## SESSION START INSTRUCTIONS

**Before doing ANY TUI work, ALWAYS:**

1. **Check latest debug logs**: `ls -la ~/.config/shelly/debug/ | tail -5`
2. **Read the most recent log**: `cat ~/.config/shelly/debug/<latest>/tui.log`
3. **Run the TUI with debug enabled**: `SHELLY_TUI_DEBUG=1 ./shelly tui`
4. **Document any new bugs** in the "Bugs Found" section below

The TUI may have many bugs - document them but don't necessarily fix all of them first. Keep a running bug list to revisit at the end.

### Known Styling Issues
- [x] **Header keys should be blue** - FIXED: Labels now use `colors.Highlight` (blue)
- [x] **Values should be white** - FIXED: Values now use `colors.Text` (white)

---

## Reference Material

- **Superfile repo**: `/tmp/superfile`
- **shelly-go library**: `/opt/repos/shelly-go`
- **Existing TUI**: `/opt/repos/shelly-cli/internal/tui/`
- **Previous plans**: `/root/.claude/plans/` (search for "superfile")

---

## Critical Discovery: Component Integration Status

The TUI has many complete, production-ready components. Status as of 2026-01-19:

| Component | LOC | Status | Notes |
|-----------|-----|--------|-------|
| `monitor/` | 617 | ✅ INTEGRATED | Monitor tab via views/monitor.go |
| `ble/` | 391 | ✅ INTEGRATED | Config view panel |
| `provisioning/` | 568 | ✅ INTEGRATED | Manage view panel |
| `devicelist/` | 785 | ✅ INTEGRATED | Used in app.go Dashboard |
| `devicedetail/` | 343 | ✅ INTEGRATED | Used in app.go overlay |
| `deviceinfo/` | 535 | ✅ INTEGRATED | Used in app.go panel |
| `jsonviewer/` | ~300 | ✅ INTEGRATED | app.go overlay with syntax highlighting |
| `confirm/` | 249 | ✅ INTEGRATED | Infrastructure in app.go |
| `energyhistory/` | 296 | ✅ INTEGRATED | Dashboard energy history sparklines |
| `protocols/` | ~300 | ✅ INTEGRATED | Config view panel |
| `security/` | ~300 | ✅ INTEGRATED | Config view panel |
| `smarthome/` | ~300 | ✅ INTEGRATED | Config view panel (Matter/Zigbee/LoRa) |

---

## Phase 0: Code Quality & DRY Refactoring ✓ COMPLETE

> All Phase 0 tasks evaluated and either completed or determined not viable.

---

### COMPLETED ✓

### 0.1 Charm huh Integration ✗ NOT VIABLE

> **Status**: Determined incompatible with existing architecture. Custom form/select.go provides the needed functionality.

### 0.2 Modal System Completion ✓

**Files**: `internal/tui/components/modal/model.go`, `internal/tui/components/confirm/model.go`

- [x] Implement backdrop/overlay compositing in modal `Overlay()` method
- [x] Add backdrop dimming (Faint style in dimBaseView)
- [x] Confirm modal `WithModalOverlay()` works correctly
- [x] Tests exist in `modal/model_test.go` and `confirm/model_test.go`

### 0.3 Slider Verification & Cleanup ✓

- [x] `form/slider.go` verified working
- [x] Unit tests exist in `form/form_test.go` (TestSliderNew, TestSliderNavigation, etc.)
- [x] No orphaned styles or duplicate code

### 0.4 Dynamic Content Width & Truncation Fix ✓

**Status**: Components already use `output.Truncate()` with dynamic widths.

Verified usage:
- `fleet/devices.go`: `output.Truncate(name, nameWidth)`, `output.Truncate(device.DeviceType, typeWidth)`
- `fleet/groups.go`: `output.Truncate(group.Name, max(available, 10))`
- `webhooks/model.go`: `output.Truncate(webhook.Event, eventWidth)`, `output.Truncate(webhook.URLs[0], urlWidth)`
- `virtuals/model.go`: `output.Truncate(name, nameWidth)`, `output.Truncate(*v.StrValue, textWidth)`
- `kvs/model.go`: `output.Truncate(item.Key, keyWidth)`
- `wifi/model.go`: `output.Truncate(netw.SSID, ssidWidth)`
- `schedules/list.go`: `output.Truncate(schedule.Timespec, timespecWidth)`
- `events/model.go`: `output.Truncate(e.Description, descWidth)`

### 0.5 Shared Message Types Adoption ✓

**Status**: 14 files use `messages.SaveResultMsg` and `messages.EditClosedMsg`:
- fleet/edit.go, kvs/edit.go, virtuals/edit.go, webhooks/edit.go
- inputs/edit.go, cloud/edit.go, wifi/edit.go, ble/edit.go
- security/edit.go, protocols/edit.go, system/edit.go
- views/fleet.go, views/config.go, views/automation.go

---

### ALSO COMPLETED ✓

### 0.6 Shared Modal Styles ✓

**File**: `internal/tui/components/editmodal/styles.go`

- [x] Extract common styles: Modal, Title, Label, LabelFocus, Error, Help, Selector, Overlay
- [x] Create `DefaultStyles()` factory with `WithLabelWidth()` option
- [x] Allow component-specific extensions via embedding
- [x] Write tests

### 0.7 Edit Modal Style Migration ✓

All 11 edit modals migrated to use `editmodal.Styles`:

- [x] `ble/edit.go` (3 toggles)
- [x] `cloud/edit.go` (1 toggle)
- [x] `security/edit.go` (2 passwords + strength) - with custom StrengthStyles
- [x] `kvs/edit.go` (2 text fields) - WithLabelWidth(12)
- [x] `webhooks/edit.go` (4 fields) - WithLabelWidth(8)
- [x] `inputs/edit.go` (5 fields + dropdown)
- [x] `protocols/edit.go` (8 MQTT fields)
- [x] `virtuals/edit.go` (type-dependent) - WithLabelWidth(10)
- [x] `system/edit.go` (6 fields + aliases) - WithLabelWidth(12)
- [x] `wifi/edit.go` (tabs, scan, 6+ fields) - WithLabelWidth(12)
- [x] `fleet/groups_edit.go` (name + members) - WithLabelWidth(10)

---

### SUPERSEDED

### ~~0.8 Stale Data Handling~~ → See CACHE-PLAN.md

**Superseded:** This functionality is now covered by [CACHE-PLAN.md](CACHE-PLAN.md) Phase 2 (TUI Cache Integration Foundation), which includes:
- CacheStatus component with "Updated X ago" indicator
- R keybinding for manual refresh
- Background refresh with spinner at panel bottom

---

### FUTURE EVALUATION

### bubbles/help Adoption

**Current**: Custom `help/help.go` (~439 lines)
**Potential savings**: ~300-400 lines
**Decision**: Evaluate after huh integration complete

### bubbles/list Adoption

**Current**: Custom `panel.Scroller` used by 13+ components
**Potential savings**: Consolidation, built-in filtering
**Decision**: Evaluate after Phase 0 complete (high effort)

---

### 0.A–0.E Generic Edit Modal Infrastructure

> **Decision (2026-01-27):** huh integration (0.1) was NOT VIABLE. The existing 11+ edit modals share
> ~100-150 lines of identical lifecycle/navigation/save boilerplate each, but each implements it
> independently. This creates unnecessary variance and maintenance burden. Build the shared
> infrastructure and migrate all edit modals to use it.
>
> **WiFi note:** Drop the Station/AP tab pattern in favor of a single scrollable field list.
> All modals should use scrollability for overflow, not tabs.

### 0.A Generic Edit Modal Base ✓

**Package**: `internal/tui/components/editmodal/`

- [x] Create shared base providing: ctx, svc, device, visible, cursor, saving, err, width, height, styles
- [x] Standardize on `cursor` (not `field`) for field index
- [x] Implement common lifecycle: Show(), Hide(), SetSize(), Visible(), Saving(), Error()
- [x] Support scrollable content when fields exceed modal height
- [x] Write tests

### 0.B Generic Field Navigation ✓

**File**: `internal/tui/components/editmodal/navigation.go`

- [x] Define `Focusable` interface with Focus()/Blur() methods
- [x] Implement `NextField()` with wrapping
- [x] Implement `PrevField()` with wrapping
- [x] Implement `BlurAll()` helper
- [x] Write tests

### 0.C Generic Key Handler ✓

**File**: `internal/tui/components/editmodal/keys.go`

- [x] Create handler for Esc/Tab/Shift+Tab/Enter/Ctrl+S
- [x] Allow component-specific key handlers
- [x] Delegate unhandled keys to focused input
- [x] Write tests

### 0.D Generic Save Wrapper ✓

**File**: `internal/tui/components/editmodal/save.go`

- [x] Create save command helper with 30s timeout, error handling, result message
- [x] Write tests

### 0.E Generic renderField Helper ✓

**File**: `internal/tui/components/editmodal/render.go`

- [x] Extract `renderField(cursor, fieldIndex, label, value)` logic
- [x] Handle selector (▶) and focus/unfocus styling
- [x] Write tests

### 0.F Migrate All Edit Modals ✓

- [x] Migrate: ble/edit.go
- [x] Migrate: cloud/edit.go
- [x] Migrate: security/edit.go
- [x] Migrate: kvs/edit.go
- [x] Migrate: webhooks/edit.go
- [x] Migrate: inputs/edit.go
- [x] Migrate: protocols/edit.go
- [x] Migrate: virtuals/edit.go
- [x] Migrate: system/edit.go
- [x] Migrate: wifi/edit.go (drop tabs, single scrollable field list)
- [x] Migrate: fleet/groups_edit.go
- [x] Migrate: smarthome/ edit modals (matter, zigbee, zwave, modbus, lora)
- [x] Remove all duplicated lifecycle/navigation/save boilerplate from migrated modals
- [x] Verify all modals pass tests and lint

---

## PHASE 1: Foundation Infrastructure (Sessions 1-3) ✓ COMPLETE

> *Building blocks needed for all subsequent features*
>
> **Completed 2025-12-28** - All core components created with tests. High-visibility loading spinners wired.

### 1.1 Loading/Spinner Component
- [x] Create `internal/tui/components/loading/` package
- [x] Use `charm.land/bubbles/v2/spinner` for animated spinners
- [x] Create `Model` with configurable spinner style and message
- [x] Implement `New()`, `Init()`, `Update()`, `View()` following bubbles pattern
- [x] Support multiple spinner styles: Dot, Line, MiniDot, Pulse, etc.
- [x] Add `SetMessage(string)` for context-specific loading text
- [x] Integrate into high-visibility components:
  - [x] Dashboard device list initial load
  - [x] Monitor tab
  - [x] JSON viewer fetch
  - Config panel data loading (deferred - panels have local loading states)
- [x] Refactor remaining components to use shared loading component (7 components: discovery, firmware, batch, backup, provisioning, energybars, energyhistory)
- [x] Write tests for loading component

### 1.2 Modal/Dialog System
- [x] Create `internal/tui/components/modal/` package
- [x] Configurable width/height (percentage or fixed)
- [x] Title bar with close hint (Esc)
- [x] Content area with scroll support
- [x] Footer with action buttons/keybindings
- [x] Esc to close, Enter to confirm (configurable)
- [x] Write tests for modal component

### 1.3 Form Input Components
- [x] `internal/tui/components/form/textinput.go` - Single-line text input
- [x] `internal/tui/components/form/password.go` - Password input (masked)
- [x] `internal/tui/components/form/dropdown.go` - Select from options
- [x] `internal/tui/components/form/toggle.go` - Boolean toggle (on/off)
- [x] `internal/tui/components/form/slider.go` - Numeric range slider
- [x] `internal/tui/components/form/textarea.go` - Multi-line text input
- [x] Tab navigation between form fields
- [x] Validation support (required, format, range)
- [x] Error message display per field
- [x] Write tests for each input type

### 1.4 Toast Notifications
- [x] Enhanced `internal/tui/components/toast/` package
- [x] Success, error, warning, info variants
- [x] Auto-dismiss with configurable duration
- [x] Stack multiple toasts (most recent on top)
- [x] Position: bottom-right of screen
- [x] Animate in/out (fade or slide)
- [x] Write tests for toast component
- [x] Queue-based display: show one toast at a time with badge "(+N)" for queued count (2025-12-29)
- [x] Timer-on-show: timer starts when toast becomes visible, not when queued (2025-12-29)
- [x] Esc×2 dismiss: first Esc dismisses current, second Esc within 500ms clears all (2025-12-29)
- [x] ViewAsInputBar rendering in input bar area with colored borders (2025-12-29)

### 1.5 Error Display Patterns
- [x] Create `internal/tui/components/errorview/` package
- [x] Consistent visual indicator for RPC errors (red border)
- [x] Consistent visual indicator for HTTP errors
- [x] Show error message on hover/select
- [x] Add retry option for failed operations
- [x] Differentiate: offline vs error vs timeout
- [x] Error details in expandable section
- [x] Write tests for errorview component

### 1.6 Confirmation Dialog Improvements
- [x] Show impact of action before confirming
- [x] Destructive actions: red confirm button
- [x] Required for: factory reset, delete script/schedule/webhook, firmware rollback

---

## PHASE 2: Keys & Navigation Consolidation (Sessions 4-5)

> *Solid navigation foundation before adding feature-specific keybindings*

### 2.1 Keys Package Consolidation ✓ COMPLETE

**Goal**: Complete the context-based keybinding system (`keys/context.go`) and remove all legacy patterns.

**Completed 2026-01-19**:
- [x] Added `ActionDetail` and `ActionRefreshAll` to context.go
- [x] Wired `ActionControl` → dispatchControlAction() → showControlPanel()
- [x] Wired `ActionDetail` → dispatchDetailAction() → showDeviceDetail()
- [x] Updated `ActionRefresh` to refresh selected device (with toast)
- [x] Added `ActionRefreshAll` for ctrl+r to refresh all devices (with toast)
- [x] Made `getCurrentKeyContext()` panel-aware: device list focused = ContextDevices
- [x] Removed hardcoded string checks (`msg.String() == "d"`, `msg.String() == "c"`)
- [x] Deleted `handleGlobalKeys()` - duplicated dispatchGlobalAction()
- [x] Deleted `handleViewSwitchKey()` - duplicated dispatchTabAction()
- [x] Deleted `handleDeviceActionKey()` - duplicated dispatchDeviceKeyAction()
- [x] Migrated help overlay to use ContextHelp via contextMap.Match()
- [x] Removed `keys KeyMap` field from Model struct
- [x] Deleted legacy `keys.go` file (KeyMap, DefaultKeyMap, KeyMapFromConfig)
- [x] Removed unused `charm.land/bubbles/v2/key` import

**Architecture**: All key presses now flow through:
1. `handleKeyPressMsg()` - entry point with overlay handling
2. `handleKeyPress()` - calls `contextMap.Match()` + `dispatchAction()`
3. `dispatchAction()` → `dispatchGlobalAction()`, `dispatchTabAction()`, `dispatchPanelAction()`, `dispatchDeviceKeyAction()`

### 2.1.1 Component-Level Key Migration ✓ COMPLETE

**Goal**: Migrate all components from raw `msg.String()` checks to using action messages from the context system.

**Problem**: Components bypass context system:
```go
// CURRENT (bad) - component checks raw key
switch msg.String() {
case "e":
    return m.openEditModal()
}
```

**Solution**: app.go translates keys to action messages, components handle actions:
```go
// NEW (good) - app.go sends action message
case keys.ActionEdit:
    return m, messages.EditRequestMsg{}, true

// Component handles action message
case messages.EditRequestMsg:
    return m.openEditModal()
```

#### Infrastructure ✓ COMPLETE
- [x] Create `internal/tui/messages/actions.go` with action message types (45+ message types defined)
- [x] Add `dispatchComponentAction()` to app.go for component-level actions
- [x] Update views to forward action messages to active panel
- [x] Add h/l/left/right to all contexts (ContextEnergy, ContextAutomation, ContextConfig, ContextManage) - 2026-01-19

#### Automation View Components (9 files) ✓ COMPLETE
- [x] `scripts/list.go` - handles NavigationMsg + action messages (has "t" for template in handleKey - legitimate)
- [x] `scripts/editor.go` - handles NavigationMsg
- [x] `scripts/create.go` - handles NavigationMsg (2026-01-20)
- [x] `scripts/console.go` - handles NavigationMsg
- [x] `scripts/eval.go` - handles NavigationMsg (2026-01-20)
- [x] `schedules/list.go` - handles NavigationMsg + action messages
- [x] `schedules/create.go` - handles NavigationMsg (2026-01-20)
- [x] `webhooks/model.go` - handles NavigationMsg + action messages
- [x] `webhooks/edit.go` - handles NavigationMsg (2026-01-20)

#### Config View Components (16 files) ✓ COMPLETE
- [x] `wifi/model.go` - handles NavigationMsg + action messages
- [x] `wifi/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `system/model.go` - handles NavigationMsg + action messages
- [x] `system/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `cloud/model.go` - handles NavigationMsg + action messages
- [x] `cloud/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `security/model.go` - handles NavigationMsg + action messages
- [x] `security/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `ble/model.go` - handles NavigationMsg + action messages
- [x] `ble/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `inputs/model.go` - handles NavigationMsg + action messages
- [x] `inputs/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `inputs/actions.go` - handles NavigationMsg (2026-01-20)
- [x] `protocols/model.go` - handles NavigationMsg + action messages
- [x] `protocols/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `smarthome/model.go` - handles NavigationMsg (2026-01-20)

#### Manage View Components (12 files) ✓ COMPLETE
- [x] `discovery/model.go` - handles NavigationMsg + action messages
- [x] `firmware/model.go` - handles NavigationMsg + action messages
- [x] `backup/model.go` - handles NavigationMsg + action messages
- [x] `batch/model.go` - handles NavigationMsg + action messages
- [x] `provisioning/model.go` - handles NavigationMsg (2026-01-20)
- [x] `templates/list.go` - handles NavigationMsg + action messages
- [x] `scenes/list.go` - handles NavigationMsg + action messages
- [x] `scenes/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `migration/wizard.go` - handles NavigationMsg
- [x] `kvs/model.go` - handles NavigationMsg + action messages
- [x] `kvs/edit.go` - handles NavigationMsg (2026-01-20)
- [x] `virtuals/model.go` - handles NavigationMsg + action messages (NavLeft/NavRight for value adjustment)
- [x] `virtuals/edit.go` - handles NavigationMsg (2026-01-20)

#### Fleet View Components (5 files) ✓ COMPLETE
- [x] `fleet/devices.go` - handles NavigationMsg + action messages (2026-01-20)
- [x] `fleet/groups.go` - handles NavigationMsg + action messages (o/f/t for group on/off/toggle - legitimate component-specific keys)
- [x] `fleet/health.go` - handles NavigationMsg + action messages (2026-01-20)
- [x] `fleet/operations.go` - handles NavigationMsg + action messages (2026-01-20)
- [x] `fleet/edit.go` - handles NavigationMsg (2026-01-20)

#### Dashboard Components (5 files) ✓ COMPLETE
- [x] `events/model.go` - handles NavigationMsg (2026-01-20)
- [x] `monitor/model.go` - handles NavigationMsg + action messages
- [x] `jsonviewer/model.go` - handles NavigationMsg (2026-01-20)
- [x] `control/*.go` (5 files) - component-specific controls (legitimate internal key handling for sliders/buttons)
- [x] `alerts/model.go` - handles NavigationMsg + action messages (2026-01-20)
- [x] `alerts/form.go` - handles NavigationMsg (2026-01-20)

#### Edit Modal Pattern (12 files) ✓ COMPLETE
**Decision (2026-01-19)**: ALL keys must be centrally managed and typed. Edit modals MUST handle NavigationMsg for j/k navigation between fields. No exceptions - consistency across entire TUI.

**Completed 2026-01-20**: All edit modals now handle NavigationMsg for j/k field navigation.
- [x] `webhooks/edit.go` - handles NavigationMsg
- [x] `wifi/edit.go` - handles NavigationMsg
- [x] `system/edit.go` - handles NavigationMsg
- [x] `cloud/edit.go` - handles NavigationMsg
- [x] `security/edit.go` - handles NavigationMsg
- [x] `ble/edit.go` - handles NavigationMsg
- [x] `inputs/edit.go` - handles NavigationMsg
- [x] `protocols/edit.go` - handles NavigationMsg
- [x] `kvs/edit.go` - handles NavigationMsg
- [x] `virtuals/edit.go` - handles NavigationMsg
- [x] `fleet/edit.go` - handles NavigationMsg
- [x] `scenes/edit.go` - handles NavigationMsg

#### Verification ✓ COMPLETE
- [x] No `msg.String() ==` patterns remain for action keys (e/n/d/r/t/etc.) in list components
  - Note: Component-specific keys (templates "d"=diff, "D"=delete; batch "n"=none; groups o/f/t) are legitimate
- [x] All list components handle NavigationMsg for j/k/g/G/pgup/pgdown/h/l
- [x] All list components handle action messages (EditRequestMsg, NewRequestMsg, etc.)
- [x] Help overlay shows accurate bindings for each context
- [x] Dead code removed from devicelist/model.go (duplicate j/k/G/pgdown handling)
- [x] Templates fixed: "D" for delete, "d" for diff (removed unreachable DeleteRequestMsg handler)

**Completed 2026-01-20**: All components migrated. Key handling is now consistent across the entire TUI.

**Principle**: ALL keys centrally managed and typed. No raw msg.String() checks for action keys. Component-specific keys (not in context system) are documented and legitimate.

### 2.2 Unified Focus State System ✓ COMPLETE

**Completed 2026-01-20**: Unified focus state system implemented as single source of truth.

- [x] Create unified `FocusState` struct (replaces multiple enums)
  - `focus.State` in `internal/tui/focus/focus.go`
  - `focus.GlobalPanelID` enum in `internal/tui/focus/panels.go` (30+ panel IDs)
- [x] Track: `ActiveTab`, `ActivePanel`, `ViewFocused`, `OverlayStack`, `Mode`
  - Methods: `ActiveTab()`, `SetActiveTab()`, `ActivePanel()`, `SetActivePanel()`
  - Methods: `NextPanel()`, `PrevPanel()`, `JumpToPanel()`, `ReturnToDeviceList()`
  - Methods: `IsPanelFocused()`, `ViewFocused()`, `HasOverlay()`, `Mode()`
- [x] Ensure focus propagates on tab switch
- [x] Ensure focus propagates on device selection
- [x] Visual indicator for focused panel (border color change)
- [x] Visual indicator for focused item (highlight/cursor)
- [x] Per-tab panel cycling order in `initTabPanels()`
- [x] Views use `focusState *focus.State` (config, automation, manage, fleet)
- [x] Removed old per-view panel enums (`ConfigPanel`, `AutomationPanel`, etc.)
- [x] Removed old `focus.Manager` type (dashboard-only)
- [x] Removed dead code from `views/dashboard.go` (`DashboardPanel`, `FocusedPanel()`, etc.)

### 2.3 Vim-Style Navigation ✓

- [x] `h`/`l` for horizontal panel movement
- [x] `j`/`k` for vertical navigation within panel
- [x] `g`/`G` for go to first/last
- [x] `Ctrl+u`/`Ctrl+d` for half-page scroll
- [x] `PageUp`/`PageDown` for full-page scroll
- [x] Arrow keys as alternatives
- [x] Navigation defined per-context in keys package

### 2.4 Global Keybindings ✓

- [x] `q` quit (works, no confirmation needed for TUI apps)
- [x] `?` help toggle - confirmed 2025-12-29
- [x] `/` filter mode - confirmed 2025-12-29
- [x] `:` command mode - confirmed 2025-12-29
- [x] `1-6` tab switching - confirmed 2025-12-29
- [x] `Tab`/`Shift+Tab` panel cycling - fixed 2025-12-26 (returns to device list from first/last panel)
- [x] `Esc`/`Ctrl+[` close overlay/cancel - fixed 2025-12-26, Ctrl+[ added 2025-12-29 for iPad (search, cmdmode, all edit modals)
- [x] `Shift+N` panel jumping - fixed 2025-12-26 (!, @, #, etc. - BubbleTea returns shifted chars)
- [x] Content loading order - fixed 2025-12-29 (Config view now loads in Shift+N order: WiFi→System→Cloud→Security→BLE→Inputs→Protocols→SmartHome)

---

## PHASE 3: Device Info Expansion (Sessions 6-7) ✓ COMPLETE

> *Visible improvement, read-only data, informs later layout decisions*
>
> **Completed 2026-01-07** - All device info fields implemented with superfile-style dividers.

### 3.1 Network Information
- [x] IP Address (from device address)
- [x] MAC Address (from DeviceInfo)
- [x] WiFi SSID (from WiFiStatusFull - lazy fetch)
- [x] Signal strength with quality indicator:
  - Excellent: > -50 dBm (●●●●)
  - Good: -50 to -60 dBm (●●●○)
  - Fair: -60 to -70 dBm (●●○○)
  - Weak: < -70 dBm (●○○○)
- [x] AP Mode status (client count when > 0)
- [x] ~~Hostname~~ (not available in Shelly API)

### 3.2 Runtime Information
- [x] Uptime (formatted: 2d5h, 3h42m)
- [x] Last seen (relative: 12s ago)
- [x] Online/offline status indicator
- [x] RAM usage (percentage)
- [x] Flash usage (percentage)

### 3.3 Firmware Status
- [x] Current firmware version
- [x] Update available indicator (▲ version)
- [x] Generation indicator (Gen1/Gen2/Gen3/Gen4 with chip type: ESP8266/ESP32/ESP32-C3)

### 3.4 Component Summary
- [x] Component list (switches, covers, lights, PM, EM, EM1)
- [x] Component list with current state
- [x] Power reading per component where applicable

### 3.5 Section Dividers (Superfile Pattern)
- [x] Use `├─ Section ─┤` dividers between sections
- [x] Identity, Network, Runtime, Components sections
- [x] Scroll support if content exceeds panel height (via renderer)

### 3.6 Cache Updates for Extended Data
- [x] Extend `cache.DeviceData` with WiFiStatus, SysStatus
- [x] Lazy fetch on device focus (not all devices at once)
- [x] Preserve WiFi/Sys data across refresh cycles (preserveExtendedStatusFromExisting)

---

## PHASE 4: Configuration Modals (Sessions 8-10)

> *Simple forms, establishes modal pattern for later phases*

### 4.1 WiFi Configuration Modal ✓ COMPLETE
- [x] View/edit STA1 network (SSID, password, static IP option)
- [x] View/edit STA2 network (backup network)
- [x] AP mode settings (SSID, password, enable/disable)
- [x] Network scan integration (scan button, select from list)
- [x] Validation: SSID required, password length
- [x] Keybindings: `e` edit from WiFi panel, `s` scan
- [x] Toast feedback on save success/failure

### 4.2 Device Naming Modal ✓ COMPLETE
- [x] Edit device name (text input with required validation)
- [x] Edit device location (lat/lng with range validation)
- [x] Validation: name required, lat -90 to 90, lng -180 to 180
- [x] Keybindings: `e` edit from System panel
- [x] Toast feedback on save ("System settings saved")
- [x] Edit timezone (searchable dropdown with common IANA timezones)
- [x] Manage device aliases (add/remove with validation) - uses config.AddDeviceAlias/RemoveDeviceAlias

### 4.3 Auth Configuration Modal ✓ COMPLETE
- [x] Set/change device password
- [x] Enable/disable authentication
- [x] Show current auth status
- [x] Password strength indicator
- [x] Keybindings: `a` from Security panel
- [x] Toast feedback on save

### 4.4 Cloud Configuration Modal ✓ COMPLETE
- [x] Enable/disable cloud connection
- [x] Login flow (if not authenticated) - Note: Device-level cloud enable/disable; cloud account login is separate
- [x] Show connection status
- [x] Server selection (if applicable) - Server shown as read-only (set by Shelly)
- [x] Keybindings: `c` from Cloud panel, `t` toggle
- [x] Toast feedback on connect/disconnect

### 4.5 MQTT Configuration Modal ✓ COMPLETE
- [x] Server address and port
- [x] Username/password
- [x] Client ID
- [x] Topic prefix
- [x] Enable/disable
- [x] TLS settings (None, TLS no verify, TLS default CA, TLS user CA)
- [x] Keybindings: `m` from Protocols panel (when MQTT selected)
- [x] Toast feedback ("MQTT settings saved")
- Note: Test connection button not implemented (would require async MQTT client)

### 4.6 Timezone/Location Modal ✓ COMPLETE
- [x] Timezone selection (dropdown with search) - Already in System edit modal
- [x] Geolocation (lat/lng) for sunrise/sunset - Already in System edit modal
- [x] Keybindings: `z` from System panel (opens edit modal focused on timezone field)

### 4.7 KVS Modal ✓ COMPLETE
- [x] View/edit KVS key-value pairs
- [x] Create new key-value entry
- [x] Delete confirmation (Y to confirm, N/Esc to cancel)
- [x] Keybindings: `e` edit, `n` new, `d` delete from KVS panel
- [x] Toast feedback on save/delete

### 4.8 Webhook Edit Modal ✓ COMPLETE
- [x] Edit existing webhook URL/events (name, event, URLs, enable status)
- [x] View webhook details (via edit modal)
- [x] Keybindings: `e` edit from Webhooks panel
- [x] Toast feedback on save/toggle/delete

### 4.9 Virtual Component Modal ✓ COMPLETE
- [x] Create virtual component (boolean, number, text, enum)
- [x] Edit virtual component value/config
- [x] Delete confirmation
- [x] Keybindings: `e` edit, `n` new, `d` delete from Virtuals panel
- [x] Toast feedback

### 4.10 Input Configuration Modal ✓ COMPLETE
- [x] Input type selection (button, switch)
- [x] Input name configuration
- [x] Enable/disable toggle
- [x] Invert logic option
- [x] Factory reset toggle (5 toggles in 60s)
- [x] Keybindings: `e` edit from Inputs panel
- [x] Toast feedback on save ("Input settings saved")
- Note: Input button actions (short/long/double press) are a separate feature for input-to-output bindings

### 4.11 BLE Configuration Modal ✓ COMPLETE
- [x] BLE enable/disable toggle
- [x] RPC service toggle (control via Bluetooth)
- [x] Observer mode toggle (receive BLU sensor broadcasts)
- [x] Keybindings: `e` edit from BLE panel
- [x] Toast feedback ("Bluetooth settings saved")

### 4.12 Fleet Groups Modal ✓ COMPLETE
- [x] Create device group
- [x] Edit group name/members
- [x] Delete group
- [x] Keybindings: `e` edit, `n` new, `d` delete from Groups panel
- [x] Toast feedback

---

## PHASE 5: Component Controls (Sessions 11-13) ✓ COMPLETE

> *Interactive device control, depends on form infrastructure*
>
> **Completed 2026-01-20** - Control panel overlay implemented for Switch, Light, Cover, RGB, and Thermostat.

### 5.1 Switch Control ✓
- [x] Toggle button with visual state indicator (ON/OFF)
- [x] Power reading display (if PM device)
- [x] Per-switch control for multi-switch devices (shows first controllable component)
- [x] Quick toggle from device list (t key)
- [x] Keybindings: `t`/space toggle, `o` on, `O` off in control panel
- [x] Service adapter calls SwitchToggle/On/Off via component service
- [x] Error display on action failure

### 5.2 Light/Dimmer Control ✓
- [x] Brightness slider (0-100%)
- [x] On/off toggle
- [x] Keybindings: `+`/`-` brightness, `t` toggle in control panel
- [x] Service adapter calls LightBrightness/Toggle via component service

### 5.3 RGB/RGBW Control ✓
- [x] Color picker (RGB sliders for R, G, B channels) - **Component exists**
- [x] Brightness slider - **Component exists**
- [x] Preset colors (quick select: red, green, blue, yellow, cyan, magenta, white, warm) - **Component exists**
- [x] Keybindings: `1-8` presets, `+`/`-` brightness, `t` toggle, Tab to navigate sliders - **Component exists**
- [x] Service adapter calls RGBColor/RGBBrightness/RGBColorAndBrightness - **Component exists**
- [x] **Wire into app**: Add RGB parsing to `cache/parser.go` - **Completed 2026-01-20**
- [x] **Wire into app**: Add `RGBs []RGBState` to `cache.DeviceData` - **Completed 2026-01-20**
- [x] **Wire into app**: Add RGB handling to `showControlPanel()` in app.go - **Completed 2026-01-20**
- [x] **Wire into app**: Call `ShowRGB()` from control panel - **Completed 2026-01-20**

### 5.4 Cover/Roller Control ✓
- [x] Open/Close/Stop buttons
- [x] Position slider (0-100%)
- [x] Calibration trigger
- [x] Keybindings: `o` open, `c` close, `s` stop, `p`/arrow keys position
- [x] Service adapter calls CoverOpen/Close/Stop/Position/Calibrate

### 5.5 Thermostat Control ✓
- [x] Target temperature setpoint (slider with +/- buttons) - **Component exists**
- [x] Mode selector (heat/cool/auto/off) - **Component exists**
- [x] Boost mode toggle with duration - **Component exists**
- [x] Current temperature display - **Component exists**
- [x] Humidity display (if available) - **Component exists**
- [x] Valve position indicator - **Component exists**
- [x] Keybindings: `+`/`-` temp, `m` mode, `b` boost, `B` cancel boost, Tab focus - **Component exists**
- [x] **Wire into app**: Add Thermostat parsing to `cache/parser.go` - **Completed 2026-01-20**
- [x] **Wire into app**: Add `Thermostats []ThermostatState` to `cache.DeviceData` - **Completed 2026-01-20**
- [x] **Wire into app**: Add Thermostat handling to `showControlPanel()` in app.go - **Completed 2026-01-20**
- [x] **Wire into app**: Call `ShowThermostat()` from control panel - **Completed 2026-01-20**

### Control Panel Integration ✓
- [x] Control panel overlay (`c` key from device list/monitor)
- [x] Service adapter pattern for all component operations
- [x] Panel.Update/View/SetSize methods
- [x] Escape to close panel
- [x] Action result message handling

---

## PHASE 6: Automation CRUD (Sessions 14-17) ✓ COMPLETE

> *Complex editors with syntax highlighting and builders*
>
> **Completed 2026-01-20** - Scripts, Schedules, Webhooks, KVS, Input Actions, Virtuals fully implemented.

### 6.1 Script Editor ✓
- [x] List all scripts with status (running/stopped)
- [x] Create new script (name, code)
- [x] Edit existing script (full editor)
- [x] Delete script (with confirmation)
- [x] Syntax highlighting (Chroma - already integrated)
- [x] Run/stop controls
- [x] Console output viewer (live script output)
- [x] Template insertion (common patterns)
- [x] Download script to file (`d` keybinding in editor)
- [x] Upload script from file (`u` keybinding in editor) - **Completed 2026-01-20**
- [x] Keybindings: `n` new, `e` edit, `d` delete, `r` run, `s` stop, `u` upload
- [x] Eval command (execute JS snippet)

### 6.2 Schedule Builder ✓
- [x] List all schedules with status
- [x] Create new schedule
- [x] Cron expression builder UI (not raw cron string)
  - Time picker (hour:minute)
  - Day of week selector
  - Day of month selector
  - Repeat options (once, daily, weekly, monthly)
- [x] Action selector (which RPC to call)
- [x] Enable/disable toggle
- [x] Delete schedule (with confirmation)
- [x] Edit existing schedule (EditModel with form for timespec, method, params) - **Completed 2026-01-20**
- [x] Keybindings: `n` new, `e` edit, `d` delete, `t` toggle enable

### 6.3 Webhook Editor ✓
- [x] Create new webhook (`n` keybinding)
- [x] HTTP method selection - N/A (Shelly API only supports GET for webhook URLs)
- [x] Headers configuration - N/A (Shelly API doesn't support custom headers)
- [x] Payload template - N/A (Shelly API doesn't support custom payloads)
- [x] Test webhook button (`T` keybinding) - Tests URL connectivity with GET request

### 6.4 KVS Editor ✓
- [x] JSON value formatting in viewer (displays raw value)
- [x] Value type detection (string, number, JSON object) - edit modal handles various types
- [x] Import KVS from file (`I` keybinding - imports from kvs/<device>_kvs.json)
- [x] Export KVS to file (`X` keybinding - exports to kvs/<device>_kvs.json)

### 6.5 Action Configuration ✓
- [x] List input actions per input (shows action count in input list)
- [x] Configure action for each event type (`a` keybinding opens modal):
  - single_push, double_push, triple_push, long_push, btn_down, btn_up
- [x] Action types: webhook URL (creates/updates webhook for input event)
- [x] Clear action (delete in action modal or leave URL empty to delete)
- [x] Test action (`T` keybinding triggers single_push event on button inputs)
- [x] Keybindings: `e` edit input, `a` actions, `T` test

### 6.6 Virtual Component Management ✓
- [x] Get/set virtual component state from list view (inline editing)
- [x] `t` toggle boolean virtuals
- [x] `h`/`l` increment/decrement number virtuals
- [x] `e` edit for full modal editing
- [x] `n` new, `d` delete with confirmation

---

## PHASE 7: Groups, Scenes & Advanced Features (Sessions 18-19) ✓ COMPLETE

> *Higher-level operations on multiple devices*
>
> **Completed 2026-01-07** - All group/scene/template/migration/firmware features implemented.

### 7.1 Group Batch Operations ✓
- [x] Batch operations on group (all on, all off, all toggle) - `o`/`f`/`t` keys in Groups panel

### 7.2 Scene Management ✓
- [x] List all scenes - Scenes panel in Manage view
- [x] Create scene (capture current device states) - `C` key captures from batch-selected devices
- [x] Edit scene (modify saved states) - `e` key
- [x] Activate scene (apply to all devices) - `a` key
- [x] Show scene details (which devices, what states) - `v` key
- [x] Export scene to file - `x` key exports to ~/.config/shelly/scenes/
- [x] Import scene from file - `i` key imports from scenes directory
- [x] Delete scene (with confirmation) - `d` key (double-press to confirm)
- [x] Keybindings: `n` new, `e` edit, `a` activate, `v` view, `d` delete, `C` capture, `x` export, `i` import

### 7.3 Template System ✓
- [x] List available templates - Templates panel in Manage view
- [x] Create template from device config - `c` key creates from batch-selected device
- [x] Apply template to device (with diff preview) - `a` key applies to batch-selected devices
- [x] Template diff viewer (compare template vs device) - `d` key shows diff with batch-selected device
- [x] Export template to file - `x` key exports to ~/.config/shelly/templates/
- [x] Import template from file - `i` key imports from templates directory
- [x] Keybindings: `c` create from device, `a` apply, `d` diff, `D` delete, `x` export, `i` import

### 7.4 Device Migration ✓
- [x] Select source device - Migration wizard in Manage view (m key from Templates panel)
- [x] Select target device - Step 2 of wizard
- [x] Show config diff (what will change) - Step 3 of wizard
- [x] Selective migration (choose which settings) - Section checkboxes in wizard
- [x] Execute migration with confirmation - Step 4 of wizard
- [x] Validation (compatible device types) - Model matching in wizard

### 7.5 Firmware Management (Advanced) ✓
- [x] Staged rollout (percentage-based batch update) - `S` key for 25% staged rollout
- [x] Firmware rollback option - `R` key for rollback
- [x] Update progress indicator per device - ↻ indicator during update
- [x] Batch update result summary - Summary shows success/failure counts
- [x] Keybindings: `u` update, `U` update all, `S` staged, `R` rollback, `c` check, `spc` select, `a` select all
- Note: Pre-download not possible (Shelly API atomically downloads and installs)

---

## PHASE 8: Monitoring & Analytics (Sessions 20-22) ✓ COMPLETE

> *Observability layer, depends on data flowing correctly*
>
> **Completed 2026-01-08** - All monitoring and analytics components created with full functionality.

### 8.1 Energy History ✓
- [x] Historical energy charts (hourly, daily, weekly, monthly) - `energyanalytics/model.go` with TimePeriod enum
- [x] Per-device energy breakdown - DeviceHistory struct with TotalEnergy, AvgPower, PeakPower
- [x] Device comparison view - sorted device list with energy/power stats
- [x] Export energy data (CSV, JSON) - `energyanalytics/export.go`
- [x] Reset energy counters (with confirmation) - ResetCountersCmd with device selection
- [x] Cost calculation (configurable rate) - EnergyConfig in config.go with CostRate/Currency
- [x] Sparklines in device list - `sensors/sparkline.go` with RenderSparkline/RenderSparklineWithThreshold

### 8.2 Alert Configuration ✓
- [x] List configured alerts - `alerts/model.go` AlertItem with status indicators
- [x] Create alert - `alerts/form.go` AlertForm with name/device/condition/action fields
- [x] Edit alert - FormModeEdit with existing alert data population
- [x] Enable/disable alert - toggle via 'e' key
- [x] Delete alert - 'd' key with confirmation
- [x] Alert history view - AlertItem tracks LastTriggered, LastValue, IsTriggered
- [x] Snooze alert - 's' (1 hour) / 'S' (24 hours) with SnoozedUntil timestamp
- [x] Test alert - 't' key triggers TestAlert with AlertTestResultMsg
- [x] Keybindings: `n` new, `e` edit/toggle, `d` delete, `s/S` snooze, `t` test

### 8.3 CoIoT Integration ✓
- [x] Integrate CoIoT into TUI monitoring view - EventStream in `automation/events.go`
- [x] Add CoIoT connection status indicator - ConnectionType enum (None/WebSocket/Polling)
- [x] Show real-time CoIoT events in Monitor tab - via EventStream subscriptions
- [x] Fallback to HTTP polling when CoIoT unavailable - automatic in EventStream
- [x] CoIoT vs WebSocket indicator - GetConnectionInfo() / GetAllConnectionInfo()

### 8.4 Sensor Monitor View ✓
- [x] Display all sensor types via `sensors/model.go`:
  - Temperature (SensorTemperature)
  - Humidity (SensorHumidity)
  - Illuminance (SensorIlluminance)
  - Flood (SensorFlood)
  - Smoke (SensorSmoke)
  - Voltmeter (SensorVoltmeter)
  - DevicePower/Battery (SensorBattery)
- [x] Historical sparklines for sensor readings - `sensors/sparkline.go` SensorHistory
- [x] Threshold alerts visualization - RenderSparklineWithThreshold with color coding
- [x] Support multi-sensor devices - HistoryManager tracks "device:sensortype:id" keys
- [x] Per-sensor detail view - SensorReading with Device, SensorType, SensorID, Value, Unit

### 8.5 Event System Completion ✓
- [x] EventStream infrastructure (internal/shelly/automation/events.go)
- [x] Subscribe to WebSocket events from shelly-go
- [x] Monitor component integration (uses EventStream for Gen2+)
- [x] Real-time status updates via WebSocket - StatusChangeEvent, FullStatusEvent
- [x] Device online/offline events - DeviceOnlineEvent, DeviceOfflineEvent
- [x] Button press events (Input component) - NotifyEvent handling
- [x] Script output events - ScriptEvent handling
- [x] Show live event feed in Dashboard - `eventfeed/model.go` with EventItem display

---

## PHASE 9: Protocol Features (Sessions 23-24)

> *Advanced connectivity, niche but complete*

### 9.1 BLE/BTHome ✓ COMPLETE
- [x] List paired BTHome devices
- [x] Pair new BTHome device (scan, select, add)
- [x] Remove BTHome device
- [x] Show BTHome sensor readings
- [x] BLE enable/disable (done in Phase 4.11)
- [x] Keybindings: `p` pair, `r` remove

### 9.2 Matter ✓ COMPLETE
- [x] Enable/disable Matter
- [x] Show Matter status
- [x] Display pairing code (QR and numeric)
- [x] Reset Matter configuration
- [x] Keybindings: `t` toggle, `c` show code, `R` reset (shift-R to avoid conflict with refresh)

### 9.3 Zigbee
- [x] List paired Zigbee devices (EUI64, coordinator, channel, PAN ID in panel view)
- [x] Pair mode (start network steering via `p` key or edit modal)
- [x] Remove Zigbee device (leave network via `R` key with double-press confirmation)
- [x] Show Zigbee network info (extended status with EUI64, coordinator EUI64)
- [x] Keybindings: `p` pair mode, `R` leave network, `t` toggle, `e` edit modal
- [x] Zigbee edit modal (enable/disable, start steering, leave network)
- [x] Full test coverage for all Zigbee TUI features

### 9.4 Z-Wave
- [x] Inclusion mode (add device)
- [x] Exclusion mode (remove device)
- [x] Z-Wave configuration
- [x] Device info
- [x] Reset Z-Wave network
- [x] Keybindings: `e` edit (opens Z-Wave config modal with all sections; Z-Wave is utility-only with no RPC — all operations are physical button presses, so a single informational modal replaces separate `i`/`e`/`r` hotkeys)
- [x] Full test coverage for Z-Wave TUI features

### 9.5 Modbus
- [x] Enable/disable Modbus
- [x] Modbus status
- [x] Register configuration — N/A (no RPC; registers auto-exposed per component)
- [x] Keybindings: `t` toggle, `e` edit (opens Modbus config modal)
- [x] Full test coverage for Modbus TUI features

### 9.6 LoRa ✓ COMPLETE
- [x] LoRa status
- [x] LoRa configuration
- [x] Send test packet
- [x] Keybindings: `T` test send

---

## PHASE 10: Superfile Alignment (Sessions 25-26) ✓ COMPLETE

> *Apply consistent patterns across all features*

### 10.1 Renderer Factory Pattern ✓ ALREADY IMPLEMENTED

> **Evaluated 2026-01-29:** `rendering.New(w,h)` and `rendering.NewModal(w,h,title,footer)` provide
> a flexible factory pattern with fluent setters (`SetFocused`, `SetPanelIndex`, `SetBadge`, etc.).
> This is more flexible than specialized factories per panel type — a single configurable renderer
> adapts to any panel. Specialized `SidebarRenderer()` etc. would be over-engineering.

- [x] Renderer factory: `rendering.New()` + `rendering.NewModal()` with fluent API
- [x] Focus-aware borders (highlight when focused, table border when blurred)
- [x] Panel hints (`⇧1`–`⇧9`) in bottom border of unfocused panels
- [x] Embedded titles in top border (`├─ Title ─┤`)
- [x] Badge support in separate border section
- [x] Footer in bottom border
- [x] Modal renderer with minimum sizes

### 10.2 Per-Panel Keybinding Hints ✓ COMPLETE

> **Completed 2026-01-29:** Created `keys.FormatHints()` with width-adaptive hint formatting that
> drops whole hints from the end when terminal is too narrow. All ~30 panels updated to use
> `theme.StyledKeybindings(keys.FormatHints(hints, keys.FooterHintWidth(m.Width)))` for consistent
> yellow-styled footer hints.

- [x] Each focused panel shows its action hints in bottom border (e.g., `e:edit n:new d:delete`)
- [x] Adapt to terminal width (show fewer hints when narrow)
- [x] Group related keybindings (e.g., `j/k:nav`)

### 10.3 Responsive Breakpoints ✓ ALREADY IMPLEMENTED

> **Evaluated 2026-01-29:** App has `LayoutNarrow`/`LayoutStandard`/`LayoutWide` (< 80, 80–120, > 120).
> StatusBar has `TierMinimal`/`TierCompact`/`TierFull`. Height is handled dynamically by
> `layout.Column` which is better than fixed breakpoints. Superfile-style `HeightBreakA/B/C/D`
> would add complexity without clear benefit.

- [x] Width breakpoints: LayoutNarrow/LayoutStandard/LayoutWide
- [x] StatusBar tiers: TierMinimal/TierCompact/TierFull
- [x] Dynamic height handling via layout.Column (better than fixed breakpoints)

### 10.4 Panel Expansion on Focus ✓ ALREADY IMPLEMENTED

> **Evaluated 2026-01-29:** `layout.Column` with `ExpandOnFocus=true` does exactly this in
> Automation/Config views. Focused panel grows to `MaxHeight`, others shrink to `MinHeight`.

- [x] Focused panel expands via layout.Column ExpandOnFocus
- [x] Other panels shrink to MinHeight
- [x] Scroll position preserved (components track their own scroll state)

### 10.5 Layout Fixes

> **Evaluated 2026-01-29:** Need manual verification of specific bugs. JSON viewer uses centered
> overlay (`lipgloss.Place`) which is arguably better DX than a three-column slide-in (doesn't
> displace content). ThreeColumnLayout exists in calculator but overlay approach is preferred.

- [ ] Verify Manage view left column height (may be resolved)
- [ ] Verify ContentDimensions double-subtraction bug (may be resolved)
- [ ] Verify all views fill available vertical space

### 10.6 Status Bar Improvements ✓ COMPLETE

> **Evaluated 2026-01-29:** Active view is already shown by the tab bar with highlighting —
> duplicating in status bar adds clutter. Active filters are shown inline in the search bar
> when active. Operation progress is shown via toast notifications and loading spinners.
> These would be redundant.

- [x] Three-tier status bar (full/compact/minimal) - TierFull/TierCompact/TierMinimal implemented
- [x] Show component state counts by type (switches on/off, lights on/off, covers open/closed/moving) (2025-12-29)
- [x] Show clock/last refresh time - Shows version + time on right side
- [x] Active view/tab indicator - Tab bar already shows active tab with highlighting (redundant in status bar)
- [x] Active filters - Search/filter bar shown inline when active (redundant in status bar)

---

## PHASE 11: Monitor Tab Redesign (Sessions 27-29)

> *The Monitor tab's main panel is a flat, read-only list of per-device metric cards that duplicates
> Dashboard in a worse format. The bottom 40% is the same EnergyBars + EnergyHistory shared with
> Dashboard. Sensor data (temperature, humidity, flood, smoke) is either buried in a metrics line
> or completely ignored. Alerts exist but are only in Automation. Events are only on Dashboard.
> Multiple built components are orphaned/unwired. This phase redesigns Monitor as a purpose-built
> operational dashboard — the tab you leave open all day.*

### Problem Statement

The current Monitor main panel renders a scrollable list of device cards:
- Line 1: `● DeviceName  192.168.1.x  DeviceType  15:04:05  [ws]`
- Line 2: `    123.4W │ 230.1V │ 0.54A │ 50.0Hz │ 45.2kWh │ 22.3°C │ 65%`

This is strictly worse than Dashboard's device list + info panel for every use case:
- No sorting, filtering, or grouping
- No interactivity (can't control devices)
- Pure-sensor devices show "no power data" — their *only* interesting data gets suppressed
- Flood and smoke alarm data is **thrown away** despite being collected by `CollectSensorData()`
- Voltmeter readings are thrown away
- Only first temperature/humidity/illuminance reading used (multi-sensor devices lose data)
- `PMEnergyCounters.ByMinute` (per-minute power trend) is available but unused
- `EMStatus` per-phase data (A/B/C voltage, current, power factor) flattened to totals

### Target Layout

```
┌─────────────────────────────────────────────────────────────┐
│  Power: 579W  Energy: 12.4kWh  Cost: $3.10  8/10 online    │  ← Summary Bar
├──────────────────────────┬──────────────────────────────────┤
│ ├─ Power Ranking ──────┤ │ ├─ Environment ───────────────┤ │
│  1. Kitchen      342W ▲  │  ├─ Temperature ─┤              │
│  2. Office       180W ─  │  Living Room    22.3°C     ok   │
│  3. Garage        45W ▼  │  Bathroom       24.1°C     ok   │
│  4. Bedroom       12W ─  │  Basement       18.5°C     ⚠    │
│  ─ Living Room     0W    │  ├─ Humidity ────┤              │
│  ✗ Outdoor   offline     │  Living Room    65%        ok   │
│                          │  Bathroom       78%        ⚠    │
│                          │  ├─ Battery ─────┤              │
│                          │  Outdoor         32%  🔋        │
│                          │  ├─ Safety ──────┤              │
│                          │  Basement Flood  ✓ OK           │
│                          │  Kitchen Smoke   ✓ OK           │
├──────────────────────────┼──────────────────────────────────┤
│ ├─ Alerts ─────────────┤ │ ├─ Event Feed ────────────────┤ │
│  🔴 High power >500W    │  15:04:02 Kitchen switch ON      │
│     Kitchen 342W         │  15:03:58 Office light 80%       │
│  🟡 Low battery <20%    │  15:03:41 Garage cover OPEN      │
│     Outdoor sensor 12%   │  15:02:15 Basement flood OK      │
│  ✓  2 alerts OK          │  15:01:33 Office switch OFF      │
└──────────────────────────┴──────────────────────────────────┘
```

### Built Infrastructure to Wire (Not New Code)

| Component/Service | Current Status | Monitor Usage |
|---|---|---|
| `components/alerts/` (model.go, form.go) | Built, Automation tab only | Alerts panel |
| `components/events/` (model.go) | Built, Dashboard only | Event Feed panel |
| `components/control/` (switch, light, cover, rgb, thermostat) | **Built, not imported by any view** | Device control from Power Ranking |
| `components/cachestatus/` (model.go) | **Built, completely orphaned** | "Updated Xs ago" in summary bar |
| `components/errorview/` (model.go) | **Built, completely orphaned** | Error display in Power Ranking |
| `monitoring.CollectDashboardData()` | Built, only used by CLI `energy dashboard` cmd | Per-component power breakdown |
| `monitoring.CollectComparisonData()` | Built, only used by CLI `energy compare` cmd | Energy comparison overlay |
| `monitoring.GetEMDataHistory()` / `GetEM1DataHistory()` | Built, only used by CLI `energy history` cmd | Historical trend overlay |
| `model.SensorData.Flood` / `.Smoke` | Collected by `CollectSensorData()` | **Currently thrown away** → Safety section |
| `model.SensorData.Voltmeter` | Collected by `CollectSensorData()` | **Currently thrown away** → Environment section |
| `PMEnergyCounters.ByMinute` | Available from PM status | **Currently unused in TUI** → Trend indicators |
| `EMStatus` per-phase (A/B/C) | Available from EM status | **Currently flattened** → Detail overlay |

### shelly-go Capabilities to Leverage

| Capability | Source | Benefit |
|---|---|---|
| 60-day energy history | `gen2/components/emdata.go`, `em1data.go` | Historical trend overlay (`h` key on device) |
| Per-minute energy breakdown | `PMEnergyCounters.ByMinute` | Micro-trend sparklines / direction arrows in Power Ranking |
| CSV download URLs | `EMDataCSVURL()`, `EM1DataCSVURL()` | One-key energy data download per device |
| Return energy (solar) | `PMStatus.RetAEnergy` | Net consumption vs generation for solar users |
| Power factor | `EMStatus.APowerFactor` / `BPowerFactor` / `CPowerFactor` | Power quality indicator for 3-phase |
| Device chip temperature | `Sys.GetStatus()` → temperature fields | Overheating detection badge |
| RAM/Flash usage | `Sys.RAMFree`/`RAMSize`, `FSFree`/`FSSize` | Device health warnings |
| WiFi RSSI | `WiFi.GetStatus()` → rssi | Connection quality indicator |
| BTHome sensor readings | `gen2/components/bthomesensor.go` | Bluetooth sensors in Environment panel |
| Illumination level | `illuminance` component interpreted level | Descriptive labels (dark/dim/bright) |

---

### Session 27: Monitor Multi-Panel Layout & Power Ranking

#### 11.1.1 Add Monitor Panel IDs to Focus System
- [x] Panel IDs: `PanelMonitorSummary`, `PanelMonitorPowerRanking`, `PanelMonitorEnvironment`, `PanelMonitorAlerts`, `PanelMonitorEventFeed`
- [x] Register in `initTabPanels()` with cycling order
- [x] Tests

#### 11.1.2 Convert Monitor View to Multi-Panel
- [x] Rewrite `views/monitor.go` as multi-panel view (follow automation/config pattern)
- [x] 2-column layout: left (Power Ranking / Alerts), right (Environment / Event Feed)
- [x] Summary Bar fixed at top (3 lines)
- [x] `layout.Column` with `ExpandOnFocus=true`
- [x] Wire focus state, Tab/Shift+Tab, h/l, Shift+N
- [x] Update `renderMonitorLayout()` in app.go, remove EnergyBars+EnergyHistory from Monitor
- [x] Tests

#### 11.1.3 Summary Bar Component
- [x] `monitor/summary.go` — condensed single/3-line bar
- [x] Total Power, Total Energy, Est. Cost, Online/Total, peak power
- [x] Wire orphaned `cachestatus` for "Updated Xs ago"
- [x] Refreshing spinner indicator
- [x] Tests

#### 11.1.4 Power Ranking Component
- [x] `monitor/powerranking.go` — sorted by power, highest first
- [x] Row: rank, name, power, trend indicator (▲/▼/─ from `ByMinute` data)
- [x] Zero-power devices muted at bottom, offline devices red at bottom
- [x] Wire orphaned `errorview` for error display
- [x] Concurrent status fetching + real-time EventStream updates
- [x] Scrollable, selectable, footer hints
- [x] Tests

#### 11.1.5 Power Ranking Interactivity
- [x] `c` opens control panel (wire orphaned `components/control/`)
- [x] `t` quick-toggles selected device
- [x] `Enter`/`d` opens device detail overlay
- [x] `j`/`J` opens JSON viewer
- [x] `x` CSV / `X` JSON export (all monitor data)
- [x] Tests

---

### Session 28: Environment & Safety Panels

#### 11.2.1 Environment Panel Component
- [x] Create `internal/tui/components/monitor/environment.go`
- [x] Grouped sections with `├─ Section ─┤` dividers:
  - **Temperature**: All readings from all devices, sorted by device name. Color-coded: blue (<15°C), green (15-25°C), orange (25-35°C), red (>35°C). Show device name + value + unit.
  - **Humidity**: All readings, comfort range indicator: green (30-60%), yellow (20-30% or 60-80%), red (<20% or >80%).
  - **Illuminance**: All readings with descriptive label from shelly-go's interpreted level (dark/dim/twilight/light/bright) if available, otherwise raw lux.
  - **Battery**: All battery-powered devices, sorted lowest-first. Red (<20%), yellow (20-40%), green (>40%). Show external power indicator if present.
  - **Voltmeter**: All voltmeter readings with device name and voltage.
- [x] Scrollable when content exceeds panel height
- [x] Real-time updates via EventStream (sensor values arrive as StatusChangeEvents)
- [x] Handle multi-sensor devices (show ALL readings, not just first)
- [x] Handle BTHome sensors (parse from `bthomesensor:` prefix in status data)
- [x] Footer hints: `j/k` scroll
- [x] Tests for each sensor type rendering, color thresholds, sorting

#### 11.2.2 Safety Section (within Environment Panel)
- [x] **Flood sensors**: Show all with device name + status (OK / ALARM / MUTED)
- [x] **Smoke sensors**: Same format as flood
- [x] ALARM state renders with red background — must be impossible to miss
- [x] Muted state renders with yellow warning indicator
- [x] Safety section always visible at bottom of Environment panel (even if empty — shows "No safety sensors configured")
- [x] Tests for alarm states, mute states, rendering

#### 11.2.3 Sensor Data Collection Enhancement
- [x] Refactor sensor fetching to collect ALL readings per device (not just first)
- [x] Include flood and smoke alarm data (currently ignored by monitor)
- [x] Include voltmeter readings (currently ignored)
- [x] Include BTHome sensor readings (parse `bthomesensor:` prefix)
- [x] Include external power status from `DevicePowerReading.External.Present`
- [x] Store sensor data in a shared data structure accessible to both Power Ranking and Environment panels
- [x] Tests for multi-sensor collection, BTHome parsing

---

### Session 29: Alerts Panel, Event Feed, & Integration

<details>
<summary><b>SESSION PROMPT</b> (delete this block when session is complete)</summary>

**Goal**: Wire the existing alerts and events components into the Monitor tab as the bottom two panels, complete all cross-panel integration, and add advanced features (energy history overlay, device health badges).

**Context**: Sessions 27-28 built the layout, Power Ranking, and Environment panels. This session adds the Alerts panel (reusing the existing `components/alerts/` that's only wired into Automation), the Event Feed panel (reusing the existing `components/events/` that's only wired into Dashboard), and adds the finishing touches: energy history overlay, device health indicators, and export enhancements.

**Read these files first:**
- `internal/tui/components/alerts/model.go` — existing alerts component (create/edit/delete/toggle/snooze/test)
- `internal/tui/components/alerts/form.go` — alert creation form
- `internal/tui/views/automation.go` — how alerts are currently wired (panel 7)
- `internal/tui/components/events/model.go` — existing events component
- `internal/tui/app.go` — how events are currently wired into Dashboard
- `internal/shelly/monitoring/data.go` — `GetEMDataHistory()`, `GetEM1DataHistory()` for historical trends
- `internal/shelly/monitoring/calc.go` — `CalculateEMMetrics()`, `CalculateEM1Metrics()` for avg/peak power
- `internal/model/energy.go` — `DeviceEnergy` (AvgPower, PeakPower, DataPoints)

**Tasks:**

#### 11.3.1 Alerts Panel (Wire Existing Component)
- [ ] Add `components/alerts/` as the bottom-left panel in Monitor view
- [ ] The alerts component already has full CRUD (create, edit, delete, toggle, snooze, test) — reuse it directly as a panel, don't rebuild
- [ ] Ensure alerts show triggered state prominently (red, with `LastTriggered` time and `LastValue`)
- [ ] Ensure snoozed alerts show remaining snooze duration
- [ ] Keybindings: `n` new, `e` edit/toggle, `d` delete, `s`/`S` snooze, `t` test (already implemented in alerts component)
- [ ] Footer hints via `keys.FormatHints()`
- [ ] Badge in Monitor tab showing count of triggered alerts (if any)
- [ ] Tests for panel integration

#### 11.3.2 Event Feed Panel (Wire Existing Component)
- [ ] Add `components/events/` as the bottom-right panel in Monitor view
- [ ] Subscribe to the shared EventStream (same one used by Power Ranking for updates)
- [ ] Show real-time event feed: timestamp, device name, event description
- [ ] Color-coded by event type (state change = default, error = red, notification = yellow)
- [ ] Scrollable with j/k
- [ ] Footer hints
- [ ] Tests for panel integration

#### 11.3.3 Energy History Overlay
- [ ] From Power Ranking, `h` key on selected device opens an energy history overlay
- [ ] Fetch historical data via `monitoring.GetEMDataHistory()` or `GetEM1DataHistory()` (up to 60 days stored on device)
- [ ] Display as a sparkline/chart showing power over the last 24 hours (or configurable period)
- [ ] Show summary: average power, peak power, total energy consumed (use `CalculateEMMetrics()` / `CalculateEM1Metrics()`)
- [ ] Esc to close overlay
- [ ] Handle devices without historical data gracefully (PM-only devices show "No historical data — only EM/EM1 devices store history")
- [ ] Tests

#### 11.3.4 Device Health Badges
- [ ] In Power Ranking rows, show health warning badges when available:
  - 🌡️ if device chip temperature > 80°C (from `Sys.GetStatus()` temperature fields)
  - 💾 if flash usage > 90% (from `Sys.FSFree`/`FSSize`)
  - 📶 if WiFi RSSI < -75 dBm (from `WiFi.GetStatus()`)
  - ⚡ if return energy detected (solar — from `PMStatus.RetAEnergy`)
- [ ] Health data fetched lazily (on first render, cached)
- [ ] Badges shown as compact icons after the power value in the ranking row
- [ ] Tests

#### 11.3.5 3-Phase Detail View
- [ ] For devices with EM (3-phase) meters, `p` key opens a per-phase detail overlay
- [ ] Show Phase A, B, C separately: voltage, current, active power, apparent power, power factor, frequency
- [ ] Show neutral current if available
- [ ] Show total row at bottom
- [ ] Esc to close
- [ ] Tests

#### 11.3.6 Export Enhancement
- [ ] `x` CSV export now includes: all power data + all sensor data + alert states + last events
- [ ] `X` JSON export same enhancement
- [ ] Add per-device CSV download URL display for EM/EM1 devices (`d` then `x` in detail overlay, using `GetEMDataCSVURL()`)
- [ ] Tests

#### 11.3.7 Context Keybinding Updates
- [ ] Update `ContextMonitor` in `keys/context.go` with all new keybindings
- [ ] Update help overlay to show Monitor-specific keybindings per panel
- [ ] Ensure no key conflicts across panels
- [ ] Tests

#### 11.3.8 Remove Old Monitor Component
- [ ] Delete the old `monitor/model.go` flat device-card implementation (replaced by powerranking.go + environment.go + summary.go)
- [ ] Remove old monitor-specific styles that are no longer used
- [ ] Clean up any dead code paths in app.go that referenced the old layout
- [ ] Update `renderMonitorLayout()` to be purely delegation (like other multi-panel tabs)
- [ ] Verify no orphaned imports
- [ ] Tests pass, lint clean

**Verification:**
- `go build ./...`
- `golangci-lint run ./...`
- `go test ./...`
- `./scripts/audit-conventions.sh > /tmp/audit.txt 2>&1` — read and fix all issues
- Manual: Full Monitor tab walkthrough — all 4 focusable panels work, alerts show triggered state, events stream live, energy history overlay works for EM devices, 3-phase detail works, export includes all data, health badges appear
- Commit with descriptive message

**When done:** Delete this prompt block and mark all checkboxes above. Leave only the checkbox list.

</details>

#### 11.3.1 Alerts Panel (Wire Existing Component)
- [ ] Wire `components/alerts/` as bottom-left Monitor panel (reuse, don't rebuild)
- [ ] Triggered alerts prominent (red, with time + value)
- [ ] Snoozed alerts show remaining duration
- [ ] Triggered alert count badge in Monitor tab
- [ ] Tests

#### 11.3.2 Event Feed Panel (Wire Existing Component)
- [ ] Wire `components/events/` as bottom-right Monitor panel
- [ ] Subscribe to shared EventStream
- [ ] Color-coded by event type, scrollable
- [ ] Tests

#### 11.3.3 Energy History Overlay
- [ ] `h` key on Power Ranking device → historical energy overlay
- [ ] Fetch via `GetEMDataHistory()` / `GetEM1DataHistory()` (60-day on-device storage)
- [ ] Sparkline/chart for last 24h, summary stats (avg/peak power, total energy)
- [ ] Graceful fallback for PM-only devices
- [ ] Tests

#### 11.3.4 Device Health Badges
- [ ] Chip temperature warning (>80°C), flash usage (>90%), WiFi RSSI (<-75dBm), solar return energy
- [ ] Lazy fetch, cached, compact badge icons in Power Ranking rows
- [ ] Tests

#### 11.3.5 3-Phase Detail View
- [ ] `p` key on EM device → per-phase overlay (A/B/C voltage, current, power, PF, frequency, neutral current)
- [ ] Tests

#### 11.3.6 Export Enhancement
- [ ] CSV/JSON exports include all panel data (power + sensors + alerts + events)
- [ ] Per-device CSV download URL for EM/EM1 via `GetEMDataCSVURL()`
- [ ] Tests

#### 11.3.7 Context Keybinding Updates
- [ ] Update `ContextMonitor` with all new keybindings
- [ ] Update help overlay for Monitor panels
- [ ] Tests

#### 11.3.8 Remove Old Monitor Component
- [ ] Delete old `monitor/model.go` flat card implementation
- [ ] Clean up dead code, old styles, orphaned imports
- [ ] Update `renderMonitorLayout()` to pure delegation
- [ ] Full audit pass

---

## PHASE 12: Help & Documentation (Session 28)

> *Discoverable features and complete docs*

### 12.1 Help Overlay Enhancement ✓ COMPLETE

> **Completed 2026-01-29:** Added `/` search within help overlay. Press `/` to activate search,
> type to filter keybindings by key or description (case-insensitive), Enter to confirm filter,
> Esc to clear and exit search. Shows "No matching keybindings" when filter produces no results.

- [x] Organize by context (Global, Navigation, Actions, View-specific) - 15 contexts defined
- [x] Show keybindings grouped by action
- [x] Add search within help (`/` in help overlay)
- [x] Context-aware: shows relevant bindings based on current view
- [x] Scrolling for long help content (j/k/up/down/g/PageUp/PageDown)

### 12.2 Per-Panel Footer Hints → Merged into 10.2

> **Evaluated 2026-01-29:** This is the same feature as 10.2. See 10.2 for implementation plan.

---

## PHASE 13: Testing & Polish (Sessions 29-30)

> *Final quality pass*

### 13.1 Testing
- [ ] Unit tests for all new components
- [ ] Integration tests for view switching
- [ ] Integration tests for keybindings
- [ ] Test all global keybindings from all views
- [ ] Test with mock data
- [ ] Test with real devices (Gen1, Gen2, Gen3, Gen4)
- [ ] Test all Dashboard panels work with integrated components
- [ ] Test provisioning wizard flow end-to-end
- [ ] Audit `internal/tui/app.go` - identify monolithic sections (1651 lines!)
- [ ] Audit `internal/tui/views/*.go` - verify all views work
- [ ] Test each tab manually (Dashboard, Automation, Config, Manage, Monitor, Fleet)
- [ ] Document bugs found during audit in "Bugs Found" section

### 13.2 VHS Documentation
- [ ] `tab-switching.tape` - Demonstrate all tabs
- [ ] `device-selection.tape` - Navigate device list
- [ ] `device-actions.tape` - Toggle, on, off, reboot
- [ ] `config-modals.tape` - WiFi, auth, cloud modals
- [ ] `config.tape` - WiFi, system, cloud, inputs panels
- [ ] `automation.tape` - Scripts, schedules, webhooks
- [ ] `manage.tape` - Discovery, batch, firmware, backup panels
- [ ] `fleet.tape` - Cloud fleet management
- [ ] `overlays.tape` - Help, JSON viewer
- [ ] `responsive.tape` - Narrow terminal layout

### 13.2.1 README.md TUI Section Update
- [ ] Add TUI section to README.md after "Installation"
- [ ] Include GIF from VHS recordings
- [ ] Navigation table with keybindings
- [ ] Link to docs/tui.md

### 13.3 Written Documentation
- [ ] Update `docs/tui.md` with all features
- [ ] Document all keybindings in table format
- [ ] Add screenshots/GIFs
- [ ] Update README TUI section

### 13.4 Performance
- [ ] Profile render performance
- [ ] Optimize hot paths (reduce allocations)
- [ ] Reduce unnecessary re-renders
- [ ] Test with 50+ devices

### 13.5 Code Cleanup
- [ ] Delete unused component files
- [ ] Remove duplicate rendering in app.go
- [ ] Clean up unused imports
- [ ] Verify no TODO/FIXME comments remain

---

## PHASE 14: Multi-Platform Support (Sessions 31-33)

> *Wire existing plugin infrastructure (`internal/plugins/`, `internal/shelly/dispatch.go`) into TUI components.*
> *The CLI plugin system is fully built — this phase is purely TUI integration.*

---

### Session 31: Device List Platform Display & Cache Integration

#### 14.1.1 Platform Badge in Device List
- [x] Badge `[T]` (first letter, uppercase) appended to name for plugin devices; no badge for native Shelly
- [x] Tests for badge rendering

#### 14.1.2 Platform in Detail Panel
- [x] "Platform" row in device detail for plugin-managed devices
- [x] Tests

#### 14.1.3 Platform Filter Key
- [x] `ActionPlatformFilter` / `PlatformFilterMsg` / `p` keybinding in `ContextDevices`
- [x] `platformFilter` field on Model, cycle through discovered platforms
- [x] Filter applied alongside existing text filter
- [x] Footer hint shows active platform filter
- [x] Tests

#### 14.1.4 Cache: Plugin Device Status Fetching
- [x] Branch in `fetchDeviceWithID()` for `device.IsPluginManaged()`
- [x] `cache/plugin_parser.go` — `ParsePluginStatus()` maps `DeviceStatusResult` → `ParsedStatus`
- [x] Minimal `DeviceInfo` construction for plugin devices
- [x] Tests for parser and cache branch

---

### Session 32: Plugin Device Control

<details>
<summary><b>SESSION PROMPT</b> (delete this block when session is complete)</summary>

**Goal**: Wire plugin control hooks into the TUI control panel so users can toggle/control plugin-managed devices the same way they control native Shelly devices.

**Context**: Session 31 completed platform display and cache integration. The control panel overlay (`internal/tui/components/control/`) has `switch.go`, `light.go`, `cover.go`, `rgb.go`, `thermostat.go` — each calls the Shelly service directly (e.g. `svc.SwitchToggle()`). For plugin devices, control must go through `svc.dispatchToPlugin()` or `hooks.ExecuteControl()`. The service layer already has `dispatchToPlugin(ctx, device, action, component, id)` → `*PluginQuickResult`.

**Read these files first:**
- `internal/tui/components/control/switch.go` — how SwitchModel calls svc for toggle/on/off
- `internal/tui/components/control/light.go` — LightModel brightness/toggle
- `internal/tui/components/control/cover.go` — CoverModel open/close/stop/position
- `internal/tui/app.go` — `showControlPanel()` function that creates control overlays
- `internal/shelly/dispatch.go` — `dispatchToPlugin()`, `PluginQuickResult`
- `internal/plugins/hooks.go` — `ExecuteControl(ctx, address, auth, action, component, id)`
- `internal/plugins/types.go` — `ControlResult` struct

**Tasks:**

#### 14.2.1 Plugin Control Adapter
- [ ] Create `internal/tui/components/control/plugin.go` — a `PluginModel` control panel
- [ ] PluginModel should display available components from `cache.DeviceData` (switches, lights, covers parsed in Session 31)
- [ ] For each component, show current state and provide toggle/on/off/open/close actions
- [ ] Actions call `svc.dispatchToPlugin(ctx, device, action, componentType, &componentID)` and display `PluginQuickResult`
- [ ] Support generic action buttons when component type is unknown (just show "toggle", "on", "off")
- [ ] Toast feedback on success/failure via `ControlResultMsg`
- [ ] Write tests

#### 14.2.2 Wire Plugin Control into App
- [ ] In `showControlPanel()` (`app.go`), add a branch: if selected device `IsPluginManaged()`, show `PluginModel` instead of native Switch/Light/Cover panels
- [ ] Ensure `c` key from device list works for plugin devices
- [ ] Quick toggle (`t` key) from device list should also dispatch through plugin for plugin devices — check the quick toggle path and add the branch
- [ ] Write tests for the dispatch branch

#### 14.2.3 Plugin Device Info Panel
- [ ] In `deviceinfo` component (or wherever the info overlay `d` key shows), handle plugin devices gracefully
- [ ] If device is plugin-managed and has no `shelly.DeviceInfo`, show what's available from config + last status
- [ ] Don't crash or show empty panels for plugin devices
- [ ] Write tests

**Verification:**
- `go build ./...`
- `golangci-lint run ./internal/tui/components/control/ ./internal/tui/app.go`
- `go test ./internal/tui/components/control/`
- Commit with descriptive message

**When done:** Delete this prompt block and mark all checkboxes above. Leave only the checkbox list.

</details>

#### 14.2.1 Plugin Control Adapter
- [ ] `control/plugin.go` — PluginModel with component list and action dispatch
- [ ] Actions call `svc.dispatchToPlugin()`, display result
- [ ] Toast feedback on success/failure
- [ ] Tests

#### 14.2.2 Wire Plugin Control into App
- [ ] Branch in `showControlPanel()` for plugin devices → PluginModel
- [ ] Quick toggle (`t`) dispatches through plugin for plugin devices
- [ ] Tests

#### 14.2.3 Plugin Device Info Panel
- [ ] Graceful handling of plugin devices in info/detail overlays
- [ ] No crashes or empty panels for plugin devices
- [ ] Tests

---

### Session 33: Firmware Multi-Platform

<details>
<summary><b>SESSION PROMPT</b> (delete this block when session is complete)</summary>

**Goal**: Include plugin-managed devices in the firmware management view with platform-aware checking and updating.

**Context**: The firmware component (`internal/tui/components/firmware/model.go`) has `DeviceFirmware` struct and `checkAllDevices()` which calls `svc.CheckFirmware()` for each device. For plugin devices, firmware checks go through `hooks.ExecuteCheckUpdates()` → `FirmwareUpdateInfo`, and updates through `hooks.ExecuteApplyUpdate()` → `UpdateResult`. The service layer dispatch is in `internal/shelly/dispatch.go`.

**Read these files first:**
- `internal/tui/components/firmware/model.go` — `DeviceFirmware` struct (line ~30), `checkAllDevices()` (line ~303), `updateDevices()` (line ~377), `renderDeviceLine()` (line ~1081), `New()` constructor
- `internal/plugins/types.go` — `FirmwareUpdateInfo`, `UpdateResult`
- `internal/plugins/hooks.go` — `ExecuteCheckUpdates()`, `ExecuteApplyUpdate()`
- `internal/shelly/dispatch.go` — check if there's already firmware dispatch, or if it needs adding

**Tasks:**

#### 14.3.1 Platform Field on DeviceFirmware
- [ ] Add `Platform string` field to `DeviceFirmware` struct
- [ ] Populate it from `device.GetPlatform()` when building the device list in `New()` or wherever devices are loaded
- [ ] Write tests

#### 14.3.2 Platform Column in Firmware View
- [ ] In `renderDeviceLine()` (line ~1081), add a platform indicator column for non-Shelly devices (same `[T]` style as device list)
- [ ] Shelly devices show no indicator (keep it clean)
- [ ] Adjust column widths to accommodate the badge
- [ ] Write tests for render output

#### 14.3.3 Plugin Firmware Checking
- [ ] In `checkAllDevices()` (line ~303), add a branch: if `device.Platform` indicates a plugin device, use the service layer's plugin firmware check instead of `svc.CheckFirmware()`
  - Check if `svc.CheckFirmware()` already handles plugin dispatch — if so, this may just work. If not, add a method or call `hooks.ExecuteCheckUpdates()` through the service
- [ ] Map `FirmwareUpdateInfo` fields to `DeviceFirmware` fields: `CurrentVersion`→`Current`, `LatestStable`→`Available`, `HasUpdate`→`HasUpdate`
- [ ] Handle case where plugin doesn't have `check_updates` hook — set `Err` to a descriptive error, don't crash
- [ ] Write tests with mock plugin returning `FirmwareUpdateInfo`

#### 14.3.4 Plugin Firmware Updates
- [ ] In `updateDevices()` (line ~377), add a branch: if device is plugin-managed, use `hooks.ExecuteApplyUpdate()` through the service instead of `svc.UpdateFirmware()`
- [ ] Map `UpdateResult` → `UpdateResult` (the TUI's `UpdateResult` struct: `Name`, `Success`, `Err`)
- [ ] Handle `Rebooting: true` — show appropriate status (same as native Shelly post-update)
- [ ] Handle case where plugin doesn't have `apply_update` hook — show error, don't crash
- [ ] Write tests

#### 14.3.5 Staged Updates for Plugin Devices
- [ ] In `updateDevicesStaged()` (line ~851), ensure plugin devices are included in staged batches
- [ ] Same dispatch logic as 14.3.4 but within the staged update flow
- [ ] Write tests

#### 14.3.6 Rollback for Plugin Devices
- [ ] Check if plugin `FirmwareUpdateInfo` provides rollback info — if `CanRollback` isn't in the plugin types, disable rollback for plugin devices
- [ ] In the rollback handler, skip or show "not supported" for plugin devices
- [ ] Write tests

**Verification:**
- `go build ./...`
- `golangci-lint run ./internal/tui/components/firmware/ ./internal/shelly/`
- `go test ./internal/tui/components/firmware/`
- `./scripts/audit-conventions.sh > /tmp/audit.txt 2>&1` — read and fix any issues
- Commit with descriptive message

**When done:** Delete this prompt block and mark all checkboxes above. Leave only the checkbox list.

</details>

#### 14.3.1 Platform Field on DeviceFirmware
- [ ] Add `Platform string` to `DeviceFirmware`, populate from `device.GetPlatform()`
- [ ] Tests

#### 14.3.2 Platform Column in Firmware View
- [ ] Platform badge `[T]` in `renderDeviceLine()` for non-Shelly devices
- [ ] Tests

#### 14.3.3 Plugin Firmware Checking
- [ ] Branch in `checkAllDevices()` for plugin devices → `ExecuteCheckUpdates()` via service
- [ ] Map `FirmwareUpdateInfo` → `DeviceFirmware` fields
- [ ] Handle missing `check_updates` hook gracefully
- [ ] Tests

#### 14.3.4 Plugin Firmware Updates
- [ ] Branch in `updateDevices()` for plugin devices → `ExecuteApplyUpdate()` via service
- [ ] Map `UpdateResult`, handle `Rebooting` state
- [ ] Handle missing `apply_update` hook gracefully
- [ ] Tests

#### 14.3.5 Staged Updates for Plugin Devices
- [ ] Plugin devices included in staged batches with same dispatch
- [ ] Tests

#### 14.3.6 Rollback for Plugin Devices
- [ ] Disable rollback for plugin devices (not in plugin types)
- [ ] Show "not supported" instead of crashing
- [ ] Tests

---

## Verification Checklist (End of All Phases)

### Functional Completeness
- [ ] All 6 tabs accessible and functional
- [ ] All panels navigable (Tab/Shift+Tab, Shift+N)
- [ ] All device types controllable (Switch, Cover, Light, RGB, Thermostat, Sensor)
- [ ] All configuration editable via modals
- [ ] All automation features work (Scripts, Schedules, Webhooks, KVS, Actions, Virtuals)
- [ ] Groups and Scenes functional
- [ ] All protocol features accessible
- [ ] JSON viewer with syntax highlighting
- [ ] Energy history with charts
- [ ] Alerts configurable
- [ ] Events streaming

### shelly-go Feature Coverage
- [ ] Gen1 devices fully supported
- [ ] Gen2 devices fully supported
- [ ] Gen3 devices fully supported (Plus, Pro)
- [ ] Gen4 devices fully supported
- [ ] All component types exposed
- [ ] Firmware management complete
- [ ] Backup/restore working
- [ ] Event streaming working
- [ ] Device provisioning working

### Code Quality
- [ ] Convention audit passes: `./scripts/audit-conventions.sh`
- [ ] No TODO/FIXME comments
- [ ] No placeholder implementations
- [ ] All error handling complete
- [ ] Factory pattern used consistently

---

## Key Files Reference

### Current TUI (to modify)
- `internal/tui/app.go` - Main app (1651 lines, needs refactor)
- `internal/tui/views/` - All 5 views
- `internal/tui/components/` - 34 components (13 unused)
- `internal/tui/layout/calculator.go` - Layout system
- `internal/tui/rendering/renderer.go` - Panel rendering
- `internal/tui/keys/` - Keybinding definitions

### Superfile Reference
- `/tmp/superfile/src/internal/ui/rendering/` - Renderer patterns
- `/tmp/superfile/src/internal/model.go` - Main model
- `/tmp/superfile/src/internal/ui/preview/` - Slide-in panel
- `/tmp/superfile/src/internal/common/style.go` - Style management

### shelly-go Features
- `/opt/repos/shelly-go/gen2/components/` - 38 component types
- `/opt/repos/shelly-go/backup/` - Backup/restore
- `/opt/repos/shelly-go/firmware/` - Firmware management
- `/opt/repos/shelly-go/discovery/` - Device discovery
- `/opt/repos/shelly-go/events/` - Event streaming

---

## Debug Logging

**Log location**: `~/.config/shelly/debug/<timestamp>/tui.log`

Each TUI session creates its own timestamped folder (e.g., `2025-12-24_15-04-05/`) containing:
- `tui.log` - Main debug log with view renders and events

---

## Code Quality: DRY Violations

> Duplicated code that should be consolidated.

### PM/EM/EM1 Aggregation Functions ✓

**Problem**: Identical aggregation functions exist in multiple files:
- `internal/tui/components/monitor/model.go`: `aggregateStatusFromPM()`, `aggregateStatusFromEM()`, `aggregateStatusFromEM1()`

**Solution**:
- [x] Consolidated with generic `aggregateMetrics[T]()` using existing `model.MeterReading` interface
- [x] Single generic function replaces all three, with `accumulateCurrent` flag for EM behavior
- [x] Tests added for all meter types and cross-type aggregation

### helpers.Sizable Location ✓

**Problem**: `internal/tui/helpers/` was created despite explicit instruction not to create a helpers directory.

**Solution**:
- [x] Moved `helpers/sizable.go` to `internal/tui/panel/sizable.go` as `panel.Sizable`
- [x] Updated all imports (39 files)
- [x] Deleted `internal/tui/helpers/` directory

---

## Code Quality: Adopt errorview Package

> `internal/tui/components/errorview/` is a well-designed error display utility that should be adopted for panel error states.

### errorview Adoption Plan

- [ ] Audit current ad-hoc error rendering patterns (uses of `theme.StatusError().Render()`)
- [ ] Identify panels that show persistent error states (device offline, fetch failed)
- [ ] Replace ad-hoc rendering with `errorview` component
- [ ] Document when to use `errorview` (panel errors) vs `toast` (transient notifications)

---

## Bugs Found (Running List)

> Document bugs here as discovered. Address periodically at user request.

### Layout Issues
- [ ] **Device detail viewport sizing** - Minor: viewport uses `height - 6` but container uses `height - 2`, causing potential whitespace inconsistency

### Resolved/Not Bugs (2026-01-19 audit)
- ~~Manage tab panels misaligned~~ - **Not reproducible**, layout calculator handles column heights correctly
- ~~Header power shows "--"~~ - **By design** when totalPower == 0 (no devices, all offline, or no PM capability)

---

## Notes

- **QUALITY OVER SPEED** - Take the time to do things right. Rushing leads to bugs and technical debt
- **DO NOT EVER SKIP TASKS WITHOUT APPROVAL** - If something seems unnecessary, ask first
- **Never cut corners** - Each checkbox must be fully complete before checking
- **Test after each change** - Build, lint, test, manual verify
- **Commit frequently** - After each completed sub-section
- **Activate existing components first** - Don't rebuild what's already built
- **Ask if unclear** - Don't guess at requirements
