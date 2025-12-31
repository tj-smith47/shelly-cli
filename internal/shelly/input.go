// Package shelly provides business logic for Shelly device operations.
package shelly

import (
	"context"
	"fmt"

	"github.com/tj-smith47/shelly-cli/internal/client"
	"github.com/tj-smith47/shelly-cli/internal/model"
	"github.com/tj-smith47/shelly-cli/internal/output"
	"github.com/tj-smith47/shelly-cli/internal/theme"
)

// InputInfo holds input information for list operations.
type InputInfo struct {
	ID    int
	Name  string
	Type  string
	State bool
}

// ListHeaders returns the column headers for the table.
func (i InputInfo) ListHeaders() []string {
	return []string{"ID", "Name", "Type", "State"}
}

// ListRow returns the formatted row values for the table.
func (i InputInfo) ListRow() []string {
	name := i.Name
	if name == "" {
		name = theme.Dim().Render("-")
	}

	state := output.RenderActive(i.State, output.CaseLower, theme.FalseError)
	return []string{fmt.Sprintf("%d", i.ID), name, i.Type, state}
}

// InputStatus gets the status of an input component.
func (s *Service) InputStatus(ctx context.Context, identifier string, inputID int) (*model.InputStatus, error) {
	var result *model.InputStatus
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		status, err := conn.Input(inputID).GetStatus(ctx)
		if err != nil {
			return err
		}
		result = status

		// Get type from config since it's not in status
		config, err := conn.Input(inputID).GetConfig(ctx)
		if err == nil {
			result.Type = config.Type
		}

		return nil
	})
	return result, err
}

// InputTrigger triggers an input event.
func (s *Service) InputTrigger(ctx context.Context, identifier string, inputID int, eventType string) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Input(inputID).Trigger(ctx, eventType)
	})
}

// InputGetConfig gets the configuration for an input component.
func (s *Service) InputGetConfig(ctx context.Context, identifier string, inputID int) (*model.InputConfig, error) {
	var result *model.InputConfig
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		config, err := conn.Input(inputID).GetConfig(ctx)
		if err != nil {
			return err
		}
		result = config
		return nil
	})
	return result, err
}

// InputSetConfig updates the configuration for an input component.
func (s *Service) InputSetConfig(ctx context.Context, identifier string, inputID int, cfg *model.InputConfig) error {
	return s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		return conn.Input(inputID).SetConfig(ctx, cfg)
	})
}

// InputList lists all input components on a device with their status.
func (s *Service) InputList(ctx context.Context, identifier string) ([]InputInfo, error) {
	var result []InputInfo
	err := s.WithConnection(ctx, identifier, func(conn *client.Client) error {
		components, err := conn.FilterComponents(ctx, model.ComponentInput)
		if err != nil {
			return err
		}

		result = make([]InputInfo, 0, len(components))
		for _, comp := range components {
			info := InputInfo{ID: comp.ID}

			// Get status for state
			status, err := conn.Input(comp.ID).GetStatus(ctx)
			if err != nil {
				continue
			}
			info.State = status.State

			// Get config for name and type
			config, err := conn.Input(comp.ID).GetConfig(ctx)
			if err == nil {
				info.Type = config.Type
				if config.Name != nil {
					info.Name = *config.Name
				}
			}

			result = append(result, info)
		}

		return nil
	})
	return result, err
}
