// Package main generates documentation for shelly CLI commands.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/tj-smith47/shelly-cli/internal/cmd"
	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
)

func main() {
	outputDir := "./docs/commands"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}
	outputDir = insideWorkingDir(outputDir)

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

	// Hide alias commands so they don't get separate doc pages
	cmdutil.HideAliases(rootCmd)

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

// insideWorkingDir confines dir to the working directory and returns its
// absolute form: the generators only ever write into the repo.
func insideWorkingDir(dir string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to read the working directory: %v", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("invalid output directory %q: %v", dir, err)
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		log.Fatalf("output directory %q is outside the working directory %q", dir, cwd)
	}
	return filepath.Join(cwd, rel)
}
