## shelly template export

Export a template to a file

### Synopsis

Export a configuration template to a JSON or YAML file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension, or can be specified with --format.

```
shelly template export <name> [file] [flags]
```

### Examples

```
  # Export to YAML file
  shelly template export my-config template.yaml

  # Export to JSON file
  shelly template export my-config template.json

  # Export to stdout as JSON
  shelly template export my-config --format json

  # Export to stdout as YAML
  shelly template export my-config
```

### Options

```
  -f, --format string   Output format: json, yaml (default "yaml")
  -h, --help            help for export
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

* [shelly template](shelly_template.md)	 - Manage device configuration templates

