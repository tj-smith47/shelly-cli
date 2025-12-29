---
title: "Alias Examples"
description: "Pre-built alias collections for common workflows"
weight: 10
---

Command alias collections ranging from simple shortcuts to advanced automation helpers.

## Usage

```bash
# Import any alias file
shelly alias import examples/aliases/shortcuts.yaml

# List imported aliases
shelly alias list

# Use an alias
shelly ls  # if 'ls' was imported as alias for 'device list'
```

---

## automation.yaml

Automation-Focused Aliases for Shelly CLI

```yaml
# Automation-Focused Aliases for Shelly CLI
#
# Import: shelly alias import examples/aliases/automation.yaml
#
# Aliases designed for scripting, scheduling, and home automation workflows.
# Shell aliases (prefixed with !) enable pipes and complex operations.

aliases:
  # Silent operations for scripts (quiet mode)
  qon: "on -q"
  qoff: "off -q"
  qtog: "toggle -q"

  # Wait for device conditions
  wait-on: "wait --state on --timeout 60s"
  wait-off: "wait --state off --timeout 60s"
  wait-online: "wait --online --timeout 30s"

  # Batch operations (quiet for cron)
  batch-on-q: "batch on -q"
  batch-off-q: "batch off -q"

  # Scene automation (define your scenes in config)
  morning: "scene activate morning"
  evening: "scene activate evening"
  night: "scene activate night"
  away: "scene activate away"
  home: "scene activate home"

  # Energy logging (shell mode for file output)
  log-power: "!shelly energy status --all -o csv >> ~/shelly-power.csv"

  # Health checks for monitoring
  check-all: "!shelly device list -o json | jq -e 'all(.online)' > /dev/null && echo OK || echo FAIL"
  check-offline: "!shelly device list -o json | jq -r '.[] | select(.online == false) | .name'"

  # Webhook testing
  hook-test: "webhook server --port 8080 --log"

  # Backup automation
  backup-daily: "!shelly backup create --all --dir ~/shelly-backups/$(date +%Y-%m-%d)"

  # Device discovery and registration
  scan-register: "discover scan --register"

  # Alert management
  alert-test: "alert test"
  alert-snooze: "alert snooze"

  # Report generation
  report-daily: "!shelly report energy --format json > ~/shelly-reports/energy-$(date +%Y-%m-%d).json"
  report-audit: "audit --check-firmware --check-auth -o json"

  # Prometheus metrics export
  metrics: "metrics export prometheus"

  # Party and fun
  party-mode: "party --duration 5m"
  sleep-mode: "sleep --duration 30m"
  wake-mode: "wake --duration 15m --simulate sunrise"

  # Conditional operations (shell mode with proper bash syntax)
  if-on: "!shelly status \"$1\" --plain 2>/dev/null | grep -q on && shelly \"$2\" \"$3\" \"$4\""
  if-off: "!shelly status \"$1\" --plain 2>/dev/null | grep -q off && shelly \"$2\" \"$3\" \"$4\""
```

---

## power-users.yaml

Power User Aliases for Shelly CLI

```yaml
# Power User Aliases for Shelly CLI
#
# Import: shelly alias import examples/aliases/power-users.yaml
#
# These aliases provide shortcuts for common command+flag combinations
# that power users type frequently. Arguments are auto-appended.
#
# Note: Reserved command names (on, off, toggle, status, device, batch, etc.)
# cannot be used as alias names.

aliases:
  # Device inspection shortcuts
  # 'shelly i kitchen' → 'shelly device info kitchen'
  i: "device info"
  ij: "device info -o json"

  # List all devices (shorter than 'device list')
  ls: "device list"
  lsj: "device list -o json"

  # Discovery with common timeout
  scan: "discover scan --timeout 30s"
  scan-reg: "discover scan --register"

  # Batch operations with all devices
  all-on: "batch on --all"
  all-off: "batch off --all"

  # Configuration export shortcuts
  cfgx: "config export -o yaml"
  cfgj: "config export -o json"

  # Firmware check all devices at once
  fw-check: "firmware check --all"
  fw-update: "firmware update --all --yes"

  # Energy monitoring shortcut
  pwr: "energy status"
  pwr-all: "energy status --all"
  pwr-hist: "energy history"

  # Quick debug/diagnostics
  methods: "debug methods"
  rpc: "debug rpc"

  # Backup shortcuts
  bk: "backup create"
  bk-all: "backup create --all"

  # TUI dashboard access
  ui: "tui dash"

  # Open device web interface
  web: "device ui"

  # Shell aliases for JSON filtering (requires jq)
  # List only online devices
  online: "!shelly device list -o json | jq -r '.[] | select(.online == true) | .name'"
  # List only offline devices
  offline: "!shelly device list -o json | jq -r '.[] | select(.online == false) | .name'"
  # Power summary for all devices
  pwr-summary: "!shelly energy status --all -o json | jq -r '.[] | \"\\(.name): \\(.power // 0)W\"'"
```

---

## shortcuts.yaml

Common Shortcut Aliases for Shelly CLI

```yaml
# Common Shortcut Aliases for Shelly CLI
#
# Import: shelly alias import examples/aliases/shortcuts.yaml
#
# Simple, memorable shortcuts for everyday device control.
# Perfect for beginners and quick command-line usage.
# Arguments are auto-appended - no placeholders needed.

aliases:
  # List all devices
  ls: "device list"

  # Quick device info
  info: "device info"

  # Reboot shortcut
  boot: "reboot"

  # Discover devices
  find: "discover scan"

  # Energy check
  power: "energy status"

  # Update firmware
  update: "firmware update"

  # Open device UI
  web: "device ui"

  # Quick config view
  cfg: "config get"

  # Scene shortcuts
  run: "scene activate"
  save: "scene save"

  # Version info
  ver: "version"
```

---

