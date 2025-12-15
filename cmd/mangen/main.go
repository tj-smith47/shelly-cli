// Package main generates man pages for shelly CLI commands.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	"github.com/tj-smith47/shelly-cli/internal/cmd"
)

func main() {
	outputDir := "./docs/man"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// Get the root command with all subcommands
	rootCmd := cmd.GetRootCmd()

	// Set up man page header
	header := &doc.GenManHeader{
		Title:   "SHELLY",
		Section: "1",
		Source:  "Shelly CLI",
		Manual:  "User Commands",
	}

	// Generate man pages
	if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
		log.Fatalf("failed to generate man pages: %v", err)
	}

	// Count generated files
	files, err := filepath.Glob(filepath.Join(outputDir, "*.1"))
	if err != nil {
		log.Printf("warning: failed to count files: %v", err)
	}

	fmt.Printf("Generated %d man pages in %s\n", len(files), outputDir)
}
