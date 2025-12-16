# Release Notes Template

Use this template when creating GitHub releases manually or for release announcements.

---

## Shelly CLI v{VERSION}

{ONE_LINE_SUMMARY}

### Highlights

- {HIGHLIGHT_1}
- {HIGHLIGHT_2}
- {HIGHLIGHT_3}

### What's New

#### Features
- {FEATURE_1}
- {FEATURE_2}

#### Improvements
- {IMPROVEMENT_1}
- {IMPROVEMENT_2}

#### Bug Fixes
- {FIX_1}
- {FIX_2}

### Breaking Changes

{BREAKING_CHANGES_OR_NONE}

### Installation

#### Homebrew (macOS/Linux)
```bash
brew install tj-smith47/tap/shelly
```

#### Go Install
```bash
go install github.com/tj-smith47/shelly-cli/cmd/shelly@v{VERSION}
```

#### Docker
```bash
docker pull ghcr.io/tj-smith47/shelly-cli:v{VERSION}
```

#### Binary Download
Download the appropriate binary for your platform from the assets below.

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | amd64 | [shelly_{VERSION}_linux_amd64.tar.gz](link) |
| Linux | arm64 | [shelly_{VERSION}_linux_arm64.tar.gz](link) |
| macOS | amd64 | [shelly_{VERSION}_darwin_amd64.tar.gz](link) |
| macOS | arm64 | [shelly_{VERSION}_darwin_arm64.tar.gz](link) |
| Windows | amd64 | [shelly_{VERSION}_windows_amd64.zip](link) |

### Checksums

See `shelly_{VERSION}_checksums.txt` in the release assets.

### Full Changelog

See [CHANGELOG.md](https://github.com/tj-smith47/shelly-cli/blob/master/CHANGELOG.md) for complete details.

### Contributors

Thanks to all contributors who made this release possible!

---

## Example: v1.0.0-rc.1

First release candidate for v1.0.0!

### Highlights

- Complete Shelly device management CLI
- TUI dashboard with 280+ themes
- Gen1, Gen2, Gen3, and Gen4 device support
- Plugin system for extensibility

### What's New

#### Features
- 347 commands covering all Shelly device operations
- Interactive TUI dashboard (`shelly dash`)
- Plugin system with manifest support
- Alias system for custom shortcuts
- Shell completions for bash, zsh, fish, PowerShell

#### Improvements
- Factory pattern architecture for testability
- Comprehensive documentation (347 command docs + man pages)
- Docker support via ghcr.io

### Breaking Changes

None - this is the first release.

### Known Issues

- Test coverage at ~19% (target: 90%+ for v1.0.0 stable)
- Config package refactor in progress

### Feedback

This is a release candidate. Please report issues at:
https://github.com/tj-smith47/shelly-cli/issues

Or use `shelly feedback` to report directly from the CLI!
