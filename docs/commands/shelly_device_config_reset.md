## shelly device config reset

Reset configuration to defaults

### Synopsis

Reset device or component configuration to factory defaults.

Without a component argument, shows available components that can be reset.
With a component argument, resets that component's configuration.

Note: This does not perform a full factory reset. For that, use:
  shelly device factory-reset <device>

```
shelly device config reset <device> [component] [flags]
```

### Examples

```
  # Reset switch:0 to defaults
  shelly config reset living-room switch:0

  # Reset with confirmation skipped
  shelly config reset living-room switch:0 --yes
```

### Options

```
  -h, --help   help for reset
  -y, --yes    Skip confirmation prompt
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

* [shelly device config](shelly_device_config.md)	 - Manage device configuration

