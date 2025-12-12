// Package export provides the backup export subcommand.
package export

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/iostreams"
	"github.com/tj-smith47/shelly-cli/internal/shelly"
)

var (
	allFlag      bool
	formatFlag   string
	parallelFlag int
)

// NewCommand creates the backup export command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <directory>",
		Short: "Export backups for all registered devices",
		Long: `Export backup files for all registered devices to a directory.

Creates one backup file per device, named by device ID.
Use --format to choose JSON or YAML output.`,
		Example: `  # Export all device backups to directory
  shelly backup export ./backups

  # Export in YAML format
  shelly backup export ./backups --format yaml

  # Export with parallel processing
  shelly backup export ./backups --parallel 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", true, "Export all registered devices")
	cmd.Flags().StringVarP(&formatFlag, "format", "f", "json", "Output format (json, yaml)")
	cmd.Flags().IntVar(&parallelFlag, "parallel", 3, "Number of parallel backups")

	return cmd
}

func run(ctx context.Context, dir string) error {
	ctx, cancel := context.WithTimeout(ctx, shelly.DefaultTimeout*10)
	defer cancel()

	ios := iostreams.System()

	// Get registered devices
	cfg := config.Get()
	if len(cfg.Devices) == 0 {
		ios.Info("No registered devices found")
		ios.Info("Use 'shelly device add' to register devices first")
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	ios.Info("Exporting backups for %d devices...", len(cfg.Devices))
	ios.Println()

	svc := shelly.NewService()
	opts := shelly.BackupOptions{}

	var (
		mu       sync.Mutex
		success  int
		failed   int
		failures []string
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(parallelFlag)

	for name, device := range cfg.Devices {
		deviceName := name
		deviceAddr := device.Address
		g.Go(func() error {
			ios.Printf("  Backing up %s (%s)...", deviceName, deviceAddr)

			backup, err := svc.CreateBackup(ctx, deviceAddr, opts)
			if err != nil {
				ios.Printf(" FAILED\n")
				mu.Lock()
				failed++
				failures = append(failures, fmt.Sprintf("%s: %v", deviceName, err))
				mu.Unlock()
				return nil // Continue with other devices
			}

			// Write backup file
			filename := sanitizeFilename(deviceName) + "." + formatFlag
			filePath := filepath.Join(dir, filename)

			if err := writeBackup(filePath, backup); err != nil {
				ios.Printf(" FAILED\n")
				mu.Lock()
				failed++
				failures = append(failures, fmt.Sprintf("%s: %v", deviceName, err))
				mu.Unlock()
				return nil
			}

			ios.Printf(" OK\n")
			mu.Lock()
			success++
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	ios.Println()
	if success > 0 {
		ios.Success("Successfully exported %d backup(s) to %s", success, dir)
	}
	if failed > 0 {
		ios.Warning("Failed to export %d device(s):", failed)
		for _, f := range failures {
			ios.Printf("  - %s\n", f)
		}
	}

	return nil
}

func sanitizeFilename(name string) string {
	// Replace problematic characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

func writeBackup(filePath string, backup *shelly.DeviceBackup) error {
	var data []byte
	var err error

	switch formatFlag {
	case "yaml", "yml":
		data, err = yaml.Marshal(backup)
	default:
		data, err = json.MarshalIndent(backup, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
