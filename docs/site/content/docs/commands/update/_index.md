---
title: "shelly update"
description: "shelly update"
weight: 810
sidebar:
  collapsed: true
---

## shelly update

Update shelly to the latest version

### Synopsis

Update shelly to the latest version.

By default, downloads and installs the latest stable release from GitHub.
Use --check to only check for updates without installing.
Use --version to install a specific version.

```
shelly update [flags]
```

### Examples

```
  # Check for updates
  shelly update --check

  # Check for updates and refresh the cache
  shelly update --check --force

  # Update to latest version
  shelly update

  # Update to a specific version
  shelly update --version v1.2.0

  # Update with pre-releases
  shelly update --include-pre

  # Update without confirmation
  shelly update --yes
```

### Options

```
      --channel string   Release channel (stable, beta) (default "stable")
  -c, --check            Check for updates without installing
  -f, --force            Force refresh and update the version cache
  -h, --help             help for update
      --include-pre      Include pre-release versions
      --rollback         Rollback to previous version
      --version string   Install a specific version
  -y, --yes              Skip confirmation prompt
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices

