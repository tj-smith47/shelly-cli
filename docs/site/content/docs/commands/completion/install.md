---
title: "shelly completion install"
description: "shelly completion install"
---

## shelly completion install

Install shell completions

### Synopsis

Automatically install shell completions for shelly.

This command detects your current shell and installs completions
to the appropriate location. It also updates your shell configuration
to source the completions on startup.

Supported shells: bash, zsh, fish, powershell

```
shelly completion install [flags]
```

### Examples

```
  # Auto-detect shell and install
  shelly completion install

  # Install for specific shell
  shelly completion install --shell bash
  shelly completion install --shell zsh
```

### Options

```
  -h, --help           help for install
      --shell string   Shell to install completions for (auto-detected if not specified)
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

* [shelly completion](shelly_completion.md)	 - Generate shell completion scripts

