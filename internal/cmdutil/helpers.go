// Package cmdutil provides command utilities and shared infrastructure for CLI commands.
package cmdutil

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tj-smith47/shelly-cli/internal/cache"
	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/config"
	"github.com/tj-smith47/shelly-cli/internal/model"
)

// AddCommandsToGroup adds multiple commands to a root command and assigns them to a group.
// This is a convenience function for organizing help output.
func AddCommandsToGroup(root *cobra.Command, groupID string, cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		cmd.GroupID = groupID
		root.AddCommand(cmd)
	}
}

// SafeConfig returns the config, logging errors to debug if load fails.
// Use this when config is optional and the command can continue without it.
// Returns nil if config loading fails.
func SafeConfig(f *Factory) *config.Config {
	cfg, err := f.Config()
	if err != nil {
		f.IOStreams().DebugErr("load config", err)
		return nil
	}
	return cfg
}

// CachedComponentList returns the component list for a device, using cache if available.
// If --refresh flag is set or cache is empty, fetches fresh data and caches it.
// This is useful for commands that need the component list and want to avoid
// repeated RPC calls for data that rarely changes.
func CachedComponentList(ctx context.Context, f *Factory, device string) ([]model.Component, error) {
	fc := f.FileCache()
	svc := f.ShellyService()

	// Try cache first unless --refresh flag is set
	if fc != nil && !viper.GetBool("refresh") {
		if entry, err := fc.Get(device, cache.TypeComponents); err == nil && entry != nil {
			var components []model.Component
			if err := entry.Unmarshal(&components); err == nil {
				return components, nil
			}
		}
	}

	// Fetch fresh component list
	var components []model.Component
	err := svc.WithConnection(ctx, device, func(conn *client.Client) error {
		var err error
		components, err = conn.ListComponents(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	// Cache the result (24 hour TTL for components)
	if fc != nil && components != nil {
		if err := fc.Set(device, cache.TypeComponents, components, cache.TTLComponents); err != nil {
			f.IOStreams().DebugErr("cache components", err)
		}
	}

	return components, nil
}
