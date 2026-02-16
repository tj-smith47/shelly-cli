## shelly backup export

Export backups for all registered devices

### Synopsis

Export backup files for all registered devices to a directory.

Creates one backup file per device, named by device ID.
Use --format to choose JSON or YAML output.

```
shelly backup export <directory> [flags]
```

### Examples

```
  # Export all device backups to directory
  shelly backup export ./backups

  # Export in YAML format
  shelly backup export ./backups --format yaml

  # Export with parallel processing
  shelly backup export ./backups --parallel 5
```

### Options

```
  -a, --all             Export all registered devices (default true)
  -f, --format string   Output format (json, yaml) (default "json")
  -h, --help            help for export
      --parallel int    Number of parallel backups (default 3)
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

