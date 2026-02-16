---
title: "shelly migrate"
description: "shelly migrate"
weight: 400
sidebar:
  collapsed: true
---

## shelly migrate

Migrate configuration between devices

### Synopsis

Migrate configuration from one Shelly device to another.

Reads the current configuration from the source device and applies it to
the target device. By default, everything is migrated including network
and authentication settings.

When network settings are migrated, the source device is factory reset
after a successful migration to prevent IP conflicts on the network.
Use --skip-network to keep both devices online with their current
network settings, or --reset-source=false to skip the factory reset
(warning: this may cause IP conflicts).

Use --dry-run to preview what would change without applying.

```
shelly migrate <source-device> <target-device> [flags]
```

### Examples

```
  # Preview migration (dry run)
  shelly migrate living-room bedroom --dry-run

  # Full migration (factory resets source after)
  shelly migrate living-room bedroom --yes

  # Migrate without network config (no factory reset needed)
  shelly migrate living-room bedroom --skip-network

  # Migrate network but skip factory reset (may cause IP conflict)
  shelly migrate living-room bedroom --reset-source=false

  # Force migration between different device types
  shelly migrate living-room bedroom --force --yes
```

### Options

```
      --dry-run          Show what would be changed without applying
      --force            Force migration between different device types
  -h, --help             help for migrate
      --reset-source     Factory reset source device after migration (default true)
      --skip-auth        Skip authentication configuration
      --skip-network     Skip network configuration (WiFi, Ethernet)
      --skip-schedules   Skip schedule migration
      --skip-scripts     Skip script migration
      --skip-webhooks    Skip webhook migration
  -y, --yes              Skip confirmation prompt
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
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
* [shelly migrate diff](shelly_migrate_diff.md)	 - Show differences between device and backup
* [shelly migrate validate](shelly_migrate_validate.md)	 - Validate a backup file

