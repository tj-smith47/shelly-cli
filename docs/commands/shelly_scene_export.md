## shelly scene export

Export a scene to file

### Synopsis

Export a scene definition to a file.

If no file is specified, outputs to stdout.
Format is auto-detected from file extension (.json, .yaml, .yml).

```
shelly scene export <name> [file] [flags]
```

### Examples

```
  # Export to YAML file
  shelly scene export movie-night scene.yaml

  # Export to JSON file
  shelly scene export movie-night scene.json

  # Export to stdout as YAML
  shelly scene export movie-night

  # Export to stdout as JSON
  shelly scene export movie-night --format json
```

### Options

```
  -f, --format string   Output format: json, yaml (default "yaml")
  -h, --help            help for export
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

* [shelly scene](shelly_scene.md)	 - Manage device scenes

