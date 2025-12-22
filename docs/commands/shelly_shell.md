## shelly shell

Interactive shell for a specific device

### Synopsis

Open an interactive shell for a specific Shelly device.

This provides direct access to execute RPC commands on the device.
It maintains a persistent connection and allows you to explore the
device's capabilities interactively. Supports readline-style line
editing (arrow keys, Ctrl+A/E, etc.) with command history.

Available commands:
  help           Show available commands
  info           Show device information
  status         Show device status
  config         Show device configuration
  methods        List available RPC methods
  components     List device components
  <method>       Execute RPC method (e.g., Switch.GetStatus, Shelly.GetConfig)
  exit           Close shell

RPC methods can be called directly by typing the method name.
For methods requiring parameters, provide JSON after the method name.

For multi-switch devices, use the RPC methods with component ID to
control specific switches.

```
shelly shell <device> [flags]
```

### Examples

```
  # Open shell for a device
  shelly shell living-room

  # Example session:
  shell> info
  shell> methods
  shell> Switch.GetStatus {"id":0}
  shell> Shelly.GetConfig
  shell> exit

  # Control specific switch on multi-switch device:
  shell> Switch.Set {"id":1,"on":true}
  shell> Switch.Toggle {"id":0}
  shell> Switch.GetStatus {"id":1}
```

### Options

```
  -h, --help   help for shell
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

