## shelly config export

Export device configuration to a file

### Synopsis

Export the complete device configuration to a file.

The configuration is saved in JSON format by default. Use --format=yaml
for YAML output.

```
shelly config export <device> <file> [flags]
```

### Examples

```
  # Export to JSON file
  shelly config export living-room config-backup.json

  # Export to YAML file
  shelly config export living-room config-backup.yaml --format=yaml

  # Export to stdout
  shelly config export living-room -
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

* [shelly config](shelly_config.md)	 - Manage device configuration

