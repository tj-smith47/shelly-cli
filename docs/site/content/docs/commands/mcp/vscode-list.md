---
title: "shelly mcp vscode list"
description: "shelly mcp vscode list"
---

## shelly mcp vscode list

Show VSCode MCP servers

### Synopsis

Show all MCP servers configured in VSCode

```
shelly mcp vscode list [flags]
```

### Options

```
      --config-path string   Path to VSCode config file
  -h, --help                 help for list
      --workspace            List from workspace settings (.vscode/mcp.json) instead of user settings
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

* [shelly mcp vscode](shelly_mcp_vscode.md)	 - Manage VSCode MCP servers

