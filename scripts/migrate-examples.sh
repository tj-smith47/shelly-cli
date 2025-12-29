#!/usr/bin/env bash
# migrate-examples.sh - Migrate examples/ directory to Hugo site content
#
# This script reads example files from examples/ and generates Hugo-compatible
# markdown files with proper front matter and embedded code blocks.

set -euo pipefail

EXAMPLES_DIR="examples"
TARGET_DIR="docs/site/content/examples"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[examples]${NC} $1"
}

success() {
    echo -e "${GREEN}[examples]${NC} $1"
}

# Clean and create target directories
log "Cleaning target directory..."
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR"/{aliases,configuration,scripts,plugins,deployments}

# =============================================================================
# Landing Page
# =============================================================================
log "Creating examples landing page..."
cat > "$TARGET_DIR/_index.md" << 'EOF'
---
title: "Examples"
description: "Example configurations, scripts, and plugins for Shelly CLI"
---

Ready-to-use examples to help you get started with Shelly CLI. Browse by category or check out the [examples directory on GitHub](https://github.com/tj-smith47/shelly-cli/tree/main/examples).

## Categories

- [Aliases](aliases/) - Pre-built command shortcuts for common workflows
- [Configuration](configuration/) - Configuration file examples from minimal to full
- [Scripts](scripts/) - Shell scripts for automation and batch operations
- [Plugins](plugins/) - Example plugin implementations
- [Deployments](deployments/) - Docker and Kubernetes deployment examples

## Quick Start

```bash
# Import example aliases
shelly alias import examples/aliases/shortcuts.yaml

# Use an example config
cp examples/config/minimal.yaml ~/.config/shelly/config.yaml
```
EOF

# =============================================================================
# Aliases Section
# =============================================================================
log "Processing alias examples..."
cat > "$TARGET_DIR/aliases/_index.md" << 'EOF'
---
title: "Alias Examples"
description: "Pre-built alias collections for common workflows"
weight: 10
---

Command alias collections ranging from simple shortcuts to advanced automation helpers.

## Usage

```bash
# Import any alias file
shelly alias import examples/aliases/shortcuts.yaml

# List imported aliases
shelly alias list

# Use an alias
shelly ls  # if 'ls' was imported as alias for 'device list'
```

---

EOF

for file in "$EXAMPLES_DIR"/aliases/*.yaml; do
    filename=$(basename "$file")
    name="${filename%.yaml}"

    # Extract description from first comment line
    desc=$(head -1 "$file" | sed 's/^# *//')

    cat >> "$TARGET_DIR/aliases/_index.md" << EOF
## $filename

$desc

\`\`\`yaml
$(cat "$file")
\`\`\`

---

EOF
done

# =============================================================================
# Configuration Section
# =============================================================================
log "Processing configuration examples..."
cat > "$TARGET_DIR/configuration/_index.md" << 'EOF'
---
title: "Configuration Examples"
description: "Configuration file examples from minimal to comprehensive"
weight: 20
---

Example configuration files showing different levels of complexity.

## Usage

```bash
# Copy a config to start with
cp examples/config/minimal.yaml ~/.config/shelly/config.yaml

# Edit to add your devices
shelly config edit

# Validate configuration
shelly init --check
```

---

EOF

for file in "$EXAMPLES_DIR"/config/*.yaml; do
    filename=$(basename "$file")
    name="${filename%.yaml}"

    # Extract description from first comment line
    desc=$(head -1 "$file" | sed 's/^# *//')

    cat >> "$TARGET_DIR/configuration/_index.md" << EOF
## $filename

$desc

\`\`\`yaml
$(cat "$file")
\`\`\`

---

EOF
done

# =============================================================================
# Scripts Section
# =============================================================================
log "Processing script examples..."
cat > "$TARGET_DIR/scripts/_index.md" << 'EOF'
---
title: "Script Examples"
description: "Shell scripts for automation and batch operations"
weight: 30
---

Shell scripts demonstrating advanced automation patterns with Shelly CLI.

## Usage

```bash
# Make executable
chmod +x examples/scripts/bulk-update.sh

# Run with dry-run first
./examples/scripts/bulk-update.sh --dry-run

# Run for real
./examples/scripts/bulk-update.sh
```

---

EOF

for file in "$EXAMPLES_DIR"/scripts/*.sh; do
    filename=$(basename "$file")
    name="${filename%.sh}"

    # Extract description from script header (first non-shebang comment)
    desc=$(grep -m1 "^# " "$file" | tail -1 | sed 's/^# *//')

    cat >> "$TARGET_DIR/scripts/_index.md" << EOF
## $filename

$desc

\`\`\`bash
$(cat "$file")
\`\`\`

---

EOF
done

# =============================================================================
# Plugins Section
# =============================================================================
log "Processing plugin examples..."
cat > "$TARGET_DIR/plugins/_index.md" << 'EOF'
---
title: "Plugin Examples"
description: "Example plugin implementations for extending Shelly CLI"
weight: 40
---

Example plugins demonstrating the plugin architecture. Each plugin includes a manifest and implementation.

## Plugin Structure

```
shelly-example/
├── manifest.json     # Plugin metadata and hooks
├── README.md         # Documentation
└── shelly-example    # Executable (bash, go binary, python, etc.)
```

## Installation

```bash
# Copy plugin to plugins directory
cp -r examples/plugins/shelly-notify ~/.config/shelly/plugins/

# Enable plugins
shelly config set plugins.enabled true

# List installed plugins
shelly plugin list
```

---

EOF

for plugin_dir in "$EXAMPLES_DIR"/plugins/*/; do
    plugin_name=$(basename "$plugin_dir")

    # Skip if no manifest
    if [[ ! -f "$plugin_dir/manifest.json" ]]; then
        continue
    fi

    # Extract description from manifest
    desc=$(jq -r '.description // "No description"' "$plugin_dir/manifest.json" 2>/dev/null || echo "Plugin example")

    cat >> "$TARGET_DIR/plugins/_index.md" << EOF
## $plugin_name

$desc

### Manifest

\`\`\`json
$(cat "$plugin_dir/manifest.json")
\`\`\`

EOF

    # Add README content if exists
    if [[ -f "$plugin_dir/README.md" ]]; then
        cat >> "$TARGET_DIR/plugins/_index.md" << EOF
### Documentation

$(tail -n +2 "$plugin_dir/README.md")

EOF
    fi

    echo "---" >> "$TARGET_DIR/plugins/_index.md"
    echo "" >> "$TARGET_DIR/plugins/_index.md"
done

# =============================================================================
# Deployments Section
# =============================================================================
log "Processing deployment examples..."
cat > "$TARGET_DIR/deployments/_index.md" << 'EOF'
---
title: "Deployment Examples"
description: "Docker and Kubernetes deployment examples"
weight: 50
---

Container and orchestration examples for running Shelly CLI as a metrics exporter.

EOF

# Add main deployments README content
if [[ -f "$EXAMPLES_DIR/deployments/README.md" ]]; then
    tail -n +2 "$EXAMPLES_DIR/deployments/README.md" >> "$TARGET_DIR/deployments/_index.md"
    echo "" >> "$TARGET_DIR/deployments/_index.md"
fi

cat >> "$TARGET_DIR/deployments/_index.md" << 'EOF'

---

## Docker Compose

EOF

if [[ -f "$EXAMPLES_DIR/deployments/docker/docker-compose.yml" ]]; then
    cat >> "$TARGET_DIR/deployments/_index.md" << EOF
\`\`\`yaml
$(cat "$EXAMPLES_DIR/deployments/docker/docker-compose.yml")
\`\`\`

EOF
fi

if [[ -f "$EXAMPLES_DIR/deployments/docker/README.md" ]]; then
    tail -n +2 "$EXAMPLES_DIR/deployments/docker/README.md" >> "$TARGET_DIR/deployments/_index.md"
    echo "" >> "$TARGET_DIR/deployments/_index.md"
fi

cat >> "$TARGET_DIR/deployments/_index.md" << 'EOF'

---

## Kubernetes

EOF

if [[ -f "$EXAMPLES_DIR/deployments/kubernetes/README.md" ]]; then
    tail -n +2 "$EXAMPLES_DIR/deployments/kubernetes/README.md" >> "$TARGET_DIR/deployments/_index.md"
    echo "" >> "$TARGET_DIR/deployments/_index.md"
fi

# Add key kubernetes manifests
for manifest in deployment service servicemonitor; do
    if [[ -f "$EXAMPLES_DIR/deployments/kubernetes/$manifest.yaml" ]]; then
        cat >> "$TARGET_DIR/deployments/_index.md" << EOF
### $manifest.yaml

\`\`\`yaml
$(cat "$EXAMPLES_DIR/deployments/kubernetes/$manifest.yaml")
\`\`\`

EOF
    fi
done

# =============================================================================
# Summary
# =============================================================================
alias_count=$(find "$EXAMPLES_DIR/aliases" -name "*.yaml" 2>/dev/null | wc -l | tr -d ' ')
config_count=$(find "$EXAMPLES_DIR/config" -name "*.yaml" 2>/dev/null | wc -l | tr -d ' ')
script_count=$(find "$EXAMPLES_DIR/scripts" -name "*.sh" 2>/dev/null | wc -l | tr -d ' ')
plugin_count=$(find "$EXAMPLES_DIR/plugins" -maxdepth 1 -type d 2>/dev/null | tail -n +2 | wc -l | tr -d ' ')

success "Examples migration complete:"
echo "  Aliases: $alias_count"
echo "  Configs: $config_count"
echo "  Scripts: $script_count"
echo "  Plugins: $plugin_count"
echo "  Deployments: docker, kubernetes"
echo ""
echo "Hugo site updated in $TARGET_DIR/"
