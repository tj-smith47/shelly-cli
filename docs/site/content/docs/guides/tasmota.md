---
title: "Tasmota Devices"
description: "Working with Tasmota-flashed devices"
weight: 50
---


The Tasmota plugin enables shelly-cli to discover, control, and update Tasmota-based smart home devices.

For full documentation, see the plugin README:

**[examples/plugins/shelly-tasmota/README.md](../examples/plugins/shelly-tasmota/README.md)**

## Quick Reference

### Installation

```bash
cd examples/plugins/shelly-tasmota
go build -o shelly-tasmota .
shelly plugin install ./
```

### Discovery

```bash
# Discover all devices (Shelly and Tasmota)
shelly discover

# Discover only Tasmota devices
shelly discover --platform tasmota
```

### Control

```bash
shelly switch on <device>
shelly switch off <device>
shelly switch toggle <device>
```

### Firmware Updates

```bash
shelly firmware check <device>
shelly firmware updates --platform tasmota
```

## See Also

- [Plugin Development Guide](plugins.md) - How to create plugins
- [Plugin README](../examples/plugins/shelly-tasmota/README.md) - Full Tasmota documentation
