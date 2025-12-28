#!/bin/bash
# =============================================================================
# MIGRATE EXISTING DOCUMENTATION TO HUGO FORMAT
# =============================================================================
# Run this script from the docs/site directory
# Usage: ./scripts/migrate-docs.sh

set -euo pipefail

SOURCE_DOCS="../../docs"  # Relative to docs/site
TARGET_DOCS="content/docs"

echo "Migrating documentation from ${SOURCE_DOCS} to ${TARGET_DOCS}..."

# Function to add Hugo front matter to a file
add_frontmatter() {
    local source_file="$1"
    local target_file="$2"
    local title="$3"
    local description="$4"
    local weight="$5"

    echo "  Migrating: $source_file -> $target_file"

    # Create target directory if needed
    mkdir -p "$(dirname "$target_file")"

    # Create file with front matter
    {
        echo "---"
        echo "title: \"$title\""
        echo "description: \"$description\""
        echo "weight: $weight"
        echo "---"
        echo ""
        # Skip first line if it's a markdown heading (we use title from front matter)
        tail -n +2 "$source_file"
    } > "$target_file"
}

# Migrate configuration.md
if [ -f "$SOURCE_DOCS/configuration.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/configuration.md" \
        "$TARGET_DOCS/configuration/_index.md" \
        "Configuration Reference" \
        "Complete reference for configuring Shelly CLI" \
        30
fi

# Migrate themes.md
if [ -f "$SOURCE_DOCS/themes.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/themes.md" \
        "$TARGET_DOCS/themes/_index.md" \
        "Theme System" \
        "Customize Shelly CLI appearance with 280+ built-in themes" \
        60
fi

# Migrate plugins.md
if [ -f "$SOURCE_DOCS/plugins.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/plugins.md" \
        "$TARGET_DOCS/plugins/_index.md" \
        "Plugin System" \
        "Extend Shelly CLI with custom plugins" \
        50
fi

# Migrate development.md to contributing/
if [ -f "$SOURCE_DOCS/development.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/development.md" \
        "$TARGET_DOCS/contributing/development.md" \
        "Development Guide" \
        "Set up your development environment for contributing" \
        10
fi

# Migrate architecture.md to reference/
if [ -f "$SOURCE_DOCS/architecture.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/architecture.md" \
        "$TARGET_DOCS/reference/architecture.md" \
        "Project Architecture" \
        "Directory structure and code organization" \
        10
fi

# Migrate testing.md to contributing/
if [ -f "$SOURCE_DOCS/testing.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/testing.md" \
        "$TARGET_DOCS/contributing/testing.md" \
        "Testing Strategy" \
        "Testing approach and coverage requirements" \
        20
fi

# Migrate dependencies.md to reference/
if [ -f "$SOURCE_DOCS/dependencies.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/dependencies.md" \
        "$TARGET_DOCS/reference/dependencies.md" \
        "Dependencies" \
        "Library dependencies and their usage" \
        40
fi

# Migrate tui.md to guides/
if [ -f "$SOURCE_DOCS/tui.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/tui.md" \
        "$TARGET_DOCS/guides/tui-dashboard.md" \
        "TUI Dashboard" \
        "Interactive terminal dashboard for device monitoring" \
        40
fi

# Migrate tasmota.md to guides/
if [ -f "$SOURCE_DOCS/tasmota.md" ]; then
    add_frontmatter \
        "$SOURCE_DOCS/tasmota.md" \
        "$TARGET_DOCS/guides/tasmota.md" \
        "Tasmota Devices" \
        "Working with Tasmota-flashed devices" \
        50
fi

echo "✓ Documentation migration complete"
echo ""
echo "Next steps:"
echo "  1. Review migrated files for formatting issues"
echo "  2. Fix any broken internal links"
echo "  3. Run 'hugo server' to verify rendering"
