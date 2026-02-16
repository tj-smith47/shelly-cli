## shelly script template show

Show script template details

### Synopsis

Show details of a script template including its code.

Displays the template metadata, configurable variables, and the
JavaScript source code.

```
shelly script template show <name> [flags]
```

### Examples

```
  # Show template details
  shelly script template show motion-light

  # Show only the code (for piping)
  shelly script template show motion-light --code

  # Output as JSON
  shelly script template show motion-light -o json
```

### Options

```
      --code            Show only the code (for piping)
  -h, --help            help for show
  -o, --output string   Output format: table, json, yaml (default "table")
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
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly script template](shelly_script_template.md)	 - Manage script templates

