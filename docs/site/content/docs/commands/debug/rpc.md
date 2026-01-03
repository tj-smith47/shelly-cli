---
title: "shelly debug rpc"
description: "shelly debug rpc"
---

## shelly debug rpc

Execute a raw RPC call

### Synopsis

Execute a raw RPC call on a Shelly device.

This command allows you to call any RPC method supported by the device.
Use 'shelly debug methods <device>' to list available methods.

Parameters should be provided as a JSON object. If no parameters are
needed, omit the params argument or use '{}'.

```
shelly debug rpc <device> <method> [params_json] [flags]
```

### Examples

```
  # Get device info
  shelly debug rpc living-room Shelly.GetDeviceInfo

  # Get switch status with ID parameter
  shelly debug rpc living-room Switch.GetStatus '{"id":0}'

  # Set switch state
  shelly debug rpc living-room Switch.Set '{"id":0,"on":true}'

  # Get all methods
  shelly debug rpc living-room Shelly.ListMethods

  # Raw output (no formatting)
  shelly debug rpc living-room Shelly.GetStatus --raw
```

### Options

```
  -h, --help   help for rpc
      --raw    Output raw JSON without formatting
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

* [shelly debug](shelly_debug.md)	 - Debug and diagnostic commands

