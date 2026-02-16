## shelly backup list

List saved backups

### Synopsis

List backup files in a directory.

By default, looks in the config directory's backups folder. Backup files
contain full device configuration snapshots that can be used to restore
device settings or migrate configurations between devices.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

Columns: Filename, Device, Model, Created, Encrypted, Size

```
shelly backup list [directory] [flags]
```

### Examples

```
  # List backups in default location
  shelly backup list

  # List backups in specific directory
  shelly backup list /path/to/backups

  # Output as JSON
  shelly backup list -o json

  # Find backups for a specific device model
  shelly backup list -o json | jq '.[] | select(.device_model | contains("Plus"))'

  # Get most recent backup filename
  shelly backup list -o json | jq -r 'sort_by(.created_at) | last | .filename'

  # Short form
  shelly backup ls
```

### Options

```
  -h, --help   help for list
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

