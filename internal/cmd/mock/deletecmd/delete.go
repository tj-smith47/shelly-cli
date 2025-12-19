// Package deletecmd provides the mock delete command.
package deletecmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tj-smith47/shelly-cli/internal/cmdutil"
	"github.com/tj-smith47/shelly-cli/internal/testutil/mock"
)

// NewCommand creates the mock delete command.
func NewCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a mock device",
		Long:    `Delete a mock device configuration.`,
		Example: `  # Delete mock device
  shelly mock delete kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), f, args[0])
		},
	}
	return cmd
}

func run(_ context.Context, f *cmdutil.Factory, name string) error {
	ios := f.IOStreams()

	mockDir, err := mock.Dir()
	if err != nil {
		return err
	}

	filename := filepath.Join(mockDir, name+".json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("mock device not found: %s", name)
	}

	if err := os.Remove(filename); err != nil {
		return err
	}

	ios.Success("Deleted mock device: %s", name)
	return nil
}
