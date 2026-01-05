---
title: "Shelly CLI"
description: "A powerful, intuitive command-line interface for managing Shelly smart home devices"
---

<div class="hero-section text-center">
<div class="hero-content">

<img src="images/shelly-gopher.png" alt="Shelly CLI Gopher" class="hero-logo">

# Shelly CLI

A powerful, intuitive command-line interface for managing Shelly smart home devices.

<p class="hero-description">
Control Gen1, Gen2, Gen3, and Gen4 Shelly devices with a unified CLI. Features batch operations, TUI dashboard, 280+ themes, and a plugin system.
</p>

<div class="hero-buttons">
{{< button href="/docs/getting-started/installation/" >}}
Get Started
{{< /button >}}

{{< button href="/docs/commands/" outline="true" >}}
Command Reference
{{< /button >}}
</div>

</div>
</div>

<div class="features-section">
<div class="container">
<div class="row">

{{< feature icon="cpu" title="Full Device Support" >}}
Control all Shelly device generations (Gen1-4) with a unified interface. Switches, lights, covers, thermostats, RGB, and more.
{{< /feature >}}

{{< feature icon="terminal" title="TUI Dashboard" >}}
Interactive terminal dashboard inspired by k9s. Real-time device status, quick controls, and keyboard-driven navigation.
{{< /feature >}}

{{< feature icon="zap" title="Batch Operations" >}}
Control multiple devices simultaneously. Create groups, define scenes, and execute batch commands with concurrent execution.
{{< /feature >}}

{{< feature icon="puzzle" title="Plugin System" >}}
Extend functionality with custom plugins using the gh-style architecture. Create, install, and share plugins easily.
{{< /feature >}}

{{< feature icon="palette" title="280+ Themes" >}}
Built-in theme support via bubbletint. Choose from Dracula, Nord, Gruvbox, and 280+ other themes, or create your own.
{{< /feature >}}

{{< feature icon="code" title="Scriptable" >}}
JSON, YAML, CSV, and template output formats. Shell completions for bash, zsh, fish, and PowerShell. Built for automation.
{{< /feature >}}

</div>
</div>
</div>

## Quick Start

```bash
# Install via Homebrew (macOS/Linux)
brew install tj-smith47/tap/shelly-cli

# Or via Go
go install github.com/tj-smith47/shelly-cli/cmd/shelly@latest

# Initialize configuration
shelly init

# Add a device to your registry
shelly device add living-room 192.168.1.100

# Control devices with quick commands
shelly on living-room
shelly off living-room
shelly toggle living-room

# Check device status
shelly status living-room

# Launch the TUI dashboard
shelly dash
```

## Features at a Glance

| Category | Features |
|----------|----------|
| **Device Control** | Switch, Light, Cover, Thermostat, RGB, Input, Sensor components |
| **Discovery** | mDNS, BLE, CoIoT automatic device discovery |
| **Automation** | Scenes, Schedules, Scripts, Webhooks, Actions |
| **Monitoring** | Energy tracking, Power monitoring, Prometheus metrics export |
| **Protocols** | BTHome, Zigbee, Matter, LoRa smart home protocol support |
| **Output** | Table, JSON, YAML, CSV, Go template output formats |
| **Extensibility** | Plugin system, command aliases, custom themes |

---

<div class="text-center">

[Documentation](/docs/) ·
[Command Reference](/docs/commands/) ·
[GitHub](https://github.com/tj-smith47/shelly-cli) ·
[Examples](/examples/)

</div>
