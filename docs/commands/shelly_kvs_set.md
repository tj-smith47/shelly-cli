## shelly kvs set

Set a KVS value

### Synopsis

Store a value in the device's Key-Value Storage.

The value is automatically parsed:
  - "true" or "false" → boolean
  - Numbers → numeric value
  - JSON arrays/objects → parsed JSON
  - Everything else → string

Use --null to set a null value.

Limits:
  - Key length: up to 42 bytes
  - Value size: up to 256 bytes (strings)

```
shelly kvs set <device> <key> <value> [flags]
```

### Examples

```
  # Set a string value
  shelly kvs set living-room my_key "my_value"

  # Set a numeric value
  shelly kvs set living-room counter 42

  # Set a boolean value
  shelly kvs set living-room enabled true

  # Set a null value
  shelly kvs set living-room cleared --null

  # Set a JSON object
  shelly kvs set living-room config '{"timeout":30}'
```

### Options

```
  -h, --help   help for set
      --null   Set null value
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

* [shelly kvs](shelly_kvs.md)	 - Manage device key-value storage

