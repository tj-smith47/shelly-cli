---
title: "shelly scene show"
description: "shelly scene show"
---

## shelly scene show

Show scene details

### Synopsis

Display detailed information about a scene including all its actions.

```
shelly scene show <name> [flags]
```

### Examples

```
  # Show scene details
  shelly scene show movie-night

  # Output as JSON
  shelly scene show movie-night --output json

  # Using alias
  shelly scene info bedtime

  # Short form
  shelly sc show morning-routine
```

### Options

```
  -h, --help            help for show
  -o, --output string   Output format: table, json, yaml (default "table")
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly scene](shelly_scene.md)	 - Manage device scenes

