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
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly template](shelly_template.md)	 - Manage device configuration templates

