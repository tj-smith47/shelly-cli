// Package create provides the extension create command.
package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/plugins"
)

// NewCommand creates the extension create command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		lang      string
		outputDir string
	)

	cmd := &cobra.Command{
		Use:     "create <name>",
		Aliases: []string{"new", "init", "scaffold"},
		Short:   "Create a new extension scaffold",
		Long: `Create a new extension scaffold with boilerplate code.

Supported languages:
  - bash (default): Shell script extension
  - go: Go language extension
  - python: Python extension

The extension will be created in the current directory or the directory
specified with --output.`,
		Example: `  # Create a bash extension
  shelly extension create myext

  # Create a Go extension
  shelly extension create myext --lang go

  # Create in specific directory
  shelly extension create myext --output ~/projects`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return run(f, args[0], lang, outputDir)
		},
	}

	cmd.Flags().StringVarP(&lang, "lang", "l", "bash", "Extension language (bash, go, python)")
	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory")

	return cmd
}

func run(f *cmdutil.Factory, name, lang, outputDir string) error {
	ios := f.IOStreams()

	// Normalize name
	name = strings.ToLower(name)
	name = strings.TrimPrefix(name, plugins.PluginPrefix)

	// Full extension name
	extName := plugins.PluginPrefix + name

	// Create output directory
	extDir := filepath.Join(outputDir, extName)
	if err := os.MkdirAll(extDir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate files based on language
	switch lang {
	case "bash", "sh":
		if err := createBashExtension(extDir, extName, name); err != nil {
			return err
		}
	case "go", "golang":
		if err := createGoExtension(extDir, extName, name); err != nil {
			return err
		}
	case "python", "py":
		if err := createPythonExtension(extDir, extName, name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported language: %s (use bash, go, or python)", lang)
	}

	ios.Success("Created extension scaffold in %s", extDir)
	ios.Printf("\nNext steps:\n")
	ios.Printf("  1. Edit the extension code in %s\n", extDir)
	ios.Printf("  2. Test locally: ./%s/%s --help\n", extName, extName)
	ios.Printf("  3. Install: shelly extension install ./%s/%s\n", extName, extName)

	return nil
}

func createBashExtension(dir, extName, name string) error {
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
	//nolint:gosec // G306: Extensions need executable permission
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	// Create README
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

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0o600); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	return nil
}

func createGoExtension(dir, extName, name string) error {
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

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0o600); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// go.mod
	goMod := fmt.Sprintf(`module %s

go 1.21
`, extName)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o600); err != nil {
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

	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(makefile), 0o600); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return createBashExtension(dir, extName, name) // Also create README
}

func createPythonExtension(dir, extName, name string) error {
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
	//nolint:gosec // G306: Extensions need executable permission
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return createBashExtension(dir, extName, name) // Also create README (overwrites bash script)
}
