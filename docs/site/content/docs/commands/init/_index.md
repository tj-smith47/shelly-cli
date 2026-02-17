---
title: "shelly init"
description: "shelly init"
weight: 300
sidebar:
  collapsed: true
---

## shelly init

Initialize shelly CLI for first-time use

### Synopsis

Initialize the shelly CLI with a guided setup wizard.

This command walks you through:
  1. Configuration (theme, output format, aliases)
  2. Device discovery (find Shelly devices on your network)
  3. Device registration (give discovered devices friendly names)
  4. Shell completions (tab completion for commands)
  5. Cloud access (optional remote control via Shelly Cloud)
  6. Telemetry (optional anonymous usage statistics)

INTERACTIVE MODE (default):
  Run without flags for the full guided wizard.

DEFAULTS MODE (--defaults):
  Runs the full wizard with sensible defaults, no prompts.
  Discovers devices, registers with default names, installs completions.
  Combine with flags to override individual defaults (e.g., --defaults --theme nord).

INDIVIDUAL FLAGS:
  Pass specific flags to set values; other steps remain interactive.

Use --check to verify your current setup without making changes.

```
shelly init [flags]
```

### Examples

```
  # Interactive setup wizard (recommended for first use)
  shelly init

  # Quick setup with sensible defaults (no prompts)
  shelly init --defaults

  # Defaults with a specific theme
  shelly init --defaults --theme nord

  # Defaults with force overwrite of existing config
  shelly init --defaults --force

  # Check current setup without changes
  shelly init --check

  # Register devices directly (other steps remain interactive)
  shelly init --device kitchen=192.168.1.100 --device bedroom=192.168.1.101

  # Run discovery without prompting (other steps remain interactive)
  shelly init --discover

  # Scan specific subnets
  shelly init --discover --network 192.168.1.0/24 --network 10.0.0.0/24

  # Full CI/CD setup (no prompts)
  shelly init --defaults \
    --device kitchen=192.168.1.100 \
    --theme dracula \
    --no-color

  # With cloud credentials
  shelly init --defaults --cloud-email user@example.com --cloud-password secret
```

### Options

```
      --aliases                     Install default command aliases (opt-in)
      --all-networks                Scan all detected subnets without prompting
      --api-mode string             API mode: local,cloud,auto (default: local)
      --check                       Verify current setup without making changes
      --cloud-email string          Shelly Cloud email (enables cloud setup)
      --cloud-password string       Shelly Cloud password (enables cloud setup)
      --completions string          Install completions for shells: bash,zsh,fish,powershell (comma-separated)
      --defaults                    Use sensible defaults for all prompts (no interactive questions)
      --device stringArray          Device spec: name=ip[:user:pass] (repeatable)
      --devices-json stringArray    JSON device(s): file path, array, or single object (repeatable)
      --discover                    Run device discovery without prompting
      --discover-modes string       Discovery modes: http,mdns,coiot,ble,all (comma-separated) (default "http")
      --discover-timeout duration   Discovery timeout (default 2m0s)
      --force                       Overwrite existing configuration
  -h, --help                        help for init
      --network stringArray         Subnet(s) for HTTP probe discovery (repeatable)
      --no-color                    Disable colors in output
      --output-format string        Set output format: table,json,yaml (default: table)
      --telemetry                   Enable anonymous usage telemetry (opt-in)
      --theme string                Set theme (default: dracula)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

