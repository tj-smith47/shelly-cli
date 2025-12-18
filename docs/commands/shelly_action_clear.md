## shelly action clear

Clear an action URL for a Gen1 device

### Synopsis

Clear (remove) an action URL for a Gen1 Shelly device.

This removes the configured URL for the specified action, disabling the
HTTP callback for that event.

Note: This feature is currently in development.

Workaround: Use curl to clear action URLs directly:
  curl "http://<device-ip>/settings?<action>="

```
shelly action clear <device> <action> [flags]
```

### Examples

```
  # Clear output on action
  shelly action clear living-room out_on_url

  # Workaround: use curl
  curl "http://192.168.1.100/settings?out_on_url="
```

### Options

```
  -h, --help   help for clear
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

* [shelly action](shelly_action.md)	 - Manage Gen1 device action URLs

