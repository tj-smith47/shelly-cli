#!/usr/bin/env bash
# =============================================================================
# MIGRATE COMMAND DOCUMENTATION
# =============================================================================
# Processes 350+ command docs into grouped subdirectories for collapsible sidebar
# Run this script from the repository root
# Usage: ./scripts/migrate-commands.sh

set -euo pipefail

SOURCE_DIR="docs/commands"
TARGET_DIR="docs/site/content/docs/commands"

# Create parent command _index.md with collapsed sidebar
create_parent_index() {
    local parent="$1"
    local description="$2"
    local weight="$3"
    local source_file="$4"
    local target_dir="$TARGET_DIR/$parent"

    mkdir -p "$target_dir"

    cat > "$target_dir/_index.md" <<EOF
---
title: "shelly ${parent}"
description: "${description}"
weight: ${weight}
sidebar:
  collapsed: true
---

EOF
    cat "$source_file" >> "$target_dir/_index.md"
}

# Create child command file
create_child_file() {
    local parent="$1"
    local subcommand="$2"
    local cmd_name="$3"
    local description="$4"
    local source_file="$5"
    local target_file="$TARGET_DIR/$parent/$subcommand.md"

    cat > "$target_file" <<EOF
---
title: "shelly ${cmd_name}"
description: "${description}"
---

EOF
    cat "$source_file" >> "$target_file"
}

# Create root command file
create_root_file() {
    local filename="$1"
    local cmd_name="$2"
    local description="$3"
    local weight="$4"
    local source_file="$5"
    local target_file="$TARGET_DIR/${filename}.md"

    cat > "$target_file" <<EOF
---
title: "shelly ${cmd_name}"
description: "${description}"
weight: ${weight}
---

EOF
    cat "$source_file" >> "$target_file"
}

echo "Migrating command documentation with grouping..."

# Clean existing command files (preserve main _index.md)
find "$TARGET_DIR" -mindepth 2 -name "*.md" -delete 2>/dev/null || true
find "$TARGET_DIR" -mindepth 1 -type d -exec rm -rf {} + 2>/dev/null || true
mkdir -p "$TARGET_DIR"

# First pass: identify parent commands
declare -A parents
declare -A parent_weights

weight=10
for file in "$SOURCE_DIR"/*.md; do
    filename=$(basename "$file" .md)
    underscores=$(echo "$filename" | tr -cd '_' | wc -c)

    if [ "$underscores" -eq 1 ]; then
        parent=$(echo "$filename" | cut -d'_' -f2)
        parents["$parent"]=1
        parent_weights["$parent"]=$weight
        ((weight += 10))
    fi
done

echo "Found ${#parents[@]} parent command groups"

# Second pass: migrate files
for file in "$SOURCE_DIR"/*.md; do
    filename=$(basename "$file" .md)
    underscores=$(echo "$filename" | tr -cd '_' | wc -c)
    cmd_name=$(echo "$filename" | sed 's/shelly_//' | tr '_' ' ')
    first_line=$(head -1 "$file" | sed 's/^#* *//' | sed 's/ - /: /' || echo "$cmd_name")

    if [ "$underscores" -eq 1 ]; then
        parent=$(echo "$filename" | cut -d'_' -f2)
        echo "  [parent] shelly $parent"
        create_parent_index "$parent" "$first_line" "${parent_weights[$parent]}" "$file"

    elif [ "$underscores" -ge 2 ]; then
        parent=$(echo "$filename" | cut -d'_' -f2)
        subcommand=$(echo "$filename" | sed "s/shelly_${parent}_//" | tr '_' '-')

        if [ -n "${parents[$parent]:-}" ]; then
            echo "  [child]  shelly $parent $subcommand"
            mkdir -p "$TARGET_DIR/$parent"
            create_child_file "$parent" "$subcommand" "$cmd_name" "$first_line" "$file"
        else
            echo "  [root]   $filename"
            create_root_file "$filename" "$cmd_name" "$first_line" 50 "$file"
        fi
    else
        echo "  [root]   $filename"
        create_root_file "$filename" "$cmd_name" "$first_line" 1 "$file"
    fi
done

echo ""
echo "Migration complete:"
echo "  Directories: $(find "$TARGET_DIR" -mindepth 1 -type d | wc -l)"
echo "  Files: $(find "$TARGET_DIR" -name "*.md" | wc -l)"
