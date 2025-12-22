## shelly template list

List templates

### Synopsis

List all configured templates.

Output is formatted as a table by default. Use -o json or -o yaml for
structured output suitable for scripting.

```
shelly template list [flags]
```

### Examples

```
  # List all templates
  shelly template list

  # Output as JSON
  shelly template list -o json

  # Output as YAML
  shelly template list -o yaml
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly template](shelly_template.md)	 - Manage device configuration templates

