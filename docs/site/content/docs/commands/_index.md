---
title: "Command Reference"
description: "Complete reference for all Shelly CLI commands"
weight: 20
sidebar:
  collapsed: true
---

Complete reference documentation for all Shelly CLI commands.

Use the sidebar to browse commands by category, or use search to find specific commands.

## Getting Help

Every command supports `--help`:

```bash
shelly --help
shelly device --help
shelly discover --help
shelly init --help
```

## Output Formats

Most commands support multiple output formats:

```bash
shelly status living-room             # Default table
shelly status living-room -o json     # JSON
shelly status living-room -o yaml     # YAML
```
