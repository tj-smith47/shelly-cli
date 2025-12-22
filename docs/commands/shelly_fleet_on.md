## shelly fleet on

Turn on devices via cloud

### Synopsis

Turn on devices through Shelly Cloud.

Uses cloud WebSocket connections to send commands, allowing control
of devices even when not on the same local network.

Requires integrator credentials configured via environment variables or config:
  SHELLY_INTEGRATOR_TAG - Your integrator tag
  SHELLY_INTEGRATOR_TOKEN - Your integrator token

```
shelly fleet on [device...] [flags]
```

### Examples

```
  # Turn on specific device
  shelly fleet on device-id

  # Turn on all devices in a group
  shelly fleet on --group living-room

  # Turn on all relay devices
  shelly fleet on --all
```

### Options

```
      --all            Turn on all relay devices
  -g, --group string   Turn on devices in group
  -h, --help           help for on
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

* [shelly fleet](shelly_fleet.md)	 - Cloud-based fleet management

