## shelly switch

Control switch components

### Synopsis

Control Shelly switch (relay) components.

Switches are the basic on/off relays found in most Shelly devices.
Use these commands to control individual switches or list all
switches on a device.

### Examples

```
  # Turn on a switch
  shelly switch on kitchen

  # Turn off a switch
  shelly sw off living-room

  # Check switch status
  shelly switch status bedroom
```

### Options

```
  -h, --help   help for switch
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly switch list](shelly_switch_list.md)	 - List switch components
* [shelly switch off](shelly_switch_off.md)	 - Turn switch off
* [shelly switch on](shelly_switch_on.md)	 - Turn switch on
* [shelly switch status](shelly_switch_status.md)	 - Show switch status
* [shelly switch toggle](shelly_switch_toggle.md)	 - Toggle switch on/off

