---
title: "shelly theme import"
description: "shelly theme import"
---

## shelly theme import

Import theme from file

### Synopsis

Import a theme configuration from a file.

Supports importing theme files that reference any of the 280+ built-in themes.

```
shelly theme import <file> [flags]
```

### Examples

```
  # Import and apply a theme
  shelly theme import mytheme.yaml --apply

  # Just validate the file
  shelly theme import mytheme.yaml
```

### Options

```
      --apply   Apply the imported theme
  -h, --help    help for import
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

* [shelly theme](shelly_theme.md)	 - Manage CLI color themes

