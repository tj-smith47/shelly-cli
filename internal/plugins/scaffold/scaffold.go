// Package scaffold provides extension scaffolding generation.
package scaffold

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// Bash creates a bash extension scaffold in the specified directory.
func Bash(dir, extName, name string) error {
	script := fmt.Sprintf(`#!/usr/bin/env bash
# %s - Shelly CLI Extension
# Description: A custom extension for Shelly CLI
#
# Environment variables available:
#   SHELLY_CONFIG_PATH   - Path to config file
#   SHELLY_DEVICES_JSON  - JSON of registered devices
#   SHELLY_OUTPUT_FORMAT - Current output format (table, json, yaml)
#   SHELLY_NO_COLOR      - Set to 1 if colors disabled
#   SHELLY_VERBOSE       - Set to 1 if verbose mode enabled

set -euo pipefail

VERSION="0.1.0"

show_help() {
    cat << EOF
%s - A custom Shelly CLI extension

Usage: shelly %s [command] [options]

Commands:
    help        Show this help message
    version     Show version information

Options:
    -h, --help      Show help
    -v, --version   Show version

Examples:
    shelly %s help
    shelly %s version
EOF
}

show_version() {
    echo "%s version $VERSION"
}

main() {
    case "${1:-help}" in
        -h|--help|help)
            show_help
            ;;
        -v|--version|version)
            show_version
            ;;
        *)
            echo "Unknown command: $1"
            echo "Run '%s --help' for usage"
            exit 1
            ;;
    esac
}

main "$@"
`, extName, extName, name, name, name, extName, extName)

	scriptPath := filepath.Join(dir, extName)
	if err := afero.WriteFile(config.Fs(), scriptPath, []byte(script), 0o700); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return Readme(dir, extName, name)
}

// Readme creates a README.md for an extension.
func Readme(dir, extName, name string) error {
	readme := fmt.Sprintf(`# %s

A custom extension for Shelly CLI.

## Installation

`+"```bash"+`
shelly extension install ./%s
`+"```"+`

## Usage

`+"```bash"+`
shelly %s --help
`+"```"+`

## Environment Variables

The following environment variables are available to extensions:

- `+"`SHELLY_CONFIG_PATH`"+` - Path to the Shelly CLI config file
- `+"`SHELLY_DEVICES_JSON`"+` - JSON string of registered devices
- `+"`SHELLY_OUTPUT_FORMAT`"+` - Current output format (table, json, yaml)
- `+"`SHELLY_NO_COLOR`"+` - Set to "1" if colors are disabled
- `+"`SHELLY_VERBOSE`"+` - Set to "1" if verbose mode is enabled
`, extName, extName, name)

	if err := afero.WriteFile(config.Fs(), filepath.Join(dir, "README.md"), []byte(readme), 0o600); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	return nil
}

// Go creates a Go extension scaffold in the specified directory.
func Go(dir, extName, name string) error {
	// main.go
	mainGo := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		showHelp()
	case "-v", "--version", "version":
		showVersion()
	case "devices":
		listDevices()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %%s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Run '%%s --help' for usage\n", os.Args[0])
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf(`+"`"+`%s - A custom Shelly CLI extension

Usage: shelly %s [command] [options]

Commands:
    help        Show this help message
    version     Show version information
    devices     List registered devices (from SHELLY_DEVICES_JSON)

Options:
    -h, --help      Show help
    -v, --version   Show version

Environment Variables:
    SHELLY_CONFIG_PATH   - Path to config file
    SHELLY_DEVICES_JSON  - JSON of registered devices
    SHELLY_OUTPUT_FORMAT - Current output format
`+"`"+`, "%s", "%s")
}

func showVersion() {
	fmt.Printf("%s version %%s\n", version)
}

func listDevices() {
	devicesJSON := os.Getenv("SHELLY_DEVICES_JSON")
	if devicesJSON == "" {
		fmt.Println("No devices found (SHELLY_DEVICES_JSON not set)")
		return
	}

	var devices map[string]interface{}
	if err := json.Unmarshal([]byte(devicesJSON), &devices); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse devices: %%v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("No devices registered")
		return
	}

	fmt.Printf("Found %%d device(s):\n", len(devices))
	for name := range devices {
		fmt.Printf("  - %%s\n", name)
	}
}
`, extName, name, extName, name, extName)

	fs := config.Fs()
	if err := afero.WriteFile(fs, filepath.Join(dir, "main.go"), []byte(mainGo), 0o600); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// go.mod
	goMod := fmt.Sprintf(`module %s

go 1.21
`, extName)

	if err := afero.WriteFile(fs, filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Makefile
	makefile := fmt.Sprintf(`# Build the extension
build:
	go build -o %s .

# Install to user plugins directory
install: build
	shelly extension install ./%s

# Clean build artifacts
clean:
	rm -f %s

.PHONY: build install clean
`, extName, extName, extName)

	if err := afero.WriteFile(fs, filepath.Join(dir, "Makefile"), []byte(makefile), 0o600); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return Readme(dir, extName, name)
}

// Python creates a Python extension scaffold in the specified directory.
func Python(dir, extName, name string) error {
	script := fmt.Sprintf(`#!/usr/bin/env python3
"""
%s - Shelly CLI Extension
"""

import json
import os
import sys

VERSION = "0.1.0"


def show_help():
    print(f"""%s - A custom Shelly CLI extension

Usage: shelly %s [command] [options]

Commands:
    help        Show this help message
    version     Show version information
    devices     List registered devices

Options:
    -h, --help      Show help
    -v, --version   Show version
""")


def show_version():
    print(f"%s version {VERSION}")


def list_devices():
    devices_json = os.environ.get("SHELLY_DEVICES_JSON", "{}")
    try:
        devices = json.loads(devices_json)
    except json.JSONDecodeError:
        print("Failed to parse devices JSON", file=sys.stderr)
        sys.exit(1)

    if not devices:
        print("No devices registered")
        return

    print(f"Found {len(devices)} device(s):")
    for name in devices:
        print(f"  - {name}")


def main():
    args = sys.argv[1:]

    if not args:
        show_help()
        return

    cmd = args[0]

    if cmd in ("-h", "--help", "help"):
        show_help()
    elif cmd in ("-v", "--version", "version"):
        show_version()
    elif cmd == "devices":
        list_devices()
    else:
        print(f"Unknown command: {cmd}", file=sys.stderr)
        print(f"Run '{sys.argv[0]} --help' for usage", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
`, extName, extName, name, extName)

	scriptPath := filepath.Join(dir, extName)
	if err := afero.WriteFile(config.Fs(), scriptPath, []byte(script), 0o700); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return Readme(dir, extName, name)
}

// NormalizeName normalizes an extension name (lowercase, prefix removed).
func NormalizeName(name string) string {
	name = strings.ToLower(name)
	return strings.TrimPrefix(name, plugins.PluginPrefix)
}

// FullName returns the full extension name with plugin prefix.
func FullName(name string) string {
	return plugins.PluginPrefix + NormalizeName(name)
}
