## shelly kvs delete

Delete a KVS key

### Synopsis

Remove a key-value pair from the device's Key-Value Storage.

This operation is irreversible. Use --yes to skip the confirmation prompt.

```
shelly kvs delete <device> <key> [flags]
```

### Examples

```
  # Delete a key (with confirmation)
  shelly kvs delete living-room my_key

  # Delete without confirmation
  shelly kvs delete living-room my_key --yes
```

### Options

```
  -h, --help   help for delete
  -y, --yes    Skip confirmation prompt
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

