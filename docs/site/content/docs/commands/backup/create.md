---
title: "shelly backup create"
description: "shelly backup create"
---

## shelly backup create

Create a device backup

### Synopsis

Create a complete backup of a Shelly device.

The backup includes configuration, scripts, schedules, and webhooks.
Backups are written as JSON. If no file is specified, the backup is saved
to ~/.config/shelly/backups/ with a name based on the device and date. Use
"-" as the file to write to stdout.

Use --encrypt to AES-encrypt the backup with a password; restore the file
with 'shelly backup restore --decrypt <password>'.

```
shelly backup create <device> [file] [flags]
```

### Examples

```
  # Create backup (auto-saved to ~/.config/shelly/backups/)
  shelly backup create living-room

  # Create backup to specific file
  shelly backup create living-room backup.json

  # Create backup to stdout
  shelly backup create living-room -

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts
```

### Options

```
  -e, --encrypt string   Password to AES-encrypt the backup
  -h, --help             help for create
      --skip-schedules   Exclude schedules from backup
      --skip-scripts     Exclude scripts from backup
      --skip-webhooks    Exclude webhooks from backup
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
      --raw                     Print the exact device response(s) as a JSON array and suppress normal output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly backup](shelly_backup.md)	 - Backup and restore device configurations

