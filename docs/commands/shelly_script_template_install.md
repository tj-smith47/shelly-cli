## shelly script template install

Install a script template on a device

### Synopsis

Install a script template on a Shelly device.

Creates a new script on the device with the template code. Template
variables are substituted with their default values, or you can use
--configure for interactive configuration.

```
shelly script template install <device> <template> [flags]
```

### Examples

```
  # Install with default values
  shelly script template install living-room motion-light

  # Install with interactive configuration
  shelly script template install living-room motion-light --configure

  # Install and enable immediately
  shelly script template install living-room motion-light --enable

  # Install with custom script name
  shelly script template install living-room motion-light --name "Motion Sensor"
```

### Options

```
      --configure     Interactive variable configuration
      --enable        Enable script after installation
  -h, --help          help for install
      --name string   Custom script name (defaults to template name)
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly script template](shelly_script_template.md)	 - Manage script templates

