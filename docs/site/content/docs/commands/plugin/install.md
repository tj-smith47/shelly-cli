---
title: "shelly plugin install"
description: "shelly plugin install"
---

## shelly plugin install

Install an extension

### Synopsis

Install an extension from a local file, URL, or GitHub repository.

Supported sources:
  - Local file: ./path/to/shelly-myext
  - GitHub repo: gh:user/shelly-myext
  - HTTP URL: https://example.com/shelly-myext

The extension must be named with the shelly- prefix.

```
shelly plugin install <source> [flags]
```

### Examples

```
  # Install from local file
  shelly extension install ./shelly-myext

  # Install from GitHub (downloads latest release binary)
  shelly extension install gh:user/shelly-myext

  # Install from URL
  shelly extension install https://example.com/shelly-myext

  # Force reinstall
  shelly extension install ./shelly-myext --force
```

### Options

```
  -f, --force   Force reinstall even if already installed
  -h, --help    help for install
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

* [shelly plugin](shelly_plugin.md)	 - Manage CLI plugins

