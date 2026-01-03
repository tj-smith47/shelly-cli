---
title: "shelly scene create"
description: "shelly scene create"
---

## shelly scene create

Create a new scene

### Synopsis

Create a new empty scene.

After creating a scene, use 'shelly scene add-action' to add device actions,
or import an existing scene definition from a file.

```
shelly scene create <name> [flags]
```

### Examples

```
  # Create a new scene
  shelly scene create movie-night

  # Create with description
  shelly scene create movie-night --description "Dim lights for movies"

  # Using alias
  shelly scene new bedtime

  # Short form
  shelly sc create morning-routine
```

### Options

```
  -d, --description string   Scene description
  -h, --help                 help for create
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

* [shelly scene](shelly_scene.md)	 - Manage device scenes

