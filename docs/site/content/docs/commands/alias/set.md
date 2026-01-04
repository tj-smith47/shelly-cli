---
title: "shelly alias set"
description: "shelly alias set"
---

## shelly alias set

Create or update a command alias

### Synopsis

Create or update a command alias.

The command can include argument placeholders:
  - $1, $2, etc. are replaced with positional arguments
  - $@ is replaced with all arguments

Shell aliases (prefixed with ! or using --shell) are executed in your shell.

```
shelly alias set <name> <command> [flags]
```

### Examples

```
  # Simple alias
  shelly alias set lights "batch on living-room kitchen bedroom"

  # Alias with argument interpolation
  shelly alias set sw "switch $1 $2"
  # Usage: shelly sw on kitchen -> switch on kitchen

  # Alias with all arguments
  shelly alias set dev "device $@"
  # Usage: shelly dev info kitchen -> device info kitchen

  # Shell alias
  shelly alias set backup --shell 'tar -czf shelly-backup.tar.gz ~/.config/shelly'
```

### Options

```
  -h, --help    help for set
  -s, --shell   Create a shell alias (executes in shell)
```

### Options inherited from parent commands

```
      --config string           Config file (default $HOME/.config/shelly/config.yaml)
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

* [shelly alias](shelly_alias.md)	 - Manage command aliases

