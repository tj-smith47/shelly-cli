# Extension Manifest System Design

> **Status**: Design Document (Not Yet Implemented)
> **Author**: Claude Code
> **Date**: 2024-12-14

## Overview

This document proposes an extension manifest system to improve upgrade reliability and metadata tracking for Shelly CLI extensions.

## Problem Statement

### Current Implementation

Extensions are stored as bare binaries in `~/.config/shelly/plugins/shelly-<name>`. The upgrade command uses heuristics to guess the GitHub source:

```go
// internal/cmd/extension/upgrade/upgrade.go
possibleSources := []string{
    fmt.Sprintf("gh:tj-smith47/%s", binaryName),  // Official plugins
    fmt.Sprintf("gh:%s/%s", name, binaryName),    // User repos
}
```

### Issues

| Problem | Impact |
|---------|--------|
| No source tracking | Extensions from local files or URLs cannot be upgraded |
| Heuristic guessing | GitHub source detection is fragile and breaks with naming variations |
| Runtime version detection | Must execute binary with `--version` (security risk, slow) |
| No integrity verification | Cannot verify binary hasn't been tampered with |
| No metadata persistence | Installation timestamp, dependencies unknown |

## Proposed Solution

### Directory Structure

Change from flat binary storage to per-extension directories:

```
~/.config/shelly/plugins/
├── shelly-myext/
│   ├── shelly-myext          # Binary executable
│   └── manifest.json         # Metadata file
├── shelly-another/
│   ├── shelly-another
│   └── manifest.json
└── .migrated                  # Migration marker file
```

### Manifest Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": ["schema_version", "name", "installed_at", "source", "binary"],
  "properties": {
    "schema_version": {
      "type": "string",
      "const": "1",
      "description": "Manifest schema version for forward compatibility"
    },
    "name": {
      "type": "string",
      "pattern": "^[a-z][a-z0-9-]*$",
      "description": "Extension name without 'shelly-' prefix"
    },
    "version": {
      "type": "string",
      "description": "Semantic version (e.g., '1.2.0')"
    },
    "description": {
      "type": "string",
      "description": "Human-readable description"
    },
    "installed_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 installation timestamp"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time",
      "description": "ISO 8601 last update timestamp"
    },
    "source": {
      "type": "object",
      "required": ["type"],
      "properties": {
        "type": {
          "type": "string",
          "enum": ["github", "url", "local", "unknown"]
        },
        "url": {
          "type": "string",
          "format": "uri"
        },
        "ref": {
          "type": "string",
          "description": "Git ref (tag/commit) for GitHub sources"
        },
        "asset": {
          "type": "string",
          "description": "Release asset filename"
        },
        "path": {
          "type": "string",
          "description": "Original file path for local sources"
        }
      }
    },
    "binary": {
      "type": "object",
      "required": ["name", "checksum"],
      "properties": {
        "name": {
          "type": "string",
          "description": "Binary filename"
        },
        "checksum": {
          "type": "string",
          "pattern": "^sha256:[a-f0-9]{64}$",
          "description": "SHA256 checksum with algorithm prefix"
        },
        "platform": {
          "type": "string",
          "description": "Platform identifier (e.g., 'linux-amd64')"
        },
        "size": {
          "type": "integer",
          "description": "File size in bytes"
        }
      }
    },
    "minimum_shelly_version": {
      "type": "string",
      "description": "Minimum compatible Shelly CLI version"
    }
  }
}
```

### Example Manifests

**GitHub-sourced extension:**
```json
{
  "schema_version": "1",
  "name": "backup",
  "version": "2.1.0",
  "description": "Enhanced backup and restore functionality",
  "installed_at": "2024-12-14T10:30:00Z",
  "updated_at": "2024-12-14T10:30:00Z",
  "source": {
    "type": "github",
    "url": "https://github.com/user/shelly-backup",
    "ref": "v2.1.0",
    "asset": "shelly-backup-linux-amd64.tar.gz"
  },
  "binary": {
    "name": "shelly-backup",
    "checksum": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "platform": "linux-amd64",
    "size": 5242880
  },
  "minimum_shelly_version": "0.1.0"
}
```

**URL-sourced extension:**
```json
{
  "schema_version": "1",
  "name": "custom-tool",
  "version": "1.0.0",
  "installed_at": "2024-12-14T11:00:00Z",
  "source": {
    "type": "url",
    "url": "https://internal.example.com/tools/shelly-custom-tool-v1.0.0"
  },
  "binary": {
    "name": "shelly-custom-tool",
    "checksum": "sha256:abc123...",
    "platform": "linux-amd64",
    "size": 3145728
  }
}
```

**Local-sourced extension:**
```json
{
  "schema_version": "1",
  "name": "dev-extension",
  "installed_at": "2024-12-14T12:00:00Z",
  "source": {
    "type": "local",
    "path": "/home/user/projects/shelly-dev-extension/build/shelly-dev-extension"
  },
  "binary": {
    "name": "shelly-dev-extension",
    "checksum": "sha256:def456...",
    "platform": "linux-amd64",
    "size": 2097152
  }
}
```

**Migrated extension (unknown source):**
```json
{
  "schema_version": "1",
  "name": "legacy-ext",
  "installed_at": "2024-12-14T09:00:00Z",
  "source": {
    "type": "unknown"
  },
  "binary": {
    "name": "shelly-legacy-ext",
    "checksum": "sha256:789abc...",
    "platform": "linux-amd64",
    "size": 4194304
  }
}
```

## Upgrade Behavior by Source Type

| Source Type | Upgrade Behavior |
|-------------|------------------|
| `github` | Fetch latest release from `source.url`, compare versions, download if newer |
| `url` | Re-fetch from `source.url`, compare checksums, replace if different |
| `local` | Display warning: "Local extension - reinstall manually with new binary" |
| `unknown` | Display warning: "Unknown source - reinstall to enable auto-upgrade" |

## Manifest Creation

Manifests are **created on install**, not fetched from source:

1. Install command parses source (GitHub/URL/local)
2. Downloads binary to temp location
3. Computes SHA256 checksum
4. Extracts version from release tag (GitHub) or `--version` output
5. Creates extension directory and manifest
6. Moves binary to final location
7. Sets executable permissions

**Rationale:**
- Not all sources provide manifests
- Keeps installation simple
- Avoids chicken-and-egg problem
- Consistent behavior across source types

## Migration Strategy

### Automatic Migration

On first CLI invocation after upgrade:

```go
// internal/plugins/migrate.go
func MigrateExtensions() error {
    pluginsDir, _ := config.PluginsDir()
    markerFile := filepath.Join(pluginsDir, ".migrated")

    // Skip if already migrated
    if fileExists(markerFile) {
        return nil
    }

    entries, _ := os.ReadDir(pluginsDir)
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasPrefix(entry.Name(), "shelly-") {
            continue
        }

        // Old format: bare binary
        oldPath := filepath.Join(pluginsDir, entry.Name())
        newDir := filepath.Join(pluginsDir, entry.Name())
        newPath := filepath.Join(newDir, entry.Name())

        // Create directory
        os.MkdirAll(newDir, 0755)

        // Move binary
        os.Rename(oldPath, newPath)

        // Create minimal manifest
        manifest := Manifest{
            SchemaVersion: "1",
            Name:          strings.TrimPrefix(entry.Name(), "shelly-"),
            InstalledAt:   time.Now().UTC().Format(time.RFC3339),
            Source:        Source{Type: "unknown"},
            Binary: Binary{
                Name:     entry.Name(),
                Checksum: computeChecksum(newPath),
                Platform: runtime.GOOS + "-" + runtime.GOARCH,
            },
        }
        writeManifest(newDir, manifest)
    }

    // Create marker file
    os.WriteFile(markerFile, []byte("1"), 0644)
    return nil
}
```

### Migration Behavior

1. **Triggered**: First run after shelly-cli upgrade
2. **Non-destructive**: Only moves files, doesn't delete
3. **Idempotent**: Marker file prevents re-migration
4. **Graceful**: Migrated extensions marked `source.type: "unknown"`
5. **User messaging**: Info message shown during migration

### User Communication

```
Migrating extensions to new format...
  shelly-backup: migrated (source unknown - reinstall to enable auto-upgrade)
  shelly-custom: migrated (source unknown - reinstall to enable auto-upgrade)
Migration complete. 2 extensions migrated.
```

## API Changes

### New Types

**File**: `internal/plugins/manifest.go`

```go
package plugins

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

const ManifestSchemaVersion = "1"
const ManifestFileName = "manifest.json"

type Manifest struct {
    SchemaVersion       string  `json:"schema_version"`
    Name                string  `json:"name"`
    Version             string  `json:"version,omitempty"`
    Description         string  `json:"description,omitempty"`
    InstalledAt         string  `json:"installed_at"`
    UpdatedAt           string  `json:"updated_at,omitempty"`
    Source              Source  `json:"source"`
    Binary              Binary  `json:"binary"`
    MinimumShellyVersion string `json:"minimum_shelly_version,omitempty"`
}

type Source struct {
    Type  string `json:"type"`            // "github", "url", "local", "unknown"
    URL   string `json:"url,omitempty"`
    Ref   string `json:"ref,omitempty"`   // Git tag/commit
    Asset string `json:"asset,omitempty"` // Release asset filename
    Path  string `json:"path,omitempty"`  // Local file path
}

type Binary struct {
    Name     string `json:"name"`
    Checksum string `json:"checksum"`
    Platform string `json:"platform,omitempty"`
    Size     int64  `json:"size,omitempty"`
}

// LoadManifest reads a manifest from an extension directory
func LoadManifest(extDir string) (*Manifest, error) {
    path := filepath.Join(extDir, ManifestFileName)
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %w", err)
    }

    var m Manifest
    if err := json.Unmarshal(data, &m); err != nil {
        return nil, fmt.Errorf("failed to parse manifest: %w", err)
    }

    return &m, nil
}

// Save writes the manifest to disk
func (m *Manifest) Save(extDir string) error {
    path := filepath.Join(extDir, ManifestFileName)
    data, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal manifest: %w", err)
    }

    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write manifest: %w", err)
    }

    return nil
}

// ComputeChecksum calculates SHA256 checksum for a file
func ComputeChecksum(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }

    return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
```

### Modified Functions

**File**: `internal/plugins/loader.go`

```go
// DiscoverPlugins finds all installed extensions
func DiscoverPlugins(pluginsDir string) ([]Plugin, error) {
    var plugins []Plugin

    entries, err := os.ReadDir(pluginsDir)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "shelly-") {
            continue
        }

        extDir := filepath.Join(pluginsDir, entry.Name())
        manifest, err := LoadManifest(extDir)
        if err != nil {
            // Fallback for directories without manifest (shouldn't happen after migration)
            continue
        }

        plugins = append(plugins, Plugin{
            Name:    manifest.Name,
            Path:    filepath.Join(extDir, manifest.Binary.Name),
            Version: manifest.Version,
        })
    }

    return plugins, nil
}
```

**File**: `internal/cmd/extension/upgrade/upgrade.go`

```go
func upgradeExtension(ctx context.Context, ios *iostreams.IOStreams, extDir string) error {
    manifest, err := plugins.LoadManifest(extDir)
    if err != nil {
        return fmt.Errorf("failed to load manifest: %w", err)
    }

    switch manifest.Source.Type {
    case "github":
        return upgradeFromGitHub(ctx, ios, extDir, manifest)
    case "url":
        return upgradeFromURL(ctx, ios, extDir, manifest)
    case "local":
        ios.Warning("Extension '%s' was installed from local file.\n", manifest.Name)
        ios.Info("  Reinstall manually: shelly extension install %s\n", manifest.Source.Path)
        return nil
    case "unknown":
        ios.Warning("Extension '%s' has unknown source (migrated from old format).\n", manifest.Name)
        ios.Info("  Reinstall to enable auto-upgrade.\n")
        return nil
    default:
        return fmt.Errorf("unsupported source type: %s", manifest.Source.Type)
    }
}
```

## Implementation Phases

### Phase 1: Core Types & Migration
1. Add `internal/plugins/manifest.go` with types
2. Add `internal/plugins/migrate.go` for migration
3. Call migration from root command initialization
4. Update `internal/plugins/loader.go` to use manifests

### Phase 2: Install Command
1. Update `internal/cmd/extension/install/install.go`
2. Create manifest on successful install
3. Support GitHub, URL, and local sources

### Phase 3: Upgrade Command
1. Update `internal/cmd/extension/upgrade/upgrade.go`
2. Read manifest for source information
3. Implement source-specific upgrade logic

### Phase 4: List & Info Commands
1. Update `internal/cmd/extension/list/list.go` to show manifest info
2. Add optional `extension info <name>` command
3. Display source, version, checksum, install date

### Phase 5: Remove Command
1. Update `internal/cmd/extension/remove/remove.go`
2. Remove entire extension directory (not just binary)

## Security Considerations

1. **Checksum Verification**: On upgrade, verify downloaded binary matches expected checksum before replacing
2. **No Arbitrary Execution**: Version is read from manifest, not by executing binary
3. **Permission Preservation**: Maintain 0700 permissions on binaries
4. **Directory Isolation**: Each extension in its own directory prevents naming conflicts

## Future Enhancements (Out of Scope)

1. **Dependencies**: Track extension dependencies
2. **Remote Manifests**: Fetch manifests from extension repositories
3. **Signature Verification**: GPG/cosign signatures
4. **Extension Registry**: Central registry for discovery
5. **Rollback**: Keep previous version for rollback
6. **Configuration**: Per-extension configuration files

## Conclusion

The proposed manifest system:
- Enables reliable upgrades from any source type
- Preserves installation metadata
- Provides integrity verification
- Maintains backwards compatibility through migration
- Keeps implementation simple and focused

Implementation should follow the phased approach, with each phase tested independently before proceeding.
