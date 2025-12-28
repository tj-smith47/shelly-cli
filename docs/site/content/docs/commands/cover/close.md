---
title: "shelly cover close"
description: "shelly cover close"
---

## shelly cover close

Close cover

### Synopsis

Close a cover/roller component on the specified device.

```
shelly cover close <device> [flags]
```

### Examples

```
  # Close cover fully
  shelly cover close bedroom

  # Close cover for 5 seconds
  shelly cover down bedroom --duration 5
```

### Options

```
  -d, --duration int   Duration in seconds (0 = full close)
  -h, --help           help for close
  -i, --id int         Cover component ID (default 0)
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

* [shelly cover](shelly_cover.md)	 - Control cover/roller components

