---
title: "shelly plugin"
description: "shelly plugin"
weight: 440
sidebar:
  collapsed: true
---

## shelly plugin

Manage CLI plugins

### Synopsis

Manage CLI plugins (extensions).

Plugins are executable programs named shelly-<name> that extend the CLI.
They can be installed from local files, GitHub repositories, or URLs.

Installed plugins are stored in ~/.config/shelly/plugins/.

### Examples

```
  # List installed plugins
  shelly plugin list

  # Install from local file
  shelly plugin install ./shelly-myext

  # Install from GitHub
  shelly plugin install gh:user/shelly-myext

  # Remove a plugin
  shelly plugin remove myext

  # Run a plugin explicitly
  shelly plugin exec myext --some-flag

  # Create a new plugin scaffold
  shelly plugin create myext
```

### Options

```
  -h, --help   help for plugin
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly plugin create](shelly_plugin_create.md)	 - Create a new extension scaffold
* [shelly plugin exec](shelly_plugin_exec.md)	 - Execute an extension
* [shelly plugin install](shelly_plugin_install.md)	 - Install an extension
* [shelly plugin list](shelly_plugin_list.md)	 - List installed extensions
* [shelly plugin remove](shelly_plugin_remove.md)	 - Remove an installed extension
* [shelly plugin upgrade](shelly_plugin_upgrade.md)	 - Upgrade extension(s)

