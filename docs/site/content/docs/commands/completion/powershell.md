---
title: "shelly completion powershell"
description: "shelly completion powershell"
---

## shelly completion powershell

Generate powershell completion script

### Synopsis

Generate the autocompletion script for PowerShell.

To load completions in your current shell session:

  shelly completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your PowerShell profile.

```
shelly completion powershell
```

### Examples

```
  shelly completion powershell > shelly.ps1
  . ./shelly.ps1
```

### Options

```
  -h, --help   help for powershell
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

* [shelly completion](shelly_completion.md)	 - Generate shell completion scripts

