---
title: "shelly dash events"
description: "shelly dash events"
---

## shelly dash events

Launch dashboard in events view

### Synopsis

Launch the TUI dashboard directly in the events view.

The events view shows a real-time stream of device events including
button presses, state changes, errors, and other notifications.

```
shelly dash events [flags]
```

### Examples

```
  # Launch events view
  shelly dash events

  # With filter
  shelly dash events --filter switch

  # With custom refresh
  shelly dash events --refresh 1
```

### Options

```
      --filter string   Filter devices by name pattern
  -h, --help            help for events
      --refresh int     Data refresh interval in seconds (default 5)
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

* [shelly dash](shelly_dash.md)	 - Launch interactive TUI dashboard

