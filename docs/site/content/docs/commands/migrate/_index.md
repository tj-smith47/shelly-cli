---
title: "shelly migrate"
description: "shelly migrate"
weight: 440
sidebar:
  collapsed: true
---

## shelly migrate

Migrate configuration between devices

### Synopsis

Migrate configuration from a source device or backup file to a target device.

Source can be a device name/address or a backup file path.
The --dry-run flag shows what would be changed without applying.

```
shelly migrate <source> <target> [flags]
```

### Examples

```
  # Migrate from one device to another
  shelly migrate living-room bedroom --dry-run

  # Migrate from backup file to device
  shelly migrate backup.json bedroom

  # Force migration between different device types
  shelly migrate backup.json bedroom --force
```

### Options

```
      --dry-run   Show what would be changed without applying
      --force     Force migration between different device types
  -h, --help      help for migrate
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

