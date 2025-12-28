---
title: "Installation"
description: "Install Shelly CLI on macOS, Linux, Windows, or Docker"
weight: 10
---

Choose your preferred installation method below.

## Homebrew (macOS/Linux) {#homebrew}

The recommended installation method for macOS and Linux users:

```bash
brew install tj-smith47/tap/shelly-cli
```

**Verify installation:**
```bash
shelly version
```

**Expected output:**
```
shelly version 1.0.0 (abc1234)
```

## Go Install {#go-install}

If you have Go 1.21+ installed:

```bash
go install github.com/tj-smith47/shelly-cli/cmd/shelly@latest
```

Ensure `$GOPATH/bin` is in your `PATH`:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Docker {#docker}

Run Shelly CLI in a container without local installation:

```bash
# Pull the image
docker pull ghcr.io/tj-smith47/shelly-cli:latest

# Run a command
docker run --rm --network host ghcr.io/tj-smith47/shelly-cli:latest device list

# With config mount (persistent configuration)
docker run --rm --network host \
  -v ~/.config/shelly:/home/shelly/.config/shelly \
  ghcr.io/tj-smith47/shelly-cli:latest status kitchen
```

**Important:** Use `--network host` to allow device discovery and communication on your local network.

## Binary Download {#binary}

Download pre-built binaries from the [releases page](https://github.com/tj-smith47/shelly-cli/releases).

### Available Platforms

| Platform | Architecture | Filename |
|----------|--------------|----------|
| Linux | amd64 | `shelly_VERSION_linux_amd64.tar.gz` |
| Linux | arm64 | `shelly_VERSION_linux_arm64.tar.gz` |
| macOS | amd64 (Intel) | `shelly_VERSION_darwin_amd64.tar.gz` |
| macOS | arm64 (Apple Silicon) | `shelly_VERSION_darwin_arm64.tar.gz` |
| Windows | amd64 | `shelly_VERSION_windows_amd64.zip` |

### Linux/macOS Installation

```bash
# Download (replace VERSION with actual version, e.g., 1.0.0)
curl -LO https://github.com/tj-smith47/shelly-cli/releases/download/v${VERSION}/shelly_${VERSION}_linux_amd64.tar.gz

# Extract
tar -xzf shelly_${VERSION}_linux_amd64.tar.gz

# Move to PATH
sudo mv shelly /usr/local/bin/

# Verify
shelly version
```

### Windows Installation

1. Download the `.zip` file from [releases](https://github.com/tj-smith47/shelly-cli/releases)
2. Extract to a directory (e.g., `C:\Program Files\shelly-cli\`)
3. Add the directory to your `PATH`:
   - Open System Properties → Advanced → Environment Variables
   - Edit `Path` under User or System variables
   - Add `C:\Program Files\shelly-cli\`
4. Open a new terminal and run `shelly version`

## Shell Completions {#completions}

Enable tab completion for your shell:

### Bash

```bash
# Add to ~/.bashrc
source <(shelly completion bash)

# Or install globally
shelly completion bash | sudo tee /etc/bash_completion.d/shelly > /dev/null
```

### Zsh

```bash
# Add to ~/.zshrc (before compinit)
source <(shelly completion zsh)

# Or install to fpath
shelly completion zsh > "${fpath[1]}/_shelly"
```

### Fish

```bash
shelly completion fish > ~/.config/fish/completions/shelly.fish
```

### PowerShell

```powershell
# Add to $PROFILE
shelly completion powershell | Out-String | Invoke-Expression

# Or save to file
shelly completion powershell > shelly.ps1
```

## Troubleshooting

### "command not found" after installation

Ensure the binary is in your `PATH`:

```bash
# Check PATH
echo $PATH

# Find shelly binary
which shelly
```

### Permission denied on /usr/local/bin

Use sudo or install to a user directory:

```bash
# Option 1: Use sudo
sudo mv shelly /usr/local/bin/

# Option 2: Install to user directory
mkdir -p ~/.local/bin
mv shelly ~/.local/bin/
export PATH="$PATH:$HOME/.local/bin"
```

### Homebrew tap not found

```bash
# Add the tap explicitly
brew tap tj-smith47/tap
brew install shelly-cli
```

## Next Steps

Continue to the [Quick Start Guide](/docs/getting-started/quickstart/) to control your first device.
