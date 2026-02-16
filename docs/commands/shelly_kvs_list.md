## shelly kvs list

List KVS keys

### Synopsis

List all keys stored in the device's Key-Value Storage (KVS).

KVS provides persistent storage on Gen2+ devices for scripts and user
data. Keys are strings and values can be strings, numbers, or booleans.
Data persists across reboots and firmware updates.

By default, only key names are listed. Use --values to also show
the stored values. Use --match for wildcard filtering.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

```
shelly kvs list <device> [flags]
```

### Examples

```
  # List all keys
  shelly kvs list living-room

  # List keys with values
  shelly kvs list living-room --values

  # List keys matching a pattern
  shelly kvs list living-room --match "sensor_*"

  # Output as JSON for scripting
  shelly kvs list living-room -o json

  # Get all values as JSON
  shelly kvs list living-room --values -o json

  # Export all KVS data to backup file
  shelly kvs list living-room --values -o json > kvs-backup.json

  # Find string-type keys only
  shelly kvs list living-room --values -o json | jq '.[] | select(.type == "string")'

  # Short form
  shelly kvs ls living-room
```

### Options

```
  -h, --help           help for list
  -m, --match string   Pattern to match keys (supports * and ? wildcards)
      --values         Show values alongside keys
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

* [shelly kvs](shelly_kvs.md)	 - Manage device key-value storage

