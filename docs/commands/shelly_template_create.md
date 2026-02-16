## shelly template create

Create a template from a device

### Synopsis

Create a configuration template by capturing settings from a device.

The template captures:
  - Device configuration and settings
  - Component configurations (switches, lights, etc.)
  - Schedules and webhooks
  - Script configurations (not code)

WiFi credentials are excluded by default for security.
Use --include-wifi to include them.

```
shelly template create <name> <device> [flags]
```

### Examples

```
  # Create a template from a device
  shelly template create my-config living-room

  # Create with description
  shelly template create my-config living-room --description "Standard switch config"

  # Include WiFi settings
  shelly template create my-config living-room --include-wifi

  # Overwrite existing template
  shelly template create my-config living-room --force
```

### Options

```
  -d, --description string   Template description
  -f, --force                Overwrite existing template
  -h, --help                 help for create
      --include-wifi         Include WiFi credentials (security risk)
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

* [shelly template](shelly_template.md)	 - Manage device configuration templates

