---
title: "shelly group set"
description: "shelly group set"
---

## shelly group set

Set light parameters on all group members

### Synopsis

Set light parameters on every device in a group.

You can set brightness, white color temperature (Gen1 white-temp bulbs such as
the Duo), and on/off state. Values not specified are left unchanged. The change
is fanned out to all members concurrently and a per-member result summary is
printed.

Unlike on/off/toggle, --id targets a single light component (default 0) on each
member rather than all components.

```
shelly group set <group> [flags]
```

### Examples

```
  # Set every member to 100% and turn on
  shelly group set guest-bath-bulbs -b 100 --on

  # Set white color temperature to 4200K on all members (Gen1 Duo)
  shelly group set master-bath -t 4200

  # Target component 1 on every member
  shelly group set living-room -b 50 --id 1
```

### Options

```
  -b, --brightness int   Brightness (0-100) (default -1)
  -c, --concurrent int   Max concurrent operations (default 5)
  -h, --help             help for set
  -i, --id int           Light component ID (default 0)
      --on               Turn on
  -t, --temp int         White color temperature in Kelvin (Gen1 Duo: 2700-6500) (default -1)
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

* [shelly group](shelly_group.md)	 - Manage device groups

