## shelly toggle

Toggle a device (auto-detects type)

### Synopsis

Toggle a device by automatically detecting its type.

Works with switches, lights, covers, and RGB devices. For covers,
this toggles between open and close based on current state.

Use --all to toggle all controllable components on the device.

```
shelly toggle <device> [flags]
```

### Examples

```
  # Toggle a switch or light
  shelly toggle living-room

  # Toggle all components on a device
  shelly toggle living-room --all

  # Toggle a cover
  shelly toggle bedroom-blinds
```

### Options

```
  -a, --all    Toggle all controllable components
  -h, --help   help for toggle
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

