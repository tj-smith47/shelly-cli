// Package main generates documentation for shelly CLI commands.
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
	outputDir := "./docs/commands"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// Clean stale .md files before regenerating
	staleFiles, err := filepath.Glob(filepath.Join(outputDir, "*.md"))
	if err != nil {
		log.Printf("warning: failed to find stale files: %v", err)
	}
	for _, f := range staleFiles {
		if err := os.Remove(f); err != nil {
			log.Printf("warning: failed to remove %s: %v", f, err)
		}
	}

	// Get the root command with all subcommands
	rootCmd := cmd.GetRootCmd()

	// Disable auto-generated timestamp to avoid churn on every regeneration
	rootCmd.DisableAutoGenTag = true

	// Generate markdown documentation
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		log.Fatalf("failed to generate markdown docs: %v", err)
	}

	// Count generated files
	files, err := filepath.Glob(filepath.Join(outputDir, "*.md"))
	if err != nil {
		log.Printf("warning: failed to count files: %v", err)
	}

	fmt.Printf("Generated %d command documentation files in %s\n", len(files), outputDir)
}
