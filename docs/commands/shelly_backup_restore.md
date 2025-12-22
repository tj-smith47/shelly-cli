## shelly backup restore

Restore a device from backup

### Synopsis

Restore a Shelly device from a backup file.

By default, all data from the backup is restored. Use --skip-* flags
to exclude specific sections.

Network configuration (WiFi, Ethernet) is skipped by default with
--skip-network to prevent losing connectivity.

```
shelly backup restore <device> <file> [flags]
```

### Examples

```
  # Restore from backup (skip network config)
  shelly backup restore living-room backup.json

  # Dry run - show what would change
  shelly backup restore living-room backup.json --dry-run

  # Restore everything including network config
  shelly backup restore living-room backup.json --skip-network=false

  # Restore encrypted backup
  shelly backup restore living-room backup.json --decrypt mysecret

  # Skip scripts during restore
  shelly backup restore living-room backup.json --skip-scripts
```

### Options

```
  -d, --decrypt string   Password to decrypt backup
      --dry-run          Show what would be restored without applying
  -h, --help             help for restore
      --skip-network     Skip network configuration (WiFi, Ethernet) (default true)
      --skip-schedules   Skip schedule restoration
      --skip-scripts     Skip script restoration
      --skip-webhooks    Skip webhook restoration
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly backup](shelly_backup.md)	 - Backup and restore device configurations

