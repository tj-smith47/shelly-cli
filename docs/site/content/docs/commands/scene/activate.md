---
title: "shelly scene activate"
description: "shelly scene activate"
---

## shelly scene activate

Activate a scene

### Synopsis

Execute all actions defined in a scene.

Actions are executed concurrently for faster execution.
Use --dry-run to preview actions without executing them.

```
shelly scene activate <name> [flags]
```

### Examples

```
  # Activate a scene
  shelly scene activate movie-night

  # Preview without executing
  shelly scene activate movie-night --dry-run

  # Using aliases
  shelly scene run bedtime
  shelly scene play morning-routine

  # Short form
  shelly sc activate party-mode
```

### Options

```
  -c, --concurrent int     Max concurrent operations (default 5)
      --dry-run            Preview actions without executing
  -h, --help               help for activate
  -t, --timeout duration   Timeout per device (default 10s)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
  -F, --fields                  Print available field names for use with --jq and --template
  -Q, --jq stringArray          Apply jq expression to filter output (repeatable, joined with |)
      --log-categories string   Filter logs by category (comma-separated: network,api,device,config,auth,plugin)
      --log-json                Output logs in JSON format
      --no-color                Disable colored output
      --no-headers              Hide table headers in output
      --offline                 Only read from cache, error on cache miss
  -o, --output string           Output format (table, json, yaml, template) (default "table")
      --plain                   Disable borders and colors (machine-readable output)
  -q, --quiet                   Suppress non-essential output
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly scene](shelly_scene.md)	 - Manage device scenes

