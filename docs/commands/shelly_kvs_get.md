## shelly kvs get

Get a KVS value

### Synopsis

Retrieve a value from the device's Key-Value Storage.

Returns the value, its type, and the etag (version identifier).
Use --raw to output only the value without formatting.

```
shelly kvs get <device> <key> [flags]
```

### Examples

```
  # Get a value
  shelly kvs get living-room my_key

  # Get raw value only (for scripting)
  shelly kvs get living-room my_key --raw

  # Output as JSON
  shelly kvs get living-room my_key -o json
```

### Options

```
  -h, --help   help for get
  -r, --raw    Output raw value only
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly kvs](shelly_kvs.md)	 - Manage device key-value storage

