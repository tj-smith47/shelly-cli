## shelly script template

Manage script templates

### Synopsis

Manage JavaScript script templates for Shelly devices.

Script templates provide reusable JavaScript code that can be installed
on Gen2+ devices. Templates include configurable variables that are
substituted during installation.

The CLI includes built-in templates for common use cases:
  - motion-light: Motion-activated lighting
  - power-monitor: Power consumption alerts
  - schedule-helper: Simple on/off scheduling
  - toggle-sync: Synchronize multiple switches
  - energy-logger: Log energy usage to KVS

### Examples

```
  # List available script templates
  shelly script template list

  # Show template details and code
  shelly script template show motion-light

  # Install a template on a device
  shelly script template install living-room motion-light

  # Install with interactive configuration
  shelly script template install living-room motion-light --configure
```

### Options

```
  -h, --help   help for template
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

* [shelly script](shelly_script.md)	 - Manage device scripts
* [shelly script template install](shelly_script_template_install.md)	 - Install a script template on a device
* [shelly script template list](shelly_script_template_list.md)	 - List available script templates
* [shelly script template show](shelly_script_template_show.md)	 - Show script template details

