---
title: "shelly backup"
description: "shelly backup"
weight: 80
sidebar:
  collapsed: true
---

## shelly backup

Backup and restore device configurations

### Synopsis

Create, restore, and manage device backups.

Backups include device configuration, scripts, schedules, and webhooks.
Use encryption to protect sensitive data in backups.

### Examples

```
  # Create a backup
  shelly backup create living-room backup.json

  # Restore from backup
  shelly backup restore living-room backup.json

  # List existing backups
  shelly backup list

  # Export all device backups
  shelly backup export ./backups
```

### Options

```
  -h, --help   help for backup
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
* [shelly backup create](shelly_backup_create.md)	 - Create a device backup
* [shelly backup export](shelly_backup_export.md)	 - Export backups for all registered devices
* [shelly backup list](shelly_backup_list.md)	 - List saved backups
* [shelly backup restore](shelly_backup_restore.md)	 - Restore a device from backup

