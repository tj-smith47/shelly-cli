## shelly kvs export

Export KVS data to file

### Synopsis

Export all key-value pairs from the device to a file.

If no file is specified, output is written to stdout.
The export format can be JSON (default) or YAML.

```
shelly kvs export <device> [file] [flags]
```

### Examples

```
  # Export to JSON file
  shelly kvs export living-room kvs-backup.json

  # Export to YAML file
  shelly kvs export living-room kvs-backup.yaml --format yaml

  # Export to stdout
  shelly kvs export living-room

  # Export to stdout as YAML
  shelly kvs export living-room --format yaml
```

### Options

```
  -f, --format string   Output format (json, yaml) (default "json")
  -h, --help            help for export
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly kvs](shelly_kvs.md)	 - Manage device key-value storage

