---
title: "shelly template import"
description: "shelly template import"
---

## shelly template import

Import a template from a file

### Synopsis

Import a configuration template from a JSON or YAML file.

If no name is specified, the template name from the file is used.
Use --force to overwrite an existing template with the same name.

```
shelly template import <file> [name] [flags]
```

### Examples

```
  # Import a template
  shelly template import template.yaml

  # Import with a different name
  shelly template import template.yaml my-new-config

  # Overwrite existing template
  shelly template import template.yaml --force
```

### Options

```
  -f, --force   Overwrite existing template
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
      --refresh                 Bypass cache and fetch fresh data from device
      --template string         Go template string for output (use with -o template)
  -v, --verbose count           Increase verbosity (-v=info, -vv=debug, -vvv=trace)
```

### SEE ALSO

* [shelly template](shelly_template.md)	 - Manage device configuration templates

