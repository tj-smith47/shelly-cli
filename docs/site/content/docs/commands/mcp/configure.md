---
title: "shelly mcp configure"
description: "shelly mcp configure"
---

## shelly mcp configure

Configure AI assistant MCP integration

### Synopsis

Configure AI assistant MCP integration.

This command writes the necessary configuration to enable AI assistants to
use the Shelly CLI via MCP (Model Context Protocol).

Supported AI assistants:
  --claude-desktop  Configure Claude Desktop app
  --claude-code     Configure Claude Code (VS Code extension / CLI)
  --gemini          Configure Gemini CLI

You can specify multiple flags to configure all at once.

```
shelly mcp configure [flags]
```

### Examples

```
  # Configure Claude Desktop
  shelly mcp configure --claude-desktop

  # Configure Claude Code
  shelly mcp configure --claude-code

  # Configure all supported assistants
  shelly mcp configure --claude-desktop --claude-code --gemini

  # Preview changes without writing (dry run)
  shelly mcp configure --claude-desktop --dry-run

  # Use custom shelly binary path
  shelly mcp configure --claude-code --shelly-path /usr/local/bin/shelly
```

### Options

```
      --claude-code          Configure Claude Code
      --claude-desktop       Configure Claude Desktop
  -n, --dry-run              Preview changes without writing files
      --gemini               Configure Gemini CLI
  -h, --help                 help for configure
      --shelly-path string   Path to shelly binary (auto-detected if not specified)
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

* [shelly mcp](shelly_mcp.md)	 - MCP server for AI assistant integration

