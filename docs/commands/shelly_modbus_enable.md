## shelly modbus enable

Enable Modbus-TCP server

### Synopsis

Enable the Modbus-TCP server on a Shelly device.

When enabled, the device exposes Modbus registers on TCP port 502.

```
shelly modbus enable <device> [flags]
```

### Examples

```
  # Enable Modbus
  shelly modbus enable kitchen
```

### Options

```
  -h, --help   help for enable
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

* [shelly modbus](shelly_modbus.md)	 - Manage Modbus-TCP configuration

