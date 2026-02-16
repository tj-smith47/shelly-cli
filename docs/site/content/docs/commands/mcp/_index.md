---
title: "shelly mcp"
description: "shelly mcp"
weight: 380
sidebar:
  collapsed: true
---

## shelly mcp

MCP server for AI assistant integration

### Synopsis

MCP (Model Context Protocol) server for AI assistant integration.

This command allows AI assistants like Claude, Gemini, and others to interact
with your Shelly devices through the CLI. The MCP server exposes CLI commands
as tools that AI assistants can invoke.

Use 'mcp start' to run the MCP server, or 'mcp configure' to set up AI
assistant configuration files automatically.

### Examples

```
  # Start the MCP server
  shelly mcp start

  # Enable in Claude Desktop
  shelly mcp claude enable

  # Enable in VS Code
  shelly mcp vscode enable

  # Configure Gemini CLI (or other assistants)
  shelly mcp configure --gemini

  # List available MCP tools
  shelly mcp tools
```

### Options

```
  -h, --help   help for mcp
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

* [shelly](shelly.md)	 - CLI for controlling Shelly smart home devices
* [shelly mcp claude](shelly_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [shelly mcp configure](shelly_mcp_configure.md)	 - Configure AI assistant MCP integration
* [shelly mcp cursor](shelly_mcp_cursor.md)	 - Manage Cursor MCP servers
* [shelly mcp start](shelly_mcp_start.md)	 - Start the MCP server
* [shelly mcp stream](shelly_mcp_stream.md)	 - Stream the MCP server over HTTP
* [shelly mcp tools](shelly_mcp_tools.md)	 - Export tools as JSON
* [shelly mcp vscode](shelly_mcp_vscode.md)	 - Manage VSCode MCP servers

