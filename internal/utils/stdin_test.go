package utils

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceStdinArg_NoDash(t *testing.T) {
	t.Parallel()
	args := []string{"status", "kitchen"}
	result, err := ReplaceStdinArg(args)
	require.NoError(t, err)
	assert.Equal(t, args, result)
}

func TestReplaceStdinArg_EmptyArgs(t *testing.T) {
	t.Parallel()
	result, err := ReplaceStdinArg(nil)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestReplaceStdinArg_NoDashInFlags(t *testing.T) {
	t.Parallel()
	// "--output" contains a dash but isn't a standalone "-"
	args := []string{"status", "--output", "json"}
	result, err := ReplaceStdinArg(args)
	require.NoError(t, err)
	assert.Equal(t, args, result)
}

func TestResolveStdinFlag_NotChanged(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	// Flag exists but wasn't changed — should be a no-op
	err := ResolveStdinFlag(cmd, "server")
	require.NoError(t, err)
}

func TestResolveStdinFlag_NotDash(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	require.NoError(t, cmd.Flags().Set("server", "mqtt://host"))
	// Flag is set but not to "-" — should be a no-op
	err := ResolveStdinFlag(cmd, "server")
	require.NoError(t, err)
	assert.Equal(t, "mqtt://host", cmd.Flags().Lookup("server").Value.String())
}

func TestResolveStdinFlag_UnknownFlag(t *testing.T) {
	t.Parallel()
	cmd := newTestCommand()
	err := ResolveStdinFlag(cmd, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag")
}

func newTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("server", "", "Server URL")
	cmd.Flags().String("value", "", "Value")
	return cmd
}
