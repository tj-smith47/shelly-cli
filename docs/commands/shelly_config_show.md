## shelly config show

Display current CLI configuration

### Synopsis

Display CLI configuration settings.

Without arguments, shows all configuration.
With a key argument, shows only that specific setting (supports dot notation).

```
shelly config show [key] [flags]
```

### Examples

```
  # Show all configuration
  shelly config show

  # Show specific setting
  shelly config show defaults.timeout

  # Show a section
  shelly config show defaults

  # Output as JSON
  shelly config show -o json
```

### Options

```
  -h, --help   help for show
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

* [shelly config](shelly_config.md)	 - Manage CLI configuration

