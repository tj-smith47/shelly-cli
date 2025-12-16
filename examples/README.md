# Shelly CLI Examples

Example configurations, scripts, and plugins to help you get started with the Shelly CLI.

## Quick Start

```bash
# Import example aliases
shelly alias import examples/aliases/shortcuts.yaml

# Use an example config
cp examples/config/minimal.yaml ~/.config/shelly/config.yaml
```

## Directory Structure

### aliases/

Pre-built alias collections for common workflows:

| File | Description |
|------|-------------|
| `shortcuts.yaml` | Common shortcuts (`ls`, `info`, `boot`, `find`, `power`) |
| `power-users.yaml` | Power user aliases with JSON output and filtering |
| `automation.yaml` | Scripting and automation aliases (`qon`, `qoff`, scenes) |

**Usage:**
```bash
# Import all shortcuts
shelly alias import examples/aliases/shortcuts.yaml

# List imported aliases
shelly alias list
```

### config/

Configuration file examples:

| File | Description |
|------|-------------|
| `minimal.yaml` | Minimal 3-device setup to get started |
| `multi-site.yaml` | Multi-location with groups, scenes, and aliases |
| `full.yaml` | Complete reference with all configuration options |

**Usage:**
```bash
# Copy a config to start with
cp examples/config/minimal.yaml ~/.config/shelly/config.yaml

# Edit to add your devices
shelly config edit

# Validate configuration
shelly init --check
```

### scripts/

Shell scripts for advanced automation:

| File | Description |
|------|-------------|
| `bulk-update.sh` | Bulk firmware updates with backups and safety checks |
| `workstation-sync.sh` | Sync desk lights with computer screen lock (macOS/Linux) |
| `presence-detect.sh` | Network presence-based home automation via ping |

**Usage:**
```bash
# Make executable
chmod +x examples/scripts/bulk-update.sh

# Run with dry-run first
./examples/scripts/bulk-update.sh --dry-run

# Run for real
./examples/scripts/bulk-update.sh
```

### plugins/

Example plugin implementations:

| Plugin | Description |
|--------|-------------|
| `shelly-notify/` | Desktop notifications for device events |

**Usage:**
```bash
# Install a plugin
cp -r examples/plugins/shelly-notify ~/.config/shelly/plugins/

# Enable plugins
shelly config set plugins.enabled true

# Test the plugin
shelly notify test
```

### deployments/

Container and orchestration examples:

- Kubernetes deployment with Prometheus metrics export
- Docker and Docker Compose examples
- Prometheus scrape configuration

## Configuration Schema

The JSON Schema for configuration validation is located at `cfg/config.schema.json`.

Use it with a YAML language server for IDE autocompletion:
```yaml
# yaml-language-server: $schema=../../cfg/config.schema.json
```

## See Also

- [Configuration Reference](../docs/configuration.md)
- [Plugin Development Guide](../docs/plugins.md)
- [Theme Customization](../docs/themes.md)
- [Full Documentation](../docs/)
