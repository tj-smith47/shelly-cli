## shelly backup create

Create a device backup

### Synopsis

Create a complete backup of a Shelly device.

The backup includes configuration, scripts, schedules, and webhooks.
If no file is specified, backup is saved to ~/.config/shelly/backups/
with a name based on the device and date. Use "-" as the file to write
to stdout.

Use --encrypt to password-protect the backup (password verification only,
sensitive data is not encrypted in the file).

```
shelly backup create <device> [file] [flags]
```

### Examples

```
  # Create backup (auto-saved to ~/.config/shelly/backups/)
  shelly backup create living-room

  # Create backup to specific file
  shelly backup create living-room backup.json

  # Create YAML backup
  shelly backup create living-room backup.yaml --format yaml

  # Create backup to stdout
  shelly backup create living-room -

  # Create encrypted backup
  shelly backup create living-room backup.json --encrypt mysecret

  # Skip scripts in backup
  shelly backup create living-room backup.json --skip-scripts
```

### Options

```
  -e, --encrypt string   Password to protect backup
  -f, --format string    Output format (json, yaml) (default "json")
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly backup](shelly_backup.md)	 - Backup and restore device configurations

