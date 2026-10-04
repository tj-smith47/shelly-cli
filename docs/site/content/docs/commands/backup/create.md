---
title: "shelly backup create"
description: "shelly backup create"
---

## shelly backup create

Create a device backup

### Synopsis

Create a complete backup of a Shelly device, or of every registered
device with --all.

The backup includes configuration, scripts, schedules, and webhooks.
Backups are written as JSON. If no file is specified, the backup is saved
to ~/.config/shelly/backups/ (or --dir) with a name based on the device,
its MAC address and the date. Use "-" as the file to write to stdout.

With --all, every registered device is backed up to its own auto-named
file. A device that fails does not stop the others; each device's result
and a summary are printed, and the command exits non-zero when any device
failed. --encrypt and the --skip-* flags apply to every device.

Use --encrypt to AES-encrypt the backup with a password; restore the file
with 'shelly backup restore <device> <file> --decrypt <password>'.

```
shelly backup create [device] [file] [flags]
```

### Examples

```
  # Create backup (auto-saved to ~/.config/shelly/backups/)
  shelly backup create living-room

  # Create backup to specific file
  shelly backup create living-room backup.json

  # Create auto-named backup in a directory
  shelly backup create living-room --dir ./backups

  # Back up every registered device
  shelly backup create --all

  # Back up every registered device into a dated directory, quietly
  shelly backup create --all --dir ./backups/$(date +%Y-%m-%d) -q

  # Create backup to stdout
  shelly backup create living-room -

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Read the encryption password from stdin, keeping it out of shell history
  shelly backup create living-room backup.json --encrypt-stdin < ~/.shelly-backup-password

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts
```

### Options

```
  -a, --all              Target all registered devices
      --dir string       Directory for auto-named backups (default ~/.config/shelly/backups/, created if missing)
  -e, --encrypt string   Password to AES-encrypt the backup
      --encrypt-stdin    Read the password to AES-encrypt the backup from stdin
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

