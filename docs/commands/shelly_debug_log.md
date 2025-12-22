## shelly debug log

Get device debug log (Gen1)

### Synopsis

Get the debug log from a Gen1 Shelly device.

This command only works with Gen1 devices. Gen2+ devices use a
different logging mechanism via WebSocket or RPC.

Gen1 debug logs can help diagnose connectivity issues, action URL problems,
and other device behavior.

```
shelly debug log <device> [flags]
```

### Examples

```
  # Get debug log from a Gen1 device
  shelly debug log living-room-gen1

  # For Gen2+ devices, use RPC instead
  shelly debug rpc living-room Sys.GetStatus
```

### Options

```
  -h, --help   help for log
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

* [shelly debug](shelly_debug.md)	 - Debug and diagnostic commands

